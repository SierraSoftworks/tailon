package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sierrasoftworks/tailon/pkg/api"
	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerSetup tests the server setup without actually starting Tailscale
func TestServerSetup(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Applications: []config.ApplicationConfig{
			{
				Name: "test-app",
				Path: "/bin/echo",
				Args: []string{"hello"},
				Env:  []string{"TEST=1"},
			},
		},
		Listen: "localhost:8080",
		Tailscale: config.TailscaleConfig{
			Name:     "test-server",
			StateDir: "/tmp/test",
		},
	}

	// Create application manager
	appManager := apps.NewManager(cfg.Applications)
	assert.NotNil(t, appManager)

	// Create API server
	apiServer := api.NewServer(appManager)
	assert.NotNil(t, apiServer)

	router := apiServer.Routes()
	assert.NotNil(t, router)

	// Create HTTP server (without Tailscale)
	server := &http.Server{
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	assert.NotNil(t, server)
}

// TestAPIIntegration tests the full API without Tailscale
func TestAPIIntegration(t *testing.T) {
	// Setup test server
	cfg := &config.Config{
		Applications: []config.ApplicationConfig{
			{
				Name: "integration-test",
				Path: "/bin/sh",
				Args: []string{"-c", "echo 'integration test'; sleep 2"}, // Longer sleep for testing
				Env:  []string{"TEST=integration"},
			},
		},
		Listen: "localhost:8080",
		Tailscale: config.TailscaleConfig{
			Name:     "test-server",
			StateDir: "/tmp/test",
		},
	}

	appManager := apps.NewManager(cfg.Applications)
	apiServer := api.NewServer(appManager)
	testServer := httptest.NewServer(apiServer.Routes())
	defer testServer.Close()

	client := testServer.Client()

	// Test 1: List applications
	resp, err := client.Get(testServer.URL + "/api/v1/apps")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test 2: Get specific application
	resp, err = client.Get(testServer.URL + "/api/v1/apps/integration-test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test 3: Start application
	resp, err = client.Post(testServer.URL+"/api/v1/apps/integration-test/start", "application/json", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Wait a bit for the app to run and generate logs
	time.Sleep(500 * time.Millisecond)

	// Test 4: Get logs
	resp, err = client.Get(testServer.URL + "/api/v1/apps/integration-test/logs")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test 5: Stop application
	resp, err = client.Post(testServer.URL+"/api/v1/apps/integration-test/stop", "application/json", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test 6: Restart application
	resp, err = client.Post(testServer.URL+"/api/v1/apps/integration-test/restart", "application/json", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Wait a bit for restart
	time.Sleep(200 * time.Millisecond)

	// Test 7: Stop the restarted application
	resp, err = client.Post(testServer.URL+"/api/v1/apps/integration-test/stop", "application/json", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test 8: Try to access non-existent application
	resp, err = client.Get(testServer.URL + "/api/v1/apps/non-existent")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

// TestGracefulShutdown tests the shutdown logic
func TestGracefulShutdown(t *testing.T) {
	server := &http.Server{
		Addr:         "localhost:0", // Use any available port
		Handler:      http.NewServeMux(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}
