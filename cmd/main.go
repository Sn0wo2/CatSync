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
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	opts := cli.Execute()
	if opts == nil { // help/version already printed
		return
	}

	if opts.ConfigPath != "" {
		config.CLIConfigPath = opts.ConfigPath
	}

	if opts.CheckOnly {
		_, path, err := config.LoadConfig("", loader.NewYAMLLoader(), loader.NewJSONLoader())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "CONFIGURATION ERROR: config file not found\n")
			} else {
				fmt.Fprintf(os.Stderr, "CONFIGURATION ERROR: %v\n", err)
			}

			os.Exit(1)
		}

		fmt.Printf("OK - %s\nConfiguration is valid\n", path)

		return
	}

	catSync, err := InitializeCatSync()
	if err != nil {
		panic(err)
	}

	logger := catSync.Logger
	server := catSync.Server

	defer func() {
		_ = logger.Sync()
	}()

	if !fiber.IsChild() {
		logger.Info("Starting CatSync...", zap.String("version", cli.GetFormatVersion()))
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		cfg := catSync.Runtime.CurrentConfig()
		if cfg == nil {
			logger.Fatal("Runtime config unavailable")
		}

		addr := cfg.Server.Address.Must()
		cert := cfg.Server.TLS.Cert.Must()
		key := cfg.Server.TLS.Key.Must()

		protocol := "http"
		if cert != "" && key != "" {
			protocol = "https"
		} else {
			logger.Warn("TLS is not enabled (no cert/key)")
		}

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			if after, found := strings.CutPrefix(addr, ":"); found {
				port = after
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

		if err := server.StartFiber(); err != nil {
			logger.Fatal("Server failed to start",
				zap.Error(err),
			)
		}
	}()

	<-shutdownChan
	logger.Info("Shutting down server...")

	if err := server.Shutdown(); err != nil {
		logger.Error("Server failed to shutdown",
			zap.Error(err),
		)
	}
}
