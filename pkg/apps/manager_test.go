package apps

import (
	"context"
	"testing"
	"time"

	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	configs := []config.ApplicationConfig{
		{Name: "app1", Path: "/bin/echo", Args: []string{"hello"}},
		{Name: "app2", Path: "/bin/echo", Args: []string{"world"}},
	}

	manager := NewManager(configs)

	assert.NotNil(t, manager)
	assert.Len(t, manager.apps, 2)
	assert.Contains(t, manager.apps, "app1")
	assert.Contains(t, manager.apps, "app2")
}

func TestManagerGetApps(t *testing.T) {
	configs := []config.ApplicationConfig{
		{Name: "test-app", Path: "/bin/echo", Args: []string{"test"}},
	}

	manager := NewManager(configs)
	apps := manager.GetApps()

	assert.Len(t, apps, 1)
	assert.Contains(t, apps, "test-app")
	assert.Equal(t, "test-app", apps["test-app"].Config.Name)
	assert.Equal(t, StateNotRunning, apps["test-app"].State)
	assert.False(t, apps["test-app"].IsRunning())
}

func TestManagerGetApp(t *testing.T) {
	configs := []config.ApplicationConfig{
		{Name: "test-app", Path: "/bin/echo", Args: []string{"test"}},
	}

	manager := NewManager(configs)

	// Test existing app
	app, err := manager.GetApp("test-app")
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "test-app", app.Config.Name)

	// Test non-existent app
	app, err = manager.GetApp("non-existent")
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "not found")
}

func TestManagerStartStopApp(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "test-echo",
			Path: "/bin/sh",
			Args: []string{"-c", "echo 'test output'; sleep 0.1"},
		},
	}

	manager := NewManager(configs)

	// Test starting non-existent app
	err := manager.StartApp(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test starting app
	err = manager.StartApp(context.Background(), "test-echo")
	assert.NoError(t, err)

	// Verify app is running
	app, err := manager.GetApp("test-echo")
	assert.NoError(t, err)
	assert.Equal(t, StateRunning, app.State)
	assert.True(t, app.IsRunning())
	assert.Greater(t, app.PID, 0)

	// Test starting already running app
	err = manager.StartApp(context.Background(), "test-echo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stopping app
	err = manager.StopApp(context.Background(), "test-echo")
	assert.NoError(t, err)

	// Wait a bit for the process to stop
	time.Sleep(200 * time.Millisecond)

	// Verify app is stopped
	app, err = manager.GetApp("test-echo")
	assert.NoError(t, err)
	assert.Equal(t, StateNotRunning, app.State)
	assert.False(t, app.IsRunning())
	assert.Equal(t, 0, app.PID)

	// Test stopping non-existent app
	err = manager.StopApp(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test stopping already stopped app
	err = manager.StopApp(context.Background(), "test-echo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestManagerGetLogs(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "test-logger",
			Path: "/bin/sh",
			Args: []string{"-c", "echo 'log line 1'; echo 'log line 2' >&2; sleep 0.1"},
		},
	}

	manager := NewManager(configs)

	// Test getting logs from non-existent app
	logs, err := manager.GetLogs("non-existent")
	assert.Error(t, err)
	assert.Nil(t, logs)

	// Start app and wait for logs
	err = manager.StartApp(context.Background(), "test-logger")
	require.NoError(t, err)

	// Wait for logs to be collected
	time.Sleep(300 * time.Millisecond)

	// Get logs
	logs, err = manager.GetLogs("test-logger")
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)

	// Verify log structure
	for _, log := range logs {
		assert.NotZero(t, log.Timestamp)
		assert.NotEmpty(t, log.Message)
		assert.NotEmpty(t, log.Source)                                        // Should have a source field
		assert.Contains(t, []string{"stdout", "stderr", "audit"}, log.Source) // Should be a valid source
	}

	// Stop the app
	manager.StopApp(context.Background(), "test-logger")
}

func TestLogCircularBuffer(t *testing.T) {
	// Create an app that generates many log lines
	configs := []config.ApplicationConfig{
		{
			Name: "log-generator",
			Path: "/bin/sh",
			Args: []string{"-c", "for i in $(seq 1 1100); do echo \"Log line $i\"; done"},
		},
	}

	manager := NewManager(configs)
	err := manager.StartApp(context.Background(), "log-generator")
	require.NoError(t, err)

	// Wait for all logs to be generated
	time.Sleep(500 * time.Millisecond)

	logs, err := manager.GetLogs("log-generator")
	assert.NoError(t, err)

	// Should not exceed maximum log lines
	assert.LessOrEqual(t, len(logs), maxLogLines)

	// Stop the app
	manager.StopApp(context.Background(), "log-generator")
}

