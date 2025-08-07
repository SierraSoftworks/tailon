package apps

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/sierrasoftworks/tail-on/pkg/config"
	"github.com/sirupsen/logrus"
)

const maxLogLines = 1000

type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type Application struct {
	Config  config.ApplicationConfig `json:"config"`
	Running bool                     `json:"running"`
	PID     int                      `json:"pid,omitempty"`
	logs    []LogLine
	logMux  sync.RWMutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
}

type Manager struct {
	apps map[string]*Application
	mux  sync.RWMutex
}

func NewManager(configs []config.ApplicationConfig) *Manager {
	apps := make(map[string]*Application)
	for _, cfg := range configs {
		apps[cfg.Name] = &Application{
			Config: cfg,
			logs:   make([]LogLine, 0, maxLogLines),
		}
	}

	return &Manager{
		apps: apps,
	}
}

func (m *Manager) GetApps() map[string]*Application {
	m.mux.RLock()
	defer m.mux.RUnlock()

	result := make(map[string]*Application)
	for name, app := range m.apps {
		result[name] = &Application{
			Config:  app.Config,
			Running: app.Running,
			PID:     app.PID,
		}
	}
	return result
}

func (m *Manager) GetApp(name string) (*Application, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	app, exists := m.apps[name]
	if !exists {
		return nil, fmt.Errorf("application %s not found", name)
	}

	return &Application{
		Config:  app.Config,
		Running: app.Running,
		PID:     app.PID,
	}, nil
}

func (m *Manager) StartApp(name string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	app, exists := m.apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	if app.Running {
		return fmt.Errorf("application %s is already running", name)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, app.Config.Path, app.Config.Args...)
	cmd.Env = append(os.Environ(), app.Config.Env...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start application: %w", err)
	}

	app.cmd = cmd
	app.cancel = cancel
	app.Running = true
	app.PID = cmd.Process.Pid

	// Start log collection
	go m.collectLogs(name, stdout, "stdout")
	go m.collectLogs(name, stderr, "stderr")

	// Monitor process
	go func() {
		if err := cmd.Wait(); err != nil {
			logrus.WithField("app", name).WithError(err).Warn("Application exited with error")
		}

		m.mux.Lock()
		app.Running = false
		app.PID = 0
		app.cmd = nil
		app.cancel = nil
		m.mux.Unlock()

		logrus.WithField("app", name).Info("Application stopped")
	}()

	logrus.WithField("app", name).WithField("pid", app.PID).Info("Application started")
	return nil
}

func (m *Manager) StopApp(name string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	app, exists := m.apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	if !app.Running {
		return fmt.Errorf("application %s is not running", name)
	}

	if app.cancel != nil {
		app.cancel()
	}

	return nil
}

func (m *Manager) GetLogs(name string) ([]LogLine, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	app, exists := m.apps[name]
	if !exists {
		return nil, fmt.Errorf("application %s not found", name)
	}

	app.logMux.RLock()
	defer app.logMux.RUnlock()

	// Return a copy of the logs
	logs := make([]LogLine, len(app.logs))
	copy(logs, app.logs)
	return logs, nil
}

func (m *Manager) collectLogs(appName string, reader io.Reader, source string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		logLine := LogLine{
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("[%s] %s", source, line),
		}

		m.mux.RLock()
		app := m.apps[appName]
		m.mux.RUnlock()

		app.logMux.Lock()
		app.logs = append(app.logs, logLine)
		if len(app.logs) > maxLogLines {
			// Remove oldest logs to maintain circular buffer
			copy(app.logs, app.logs[1:])
			app.logs = app.logs[:maxLogLines]
		}
		app.logMux.Unlock()
	}
}
