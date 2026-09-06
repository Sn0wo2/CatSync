package runtime

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/config"
)

type Snapshot struct {
	Config   *config.Config
	Executor *execute.Executor
}

type Manager struct {
	CurrentSnapshot atomic.Pointer[Snapshot]
	ReloadMu        sync.Mutex
	Logger          *slog.Logger
	ConfigPath      string
}

func NewManager(initial *config.LoadResult, logger *slog.Logger) (*Manager, error) {
	if initial == nil || initial.Config == nil {
		return nil, errors.New("nil config")
	}

	if logger == nil {
		return nil, errors.New("nil logger")
	}

	executor, err := execute.New().WithConfig(initial.Config).Build()
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		Logger:     logger,
		ConfigPath: initial.Path,
	}
	manager.CurrentSnapshot.Store(&Snapshot{Config: initial.Config, Executor: executor})

	return manager, nil
}

func (m *Manager) Current() *Snapshot {
	if m == nil {
		return nil
	}

	return m.CurrentSnapshot.Load()
}

func (m *Manager) CurrentConfig() *config.Config {
	snapshot := m.Current()
	if snapshot == nil {
		return nil
	}

	return snapshot.Config
}

func (m *Manager) Reload() error {
	if m == nil {
		return errors.New("nil runtime manager")
	}

	m.ReloadMu.Lock()
	defer m.ReloadMu.Unlock()

	cfg, _, err := config.LoadConfig(m.ConfigPath)
	if err != nil {
		return err
	}

	cfg.LogWarnings(m.Logger)

	executor, err := execute.New().WithConfig(cfg).Build()
	if err != nil {
		return err
	}

	m.CurrentSnapshot.Store(&Snapshot{Config: cfg, Executor: executor})

	return nil
}
