package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/go-common/helper"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Actions(c *params.Ctx, act config.Action) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		auth := act.Auth
		if auth.UA != "" {
			re, err := regexp.Compile(auth.UA)
			if err != nil {
				return fmt.Errorf("invalid user agent regexp %q: %w", auth.UA, err)
			}

			if !re.MatchString(helper.BytesToString(ctx.Request().Header.UserAgent())) {
				c.GetLogger().Info("Router >> User agent not matched",
					zap.String("ua", auth.UA),
					zap.String("ctx", util.FiberContextString(ctx)),
				)

				return ctx.Next()
			}
		}

		for k, v := range auth.Query.Map {
			if auth.Query.IgnoreCaseCase {
				k = strings.ToLower(k)
				v = strings.ToLower(v)
			}

			if ctx.Query(k) != v {
				c.GetLogger().Info("Router >> Query not matched",
					zap.String("key", k),
					zap.String("expected", v),
					zap.String("actual", ctx.Query(k)),
					zap.String("ctx", util.FiberContextString(ctx)),
				)

				return ctx.Next()
			}
		}

		for k, v := range act.ResponseHeader {
			ctx.Append(k, v...)
		}

		handler, ok := action.HandlerRegistry[act.Operation]
		if !ok {
			c.GetLogger().Info("Router >> Unknown action", zap.String("operation", string(act.Operation)), zap.String("ctx", util.FiberContextString(ctx)))

			return ctx.Next()
		}

		var actionData any

		switch act.Operation {
		case config.OperationFile:
			actionData = act.ActionOperationFile
		case config.OperationString:
			actionData = act.ActionOperationString
		}

		return handler.Execute(c, ctx, actionData)
	}
}
