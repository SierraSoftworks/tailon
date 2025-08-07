package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetApps(t *testing.T) {
	server, _ := SetupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/apps", nil)
	recorder := httptest.NewRecorder()

	server.HandleGetApps(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "test-app")
	assert.Contains(t, response, "test-logger")
}

func TestHandleGetApp(t *testing.T) {
	server, _ := SetupTestServer()

	// Test existing app
	req := httptest.NewRequest("GET", "/api/v1/apps/test-app", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
	recorder := httptest.NewRecorder()

	server.HandleGetApp(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	config, ok := response["config"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test-app", config["name"])

	// Test non-existent app
	req = httptest.NewRequest("GET", "/api/v1/apps/non-existent", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder = httptest.NewRecorder()

	server.HandleGetApp(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
