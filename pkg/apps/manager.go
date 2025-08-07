package apps

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
	"github.com/sirupsen/logrus"
)

const maxLogLines = 1000

// ApplicationState represents the current state of an application
type ApplicationState string

const (
	StateNotRunning ApplicationState = "not_running"
	StateRunning    ApplicationState = "running"
	StateStopping   ApplicationState = "stopping"
)

type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

type Application struct {
	Config         config.ApplicationConfig `json:"config"`
	State          ApplicationState         `json:"state"`
	PID            int                      `json:"pid,omitempty"`
	LastExitCode   int                      `json:"last_exit_code"`
	StateChangedBy *userctx.User            `json:"state_changed_by,omitempty"`
	StateChangedAt *time.Time               `json:"state_changed_at,omitempty"`
	logs           []LogLine
	logMux         sync.RWMutex
	cmd            *exec.Cmd
	cancel         context.CancelFunc
}

// IsRunning returns true if the application is currently running
func (a *Application) IsRunning() bool {
	return a.State == StateRunning
}

type Manager struct {
	apps map[string]*Application
	mux  sync.RWMutex
}

func NewManager(configs []config.ApplicationConfig) *Manager {
	apps := make(map[string]*Application)
	for _, cfg := range configs {
		apps[cfg.Name] = &Application{
			Config:       cfg,
			State:        StateNotRunning,
			LastExitCode: 0,
			logs:         make([]LogLine, 0, maxLogLines),
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
			Config:         app.Config,
			State:          app.State,
			PID:            app.PID,
			StateChangedBy: app.StateChangedBy,
			StateChangedAt: app.StateChangedAt,
			LastExitCode:   app.LastExitCode,
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
		Config:         app.Config,
		State:          app.State,
		PID:            app.PID,
		StateChangedBy: app.StateChangedBy,
		StateChangedAt: app.StateChangedAt,
		LastExitCode:   app.LastExitCode,
	}, nil
}

func (m *Manager) StartApp(ctx context.Context, name string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	app, exists := m.apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	if app.IsRunning() {
		return fmt.Errorf("application %s is already running", name)
	}

	// Get user from context
	user := userctx.FromContext(ctx)
	logger := userctx.GetLoggerFromContext(ctx)

	// Log the start event
	event := userctx.NewUserEvent(user, "start", name, "")
	logger.WithFields(logrus.Fields{
		"action": event.Action,
		"target": event.Target,
		"event":  event,
	}).Info("User started application")

	cmdCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(cmdCtx, app.Config.Path, app.Config.Args...)
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

	// Update application state and user tracking
	now := time.Now()
	app.cmd = cmd
	app.cancel = cancel
	app.State = StateRunning
	app.PID = cmd.Process.Pid
	app.StateChangedBy = user
	app.StateChangedAt = &now
	app.LastExitCode = 0 // Reset exit code when starting

	// Add audit log entry
	m.addAuditLog(app, user, "Started application")

	// Start log collection
	go m.collectLogs(name, stdout, "stdout")
	go m.collectLogs(name, stderr, "stderr")

	// Monitor process
	go func() {
		// Capture user for the goroutine closure
		currentUser := user

		var exitCode int
		if err := cmd.Wait(); err != nil {
			logger.WithField("app", name).WithError(err).Warn("Application exited with error")
			// Extract exit code from error if possible
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1 // Default to 1 for other errors
			}
		} else {
			exitCode = 0 // Successful exit
		}

		m.mux.Lock()
		now := time.Now()
		app.State = StateNotRunning
		app.PID = 0
		app.cmd = nil
		app.cancel = nil
		app.StateChangedBy = currentUser
		app.StateChangedAt = &now
		app.LastExitCode = exitCode
		m.mux.Unlock()

		// Add audit log for process exit
		auditMsg := fmt.Sprintf("Application process exited with code %d", exitCode)
		m.addAuditLog(app, currentUser, auditMsg)

		logger.WithField("app", name).WithField("exit_code", exitCode).Info("Application stopped")
	}()

	logger.WithField("app", name).WithField("pid", app.PID).Info("Application started")
	return nil
}

