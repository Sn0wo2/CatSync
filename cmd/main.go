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
	"github.com/Sn0wo2/CatSync/config/file"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func init() {
	// _ = godotenv.Load()
	cli.Execute()

	if err := config.Init(file.NewYAMLLoader(), file.NewJSONLoader()); err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			envPath := config.Path
			if envPath == "" {
				envPath = "./data/config.yml"
			}

			config.Instance = config.DefaultConfig
			if err := file.NewYAMLLoader().Save(config.DefaultConfig, envPath); err != nil {
				panic(err)
			}

			config.Default = true
			config.Path = envPath
		} else {
			panic(err)
		}
	}
}

func main() {
	log.Init()

	defer func() {
		_ = log.Instance.Sync()
	}()

	if !fiber.IsChild() {
		log.Instance.Info("Starting CatSync...", zap.String("version", version.GetFormatVersion()))
	}

	if config.Default {
		log.Instance.Warn("You have not set a configuration file, using the default!", zap.String("path", config.Path))
	}

	app := framework.Fiber()

	router.Init(app)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		addr, cert, key := config.Instance.Server.Address, config.Instance.Server.TLS.Cert, config.Instance.Server.TLS.Key

		protocol := "http"
		if cert != "" && key != "" {
			protocol = "https"
		}

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			if strings.HasPrefix(addr, ":") {
				port = strings.TrimPrefix(addr, ":")
				host = ""
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

		log.Instance.Info("Server listening on: " + strings.Join(logAddresses, ", "))

		if err := framework.Start(app, addr, cert, key); err != nil {
			log.Instance.Fatal("Server failed to start",
				zap.Error(err),
			)
		}
	}()

	<-shutdownChan

	if err := app.Shutdown(); err != nil {
		log.Instance.Error("Server failed to shutdown",
			zap.Error(err),
		)
	}
}
