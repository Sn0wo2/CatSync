package framework

import (
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/router/errorhandler"
	"github.com/gofiber/fiber/v3"
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
	return &Framework{
		Provider: p,
		App: fiber.New(fiber.Config{
			AppName:       "CatSync",
			CaseSensitive: true,
			ErrorHandler:  errorhandler.Error(p.GetLogger()),
			IdleTimeout:   5 * time.Second,
			ReadTimeout:   10 * time.Second,
			StrictRouting: true,
			WriteTimeout:  10 * time.Second,
		}),
	}
}

func (ctx *Framework) StartFiber() error {
	cfg := ctx.GetConfig()
	logger := ctx.GetLogger()
	addr := cfg.Server.Address.Must()

	cert := cfg.Server.TLS.Cert.Must()
	key := cfg.Server.TLS.Key.Must()

	fl := fiber.ListenConfig{}

	if cfg.Server.Prefork {
		fl.EnablePrefork = !debug.IsDebugging()
	}

	if cert != "" && key != "" {
		logger.Info("TLS listening", zap.String("addr", addr))

		fl.CertFile = cert
		fl.CertKeyFile = key

		return ctx.Listen(addr, fl)
	}

	logger.Info("HTTP listening", zap.String("addr", addr))

	return ctx.Listen(addr, fl)
}
