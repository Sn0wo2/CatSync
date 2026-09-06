package execute

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/cstx"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v3"
)

type ResultStatus uint8

const (
	StatusNotMatched ResultStatus = iota
	StatusMatched
	StatusJump
)

type Result struct {
	Status ResultStatus
	JumpTo int
}

type ActionEntry struct {
	Exact   string
	Re      *regexp.Regexp
	Handler action.Handler
	Payload config.ActionData
}

type Executor struct {
	Cfg              *config.Config
	Registry         action.Registry
	Entries          []ActionEntry
	LabelIndex       map[string]int
	PreResolvedJumps map[int]int
}

func New() *Executor {
	return &Executor{Registry: action.NewRegistry()}
}

func (e *Executor) WithConfig(cfg *config.Config) *Executor {
	e.Cfg = cfg

	return e
}

func (e *Executor) Build() (*Executor, error) {
	if e == nil || e.Cfg == nil {
		return nil, errors.New("nil executor or config")
	}

	actions := e.Cfg.Actions
	e.Entries = make([]ActionEntry, len(actions))

	e.LabelIndex = make(map[string]int, len(actions))
	for i := range actions {
		if label := actions[i].Label; label != "" {
			e.LabelIndex[label] = i
		}
	}

	for i := range actions {
		act := &actions[i]
		entry := &e.Entries[i]

		route, ok := act.Route.Literal()
		if !ok {
			return nil, fmt.Errorf("actions[%d].route must be a literal string", i)
		}

		switch {
		case route == "":
		case strings.ContainsAny(route, `.*+?()[]{}|\^$`):
			re, err := util.GetCompiledRegexp(route)
			if err != nil {
				return nil, fmt.Errorf("invalid action route regexp at actions[%d].route (%q): %w", i, route, err)
			}

			entry.Re = re
		default:
			entry.Exact = route
		}

		baseHandler, ok := e.Registry[act.TypeName()]
		if !ok || baseHandler == nil {
			return nil, fmt.Errorf("unknown action handler at actions[%d]: %s", i, act.TypeName())
		}

		payload := act.GetPayload()
		if payload == nil {
			return nil, fmt.Errorf("actions[%d] type=%s but payload is nil", i, act.TypeName())
		}

		entry.Payload = payload

		h := action.WrapHandler(baseHandler)

		add := func(source string, mods []action.Modifier, err error) error {
			if err != nil {
				return fmt.Errorf("build %s modifiers at actions[%d]: %w", source, i, err)
			}

			for _, m := range mods {
				h.WithModifier(m)
			}

			return nil
		}

		if !act.SkipGlobalModifiers {
			mods, err := action.BuildGlobalModifiers(e.Cfg)
			if err := add("global", mods, err); err != nil {
				return nil, err
			}
		}

		mods, err := action.BuildModifiers(&act.GlobalModifier)
		if err := add("action", mods, err); err != nil {
			return nil, err
		}

		mods, err = action.BuildModifiers(payload.GetGlobalModifier())
		if err := add("payload", mods, err); err != nil {
			return nil, err
		}

		entry.Handler = h
	}

	e.PreResolvedJumps = make(map[int]int)

	for i := range actions {
		auth := actions[i].ActionModifierAuth
		if auth == nil || auth.Fallback == nil || auth.Fallback.JumpLabel == "" {
			continue
		}

		if idx, ok := e.LabelIndex[auth.Fallback.JumpLabel]; ok {
			e.PreResolvedJumps[i] = idx
		}
	}

	return e, nil
}

type RequestContext struct {
	Ctx      *cstx.Ctx
	FiberCtx fiber.Ctx
	Reloader action.Reloader
}

func (e *Executor) Dispatch(rc *RequestContext) (bool, error) {
	if e == nil || e.Cfg == nil {
		return false, errors.New("nil executor or config")
	}

	if rc == nil || rc.Ctx == nil || rc.FiberCtx == nil {
		return false, errors.New("nil request context")
	}

	n := len(e.Cfg.Actions)
	if n == 0 {
		return false, nil
	}

	visited := make(map[int]struct{}, n)

	for i := range n {
		res, err := e.ExecuteOne(rc, i, false)
		if err != nil {
			return false, err
		}

		switch res.Status {
		case StatusNotMatched:
			continue

		case StatusMatched:
			return true, nil

		case StatusJump:
			target := res.JumpTo
			for {
				if target < 0 || target >= n {
					return false, fmt.Errorf("invalid auth fallback jumpTo index: %d", target)
				}

				if _, dup := visited[target]; dup {
					return false, fmt.Errorf("auth fallback jump loop detected at index: %d", target)
				}

				visited[target] = struct{}{}

				jres, jerr := e.ExecuteOne(rc, target, true)
				if jerr != nil {
					return false, jerr
				}

				if jres.Status == StatusMatched {
					return true, nil
				}

				if jres.Status == StatusNotMatched {
					break
				}

				target = jres.JumpTo
			}
		}
	}

	lastRes, lastErr := e.ExecuteOne(rc, n-1, true)
	if lastErr != nil {
		return false, lastErr
	}

	return lastRes.Status == StatusMatched, nil
}

func (e *Executor) ExecuteOne(rc *RequestContext, index int, skipRoute bool) (Result, error) {
	if index < 0 || index >= len(e.Entries) {
		return Result{}, fmt.Errorf("invalid action index: %d", index)
	}

	entry := &e.Entries[index]
	path := rc.FiberCtx.Path()

	if !skipRoute {
		if entry.Exact != "" {
			if entry.Exact != path {
				return Result{Status: StatusNotMatched}, nil
			}
		} else if entry.Re == nil || !entry.Re.MatchString(path) {
			return Result{Status: StatusNotMatched}, nil
		}
	}

	if entry.Handler == nil || entry.Payload == nil {
		return Result{Status: StatusNotMatched}, nil
	}

	result := entry.Handler.ProcessAction(&action.ProcessData{
		CStx:     rc.Ctx,
		FCtx:     rc.FiberCtx,
		Payload:  entry.Payload,
		Reloader: rc.Reloader,
	})
	if result.Err != nil {
		return Result{Status: StatusMatched}, result.Err
	}

	if resolved, ok := e.PreResolvedJumps[index]; ok {
		result.JumpTo = resolved
	}

	switch result.Status {
	case action.ExecutionCompleted:
		return Result{Status: StatusMatched}, nil
	case action.ExecutionFallbackNext:
		return Result{Status: StatusNotMatched}, nil
	case action.ExecutionFallbackJump:
		return Result{Status: StatusJump, JumpTo: result.JumpTo}, nil
	default:
		return Result{Status: StatusMatched}, fmt.Errorf("unsupported action execution status: %d", result.Status)
	}
}
