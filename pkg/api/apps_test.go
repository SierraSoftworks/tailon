package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
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
}

func TestHandleGetAppEnvironmentFiltering(t *testing.T) {
	server, _ := SetupTestServer()

	tests := []struct {
		name           string
		user           *userctx.User
		shouldSeeEnv   bool
		description    string
	}{
		{
			name: "admin user sees env vars",
			user: &userctx.User{
				ID:          "admin-user",
				DisplayName: "Admin User",
				IsAnonymous: false,
				ApplicationRoles: map[string]string{
					"test-app": userctx.RoleAdmin,
				},
			},
			shouldSeeEnv: true,
			description:  "Admin users should see environment variables",
		},
		{
			name: "operator user does not see env vars",
			user: &userctx.User{
				ID:          "operator-user",
				DisplayName: "Operator User",
				IsAnonymous: false,
				ApplicationRoles: map[string]string{
					"test-app": userctx.RoleOperator,
				},
			},
			shouldSeeEnv: false,
			description:  "Operator users should not see environment variables",
		},
		{
			name: "viewer user does not see env vars",
			user: &userctx.User{
				ID:          "viewer-user",
				DisplayName: "Viewer User",
				IsAnonymous: false,
				ApplicationRoles: map[string]string{
					"test-app": userctx.RoleViewer,
				},
			},
			shouldSeeEnv: false,
			description:  "Viewer users should not see environment variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/apps/test-app", nil)
			req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
			
			// Add user context
			ctx := userctx.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)
			
			recorder := httptest.NewRecorder()
			server.HandleGetApp(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code, tt.description)

			var response ApplicationResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.shouldSeeEnv {
				assert.NotNil(t, response.Config.Env, "Admin user should see environment variables")
				assert.Greater(t, len(response.Config.Env), 0, "Environment variables should not be empty for admin")
			} else {
				assert.Nil(t, response.Config.Env, "Non-admin users should not see environment variables")
			}
		})
	}
}

func TestHandleGetAppNotFound(t *testing.T) {
	server, _ := SetupTestServer()

	// Test non-existent app
	req := httptest.NewRequest("GET", "/api/v1/apps/non-existent", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "non-existent"})
	recorder := httptest.NewRecorder()

	server.HandleGetApp(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
