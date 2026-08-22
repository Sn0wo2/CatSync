package runtime

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/config"
	"go.uber.org/zap"
)

type Snapshot struct {
	Config   *config.Config
	Executor *execute.Executor
}

type Manager struct {
	CurrentSnapshot atomic.Pointer[Snapshot]
	ReloadMu        sync.Mutex
	Logger          *zap.Logger
	Loaders         []config.Loader
	Builders        execute.Builders
	ConfigPath      string
}

func NewManager(initial *config.LoadResult, logger *zap.Logger, loaders []config.Loader, builders execute.Builders) (*Manager, error) {
	if initial == nil || initial.Config == nil {
		return nil, errors.New("nil config")
	}

	if logger == nil {
		return nil, errors.New("nil logger")
	}

	if len(loaders) == 0 {
		return nil, errors.New("no config loaders")
	}

	executor, err := execute.New().WithConfig(initial.Config).WithBuilders(builders).Build()
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		Logger:     logger,
		Loaders:    loaders,
		Builders:   builders,
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

	cfg, _, err := config.LoadConfig(m.ConfigPath, m.Loaders...)
	if err != nil {
		return err
	}

	cfg.LogWarnings(m.Logger)

	executor, err := execute.New().WithConfig(cfg).WithBuilders(m.Builders).Build()
	if err != nil {
		return err
	}

	m.CurrentSnapshot.Store(&Snapshot{Config: cfg, Executor: executor})

	return nil
}
