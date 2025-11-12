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

type Context struct {
	*fiber.App
	Logger *zap.Logger
	Config *config.Config
}

func NewFiber(logger *zap.Logger, cfg *config.Config) *Context {
	app := &Context{
		Logger: logger,
		Config: cfg,
	}

	app.App = fiber.New(fiber.Config{
		AppName:               "CatSync",
		CaseSensitive:         true,
		DisableStartupMessage: false,
		ErrorHandler:          errorhandler.Error(logger),
		IdleTimeout:           5 * time.Second,
		// dlv cant debug multiple process(prefork)
		Prefork:           !debug.IsDebugging(),
		ReadTimeout:       10 * time.Second,
		ReduceMemoryUsage: true,
		StrictRouting:     true,
		WriteTimeout:      10 * time.Second,
		ServerHeader:      cfg.Server.Header,
	})

	return app
}

func (ctx *Context) StartFiber(addr, cert, key string) error {
	server := &http.Server{
		Addr:         addr,
		Handler:      adaptor.FiberApp(ctx.App),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if cert != "" && key != "" {
		return server.ListenAndServeTLS(cert, key)
	}

	return server.ListenAndServe()
}
