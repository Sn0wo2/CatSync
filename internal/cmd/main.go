package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/Sn0wo2/FileSync/internal/config"
	"github.com/Sn0wo2/FileSync/internal/framework"
	"github.com/Sn0wo2/FileSync/internal/log"
	"github.com/Sn0wo2/FileSync/internal/router"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	debug.SetGCPercent(50)

	_ = godotenv.Load()

	if err := config.Init(config.NewYAMLLoader(), config.NewJSONLoader()); err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	log.Init()
}

func main() {
	defer func() {
		_ = log.Instance.Sync()
	}()

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
		log.Instance.Error("Server shutdown error",
			zap.Error(err),
		)
	}
}
