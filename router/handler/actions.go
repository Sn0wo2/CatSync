package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Actions(logger *zap.Logger, cfg *config.Config, act config.Action) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		auth := act.Auth
		if auth.UA != "" {
			re, err := regexp.Compile(auth.UA)
			if err != nil {
				return fmt.Errorf("invalid user agent regexp %q: %w", auth.UA, err)
			}

			if !re.MatchString(util.BytesToString(ctx.Request().Header.UserAgent())) {
				logger.Info("Router >> User agent not matched",
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
				logger.Info("Router >> Query not matched",
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
			logger.Info("Router >> Unknown action", zap.Int("action", int(act.Action)), zap.String("ctx", util.FiberContextString(ctx)))

			return ctx.Next()
		}

		return handler.Execute(logger, ctx, act.ActionData)
	}
}
