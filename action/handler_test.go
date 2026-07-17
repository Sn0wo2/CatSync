package action

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
)

func TestModifierBuildersReturnNilForNilInputs(t *testing.T) {
	t.Parallel()

	// Given: absent configuration at every builder boundary.
	// When: modifiers are built.
	// Then: each builder returns no modifiers without panicking.
	if got := BuildGlobalModifiers(nil); got != nil {
		t.Fatalf("global modifiers = %v, want nil", got)
	}

	if got := BuildActionModifiers(nil); got != nil {
		t.Fatalf("action modifiers = %v, want nil", got)
	}

	if got := BuildPayloadModifiers(nil); got != nil {
		t.Fatalf("payload modifiers = %v, want nil", got)
	}
}

func TestReloadHandlerRejectsUnavailableReloader(t *testing.T) {
	t.Parallel()

	// Given: a reload action without the runtime reloader.
	handler := NewReload()

	// When: the action is processed.
	result := handler.ProcessAction(&ProcessData{Payload: &config.ActionReloadData{}})

	// Then: it reports the unavailable runtime dependency.
	if result.Err == nil || result.Err.Error() != "runtime reloader is unavailable" {
		t.Fatalf("result error = %v, want unavailable reloader error", result.Err)
	}
}

//nolint:paralleltest // t.Chdir changes the process-wide working directory.
func TestFileHandlerRejectsPathsOutsideDataDirectory(t *testing.T) {
	// Given: an isolated working directory and a path outside its data directory.
	t.Chdir(t.TempDir())
	outsidePath := filepath.Join(t.TempDir(), "outside.txt")
	handler := NewFile()

	// When: the file action is processed.
	result := handler.ProcessAction(&ProcessData{Payload: &config.ActionFileData{Path: reader.Str(outsidePath)}})

	// Then: it rejects the path before it can create a file outside data.
	if result.Err == nil {
		t.Fatal("expected an outside-data error")
	}

	if _, err := os.Stat(outsidePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("outside path stat error = %v, want not exist", err)
	}
}

//nolint:paralleltest // t.Chdir changes the process-wide working directory.
func TestServerHandlerRejectsTraversalPath(t *testing.T) {
	// Given: a server action rooted below data/server.
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(filepath.Join("data", "server", "site"), 0o750); err != nil {
		t.Fatalf("create server root: %v", err)
	}

	handler := NewServer()

	// When: a request path contains traversal syntax.
	response := executeAction(t, "/../secret", func(data *ProcessData) ExecutionResult {
		data.Payload = &config.ActionServerData{Directory: reader.Str("server/site")}

		return handler.ProcessAction(data)
	})

	// Then: the filesystem boundary is exposed as a forbidden response.
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.StatusCode)
	}

	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	})
}
