package userctx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserContext(t *testing.T) {
	// Test anonymous user
	anonymous := Anonymous(RoleNone)
	assert.True(t, anonymous.IsAnonymous)
	assert.Equal(t, "$anonymous$", anonymous.ID)
	assert.Equal(t, "Anonymous", anonymous.DisplayName)

	// Test context with user
	ctx := context.Background()
	ctx = WithUser(ctx, anonymous)

	retrievedUser := FromContext(ctx)
	assert.Equal(t, anonymous, retrievedUser)

	// Test context without user
	emptyCtx := context.Background()
	defaultUser := FromContext(emptyCtx)
	assert.True(t, defaultUser.IsAnonymous)
	assert.Equal(t, "$anonymous$", defaultUser.ID)
}

func TestTailscaleUser(t *testing.T) {
	// Test with nil user info
	user := NewTailscaleUser(nil, RoleViewer)
	assert.True(t, user.IsAnonymous)
	assert.Equal(t, "$anonymous$", user.ID)

	// Test with valid user info
	userInfo := &TailscaleUserInfo{
		ID:          "user123",
		DisplayName: "John Doe",
		LoginName:   "john@example.com",
		Node:        "johns-laptop",
		Grants: []RoleAssignment{
			RoleAssignment{
				Role: RoleAdmin,
				Applications: []string{
					"test-app",
				},
			},
		},
	}

	user = NewTailscaleUser(userInfo, RoleViewer)
	assert.False(t, user.IsAnonymous)
	assert.Equal(t, "user123", user.ID)
	assert.Equal(t, "John Doe", user.DisplayName)
	assert.Equal(t, "john@example.com", user.LoginName)
	assert.Equal(t, "johns-laptop", user.Node)
	assert.Equal(t, RoleAdmin, user.GetRole("test-app"))
}

func TestUserEvent(t *testing.T) {
	user := &User{
		ID:          "test-user",
		DisplayName: "Test User",
		IsAnonymous: false,
	}

	event := NewUserEvent(user, "start", "my-app", "Started application")

	assert.Equal(t, user, event.User)
	assert.Equal(t, "start", event.Action)
	assert.Equal(t, "my-app", event.Target)
	assert.Equal(t, "Started application", event.Details)
	assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
}
