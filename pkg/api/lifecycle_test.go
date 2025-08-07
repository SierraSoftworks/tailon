package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleStartApp(t *testing.T) {
	server, _ := SetupTestServer()

	// Test starting existing app
	req := httptest.NewRequest("POST", "/api/v1/apps/test-app/start", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder := httptest.NewRecorder()

	server.HandleStartApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "started", response["status"])

	// Test starting non-existent app
	req = httptest.NewRequest("POST", "/api/v1/apps/non-existent/start", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder = httptest.NewRecorder()

	server.HandleStartApp(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleStopApp(t *testing.T) {
	server, manager := SetupTestServer()

	// First start the app
	err := manager.StartApp("test-app")
	require.NoError(t, err)

	// Test stopping the app
	req := httptest.NewRequest("POST", "/api/v1/apps/test-app/stop", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder := httptest.NewRecorder()

	server.HandleStopApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "stopped", response["status"])

	// Test stopping non-existent app
	req = httptest.NewRequest("POST", "/api/v1/apps/non-existent/stop", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder = httptest.NewRecorder()

	server.HandleStopApp(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleRestartApp(t *testing.T) {
	server, manager := SetupTestServer()

	// Test restarting existing app (not currently running)
	req := httptest.NewRequest("POST", "/api/v1/apps/test-app/restart", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder := httptest.NewRecorder()

	server.HandleRestartApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "restarted", response["status"])

	// Verify app is running
	app, err := manager.GetApp("test-app")
	assert.NoError(t, err)
	assert.True(t, app.IsRunning())

	// Test restarting app that is already running
	req = httptest.NewRequest("POST", "/api/v1/apps/test-app/restart", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder = httptest.NewRecorder()

	server.HandleRestartApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "restarted", response["status"])

	// Test restarting non-existent app
	req = httptest.NewRequest("POST", "/api/v1/apps/non-existent/restart", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder = httptest.NewRecorder()

	server.HandleRestartApp(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	// Clean up - stop the app
	manager.StopApp("test-app")
}

func TestHandleStopAppWithForce(t *testing.T) {
	server, manager := SetupTestServer()

	// Test normal stop
	// First start the app
	err := manager.StartApp("test-app")
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/apps/test-app/stop", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder := httptest.NewRecorder()

	server.HandleStopApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	var response map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "stopped", response["status"])

	// Wait for app to stop completely
	time.Sleep(100 * time.Millisecond)

	// Start the app again for force stop test  
	err = manager.StartApp("test-app")
	require.NoError(t, err)

	// Test force stop
	req = httptest.NewRequest("POST", "/api/v1/apps/test-app/stop?force=true", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder = httptest.NewRecorder()

	server.HandleStopApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "force_stopped", response["status"])

	// Test force stop on non-existent app
	req = httptest.NewRequest("POST", "/api/v1/apps/non-existent/stop?force=true", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder = httptest.NewRecorder()

	server.HandleStopApp(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
