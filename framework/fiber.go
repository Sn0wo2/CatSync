package framework

import (
	"net/http"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/router/errorhandler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"go.uber.org/zap"
)

type Provider interface {
	GetLogger() *zap.Logger
	GetConfig() *config.Config
}

type Framework struct {
	Provider
	*fiber.App
}

func NewFiber(p Provider) *Framework {
	app := &Framework{
		Provider: p,
		App: fiber.New(fiber.Config{
			AppName:               "CatSync",
			CaseSensitive:         true,
			DisableStartupMessage: false,
			ErrorHandler:          errorhandler.Error(p.GetLogger()),
			IdleTimeout:           5 * time.Second,
			// dlv cant debug multiple process(prefork)
			Prefork:           !debug.IsDebugging(),
			ReadTimeout:       10 * time.Second,
			ReduceMemoryUsage: true,
			StrictRouting:     true,
			WriteTimeout:      10 * time.Second,
			ServerHeader:      p.GetConfig().Server.Header,
		}),
	}

	return app
}

func (ctx *Framework) StartFiber() error {
	server := &http.Server{
		Addr:         ctx.GetConfig().Server.Address,
		Handler:      adaptor.FiberApp(ctx.App),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorLog:     zap.NewStdLog(ctx.GetLogger()),
	}

	if ctx.GetConfig().Server.TLS.Cert != "" && ctx.GetConfig().Server.TLS.Key != "" {
		return server.ListenAndServeTLS(ctx.GetConfig().Server.TLS.Cert, ctx.GetConfig().Server.TLS.Key)
	}

	return server.ListenAndServe()
}
