package userctx

import (
	"context"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"tailscale.com/client/local"
	"tailscale.com/tailcfg"
)

// Middleware provides user context extraction middleware
type Middleware struct {
	localClient *local.Client
}

// NewMiddleware creates a new user context middleware
func NewMiddleware(localClient *local.Client) *Middleware {
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
			"user_id":      user.ID,
			"user_name":    user.DisplayName,
			"is_anonymous": user.IsAnonymous,
			"ip_address":   user.IPAddress,
		})

		// Add logger to context for downstream use
		ctx = context.WithValue(ctx, "logger", logger)

		// Continue with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractUser extracts user information from the request
func (m *Middleware) extractUser(r *http.Request) *User {
	// Extract IP address from request
	remoteAddr := r.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	defaultRole := getDefaultRoleFromContext(r.Context())

	// If no LocalClient is available, return anonymous user with IP tracking
	if m.localClient == nil {
		return AnonymousFromIP(remoteAddr, defaultRole)
	}

	// Extract user information from Tailscale
	whois, err := m.localClient.WhoIs(r.Context(), remoteAddr)
	if err != nil {
		logrus.WithError(err).WithField("remote_addr", remoteAddr).Debug("Failed to get Tailscale user info, using anonymous")
		return AnonymousFromIP(remoteAddr, defaultRole)
	}

	grants, err := tailcfg.UnmarshalCapJSON[RoleAssignment](whois.CapMap, tailcfg.PeerCapability("https://sierrasoftworks.com/cap/tailon"))

	// Convert Tailscale user info to our user format
	if whois.UserProfile != nil {
		return NewTailscaleUser(&TailscaleUserInfo{
			ID:          whois.UserProfile.ID.String(),
			DisplayName: whois.UserProfile.DisplayName,
			LoginName:   whois.UserProfile.LoginName,
			Node:        whois.Node.Name,
			IPAddress:   host,
			Grants:      grants,
		}, defaultRole)
	}

	return AnonymousFromIP(remoteAddr, defaultRole)
}
