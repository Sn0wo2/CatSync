package execute

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
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

type Builders struct {
	Global  func(*config.Config) []action.Modifier
	Action  func(*config.Action) []action.Modifier
	Payload func(config.ActionData) []action.Modifier
}

type routeEntry struct {
	empty bool
	exact string
	re    *regexp.Regexp
}

type actionEntry struct {
	route   routeEntry
	handler action.Handler
	payload config.ActionData
}

const regexMetaChars = `.*+?()[]{}|\^$`

func isExactRoute(pattern string) bool {
	return !strings.ContainsAny(pattern, regexMetaChars)
}

type Executor struct {
	cfg      *config.Config
	builders Builders
	entries  []actionEntry // prebuilt per-action, indexed same as cfg.Actions
}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) WithConfig(cfg *config.Config) *Executor {
	e.cfg = cfg

	return e
}

func (e *Executor) WithBuilders(builders Builders) *Executor {
	e.builders = builders

	return e
}

func (e *Executor) Build() (*Executor, error) {
	if e == nil || e.cfg == nil {
		return nil, errors.New("nil executor or config")
	}

	actions := e.cfg.Actions
	e.entries = make([]actionEntry, len(actions))

	for i := range actions {
		act := &actions[i]
		entry := &e.entries[i]

		route, ok := act.Route.Literal()
		if !ok {
			return nil, fmt.Errorf("actions[%d].route must be a literal string", i)
		}

		switch {
		case route == "":
			entry.route.empty = true
		case isExactRoute(route):
			entry.route.exact = route
		default:
			re, err := util.GetCompiledRegexp(route)
			if err != nil {
				return nil, fmt.Errorf("invalid action route regexp at actions[%d].route (%q): %w", i, route, err)
			}

			entry.route.re = re
		}

		baseHandler, ok := action.HandlerRegistry[act.Type]
		if !ok || baseHandler == nil {
			return nil, fmt.Errorf("unknown action handler at actions[%d]: %s", i, act.Type)
		}

		payload := act.GetPayload()
		if payload == nil {
			return nil, fmt.Errorf("actions[%d] type=%s but payload is nil", i, act.Type)
		}

		entry.payload = payload

		h := action.WrapHandler(baseHandler)

		add := func(ms []action.Modifier) {
			for _, m := range ms {
				h.WithModifier(m)
			}
		}

		if !act.SkipGlobalModifiers && e.builders.Global != nil {
			add(e.builders.Global(e.cfg))
		}

		if e.builders.Action != nil {
			add(e.builders.Action(act))
		}

		if e.builders.Payload != nil {
			add(e.builders.Payload(payload))
		}

		entry.handler = h
	}

	return e, nil
}

type RequestContext struct {
	Ctx      *params.Ctx
	FiberCtx fiber.Ctx
	Reloader action.Reloader
}

const maxJumpDepth = 16

func (e *Executor) Dispatch(rc *RequestContext) (bool, error) {
	if e == nil || e.cfg == nil {
		return false, errors.New("nil executor or config")
	}

	if rc == nil || rc.Ctx == nil || rc.FiberCtx == nil {
		return false, errors.New("nil request context")
	}

	n := len(e.cfg.Actions)
	if n == 0 {
		return false, nil
	}

	var (
		visitedBits uint64
		visitedMap  map[int]struct{}
	)

	markVisited := func(idx int) bool {
		if n <= 64 {
			bit := uint64(1) << idx
			if visitedBits&bit != 0 {
				return false
			}

			visitedBits |= bit

			return true
		}

		if visitedMap == nil {
			visitedMap = make(map[int]struct{}, maxJumpDepth)
		}

		if _, dup := visitedMap[idx]; dup {
			return false
		}

		visitedMap[idx] = struct{}{}

		return true
	}

	for i := range n {
		res, err := e.executeOne(rc, i, false)
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
			for range maxJumpDepth {
				if target < 0 || target >= n {
					return false, fmt.Errorf("invalid auth fallback jumpTo index: %d", target)
				}

				if !markVisited(target) {
					return false, fmt.Errorf("auth fallback jump loop detected at index: %d", target)
				}

				jres, jerr := e.executeOne(rc, target, true)
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

	lastRes, lastErr := e.executeOne(rc, n-1, true)
	if lastErr != nil {
		return false, lastErr
	}

	return lastRes.Status == StatusMatched, nil
}

func (e *Executor) matchRoute(index int, path string) bool {
	entry := &e.entries[index]

	if entry.route.empty {
		return false
	}

	if entry.route.exact != "" {
		return entry.route.exact == path
	}

	if entry.route.re != nil {
		return entry.route.re.MatchString(path)
	}

	return false
}

func (e *Executor) executeOne(rc *RequestContext, index int, skipRoute bool) (Result, error) {
	if index < 0 || index >= len(e.entries) {
		return Result{}, fmt.Errorf("invalid action index: %d", index)
	}

	if !skipRoute && !e.matchRoute(index, rc.FiberCtx.Path()) {
		return Result{Status: StatusNotMatched}, nil
	}

	entry := &e.entries[index]
	if entry.handler == nil || entry.payload == nil {
		return Result{Status: StatusNotMatched}, nil
	}

	act := e.cfg.Actions[index]

	result := entry.handler.ProcessAction(&action.ProcessData{
		Ctx:      rc.Ctx,
		C:        rc.FiberCtx,
		Action:   &act,
		Payload:  entry.payload,
		Reloader: rc.Reloader,
	})
	if result.Err != nil {
		return Result{Status: StatusMatched}, result.Err
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
