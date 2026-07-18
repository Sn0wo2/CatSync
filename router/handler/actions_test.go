package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sn0wo2/CatSync/internal/appctx"
	"github.com/Sn0wo2/CatSync/runtime"
	"github.com/gofiber/fiber/v3"
)

type unavailableRuntime struct {
	snapshot *runtime.Snapshot
}

func (r unavailableRuntime) Current() *runtime.Snapshot {
	return r.snapshot
}

func (unavailableRuntime) Reload() error {
	return nil
}

func TestActions_ReturnsInternalServerErrorWhenSnapshotIsUnavailable(t *testing.T) {
	t.Parallel()

	// Given: a handler whose runtime has no current snapshot.
	app := newUnavailableRuntimeApp(t, unavailableRuntime{})

	// When: it receives a request.
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	})

	// Then: it reports that the runtime snapshot is unavailable.
	assertSnapshotUnavailable(t, response)
}

func TestActions_ReturnsInternalServerErrorWhenSnapshotExecutorIsUnavailable(t *testing.T) {
	t.Parallel()

	// Given: a runtime snapshot without an executor.
	app := newUnavailableRuntimeApp(t, unavailableRuntime{snapshot: &runtime.Snapshot{}})

	// When: it receives a request.
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	})

	// Then: it reports that the runtime snapshot is unavailable.
	assertSnapshotUnavailable(t, response)
}

func newUnavailableRuntimeApp(t *testing.T, runtime unavailableRuntime) *fiber.App {
	t.Helper()

	app := fiber.New()
	app.Get("*", Actions(appctx.New(), runtime))

	return app
}

func assertSnapshotUnavailable(t *testing.T, response *http.Response) {
	t.Helper()

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusInternalServerError)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if !strings.Contains(string(body), "runtime snapshot unavailable") {
		t.Fatalf("body = %q, want unavailable snapshot message", body)
	}
}
