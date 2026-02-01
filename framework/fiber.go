package framework

import (
	"context"
	"net/http"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
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

func sval(r *reader.String) string {
	if r == nil {
		return ""
	}

	s, _ := r.ReadString(context.Background())

	return s
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
		}),
	}

	return app
}

func (ctx *Framework) StartFiber() error {
	cfg := ctx.GetConfig()
	logger := ctx.GetLogger()

	addr, _ := cfg.Server.Address.ReadString(context.Background())

	server := &http.Server{
		Addr:         addr,
		Handler:      adaptor.FiberApp(ctx.App),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorLog:     zap.NewStdLog(logger),
	}

	if cfg.Server.ACME != nil && cfg.Server.ACME.Enable {
		acmeCfg := cfg.Server.ACME

		if acmeCfg.DNS01 != nil {
			return startACMEDNS01(server, acmeCfg, logger)
		}

		return startACMEHTTP01(server, cfg, acmeCfg, logger)
	}

	cert := ""
	key := ""

	if cfg.Server.TLS.Cert != nil {
		cert = sval(cfg.Server.TLS.Cert)
	}

	if cfg.Server.TLS.Key != nil {
		key = sval(cfg.Server.TLS.Key)
	}

	if cert != "" && key != "" {
		return server.ListenAndServeTLS(cert, key)
	}

	return server.ListenAndServe()
}
