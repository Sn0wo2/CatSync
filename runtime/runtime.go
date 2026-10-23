package runtime

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/Sn0wo2/CatSync/action/execute"
	"github.com/Sn0wo2/CatSync/config"
)

type Manager struct {
	Executor   atomic.Pointer[execute.Executor]
	reloadMu   sync.Mutex
	Logger     *slog.Logger
	ConfigPath string
}

func (m *Manager) Reload() error {
	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	cfg, _, err := config.LoadConfig(m.ConfigPath)
	if err != nil {
		return err
	}

	cfg.LogWarnings(m.Logger)

	executor, err := execute.Build(cfg)
	if err != nil {
		return err
	}

	if previous := m.Executor.Swap(executor); previous != nil {
		previous.Stop()
	}

	return nil
}

func (m *Manager) Stop() {
	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	if executor := m.Executor.Swap(nil); executor != nil {
		executor.Stop()
	}
}
