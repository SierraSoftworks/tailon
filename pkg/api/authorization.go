package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
)

// AuthorizationRule defines an interface for authorization rules
type AuthorizationRule interface {
	IsAllowed(vars map[string]string, user *userctx.User) bool
}

// RequireAuthorization checks if the user meets any of the provided authorization rules
// Returns true if authorized, false otherwise and writes 403 Forbidden to the response
func (s *Server) RequireAuthorization(w http.ResponseWriter, r *http.Request, rules ...AuthorizationRule) bool {
	user := userctx.FromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Get URL variables for rule evaluation
	vars := mux.Vars(r)

	// Check if any rule allows access
	for _, rule := range rules {
		if rule.IsAllowed(vars, user) {
			return true
		}
	}

	http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
	return false
}

// AppRole creates a rule that requires a specific role or higher for an application
type AppRole struct {
	roles map[string]bool
}

func (rule AppRole) IsAllowed(vars map[string]string, user *userctx.User) bool {
	appName := vars["app_name"]
	userRole := user.GetRole(appName)
	return rule.roles[userRole]
}

// AppViewer creates a rule that requires viewer role or higher for the application in the URL
func AppViewer() AuthorizationRule {
	return AppRole{
		roles: map[string]bool{
			userctx.RoleViewer:   true,
			userctx.RoleOperator: true,
			userctx.RoleAdmin:    true,
		},
	}
}

// AppOperator creates a rule that requires operator role or higher for the application in the URL
func AppOperator() AuthorizationRule {
	return AppRole{
		roles: map[string]bool{
			userctx.RoleOperator: true,
			userctx.RoleAdmin:    true,
		},
	}
}

// AppAdmin creates a rule that requires admin role for the application in the URL
func AppAdmin() AuthorizationRule {
	return AppRole{
		roles: map[string]bool{
			userctx.RoleAdmin: true,
		},
	}
}

// SetupTestServerWithDefaultRole creates a test server with a specific default role
func SetupTestServerWithDefaultRole(defaultRole string) (*Server, *apps.Manager) {
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
