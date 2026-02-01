package framework

import (
	"net/http"
	"strings"
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
		}),
	}

	return app
}

func (ctx *Framework) StartFiber() error {
	cfg := ctx.GetConfig()
	logger := ctx.GetLogger()

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      adaptor.FiberApp(ctx.App),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorLog:     zap.NewStdLog(logger),
	}

	if cfg.Server.ACME != nil && cfg.Server.ACME.Enable {
		acmeCfg := cfg.Server.ACME
		challenge := strings.ToLower(strings.TrimSpace(acmeCfg.Challenge))
		if challenge == "" {
			if acmeCfg.DNS != nil {
				challenge = "dns-01"
			} else {
				challenge = "http-01"
			}
		}

		if challenge == "dns-01" {
			return startACMEDNS01(server, acmeCfg, logger)
		}

		return startACMEHTTP01(server, cfg, acmeCfg, logger)
	}

	if cfg.Server.TLS.Cert != "" && cfg.Server.TLS.Key != "" {
		return server.ListenAndServeTLS(cfg.Server.TLS.Cert, cfg.Server.TLS.Key)
	}

	return server.ListenAndServe()
}
