package api

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleSwaggerUI(t *testing.T) {
	server, _ := SetupTestServer()

	req := httptest.NewRequest("GET", "/docs/swagger/", nil)
	recorder := httptest.NewRecorder()

	server.HandleSwaggerUI(recorder, req)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/html", recorder.Header().Get("Content-Type"))

	body := recorder.Body.String()
	assert.Contains(t, body, "SwaggerUIBundle")
	assert.Contains(t, body, "swagger-ui-dist")
	assert.Contains(t, body, "tail-on API Documentation")
}

func TestHandleOpenAPISpecYAML(t *testing.T) {
	server, _ := SetupTestServer()

	req := httptest.NewRequest("GET", "/docs/openapi.yaml", nil)
	recorder := httptest.NewRecorder()

	server.HandleOpenAPISpec(recorder, req)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/yaml", recorder.Header().Get("Content-Type"))

	body := recorder.Body.String()
	assert.Contains(t, body, "openapi:")
	assert.Contains(t, body, "tail-on")
}

func TestHandleAPIDocsWithNavigation(t *testing.T) {
	server, _ := SetupTestServer()

	req := httptest.NewRequest("GET", "/docs/", nil)
	recorder := httptest.NewRecorder()

	server.HandleAPIDocs(recorder, req)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/html", recorder.Header().Get("Content-Type"))

	body := recorder.Body.String()
	assert.Contains(t, body, "tail-on API")
	assert.Contains(t, body, "/docs/swagger/")
	assert.Contains(t, body, "API Explorer")
	assert.Contains(t, body, "nav-link")
}

func TestBaseURLGeneration(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		proto    string
		expected string
	}{
		{
			name:     "HTTP localhost",
			host:     "localhost:8080",
			proto:    "",
			expected: "http://localhost:8080",
		},
		{
			name:     "HTTPS with TLS",
			host:     "api.example.com",
			proto:    "https",
			expected: "https://api.example.com",
		},
		{
			name:     "HTTP with custom port",
			host:     "192.168.1.100:3000",
			proto:    "",
			expected: "http://192.168.1.100:3000",
		},
	}

	server, _ := SetupTestServer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/docs/", nil)
			req.Host = tt.host
			if tt.proto == "https" {
				req.Header.Set("X-Forwarded-Proto", "https")
			}

			recorder := httptest.NewRecorder()
			server.HandleAPIDocs(recorder, req)

			assert.Equal(t, 200, recorder.Code)
			body := recorder.Body.String()
			assert.Contains(t, body, tt.expected)
		})
	}
}