func (m *Manager) StopApp(ctx context.Context, name string) error {
	return m.stopApp(ctx, name, false)
}

func (m *Manager) ForceStopApp(ctx context.Context, name string) error {
	return m.stopApp(ctx, name, true)
}

func (m *Manager) stopApp(ctx context.Context, name string, force bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	app, exists := m.apps[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	if !app.IsRunning() {
		return fmt.Errorf("application %s is not running", name)
	}

	// Get user from context
	user := userctx.FromContext(ctx)
	logger := userctx.GetLoggerFromContext(ctx)

	// Log the stop event
	action := "stop"
	details := ""
	if force {
		action = "force_stop"
		details = "Using SIGKILL"
	} else {
		signal := app.Config.StopSignal
		if signal == "" {
			signal = "SIGINT"
		}
		details = fmt.Sprintf("Using %s", signal)
	}

	event := userctx.NewUserEvent(user, action, name, details)
	logger.WithFields(logrus.Fields{
		"action":  event.Action,
		"target":  event.Target,
		"details": event.Details,
		"event":   event,
	}).Info("User stopped application")

	// Set state to stopping and record who initiated the stop
	now := time.Now()
	app.State = StateStopping
	app.StateChangedBy = user
	app.StateChangedAt = &now

	// Add audit log entry
	auditMsg := fmt.Sprintf("Stopped application (%s)", details)
	m.addAuditLog(app, user, auditMsg)

	if force {
		// Force stop with SIGKILL
		if app.cmd != nil && app.cmd.Process != nil {
			if err := app.cmd.Process.Kill(); err != nil {
				logger.WithField("app", name).WithError(err).Warn("Failed to force kill application")
			}
		}
	} else {
		// Graceful stop with configured signal or SIGINT
		if app.cmd != nil && app.cmd.Process != nil {
			sig := m.parseStopSignal(app.Config.StopSignal)
			if err := app.cmd.Process.Signal(sig); err != nil {
				logger.WithField("app", name).WithError(err).Warn("Failed to signal application")
			}
		}
	}

	// Also cancel the context as a backup
	if app.cancel != nil {
		app.cancel()
	}

	return nil
}

// parseStopSignal converts a string signal name to syscall.Signal
func (m *Manager) parseStopSignal(signalName string) os.Signal {
	if signalName == "" {
		return syscall.SIGINT // default
	}

	switch signalName {
	case "SIGINT":
		return syscall.SIGINT
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGQUIT":
		return syscall.SIGQUIT
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGHUP":
		return syscall.SIGHUP
	default:
		logrus.WithField("signal", signalName).Warn("Unknown signal, defaulting to SIGINT")
		return syscall.SIGINT
	}
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
			Message:   line,
			Source:    source,
		}

		// Safely add log line by looking up the app each time
		m.mux.RLock()
		app := m.apps[appName]
		m.mux.RUnlock()

		if app != nil {
			m.addLogLine(app, logLine)
		}
	}
}

// addLogLine adds a log line to the application's log buffer
func (m *Manager) addLogLine(app *Application, logLine LogLine) {
	if app == nil {
		return
	}

	app.logMux.Lock()
	app.logs = append(app.logs, logLine)
	if len(app.logs) > maxLogLines {
		// Remove oldest logs to maintain circular buffer
		copy(app.logs, app.logs[1:])
		app.logs = app.logs[:maxLogLines]
	}
	app.logMux.Unlock()
}

// addAuditLog adds an audit log entry to the application's log buffer
func (m *Manager) addAuditLog(app *Application, user *userctx.User, message string) {
	userName := "Anonymous"
	if user != nil && !user.IsAnonymous {
		userName = user.DisplayName
	}

	auditMessage := fmt.Sprintf("%s: %s", userName, message)
	logLine := LogLine{
		Timestamp: time.Now(),
		Message:   auditMessage,
		Source:    "audit",
	}

	m.addLogLine(app, logLine)
}
