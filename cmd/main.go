package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	config2 "github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	log2 "github.com/Sn0wo2/CatSync/log"
	"github.com/Sn0wo2/CatSync/router"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	debug.SetGCPercent(50)

	_ = godotenv.Load()

	if err := config2.Init(config2.NewYAMLLoader(), config2.NewJSONLoader()); err != nil {
		panic(fmt.Errorf("failed to initialize config: %w", err))
	}

	log2.Init()
}

func main() {
	defer func() {
		_ = log2.Instance.Sync()
	}()

	app := framework.Fiber()

	router.Init(app)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := framework.Start(app); err != nil {
			log2.Instance.Fatal("Server failed to start",
				zap.Error(err),
			)
		}
	}()

	<-shutdownChan

	if err := app.Shutdown(); err != nil {
		log2.Instance.Error("Server shutdown error",
			zap.Error(err),
		)
	}
}
