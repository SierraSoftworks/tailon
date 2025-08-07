package api

import (
	"net/http"
	"time"

	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware logs HTTP requests
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		wrapper := &ResponseWrapper{ResponseWriter: w, StatusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		logrus.WithFields(logrus.Fields{
			"method":     r.Method,
			"url":        r.URL.Path,
			"status":     wrapper.StatusCode,
			"duration":   time.Since(start),
			"remote":     r.RemoteAddr,
			"user_agent": r.UserAgent(),
		}).Info("HTTP request")
	})
}

// CORSMiddleware adds CORS headers
func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ResponseWrapper wraps http.ResponseWriter to capture status codes
type ResponseWrapper struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader captures the status code
func (rw *ResponseWrapper) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher interface
func (rw *ResponseWrapper) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Write passes through to the wrapped ResponseWriter
func (rw *ResponseWrapper) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// SetupTestServer creates a test server with predefined applications
func SetupTestServer() (*Server, *apps.Manager) {
	configs := []config.ApplicationConfig{
		{
			Name: "test-app",
			Path: "/bin/echo",
			Args: []string{"hello"},
			Env:  []string{"TEST=1"},
		},
		{
			Name: "test-logger",
			Path: "/bin/sh",
			Args: []string{"-c", "echo 'test log'; sleep 0.1"},
		},
	}

	manager := apps.NewManager(configs)
	server := NewServer(manager)
	return server, manager
}
