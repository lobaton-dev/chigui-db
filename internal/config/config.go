// Package config manages application configuration: loading, saving, merging,validating, and persisting session state (YAML-based).
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lobaton-dev/chigui-db/internal/models"
	"gopkg.in/yaml.v3"
)

var defaultDir = filepath.Join(os.Getenv("HOME"), ".config", "chiguidb")

type Session struct {
	LastConnectionID string `json:"last_connection_id,omitempty"`
	LastScreen       string `json:"last_screen,omitempty"`
	WindowWidth      int    `json:"window_width,omitempty"`
	WindowHeight     int    `json:"window_height,omitempty"`
}

type Manager struct {
	path        string
	sessionPath string
	config      *models.Config
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func Load() (*Manager, error) {
	return LoadFrom(defaultDir)
}

func LoadFrom(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("creating config dir: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		path:        filepath.Join(dir, "config.yaml"),
		sessionPath: filepath.Join(dir, "session.yaml"),
		config:      models.DefaultConfig(),
		ctx:         ctx,
		cancel:      cancel,
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		cancel()
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, m.config); err != nil {
		cancel()
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return m, nil
}

func (m *Manager) GetContext() context.Context {
	return m.ctx
}

func (m *Manager) Close() {
	m.cancel()
}

func (m *Manager) GetConfig() *models.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

func (m *Manager) AddConnection(cfg *models.ConnectionConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Connections {
		if c.ID == cfg.ID {
			m.config.Connections[i] = cfg
			return
		}
	}
	m.config.Connections = append(m.config.Connections, cfg)
}

func (m *Manager) RemoveConnection(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.config.Connections {
		if c.ID == id {
			m.config.Connections = append(
				m.config.Connections[:i],
				m.config.Connections[i+1:]...,
			)
			return
		}
	}
}

func (m *Manager) GetConnection(id string) *models.ConnectionConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.config.Connections {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (m *Manager) ListConnections() []*models.ConnectionConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*models.ConnectionConfig, len(m.config.Connections))
	copy(result, m.config.Connections)
	return result
}

func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(m.path, data, 0o640); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

func (m *Manager) Validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.config.Connections {
		if c.Name == "" {
			return fmt.Errorf("connection name is required")
		}
		if c.Type == "" {
			return fmt.Errorf("connection type is required for %q", c.Name)
		}
	}
	return nil
}

func (m *Manager) Merge(cfg *models.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cfg.UI.Theme != "" {
		m.config.UI = cfg.UI
	}
	if cfg.Editor.Theme != "" {
		m.config.Editor = cfg.Editor
	}
	if cfg.Export.DefaultFormat != "" {
		m.config.Export = cfg.Export
	}
	if len(cfg.Connections) > 0 {
		m.config.Connections = append(m.config.Connections, cfg.Connections...)
	}
	return nil
}

func (m *Manager) LoadSession() (*Session, error) {
	data, err := os.ReadFile(m.sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Session{}, nil
		}
		return nil, fmt.Errorf("reading session: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return &Session{}, nil
	}
	return &s, nil
}

func (m *Manager) SaveSession(s *Session) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	if err := os.WriteFile(m.sessionPath, data, 0o640); err != nil {
		return fmt.Errorf("writing session: %w", err)
	}
	return nil
}
