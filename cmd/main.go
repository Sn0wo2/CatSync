package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sn0wo2/CatSync/cli"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func init() {
	cli.Execute()
}

func main() {
	appCtx, err := InitializeApp()
	if err != nil {
		panic(err)
	}

	cfg := appCtx.Cfg
	logger := appCtx.Logger
	app := appCtx.App

	defer func() {
		_ = logger.Sync()
	}()

	if !fiber.IsChild() {
		logger.Info("Starting CatSync...", zap.String("version", version.GetFormatVersion()))
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		addr, cert, key := cfg.Server.Address, cfg.Server.TLS.Cert, cfg.Server.TLS.Key

		protocol := "http"
		if (cert != "" && key != "") || (cfg.Server.ACME != nil && cfg.Server.ACME.Enable) {
			protocol = "https"
		} else {
			logger.Warn("TLS is not enabled (no cert/key and acme disabled)")
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
