package execute

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
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
	Registry         map[config.ActionType]action.Handler
	Entries          []ActionEntry
	LabelIndex       map[string]int
	PreResolvedJumps map[int]int
}

func Build(cfg *config.Config) (*Executor, error) {
	if cfg == nil {
		return nil, errors.New("nil executor or config")
	}

	e := &Executor{Cfg: cfg, Registry: action.NewRegistry()}
	actions := cfg.Actions
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

		h := &action.ModifiableHandler{Handler: baseHandler}

		if !act.SkipGlobalModifiers {
			mods := make([]action.Modifier, 0, len(e.Cfg.Modifiers))
			for j := range e.Cfg.Modifiers {
				global, err := action.BuildModifiers(&e.Cfg.Modifiers[j])
				if err != nil {
					return nil, fmt.Errorf("build global modifiers at actions[%d]: modifiers[%d]: %w", i, j, err)
				}

				mods = append(mods, global...)
			}

			h.Modifiers = append(h.Modifiers, mods...)
		}

		mods, err := action.BuildModifiers(&act.GlobalModifier)
		if err != nil {
			return nil, fmt.Errorf("build action modifiers at actions[%d]: %w", i, err)
		}

		h.Modifiers = append(h.Modifiers, mods...)

		mods, err = action.BuildModifiers(payload.GetGlobalModifier())
		if err != nil {
			return nil, fmt.Errorf("build payload modifiers at actions[%d]: %w", i, err)
		}

		h.Modifiers = append(h.Modifiers, mods...)

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

func (e *Executor) Stop() {
	if fileHandler, ok := e.Registry[config.ActionFile].(*action.FileHandler); ok {
		fileHandler.Stop()
	}
}

type RequestContext struct {
	Logger   *slog.Logger
	FiberCtx fiber.Ctx
	Reloader action.Reloader
}

func (e *Executor) Dispatch(rc *RequestContext) (bool, error) {
	if e == nil || e.Cfg == nil {
		return false, errors.New("nil executor or config")
	}

	if rc == nil || rc.Logger == nil || rc.FiberCtx == nil {
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
		Logger:   rc.Logger,
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
