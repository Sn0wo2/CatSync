package action

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sn0wo2/CatSync/internal/appctx"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func executeAction(t *testing.T, path string, process func(*ProcessData) ExecutionResult) *http.Response {
	t.Helper()

	ctx := appctx.New()
	ctx.Logger = zap.NewNop()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		return process(&ProcessData{
			Ctx: ctx,
			C:   c,
		}).Err
	})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("execute action: %v", err)
	}

	return response
}
