package main

import (
	"errors"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/file"
	"github.com/Sn0wo2/CatSync/framework"
	"github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/Sn0wo2/CatSync/version"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	debug.SetGCPercent(50)

	_ = godotenv.Load()

	if err := config.Init(file.NewYAMLLoader(), file.NewJSONLoader()); err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			config.Instance = config.DefaultConfig
			if err := file.NewYAMLLoader().Save("./data/config.yml", config.DefaultConfig); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	log.Init()
}

func main() {
	defer func() {
		_ = log.Instance.Sync()
	}()

	if !fiber.IsChild() {
		log.Instance.Info("Starting CatSync...", zap.String("version", version.GetFormatVersion()))
	}

	app := framework.Fiber()

	router.Init(app)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := framework.Start(app); err != nil {
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
