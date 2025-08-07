package apps

import (
	"testing"
	"time"

	"github.com/sierrasoftworks/tail-on/pkg/config"
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
	assert.False(t, apps["test-app"].Running)
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
	err := manager.StartApp("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test starting app
	err = manager.StartApp("test-echo")
	assert.NoError(t, err)

	// Verify app is running
	app, err := manager.GetApp("test-echo")
	assert.NoError(t, err)
	assert.True(t, app.Running)
	assert.Greater(t, app.PID, 0)

	// Test starting already running app
	err = manager.StartApp("test-echo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stopping app
	err = manager.StopApp("test-echo")
	assert.NoError(t, err)

	// Wait a bit for the process to stop
	time.Sleep(200 * time.Millisecond)

	// Verify app is stopped
	app, err = manager.GetApp("test-echo")
	assert.NoError(t, err)
	assert.False(t, app.Running)
	assert.Equal(t, 0, app.PID)

	// Test stopping non-existent app
	err = manager.StopApp("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test stopping already stopped app
	err = manager.StopApp("test-echo")
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
	err = manager.StartApp("test-logger")
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
		assert.True(t, log.Message[0] == '[') // Should start with [stdout] or [stderr]
	}

	// Stop the app
	manager.StopApp("test-logger")
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
	err := manager.StartApp("log-generator")
	require.NoError(t, err)

	// Wait for all logs to be generated
	time.Sleep(500 * time.Millisecond)

	logs, err := manager.GetLogs("log-generator")
	assert.NoError(t, err)

	// Should not exceed maximum log lines
	assert.LessOrEqual(t, len(logs), maxLogLines)

	// Stop the app
	manager.StopApp("log-generator")
}
