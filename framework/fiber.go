package framework

import (
	"log/slog"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/router/errorhandler"
	"github.com/gofiber/fiber/v3"
)

type Provider interface {
	GetLogger() *slog.Logger
	GetConfig() *config.Config
}

type FB struct {
	Provider
	*fiber.App
}

func NewFiber(p Provider) *FB {
	return &FB{
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

func (fb *FB) StartFiber() error {
	cfg := fb.GetConfig()
	logger := fb.GetLogger()
	addr := cfg.Server.Address.Must()

	cert := cfg.Server.TLS.Cert.Must()
	key := cfg.Server.TLS.Key.Must()

	fl := fiber.ListenConfig{}

	if cfg.Server.Prefork {
		fl.EnablePrefork = !debug.IsDebugging()
	}

	if cert != "" && key != "" {
		logger.Info("TLS listening", "addr", addr)

		fl.CertFile = cert
		fl.CertKeyFile = key

		return fb.Listen(addr, fl)
	}

	logger.Info("HTTP listening", "addr", addr)

	return fb.Listen(addr, fl)
}
