package runtime

import (
	"errors"
	"testing"

	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"go.uber.org/zap"
)

type loaderStub struct {
	err error
}

const nilConfigError = "nil config"

func (s loaderStub) GetTag() string {
	return "stub"
}

func (s loaderStub) Load(_ *config.Config, _ string) error {
	return s.err
}

func (s loaderStub) Save(_ *config.Config, _ string) error {
	return nil
}

func (s loaderStub) GetAllowFileExtensions() []string {
	return []string{"test"}
}

func validLoadResult() *config.LoadResult {
	return &config.LoadResult{
		Config: &config.Config{
			Log:    config.Log{FileFormat: reader.Str("test.log")},
			Server: config.Server{Address: reader.Str(":3000")},
			Actions: []config.Action{{
				Route:        reader.Str("/"),
				Type:         config.ActionString,
				ActionString: &config.ActionStringData{Content: reader.Str("ok")},
			}},
		},
		Path: "config.test",
	}
}

func TestNewManagerValidatesRequiredInputs(t *testing.T) {
	t.Parallel()

	loaders := []config.Loader{loaderStub{}}
	logger := zap.NewNop()

	tests := []struct {
		name    string
		initial *config.LoadResult
		logger  *zap.Logger
		loaders []config.Loader
		want    string
	}{
		{name: "nil result", logger: logger, loaders: loaders, want: nilConfigError},
		{name: "nil config", initial: &config.LoadResult{}, logger: logger, loaders: loaders, want: nilConfigError},
		{name: "nil logger", initial: validLoadResult(), loaders: loaders, want: "nil logger"},
		{name: "no loaders", initial: validLoadResult(), logger: logger, want: "no config loaders"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			manager, err := NewManager(test.initial, test.logger, test.loaders, execute.Builders{})
			if manager != nil {
				t.Fatal("NewManager() manager = non-nil, want nil")
			}

			if err == nil || err.Error() != test.want {
				t.Fatalf("NewManager() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestNewManagerExposesInitialSnapshot(t *testing.T) {
	t.Parallel()

	initial := validLoadResult()

	manager, err := NewManager(initial, zap.NewNop(), []config.Loader{loaderStub{}}, execute.Builders{})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	snapshot := manager.Current()
	if snapshot == nil {
		t.Fatal("Current() = nil, want initial snapshot")
	}

	if snapshot.Config != initial.Config {
		t.Fatalf("Current().Config = %p, want %p", snapshot.Config, initial.Config)
	}

	if snapshot.Executor == nil {
		t.Fatal("Current().Executor = nil, want built executor")
	}

	if manager.CurrentConfig() != initial.Config {
		t.Fatal("CurrentConfig() did not return the initial config")
	}
}

func TestNilManagerHasNilStateAndReloadError(t *testing.T) {
	t.Parallel()

	var manager *Manager

	if manager.Current() != nil {
		t.Fatal("Current() = non-nil for nil manager")
	}

	if manager.CurrentConfig() != nil {
		t.Fatal("CurrentConfig() = non-nil for nil manager")
	}

	if err := manager.Reload(); err == nil || err.Error() != "nil runtime manager" {
		t.Fatalf("Reload() error = %v, want nil runtime manager", err)
	}
}

func TestReloadFailurePreservesCurrentSnapshot(t *testing.T) {
	t.Parallel()

	loadErr := errors.New("load failed")
	initial := validLoadResult()

	manager, err := NewManager(initial, zap.NewNop(), []config.Loader{loaderStub{err: loadErr}}, execute.Builders{})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	before := manager.Current()

	err = manager.Reload()
	if !errors.Is(err, loadErr) {
		t.Fatalf("Reload() error = %v, want wrapped %v", err, loadErr)
	}

	if after := manager.Current(); after != before {
		t.Fatalf("Current() changed after failed reload: got %p, want %p", after, before)
	}
}
