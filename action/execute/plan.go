package execute

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
)

type Plan struct {
	actions []plannedAction
}

type plannedAction struct {
	act   *config.Action
	re    *regexp.Regexp
	h     action.Handler
	aType config.ActionType
}

func Compile(cfg *config.Config, builders Builders) (*Plan, error) {
	if cfg == nil {
		return nil, errors.New("nil config")
	}

	var global []action.Modifier
	if builders.Global != nil {
		global = builders.Global(cfg)
	}

	pl := &Plan{actions: make([]plannedAction, 0, len(cfg.Actions))}

	for i := range cfg.Actions {
		act := &cfg.Actions[i]

		baseHandler, ok := action.HandlerRegistry[act.Type]
		if !ok {
			// keep behavior: unknown action is treated as not matched
			pl.actions = append(pl.actions, plannedAction{act: act, aType: act.Type})

			continue
		}

		// Compile route once.
		route, ok := reader.LiteralTrim(act.Route)
		if !ok {
			return nil, fmt.Errorf("actions[%d].route must be literal string", i)
		}

		var re *regexp.Regexp
		if route != "" {
			re, _ = util.GetCompiledRegexp(route)
		}

		h := action.WrapHandler(baseHandler)

		add := func(ms []action.Modifier) {
			for _, m := range ms {
				h.WithModifier(m)
			}
		}

		if !act.SkipGlobalModifiers {
			add(global)
		}

		if builders.Action != nil {
			add(builders.Action(act))
		}

		var payload config.ActionData

		switch act.Type {
		case config.ActionFile:
			payload = act.ActionFile
		case config.ActionString:
			payload = act.ActionString
		}

		if builders.Payload != nil {
			add(builders.Payload(payload))
		}

		// Version placeholder modifier is pure and can be pre-built.
		ph := ""

		switch act.Type {
		case config.ActionString:
			if act.ActionString != nil && act.ActionString.ActionModifierVersion != nil {
				ph = strings.TrimSpace(reader.Must(act.ActionString.Placeholder))
			}
		case config.ActionFile:
			if act.ActionFile != nil && act.ActionFile.ActionModifierVersion != nil {
				ph = strings.TrimSpace(reader.Must(act.ActionFile.Placeholder))
			}
		}

		if ph != "" {
			h.WithModifier(action.NewPlaceholderModifier().WithPlaceholder(ph).WithValue(version.GetFormatVersion()))
		}

		pl.actions = append(pl.actions, plannedAction{act: act, re: re, h: h.Build(), aType: act.Type})
	}

	return pl, nil
}

type Runner struct {
	pl        *Plan
	ctx       *params.Ctx
	fiberCtx  *fiber.Ctx
	skipRoute bool
}

func (p *Plan) Runner(ctx *params.Ctx, fiberCtx *fiber.Ctx) *Runner {
	return &Runner{pl: p, ctx: ctx, fiberCtx: fiberCtx}
}

func (r *Runner) WithSkipRouteCheck(skip bool) *Runner {
	r.skipRoute = skip

	return r
}

func (r *Runner) ExecuteAt(index int) (Result, error) {
	if r == nil || r.pl == nil {
		return Result{}, errors.New("nil runner")
	}

	if index < 0 || index >= len(r.pl.actions) {
		return Result{}, fmt.Errorf("invalid action index: %d", index)
	}

	pa := r.pl.actions[index]
	if pa.h == nil {
		return Result{NotMatched: true}, nil
	}

	if !r.skipRoute {
		if pa.re == nil {
			return Result{NotMatched: true}, nil
		}

		if !pa.re.MatchString(r.fiberCtx.Path()) {
			return Result{NotMatched: true}, nil
		}
	}

	var payload config.ActionData

	switch pa.aType {
	case config.ActionFile:
		payload = pa.act.ActionFile
	case config.ActionString:
		payload = pa.act.ActionString
	}

	err := pa.h.ProcessAction(&action.ProcessData{Ctx: r.ctx, C: r.fiberCtx, Action: pa.act, PayLoad: payload})
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
