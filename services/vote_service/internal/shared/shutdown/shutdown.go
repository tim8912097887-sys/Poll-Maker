package shutdown

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Manager struct {
	logger *slog.Logger
    mu       sync.Mutex
	handlers []func(context.Context) error
}

func NewManager(logger *slog.Logger) *Manager {
	return &Manager{logger: logger}
}

func (m *Manager) Register(handler func(context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

func (m *Manager) Shutdown(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

	handlers := make([]func(context.Context) error, len(m.handlers))

	copy(handlers, m.handlers)
	for i := len(handlers) - 1; i >= 0; i-- {
		if err := handlers[i](ctx); err != nil {
			m.logger.Error("failed to shutdown",slog.Any("error", err))
		}
	}

	return nil
}