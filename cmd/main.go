package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/cli"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func main() {
	opts := cli.Execute()
	if opts == nil { // help/version already printed
		return
	}

	config.CLIConfigPath = opts.ConfigPath
	cfg, path, err := config.LoadConfig("")

	if opts.CheckOnly {
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

	if errors.Is(err, os.ErrNotExist) {
		cfg = config.GetDefaultConfig()
		if err := cfg.Validate(); err != nil {
			panic(err)
		}

		if err := config.SaveConfig(cfg, path); err != nil {
			panic(err)
		}

		log.Writef("Using default config, saved to: %s", path)
	} else if err != nil {
		panic(err)
	}

	logger, logFile := log.NewLog(cfg.Log.Dir.Must(), cfg.Log.Level.Must(), cfg.Log.FileFormat.Must())
	defer func() {
		_ = logger.Flush()

		if logFile != nil {
			_ = logFile.Close()
		}
	}()

	log.Flush(logger.Logger)
	cfg.LogWarnings(logger.Logger)

	executor, err := execute.Build(cfg)
	if err != nil {
		panic(err)
	}

	manager := &runtime.Manager{Logger: logger.Logger, ConfigPath: path}
	manager.Executor.Store(executor)

	defer manager.Stop()

	server := fiber.New(fiber.Config{
		AppName:       "CatSync",
		CaseSensitive: true,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			traceID := uuid.NewString()
			logger.Error(fmt.Sprintf("EH >> %+v", err),
				"traceID", traceID,
				"stack", util.LazyStack{},
				"ctx", util.LazyFiberContext{Ctx: ctx},
			)
			ctx.Status(fiber.StatusInternalServerError)

			return ctx.JSON(fiber.Map{
				"msg":  "oops, something went wrong",
				"data": fiber.Map{"traceID": traceID},
			})
		},
		IdleTimeout:   5 * time.Second,
		ReadTimeout:   10 * time.Second,
		StrictRouting: true,
		WriteTimeout:  10 * time.Second,
	})
	router.Init(server, manager)

	if !fiber.IsChild() {
		logger.Info("Starting CatSync...", "version", cli.GetFormatVersion())
	}

	shutdown, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := cfg.Server.Address.Must()
		cert := cfg.Server.TLS.Cert.Must()
		key := cfg.Server.TLS.Key.Must()
		listen := fiber.ListenConfig{EnablePrefork: cfg.Server.Prefork && !debug.Enabled}

		protocol := "http"
		if cert != "" && key != "" {
			protocol = "https"
			listen.CertFile = cert
			listen.CertKeyFile = key
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

		if listen.CertFile != "" {
			logger.Info("TLS listening", "addr", addr)
		} else {
			logger.Info("HTTP listening", "addr", addr)
		}

		if err := server.Listen(addr, listen); err != nil {
			logger.Error("Server failed to start",
				"error", err,
				"stack", util.LazyStack{},
			)
			_ = logger.Flush()

			if logFile != nil {
				_ = logFile.Close()
			}

			os.Exit(1)
		}
	}()

	<-shutdown.Done()
	logger.Info("Shutting down server...")

	if err := server.Shutdown(); err != nil {
		logger.Error("Server failed to shutdown",
			"error", err,
			"stack", util.LazyStack{},
		)
	}
}
