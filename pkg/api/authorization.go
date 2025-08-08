package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
)

// AuthorizationRule defines an interface for authorization rules
type AuthorizationRule interface {
	GetActiveRole(vars map[string]string, user *userctx.User) userctx.Role
}

// RequireAuthorization checks if the user meets any of the provided authorization rules
// Returns true if authorized, false otherwise and writes 403 Forbidden to the response
func (s *Server) RequireAuthorization(w http.ResponseWriter, r *http.Request, rules ...AuthorizationRule) userctx.Role {
	user := userctx.FromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return userctx.RoleNone
	}

	// Get URL variables for rule evaluation
	vars := mux.Vars(r)

	// Check if any rule allows access
	for _, rule := range rules {
		if role := rule.GetActiveRole(vars, user); role.IsAllowed() {
			return role
		}
	}

	http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
	return ""
}

// AppRole creates a rule that requires a specific role or higher for an application
type appRole struct {
	roles map[userctx.Role]bool
}

func (rule appRole) GetActiveRole(vars map[string]string, user *userctx.User) userctx.Role {
	appName := vars["app_name"]
	userRole := user.GetRole(appName)
	if rule.roles[userRole] {
		return userRole
	}

	return userctx.RoleNone
}

func AppRole(roles ...userctx.Role) AuthorizationRule {
	roleMap := make(map[userctx.Role]bool)
	for _, role := range roles {
		roleMap[role] = true
	}
	return appRole{
		roles: roleMap,
	}
}

// AppViewer creates a rule that requires viewer role or higher for the application in the URL
func AppViewer() AuthorizationRule {
	return AppRole(userctx.RoleAdmin, userctx.RoleOperator, userctx.RoleViewer)
}

// AppOperator creates a rule that requires operator role or higher for the application in the URL
func AppOperator() AuthorizationRule {
	return AppRole(userctx.RoleAdmin, userctx.RoleOperator)
}

// AppAdmin creates a rule that requires admin role for the application in the URL
func AppAdmin() AuthorizationRule {
	return AppRole(userctx.RoleAdmin)
}
