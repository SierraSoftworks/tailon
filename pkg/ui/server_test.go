package ui

import (
	"net/http/httptest"
	"testing"

	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/stretchr/testify/assert"
)

// setupTestServer creates a test UI server
func setupTestServer() *Server {
	manager := apps.NewManager([]config.ApplicationConfig{})
	return NewServer(manager)
}

func TestHandleOpenAPISpecYAML(t *testing.T) {
	server := setupTestServer()

	req := httptest.NewRequest("GET", "/docs/openapi.yaml", nil)
	recorder := httptest.NewRecorder()

	server.Routes().ServeHTTP(recorder, req)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/yaml", recorder.Header().Get("Content-Type"))

	body := recorder.Body.String()
	assert.Contains(t, body, "openapi:")
	assert.Contains(t, body, "tailon")
}
