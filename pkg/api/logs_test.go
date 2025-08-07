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

func TestHandleLogs(t *testing.T) {
	server, manager := SetupTestServer()

	// Start the logger app and wait for logs
	err := manager.StartApp("test-logger")
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	// Test getting logs as JSON
	req := httptest.NewRequest("GET", "/api/v1/apps/test-logger/logs", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-logger"})
	recorder := httptest.NewRecorder()

	server.HandleLogs(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var logs []map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &logs)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)

	// Test non-existent app
	req = httptest.NewRequest("GET", "/api/v1/apps/non-existent/logs", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder = httptest.NewRecorder()

	server.HandleLogs(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	// Stop the app
	manager.StopApp("test-logger")
}

func TestHandleLogsSSERouting(t *testing.T) {
	server, manager := SetupTestServer()

	// Start the logger app
	err := manager.StartApp("test-logger")
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	// Test that the regular logs endpoint works
	req := httptest.NewRequest("GET", "/api/v1/apps/test-logger/logs", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-logger"})
	recorder := httptest.NewRecorder()

	server.HandleLogs(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var logs []map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &logs)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)

	// Stop the app
	manager.StopApp("test-logger")
}
