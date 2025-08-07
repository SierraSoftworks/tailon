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

func TestApplicationUserTracking(t *testing.T) {
	// Create test applications including a long-running one
	configs := []config.ApplicationConfig{
		{
			Name: "long-running-test",
			Path: "/bin/sleep",
			Args: []string{"10"}, // Simple sleep command that definitely runs for 10 seconds
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

	// Get the application and verify user tracking
	app, err := manager.GetApp("long-running-test")
	require.NoError(t, err)

	// First check that the app is actually running
	assert.Equal(t, apps.StateRunning, app.State)
	assert.Greater(t, app.PID, 0)

	// Then check user tracking (this is where the test is failing)
	assert.NotNil(t, app.StartedBy)
	assert.Equal(t, testUser.ID, app.StartedBy.ID)
	assert.Equal(t, testUser.DisplayName, app.StartedBy.DisplayName)
	assert.Equal(t, testUser.LoginName, app.StartedBy.LoginName)

	assert.NotNil(t, app.StartedAt)
	assert.WithinDuration(t, time.Now(), *app.StartedAt, time.Minute)

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

	// Verify the API response includes user tracking info
	assert.Contains(t, response, "started_by")
	assert.Contains(t, response, "started_at")

	startedBy, ok := response["started_by"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, testUser.ID, startedBy["id"])
	assert.Equal(t, testUser.DisplayName, startedBy["display_name"])

	// Stop the application
	err = manager.StopApp(ctx, "long-running-test")
	require.NoError(t, err)

	// Wait for app to stop
	time.Sleep(200 * time.Millisecond)

	// Verify user tracking is cleared when app stops
	app, err = manager.GetApp("long-running-test")
	require.NoError(t, err)
	assert.Nil(t, app.StartedBy)
	assert.Nil(t, app.StartedAt)
}
