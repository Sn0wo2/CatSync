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
	current    atomic.Pointer[Snapshot]
	reloadMu   sync.Mutex
	logger     *zap.Logger
	loaders    []config.Loader
	builders   execute.Builders
	configPath string
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
		logger:     logger,
		loaders:    loaders,
		builders:   builders,
		configPath: initial.Path,
	}
	manager.current.Store(&Snapshot{Config: initial.Config, Executor: executor})

	return manager, nil
}

func (m *Manager) Current() *Snapshot {
	if m == nil {
		return nil
	}

	return m.current.Load()
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

	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	result, err := config.LoadConfig(m.configPath, m.loaders...)
	if err != nil {
		return err
	}

	candidate := result.Config
	candidate.LogWarnings(m.logger)

	executor, err := execute.New().WithConfig(candidate).WithBuilders(m.builders).Build()
	if err != nil {
		return err
	}

	m.current.Store(&Snapshot{Config: candidate, Executor: executor})

	return nil
}