func TestManagerForceStopApp(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "long-running",
			Path: "/bin/sh",
			Args: []string{"-c", "trap 'echo trapped; sleep 1' TERM; sleep 10"},
		},
	}

	manager := NewManager(configs)

	// Test force stopping non-existent app
	err := manager.ForceStopApp(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Start app
	err = manager.StartApp(context.Background(), "long-running")
	assert.NoError(t, err)

	// Verify app is running
	app, err := manager.GetApp("long-running")
	assert.NoError(t, err)
	assert.Equal(t, StateRunning, app.State)

	// Force stop the app
	err = manager.ForceStopApp(context.Background(), "long-running")
	assert.NoError(t, err)

	// Wait for process to be killed
	time.Sleep(200 * time.Millisecond)

	// Verify app is stopped
	app, err = manager.GetApp("long-running")
	assert.NoError(t, err)
	assert.Equal(t, StateNotRunning, app.State)
	assert.Equal(t, 0, app.PID)

	// Test force stopping already stopped app
	err = manager.ForceStopApp(context.Background(), "long-running")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestManagerStopSignalConfiguration(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name:       "sigterm-app",
			Path:       "/bin/sh",
			Args:       []string{"-c", "trap 'echo got SIGTERM; exit' TERM; sleep 10"},
			StopSignal: "SIGTERM",
		},
		{
			Name:       "sigint-app",
			Path:       "/bin/sh",
			Args:       []string{"-c", "trap 'echo got SIGINT; exit' INT; sleep 10"},
			StopSignal: "SIGINT",
		},
	}

	manager := NewManager(configs)

	// Start both apps
	err := manager.StartApp(context.Background(), "sigterm-app")
	assert.NoError(t, err)
	err = manager.StartApp(context.Background(), "sigint-app")
	assert.NoError(t, err)

	// Stop both apps (should use their configured signals)
	err = manager.StopApp(context.Background(), "sigterm-app")
	assert.NoError(t, err)
	err = manager.StopApp(context.Background(), "sigint-app")
	assert.NoError(t, err)

	// Wait for processes to handle signals
	time.Sleep(200 * time.Millisecond)

	// Verify both apps are stopped
	app1, err := manager.GetApp("sigterm-app")
	assert.NoError(t, err)
	assert.Equal(t, StateNotRunning, app1.State)

	app2, err := manager.GetApp("sigint-app")
	assert.NoError(t, err)
	assert.Equal(t, StateNotRunning, app2.State)
}

func TestParseStopSignal(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		input    string
		expected string
	}{
		{"", "interrupt"}, // default SIGINT
		{"SIGINT", "interrupt"},
		{"SIGTERM", "terminated"},
		{"SIGQUIT", "quit"},
		{"SIGKILL", "killed"},
		{"SIGHUP", "hangup"},
		{"INVALID", "interrupt"}, // fallback to SIGINT
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			signal := manager.parseStopSignal(test.input)
			assert.Contains(t, signal.String(), test.expected)
		})
	}
}

func TestApplicationStates(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "state-test",
			Path: "/bin/sh",
			Args: []string{"-c", "sleep 0.5"},
		},
	}

	manager := NewManager(configs)

	// Initial state should be not running
	app, err := manager.GetApp("state-test")
	assert.NoError(t, err)
	assert.Equal(t, StateNotRunning, app.State)
	assert.False(t, app.IsRunning())

	// Start app - should be running
	err = manager.StartApp(context.Background(), "state-test")
	assert.NoError(t, err)

	app, err = manager.GetApp("state-test")
	assert.NoError(t, err)
	assert.Equal(t, StateRunning, app.State)
	assert.True(t, app.IsRunning())

	// Stop app - should transition to stopping then not running
	err = manager.StopApp(context.Background(), "state-test")
	assert.NoError(t, err)

	// Wait for process to exit
	time.Sleep(200 * time.Millisecond)

	app, err = manager.GetApp("state-test")
	assert.NoError(t, err)
	assert.Equal(t, StateNotRunning, app.State)
	assert.False(t, app.IsRunning())
}

