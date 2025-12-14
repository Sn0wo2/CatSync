package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sn0wo2/CatSync/cli"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/loader"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/params"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func init() {
	// _ = godotenv.Load()
	cli.Execute()
}

func main() {
	var cfgDefault bool

	cfg, err := config.New(loader.NewYAMLLoader(), loader.NewJSONLoader())
	if err != nil {
		if !errors.Is(err, config.ErrConfigNotFound) {
			panic(err)
		}

		cfg = config.GetDefaultConfig()
		if err := loader.NewYAMLLoader().Save(cfg, config.Path); err != nil {
			panic(err)
		}

		cfgDefault = true
	}

	logger := log.NewLog(cfg.Log.Dir, cfg.Log.Level, cfg.Log.FileFormat)

	defer func() {
		_ = logger.Sync()
	}()

	p := params.New()
	p.SetConfig(cfg)
	p.SetLogger(logger)

	if !fiber.IsChild() {
		logger.Info("Starting CatSync...", zap.String("version", version.GetFormatVersion()))
	}

	if cfgDefault {
		logger.Warn("You have not set a configuration file, using the default!", zap.String("path", config.Path))
	}

	app := framework.NewFiber(p)
	p.SetFramework(app)

	router.Init(p)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		addr, cert, key := cfg.Server.Address, cfg.Server.TLS.Cert, cfg.Server.TLS.Key

		protocol := "http"
		if cert != "" && key != "" {
			protocol = "https"
		} else {
			logger.Warn("TLS is not enabled, because cert or key is empty!")
		}

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			if strings.HasPrefix(addr, ":") {
				port = strings.TrimPrefix(addr, ":")
			}
		}

		var logAddresses []string

		if host == "" || host == "0.0.0.0" {
			logAddresses = append(logAddresses, fmt.Sprintf("%s://localhost:%s", protocol, port))

			if ifaces, err := net.InterfaceAddrs(); err == nil {
				for _, i := range ifaces {
					if ipnet, ok := i.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
						logAddresses = append(logAddresses, fmt.Sprintf("%s://%s:%s", protocol, ipnet.IP.String(), port))
					}
				}
			}
		} else {
			logAddresses = append(logAddresses, fmt.Sprintf("%s://%s:%s", protocol, host, port))
		}

		logger.Info("Server listening on: " + strings.Join(logAddresses, ", "))

		if err := app.StartFiber(); err != nil {
			logger.Fatal("Server failed to start",
				zap.Error(err),
			)
		}
	}()

	<-shutdownChan

	logger.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		logger.Error("Server failed to shutdown",
			zap.Error(err),
		)
	}
}
