package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	server, _ := SetupTestServer()

	// Create a simple handler to test middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Test logging middleware
	wrappedHandler := server.LoggingMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "test", recorder.Body.String())

	// Test CORS middleware
	corsHandler := server.CORSMiddleware(handler)
	req = httptest.NewRequest("GET", "/test", nil)
	recorder = httptest.NewRecorder()

	corsHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))

	// Test CORS preflight
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	recorder = httptest.NewRecorder()

	corsHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestResponseWrapper(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapper := &ResponseWrapper{ResponseWriter: recorder, StatusCode: http.StatusOK}

	// Test WriteHeader
	wrapper.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, wrapper.StatusCode)

	// Test Write
	data := []byte("test data")
	n, err := wrapper.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, "test data", recorder.Body.String())
}