func TestAuditLogging(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "audit-test",
			Path: "/bin/echo",
			Args: []string{"Hello from audit test"},
		},
	}

	manager := NewManager(configs)

	// Create a test user context
	user := &userctx.User{
		ID:          "test-user-123",
		DisplayName: "Test User",
		LoginName:   "test@example.com",
		IsAnonymous: false,
	}
	ctx := userctx.WithUser(context.Background(), user)

	// Start the application
	err := manager.StartApp(ctx, "audit-test")
	require.NoError(t, err)

	// Wait for process to complete and logs to be collected
	time.Sleep(300 * time.Millisecond)

	// Get logs
	logs, err := manager.GetLogs("audit-test")
	require.NoError(t, err)

	// Verify we have logs
	assert.Greater(t, len(logs), 0, "Should have some logs")

	// Find the audit logs
	var auditLogs []LogLine
	var stdoutLogs []LogLine

	for _, log := range logs {
		switch log.Source {
		case "audit":
			auditLogs = append(auditLogs, log)
		case "stdout":
			stdoutLogs = append(stdoutLogs, log)
		}
	}

	// Should have at least one audit log for starting
	assert.Greater(t, len(auditLogs), 0, "Should have audit logs")

	// Check the start audit log
	found := false
	for _, log := range auditLogs {
		if log.Source == "audit" && log.Message == "Test User: Started application" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have audit log for application start")

	// Should have stdout logs too
	assert.Greater(t, len(stdoutLogs), 0, "Should have stdout logs")

	// Check if the application is still running before trying to stop it
	app, err := manager.GetApp("audit-test")
	require.NoError(t, err)

	if app.State == StateRunning {
		// Stop the application
		err = manager.StopApp(ctx, "audit-test")
		require.NoError(t, err)

		// Wait for stop to complete
		time.Sleep(200 * time.Millisecond)

		// Get logs again to check for stop audit log
		logs, err = manager.GetLogs("audit-test")
		require.NoError(t, err)

		// Find stop audit logs
		auditLogs = []LogLine{}
		for _, log := range logs {
			if log.Source == "audit" {
				auditLogs = append(auditLogs, log)
			}
		}

		// Should have audit logs for both start and stop
		assert.GreaterOrEqual(t, len(auditLogs), 2, "Should have audit logs for start and stop")

		// Check for stop audit log
		foundStop := false
		for _, log := range auditLogs {
			if log.Source == "audit" && log.Message == "Test User: Stopped application (Using SIGINT)" {
				foundStop = true
				break
			}
		}
		assert.True(t, foundStop, "Should have audit log for application stop")
	}
}

func TestAuditLoggingAnonymousUser(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "anonymous-test",
			Path: "/bin/echo",
			Args: []string{"Hello from anonymous test"},
		},
	}

	manager := NewManager(configs)

	// Use context without user (should default to anonymous)
	ctx := context.Background()

	// Start the application
	err := manager.StartApp(ctx, "anonymous-test")
	require.NoError(t, err)

	// Wait for process to complete
	time.Sleep(300 * time.Millisecond)

	// Get logs
	logs, err := manager.GetLogs("anonymous-test")
	require.NoError(t, err)

	// Find audit logs
	var auditLogs []LogLine
	for _, log := range logs {
		if log.Source == "audit" {
			auditLogs = append(auditLogs, log)
		}
	}

	// Should have audit log with anonymous user
	assert.Greater(t, len(auditLogs), 0, "Should have audit logs")

	found := false
	for _, log := range auditLogs {
		if log.Source == "audit" && log.Message == "Anonymous: Started application" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have audit log with anonymous user")
}

func TestLogSourceTypes(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "source-test",
			Path: "/bin/sh",
			Args: []string{"-c", "echo 'stdout message'; echo 'stderr message' >&2"},
		},
	}

	manager := NewManager(configs)

	// Create a test user context
	user := &userctx.User{
		ID:          "test-user",
		DisplayName: "Source Test User",
		IsAnonymous: false,
	}
	ctx := userctx.WithUser(context.Background(), user)

	// Start the application
	err := manager.StartApp(ctx, "source-test")
	require.NoError(t, err)

	// Wait for process to complete and logs to be collected
	time.Sleep(300 * time.Millisecond)

	// Get logs
	logs, err := manager.GetLogs("source-test")
	require.NoError(t, err)

	// Categorize logs by source
	sourceTypes := make(map[string]int)
	for _, log := range logs {
		sourceTypes[log.Source]++
	}

	// Should have all three source types
	assert.Greater(t, sourceTypes["audit"], 0, "Should have audit logs")
	assert.Greater(t, sourceTypes["stdout"], 0, "Should have stdout logs")
	assert.Greater(t, sourceTypes["stderr"], 0, "Should have stderr logs")

	// Verify specific log contents
	var foundStdout, foundStderr, foundAudit bool
	for _, log := range logs {
		switch log.Source {
		case "stdout":
			if log.Message == "stdout message" {
				foundStdout = true
			}
		case "stderr":
			if log.Message == "stderr message" {
				foundStderr = true
			}
		case "audit":
			if log.Message == "Source Test User: Started application" {
				foundAudit = true
			}
		}
	}

	assert.True(t, foundStdout, "Should find stdout message")
	assert.True(t, foundStderr, "Should find stderr message")
	assert.True(t, foundAudit, "Should find audit message")
}
