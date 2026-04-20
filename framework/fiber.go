package framework

import (
	"crypto/tls"
	"net"
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

	if cfg.Server.ACME != nil && cfg.Server.ACME.Enable {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
		httpServer := &httpServerToStd{TLSConfig: tlsCfg}

		acmeCfg := cfg.Server.ACME
		if acmeCfg.DNS01 != nil {
			if err := startACMEDNS01(httpServer, acmeCfg, logger); err != nil {
				return err
			}
		} else {
			if err := startACMEHTTP01(httpServer, cfg, acmeCfg, logger); err != nil {
				return err
			}
		}

		tlsLn := tls.NewListener(ln, tlsCfg)

		logger.Info("TLS listening", zap.String("addr", addr))

		return ctx.Listener(tlsLn, fiber.ListenConfig{EnablePrefork: !debug.IsDebugging()})
	}

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

type httpServerToStd struct {
	TLSConfig *tls.Config
}

func (s *httpServerToStd) SetTLSConfig(cfg *tls.Config) {
	s.TLSConfig = cfg
}
