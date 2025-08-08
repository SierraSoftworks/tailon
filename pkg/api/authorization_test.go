package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationRules(t *testing.T) {
	// Test user with different roles
	user := &userctx.User{
		ID:          "test-user",
		DisplayName: "Test User",
		IsAnonymous: false,
		ApplicationRoles: map[string]userctx.Role{
			"app1": userctx.RoleAdmin,
			"app2": userctx.RoleOperator,
			"app3": userctx.RoleViewer,
			// No role for "app4"
		},
	}

	vars := map[string]string{}

	tests := []struct {
		name     string
		rule     AuthorizationRule
		appName  string
		expected bool
	}{
		{"AppViewer allows admin", AppViewer(), "app1", true},
		{"AppViewer allows operator", AppViewer(), "app2", true},
		{"AppViewer allows viewer", AppViewer(), "app3", true},
		{"AppViewer denies no role", AppViewer(), "app4", false},

		{"AppOperator allows admin", AppOperator(), "app1", true},
		{"AppOperator allows operator", AppOperator(), "app2", true},
		{"AppOperator denies viewer", AppOperator(), "app3", false},
		{"AppOperator denies no role", AppOperator(), "app4", false},

		{"AppAdmin allows admin", AppAdmin(), "app1", true},
		{"AppAdmin denies operator", AppAdmin(), "app2", false},
		{"AppAdmin denies viewer", AppAdmin(), "app3", false},
		{"AppAdmin denies no role", AppAdmin(), "app4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.appName != "" {
				vars["app_name"] = tt.appName
			}
			result := tt.rule.GetActiveRole(vars, user).IsAllowed()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppRulesFromURL(t *testing.T) {
	user := &userctx.User{
		ID:          "test-user",
		DisplayName: "Test User",
		IsAnonymous: false,
		ApplicationRoles: map[string]userctx.Role{
			"test-app": userctx.RoleOperator,
		},
	}

	vars := map[string]string{"app_name": "test-app"}

	// Test rules that get app name from URL vars
	viewerRule := AppViewer()
	operatorRule := AppOperator()
	adminRule := AppAdmin()

	assert.Equal(t, viewerRule.GetActiveRole(vars, user), userctx.RoleOperator)
	assert.Equal(t, operatorRule.GetActiveRole(vars, user), userctx.RoleOperator)
	assert.Equal(t, adminRule.GetActiveRole(vars, user), userctx.RoleNone)
}

func TestHandleGetAppsAuthorization(t *testing.T) {
	server, _ := SetupTestServer()

	// Test with user having access to only one app
	user := &userctx.User{
		ID:          "test-user",
		DisplayName: "Test User",
		IsAnonymous: false,
		ApplicationRoles: map[string]userctx.Role{
			"test-app": userctx.RoleViewer,
			// No access to "test-logger"
		},
	}

	req := httptest.NewRequest("GET", "/api/v1/apps", nil)
	ctx := userctx.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()

	server.HandleGetApps(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// The response should only include apps the user has access to
	// This would need more detailed testing with actual JSON parsing
}

func TestHandleGetAppAuthorization(t *testing.T) {
	server, _ := SetupTestServer()

	tests := []struct {
		name           string
		user           *userctx.User
		appName        string
		expectedStatus int
	}{
		{
			name: "User with viewer role can access app",
			user: &userctx.User{
				ID:          "test-user",
				DisplayName: "Test User",
				IsAnonymous: false,
				ApplicationRoles: map[string]userctx.Role{
					"test-app": userctx.RoleViewer,
				},
			},
			appName:        "test-app",
			expectedStatus: http.StatusOK,
		},
		{
			name: "User without role cannot access app",
			user: &userctx.User{
				ID:          "test-user",
				DisplayName: "Test User",
				IsAnonymous: false,
				ApplicationRoles: map[string]userctx.Role{
					"other-app": userctx.RoleViewer,
					// No role for "test-app", and no wildcard role
				},
			},
			appName:        "test-app",
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Anonymous user with default admin role can access app",
			user: &userctx.User{
				ID:          "$anonymous$",
				DisplayName: "Anonymous",
				IsAnonymous: true,
				ApplicationRoles: map[string]userctx.Role{
					"*": userctx.RoleAdmin, // Default admin role for anonymous users
				},
			},
			appName:        "test-app",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/apps/test-app", nil)
			req = mux.SetURLVars(req, map[string]string{"app_name": tt.appName})

			ctx := userctx.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			server.HandleGetApp(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestHandleStartAppAuthorization(t *testing.T) {
	server, manager := SetupTestServer()

	// Ensure test app is stopped before testing
	_ = manager.StopApp(context.Background(), "test-app")
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name           string
		user           *userctx.User
		expectedStatus int
	}{
		{
			name: "User with operator role can start app",
			user: &userctx.User{
				ID:          "test-user",
				DisplayName: "Test User",
				IsAnonymous: false,
				ApplicationRoles: map[string]userctx.Role{
					"test-app": userctx.RoleOperator,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User with admin role can start app",
			user: &userctx.User{
				ID:          "test-user",
				DisplayName: "Test User",
				IsAnonymous: false,
				ApplicationRoles: map[string]userctx.Role{
					"test-app": userctx.RoleAdmin,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User with viewer role cannot start app",
			user: &userctx.User{
				ID:          "test-user",
				DisplayName: "Test User",
				IsAnonymous: false,
				ApplicationRoles: map[string]userctx.Role{
					"test-app": userctx.RoleViewer,
				},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "User without role cannot start app",
			user: &userctx.User{
				ID:          "test-user",
				DisplayName: "Test User",
				IsAnonymous: false,
				ApplicationRoles: map[string]userctx.Role{
					"other-app": userctx.RoleOperator,
					// No role for "test-app", and no wildcard role
				},
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure app is stopped before each test
			_ = manager.StopApp(context.Background(), "test-app")
			time.Sleep(50 * time.Millisecond)

			req := httptest.NewRequest("POST", "/api/v1/apps/test-app/start", nil)
			req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})
			ctx := userctx.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			server.HandleStartApp(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Clean up if the app was started successfully
			if recorder.Code == http.StatusOK {
				_ = manager.StopApp(context.Background(), "test-app")
				time.Sleep(50 * time.Millisecond)
			}
		})
	}
}

func TestRequireAuthorizationWithMultipleRules(t *testing.T) {
	server, _ := SetupTestServer()

	// Test user with only viewer role for app1, but admin role for app2
	user := &userctx.User{
		ID:          "test-user",
		DisplayName: "Test User",
		IsAnonymous: false,
		ApplicationRoles: map[string]userctx.Role{
			"app1": userctx.RoleViewer,
			"app2": userctx.RoleAdmin,
		},
	}

	// Test multiple rules on app2 (where user has admin role)
	req := httptest.NewRequest("POST", "/api/v1/apps/app2/start", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "app2"})
	ctx := userctx.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	// Test with multiple rules - should succeed if ANY rule allows access
	// User has admin role for app2, so any of these should succeed:
	// 1. Operator role for app2 (should succeed), OR
	// 2. Admin role for app2 (should succeed)
	allowed := server.RequireAuthorization(recorder, req,
		AppOperator(), // This will succeed (user has admin for app2, which >= operator)
		AppAdmin(),    // This will also succeed (user has admin for app2)
	).IsAllowed()

	assert.True(t, allowed, "Should be allowed because user has admin role for app2")
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRestrictiveDefaultRole(t *testing.T) {
	// Test with a server that has no default permissions for anonymous users
	server, _ := SetupTestServer()

	// Create a request without user context (will get anonymous user with RoleNone)
	req := httptest.NewRequest("GET", "/api/v1/apps/test-app", nil)
	req = mux.SetURLVars(req, map[string]string{"app_name": "test-app"})

	// Set up context with restrictive default role
	ctx := userctx.WithDefaultRole(req.Context(), userctx.RoleNone)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	server.HandleGetApp(recorder, req)

	// Should be forbidden because anonymous user has RoleNone
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}
