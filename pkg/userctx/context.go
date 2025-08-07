package userctx

import (
	"context"
	"fmt"
	"time"
)

// User represents a user making requests to the system
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	LoginName   string `json:"login_name,omitempty"`
	Node        string `json:"node,omitempty"`
	IsAnonymous bool   `json:"is_anonymous"`
}

// Anonymous returns the default anonymous user
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
	}
}

// TailscaleUserInfo represents user information from Tailscale LocalClient
type TailscaleUserInfo struct {
	ID          string
	DisplayName string
	LoginName   string
	Node        string
}

// UserEvent represents an action performed by a user
type UserEvent struct {
	User      *User     `json:"user"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Details   string    `json:"details,omitempty"`
}

// NewUserEvent creates a new user event
func NewUserEvent(user *User, action, target, details string) *UserEvent {
	return &UserEvent{
		User:      user,
		Timestamp: time.Now(),
		Action:    action,
		Target:    target,
		Details:   details,
	}
}

// String implements the fmt.Stringer interface for better logging
func (e *UserEvent) String() string {
	userName := "Anonymous"
	if e.User != nil && !e.User.IsAnonymous {
		userName = e.User.DisplayName
	}

	if e.Details != "" {
		return fmt.Sprintf("%s %s %s (%s) at %s", userName, e.Action, e.Target, e.Details, e.Timestamp.Format("15:04:05"))
	}
	return fmt.Sprintf("%s %s %s at %s", userName, e.Action, e.Target, e.Timestamp.Format("15:04:05"))
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
