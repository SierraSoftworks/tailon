package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sierrasoftworks/tail-on/pkg/userctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleWhoAmI(t *testing.T) {
	server, _ := SetupTestServer()

	// Test with anonymous user (default)
	req := httptest.NewRequest("GET", "/api/v1/whoami", nil)
	recorder := httptest.NewRecorder()

	server.HandleWhoAmI(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var user userctx.User
	err := json.Unmarshal(recorder.Body.Bytes(), &user)
	require.NoError(t, err)

	// Should be anonymous user
	assert.True(t, user.IsAnonymous)
	assert.Equal(t, "anonymous", user.ID)
	assert.Equal(t, "Anonymous", user.DisplayName)
}

func TestHandleWhoAmIWithTailscaleUser(t *testing.T) {
	server, _ := SetupTestServer()

	// Create a test Tailscale user
	testUser := &userctx.User{
		ID:          "user123",
		DisplayName: "John Doe",
		LoginName:   "john@example.com",
		Node:        "johns-laptop",
		IsAnonymous: false,
	}

	// Create a request with user context
	req := httptest.NewRequest("GET", "/api/v1/whoami", nil)
	ctx := userctx.WithUser(context.Background(), testUser)
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()

	server.HandleWhoAmI(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var user userctx.User
	err := json.Unmarshal(recorder.Body.Bytes(), &user)
	require.NoError(t, err)

	// Should be the test user
	assert.False(t, user.IsAnonymous)
	assert.Equal(t, "user123", user.ID)
	assert.Equal(t, "John Doe", user.DisplayName)
	assert.Equal(t, "john@example.com", user.LoginName)
	assert.Equal(t, "johns-laptop", user.Node)
}
