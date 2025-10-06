package framework

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/debug"
	"github.com/Sn0wo2/CatSync/router/errorhandler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func Fiber() *fiber.App {
	return fiber.New(fiber.Config{
		AppName:               "CatSync",
		CaseSensitive:         true,
		DisableStartupMessage: false,
		ErrorHandler:          errorhandler.Error,
		IdleTimeout:           5 * time.Second,
		// dlv cant debug multiple process
		Prefork:           !debug.IsDebugging(),
		ReadTimeout:       10 * time.Second,
		ReduceMemoryUsage: true,
		StrictRouting:     true,
		WriteTimeout:      10 * time.Second,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
		ServerHeader:      config.Instance.Server.Header,
	})
}

func Start(app *fiber.App) error {
	addr := config.Instance.Server.Address

	httpHandler := adaptor.FiberApp(app)

	if config.Instance.Server.TLS.Cert != "" && config.Instance.Server.TLS.Key != "" {
		return http.ListenAndServeTLS(addr, config.Instance.Server.TLS.Cert, config.Instance.Server.TLS.Key, httpHandler)
	}

	return http.ListenAndServe(addr, httpHandler)
}
