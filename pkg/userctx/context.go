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
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	LoginName   string `json:"login_name,omitempty"`
	Node        string `json:"node,omitempty"`
	IsAnonymous bool   `json:"is_anonymous"`
	IPAddress   string `json:"ip_address,omitempty"` // Track IP for audit logging
}

// AnonymousFromIP creates an anonymous user with IP-based tracking
func AnonymousFromIP(remoteAddr string) *User {
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
	}
}

// Anonymous returns the default anonymous user (for backward compatibility)
func Anonymous() *User {
	return &User{
		ID:          "$anonymous$",
		DisplayName: "Anonymous",
		IsAnonymous: true,
	}
}

// NewTailscaleUser creates a user from Tailscale user information
func NewTailscaleUser(userInfo *TailscaleUserInfo) *User {
	if userInfo == nil {
		return Anonymous()
	}

	return &User{
		ID:          userInfo.ID,
		DisplayName: userInfo.DisplayName,
		LoginName:   userInfo.LoginName,
		Node:        userInfo.Node,
		IsAnonymous: false,
		IPAddress:   userInfo.IPAddress,
	}
}

// TailscaleUserInfo represents user information from Tailscale LocalClient
type TailscaleUserInfo struct {
	ID          string
	DisplayName string
	LoginName   string
	Node        string
	IPAddress   string
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

// WithUser adds a user to the context
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// FromContext extracts the user from the context, returning Anonymous if not found
func FromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return Anonymous()
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
