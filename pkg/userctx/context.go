package userctx

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// User represents a user making requests to the system
type User struct {
	ID               string            `json:"id"`
	DisplayName      string            `json:"display_name"`
	LoginName        string            `json:"login_name,omitempty"`
	Node             string            `json:"node,omitempty"`
	IsAnonymous      bool              `json:"is_anonymous"`
	IPAddress        string            `json:"ip_address,omitempty"` // Track IP for audit logging
	ApplicationRoles map[string]string `json:"app_roles,omitempty"`
}

func (u *User) GetRole(app string) string {
	if role, ok := u.ApplicationRoles[app]; ok {
		if role == RoleAdmin || role == RoleOperator || role == RoleViewer {
			return role
		}
	}

	if role, ok := u.ApplicationRoles["*"]; ok {
		if role == RoleAdmin || role == RoleOperator || role == RoleViewer {
			return role
		}
	}

	return RoleNone
}

// AnonymousFromIP creates an anonymous user with IP-based tracking
func AnonymousFromIP(remoteAddr, defaultRole string) *User {
	// Extract IP from address (remove port)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If we can't parse it, use the full address
		host = remoteAddr
	}

	return &User{
		ID:          fmt.Sprintf("$anonymous-%s$", host),
		DisplayName: fmt.Sprintf("Anonymous (%s)", host),
		IsAnonymous: true,
		IPAddress:   host,
		ApplicationRoles: map[string]string{
			"*": defaultRole,
		},
	}
}

// Anonymous returns the default anonymous user (for backward compatibility)
func Anonymous(defaultRole string) *User {
	return &User{
		ID:          "$anonymous$",
		DisplayName: "Anonymous",
		IsAnonymous: true,
		ApplicationRoles: map[string]string{
			"*": defaultRole,
		},
	}
}

// NewTailscaleUser creates a user from Tailscale user information
func NewTailscaleUser(userInfo *TailscaleUserInfo, defaultRole string) *User {
	if userInfo == nil {
		return Anonymous(defaultRole)
	}

	roles := map[string]string{
		"*": defaultRole,
	}

	for _, grant := range userInfo.Grants {
		for _, app := range grant.Applications {
			roles[app] = grant.Role
		}
	}

	return &User{
		ID:               userInfo.ID,
		DisplayName:      userInfo.DisplayName,
		LoginName:        userInfo.LoginName,
		Node:             userInfo.Node,
		IsAnonymous:      false,
		IPAddress:        userInfo.IPAddress,
		ApplicationRoles: roles,
	}
}

// TailscaleUserInfo represents user information from Tailscale LocalClient
type TailscaleUserInfo struct {
	ID          string
	DisplayName string
	LoginName   string
	Node        string
	IPAddress   string
	Grants      []RoleAssignment
}

// UserEvent represents an action performed by a user
type UserEvent struct {
	User      *User     `json:"user"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Details   string    `json:"details,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"` // Always track IP for audit
}

// NewUserEvent creates a new user event
func NewUserEvent(user *User, action, target, details string) *UserEvent {
	event := &UserEvent{
		User:      user,
		Timestamp: time.Now(),
		Action:    action,
		Target:    target,
		Details:   details,
	}

	// Always include IP address for audit trail
	if user != nil && user.IPAddress != "" {
		event.IPAddress = user.IPAddress
	}

	return event
}

// String implements the fmt.Stringer interface for better logging
func (e *UserEvent) String() string {
	userName := "Anonymous"
	if e.User != nil && !e.User.IsAnonymous {
		userName = e.User.DisplayName
	}

	// Include IP address in audit string
	ipInfo := ""
	if e.IPAddress != "" {
		ipInfo = fmt.Sprintf(" from %s", e.IPAddress)
	}

	if e.Details != "" {
		return fmt.Sprintf("%s %s %s (%s)%s at %s", userName, e.Action, e.Target, e.Details, ipInfo, e.Timestamp.Format("15:04:05"))
	}
	return fmt.Sprintf("%s %s %s%s at %s", userName, e.Action, e.Target, ipInfo, e.Timestamp.Format("15:04:05"))
}

// Key type for context values to avoid collisions
type contextKey string

const userContextKey contextKey = "user"
const defaultRoleContextKey contextKey = "defaultRole"

// WithUser adds a user to the context
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func WithDefaultRole(ctx context.Context, defaultRole string) context.Context {
	return context.WithValue(ctx, defaultRoleContextKey, defaultRole)
}

// FromContext extracts the user from the context, returning Anonymous if not found
func FromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}

	return Anonymous(getDefaultRoleFromContext(ctx))
}

// GetLoggerFromContext extracts the enriched logger from context
func GetLoggerFromContext(ctx context.Context) *logrus.Entry {
	if logger := ctx.Value("logger"); logger != nil {
		if entry, ok := logger.(*logrus.Entry); ok {
			return entry
		}
	}
	// Fallback to default logger with anonymous user
	return logrus.WithFields(logrus.Fields{
		"user_id":      "$anonymous$",
		"is_anonymous": true,
	})
}

func getDefaultRoleFromContext(ctx context.Context) string {
	if defaultRole, ok := ctx.Value(defaultRoleContextKey).(string); ok {
		return defaultRole
	}

	return RoleAdmin
}
