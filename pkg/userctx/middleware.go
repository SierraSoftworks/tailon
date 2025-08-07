package userctx

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
	"tailscale.com/client/tailscale"
)

// Middleware provides user context extraction middleware
type Middleware struct {
	localClient *tailscale.LocalClient
}

// NewMiddleware creates a new user context middleware
func NewMiddleware(localClient *tailscale.LocalClient) *Middleware {
	return &Middleware{
		localClient: localClient,
	}
}

// Handler wraps an HTTP handler to extract and inject user context
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := m.extractUser(r)

		// Add user to context
		ctx = WithUser(ctx, user)

		// Add user information to logger context
		logger := logrus.WithFields(logrus.Fields{
			"user_id": user.ID,
		})

		// Add logger to context for downstream use
		ctx = context.WithValue(ctx, "logger", logger)

		// Continue with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractUser extracts user information from the request
func (m *Middleware) extractUser(r *http.Request) *User {
	// If no LocalClient is available, return anonymous user
	if m.localClient == nil {
		return Anonymous()
	}

	// Extract user information from Tailscale
	whois, err := m.localClient.WhoIs(r.Context(), r.RemoteAddr)
	if err != nil {
		logrus.WithError(err).Debug("Failed to get Tailscale user info, using anonymous")
		return Anonymous()
	}

	// Convert Tailscale user info to our user format
	if whois.UserProfile != nil {
		return NewTailscaleUser(&TailscaleUserInfo{
			ID:          whois.UserProfile.ID.String(),
			DisplayName: whois.UserProfile.DisplayName,
			LoginName:   whois.UserProfile.LoginName,
			Node:        whois.Node.Name,
		})
	}

	return Anonymous()
}

// GetLoggerFromContext extracts the enriched logger from context
func GetLoggerFromContext(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value("logger").(*logrus.Entry); ok {
		return logger
	}
	// Fallback to default logger with anonymous user
	return logrus.WithFields(logrus.Fields{
		"user_id": "$anonymous$",
	})
}
