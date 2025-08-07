package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tail-on/pkg/apps"
	"github.com/sierrasoftworks/tail-on/pkg/config"
	"github.com/sierrasoftworks/tail-on/pkg/userctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationStateTracking(t *testing.T) {
	// Create test applications including a long-running one
	configs := []config.ApplicationConfig{
		{
			Name: "long-running-test",
			Path: "/bin/sleep",
			Args: []string{"30"}, // Use a simple sleep command
			Env:  []string{"TEST=1"},
		},
	}

	manager := apps.NewManager(configs)
	server := NewServer(manager)

	// Create a test user
	testUser := &userctx.User{
		ID:          "test-user-123",
		DisplayName: "Test User",
		LoginName:   "test@example.com",
		Node:        "test-node",
		IsAnonymous: false,
	}

	// Create context with user
	ctx := userctx.WithUser(context.Background(), testUser)

	// Start the long-running application with user context
	err := manager.StartApp(ctx, "long-running-test")
	require.NoError(t, err)

	// Wait for app to start (shorter wait since sleep command starts immediately)
	time.Sleep(100 * time.Millisecond)

	// Get the application and verify state change tracking
	app, err := manager.GetApp("long-running-test")
	require.NoError(t, err)

	// First check that the app is actually running
	assert.Equal(t, apps.StateRunning, app.State)
	assert.Greater(t, app.PID, 0)

	// Check that exit code is properly initialized for a running app
	assert.Equal(t, 0, app.LastExitCode, "LastExitCode should be 0 for a running application")

	// Then check state change tracking
	assert.NotNil(t, app.StateChangedBy)
	assert.Equal(t, testUser.ID, app.StateChangedBy.ID)
	assert.Equal(t, testUser.DisplayName, app.StateChangedBy.DisplayName)
	assert.Equal(t, testUser.LoginName, app.StateChangedBy.LoginName)

	assert.NotNil(t, app.StateChangedAt)
	assert.WithinDuration(t, time.Now(), *app.StateChangedAt, time.Minute)

	// Test that the API returns this information
	req := httptest.NewRequest("GET", "/api/v1/apps/long-running-test", nil)
	req = req.WithContext(ctx) // Add user context to request FIRST
	req = mux.SetURLVars(req, map[string]string{"app_name": "long-running-test"})

	recorder := httptest.NewRecorder()

	server.HandleGetApp(recorder, req)

	assert.Equal(t, 200, recorder.Code)

	var response map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify the API response includes state change tracking info
	assert.Contains(t, response, "state_changed_by")
	assert.Contains(t, response, "state_changed_at")
	assert.Contains(t, response, "last_exit_code")

	stateChangedBy, ok := response["state_changed_by"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, testUser.ID, stateChangedBy["id"])
	assert.Equal(t, testUser.DisplayName, stateChangedBy["display_name"])

	// Verify exit code is 0 for running app
	assert.Equal(t, float64(0), response["last_exit_code"])

	// Stop the application
	err = manager.StopApp(ctx, "long-running-test")
	require.NoError(t, err)

	// Wait for app to stop
	time.Sleep(500 * time.Millisecond)

	// Verify state change tracking persists and shows who stopped the app
	app, err = manager.GetApp("long-running-test")
	require.NoError(t, err)
	assert.Equal(t, apps.StateNotRunning, app.State)
	assert.NotNil(t, app.StateChangedBy)
	assert.Equal(t, testUser.ID, app.StateChangedBy.ID)
	assert.NotNil(t, app.StateChangedAt)
	// Exit code will be signal-related (not 0) since we terminated the process
	assert.NotEqual(t, 0, app.LastExitCode, "Process terminated by signal should not have exit code 0")
}
