package handler

import (
	"fmt"
	"regexp"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Actions(c *params.Ctx) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		for _, act := range c.GetConfig().Actions {
			// router matcher
			re, err := regexp.Compile(act.Route)
			if err != nil {
				return fmt.Errorf("invalid route regexp %q: %w", act.Route, err)
			}

			if !re.MatchString(ctx.Path()) {
				c.GetLogger().Info("Router >> Router not matched",
					zap.String("route", act.Route),
					zap.String("ctx", util.FiberContextString(ctx)),
				)

				continue
			}

			// start verify auth
			auth := act.Auth
			for k, v := range auth.Header {
				for k1, v1 := range ctx.GetReqHeaders() {
					if k != k1 {
						continue
					}

					for _, vv := range v {
						for _, vv1 := range v1 {
							re, err := regexp.Compile(vv)
							if err != nil {
								return fmt.Errorf("invalid header value regexp %q: %w", auth.Header, err)
							}

							if !re.MatchString(vv1) {
								c.GetLogger().Info("Router >> Header value not matched",
									zap.String("header", k1),
									zap.String("ctx", util.FiberContextString(ctx)),
								)

								return ctx.Next()
							}
						}
					}
				}
			}

			for k, v := range auth.Query {
				re, err := regexp.Compile(v)
				if err != nil {
					return fmt.Errorf("invalid query value regexp %q: %w", v, err)
				}

				if !re.MatchString(ctx.Query(k)) {
					c.GetLogger().Info("Router >> Query value not matched",
						zap.String("key", k),
						zap.String("expected_regex", v),
						zap.String("actual", ctx.Query(k)),
						zap.String("ctx", util.FiberContextString(ctx)),
					)

					return ctx.Next()
				}
			}

			// set cfg response header
			for k, v := range act.ResponseHeader {
				ctx.Append(k, v...)
			}

			// get action handler from map
			handler, ok := action.HandlerRegistry[act.Type]
			if !ok {
				c.GetLogger().Info("Router >> Unknown action", zap.String("type", string(act.Type)), zap.String("ctx", util.FiberContextString(ctx)))

				return ctx.Next()
			}

			var actionData config.ActionData

			switch act.Type {
			case config.ActionFile:
				actionData = act.ActionFile
			case config.ActionString:
				actionData = act.ActionString
			}

			p := &action.ProcessData{
				Ctx:     c,
				C:       ctx,
				Action:  &act,
				PayLoad: &actionData,
			}

			// "p" can change by hook!
			if hook := handler.HookProcessData(); hook != nil {
				var err error

				p, err = hook(p)
				if err != nil {
					return fmt.Errorf("hook process data error: %w", err)
				}
			}

			return handler.ProcessAction(p)
		}

		return nil
	}
}
