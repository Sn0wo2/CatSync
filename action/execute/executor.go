package execute

import (
	"errors"
	"fmt"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
)

// Result indicates how the caller should continue processing.
//
// - Matched means the action was selected and executed successfully.
// - NotMatched means the action does not match current request and caller should try next action.
// - Jump means auth fallback requested a jump to another action index; caller should execute that index.
type Result struct {
	Matched    bool
	NotMatched bool
	JumpTo     *int
}

type Builders struct {
	Global  func(*config.Config) []action.Modifier
	Action  func(*config.Action) []action.Modifier
	Payload func(config.ActionData) []action.Modifier
}

// Executor is a per-request action executor.
// Routers should create one and then call ExecuteAt.
type Executor struct {
	cfg      *config.Config
	builders Builders

	ctx      *params.Ctx
	fiberCtx *fiber.Ctx

	skipRouteCheck bool
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

func (e *Executor) WithGlobalBuilder(fn func(*config.Config) []action.Modifier) *Executor {
	e.builders.Global = fn

	return e
}

func (e *Executor) WithActionBuilder(fn func(*config.Action) []action.Modifier) *Executor {
	e.builders.Action = fn

	return e
}

func (e *Executor) WithPayloadBuilder(fn func(config.ActionData) []action.Modifier) *Executor {
	e.builders.Payload = fn

	return e
}

func (e *Executor) WithContext(ctx *params.Ctx, fiberCtx *fiber.Ctx) *Executor {
	e.ctx = ctx
	e.fiberCtx = fiberCtx

	return e
}

// WithSkipRouteCheck forces execution regardless of action.route.
//
// This is used for jump-only actions (route is empty) and auth fallback jumps.
func (e *Executor) WithSkipRouteCheck(skip bool) *Executor {
	e.skipRouteCheck = skip

	return e
}

func (e *Executor) ExecuteAt(index int) (Result, error) {
	if e == nil {
		return Result{}, errors.New("nil executor")
	}

	if e.cfg == nil {
		return Result{}, errors.New("nil config")
	}

	if e.ctx == nil {
		return Result{}, errors.New("nil params ctx")
	}

	if e.fiberCtx == nil {
		return Result{}, errors.New("nil fiber ctx")
	}

	if index < 0 || index >= len(e.cfg.Actions) {
		return Result{}, fmt.Errorf("invalid action index: %d", index)
	}

	act := e.cfg.Actions[index]

	if !e.skipRouteCheck {
		route := reader.Must(act.Route)
		if route == "" {
			return Result{NotMatched: true}, nil
		}

		re, err := util.GetCompiledRegexp(route)
		if err != nil {
			return Result{}, fmt.Errorf("invalid route regexp %q: %w", route, err)
		}

		if !re.MatchString(e.fiberCtx.Path()) {
			return Result{NotMatched: true}, nil
		}
	}

	baseHandler, ok := action.HandlerRegistry[act.Type]
	if !ok {
		return Result{NotMatched: true}, nil
	}

	// Validate payload presence early to avoid nil deref later.
	switch act.Type {
	case config.ActionString:
		if act.ActionString == nil {
			return Result{}, fmt.Errorf("action[%d] type=string but string is nil", index)
		}
	case config.ActionFile:
		if act.ActionFile == nil {
			return Result{}, fmt.Errorf("action[%d] type=file but file is nil", index)
		}
	case config.ActionServer:
		if act.ActionServer == nil {
			return Result{}, fmt.Errorf("action[%d] type=server but server is nil", index)
		}
	}

	h := action.WrapHandler(baseHandler)

	add := func(ms []action.Modifier) {
		for _, m := range ms {
			h.WithModifier(m)
		}
	}

	if !act.SkipGlobalModifiers {
		if e.builders.Global != nil {
			add(e.builders.Global(e.cfg))
		}
	}

	if e.builders.Action != nil {
		add(e.builders.Action(&act))
	}

	var payload config.ActionData

	switch act.Type {
	case config.ActionFile:
		payload = act.ActionFile
	case config.ActionString:
		payload = act.ActionString
	case config.ActionServer:
		payload = act.ActionServer
	case config.ActionReload:
		payload = act.ActionReload
	}

	if e.builders.Payload != nil {
		add(e.builders.Payload(payload))
	}

	switch act.Type {
	case config.ActionString:
		if act.ActionString != nil && act.ActionString.ActionModifierVersion != nil {
			ph := reader.Must(act.ActionString.Placeholder)
			if ph != "" {
				h.WithModifier(action.NewPlaceholderModifier().WithPlaceholder(ph).WithValue(version.GetFormatVersion()))
			}
		}
	case config.ActionFile:
		if act.ActionFile != nil && act.ActionFile.ActionModifierVersion != nil {
			ph := reader.Must(act.ActionFile.Placeholder)
			if ph != "" {
				h.WithModifier(action.NewPlaceholderModifier().WithPlaceholder(ph).WithValue(version.GetFormatVersion()))
			}
		}
	case config.ActionServer:
		if act.ActionServer != nil && act.ActionServer.ActionModifierVersion != nil {
			ph := reader.Must(act.ActionServer.Placeholder)
			if ph != "" {
				h.WithModifier(action.NewPlaceholderModifier().WithPlaceholder(ph).WithValue(version.GetFormatVersion()))
			}
		}
	}

	err := h.Build().ProcessAction(&action.ProcessData{Ctx: e.ctx, C: e.fiberCtx, Action: &act, PayLoad: payload})
	if err == nil {
		return Result{Matched: true}, nil
	}

	var jumpErr *action.AuthFallbackJumpError
	if errors.As(err, &jumpErr) {
		return Result{Matched: true, JumpTo: &jumpErr.JumpTo}, nil
	}

	var nextErr *action.AuthFallbackNextError
	if errors.As(err, &nextErr) {
		return Result{NotMatched: true}, nil
	}

	return Result{Matched: true}, err
}
