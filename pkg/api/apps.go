package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
	"github.com/sirupsen/logrus"
)

// ApplicationResponse represents the JSON response for application details
type ApplicationResponse struct {
	Config         config.ApplicationConfig `json:"config"`
	State          apps.ApplicationState    `json:"state"`
	PID            int                      `json:"pid,omitempty"`
	LastExitCode   int                      `json:"last_exit_code"`
	StateChangedBy *userctx.User            `json:"state_changed_by,omitempty"`
	StateChangedAt *time.Time               `json:"state_changed_at,omitempty"`
}

// HandleGetApps returns all configured applications (filtered by user permissions)
func (s *Server) HandleGetApps(w http.ResponseWriter, r *http.Request) {
	user := userctx.FromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	allApps := s.manager.GetApps()
	authorizedApps := make(map[string]ApplicationResponse)

	// Filter applications based on user permissions
	for appName, appData := range allApps {
		// Create a dummy vars map for rule evaluation
		vars := map[string]string{"app_name": appName}

		// Check if user has viewer role or higher for this app
		viewerRule := AppViewer()
		if viewerRule.IsAllowed(vars, user) {
			// Create response object
			response := ApplicationResponse{
				Config:         appData.Config,
				State:          appData.State,
				PID:            appData.PID,
				LastExitCode:   appData.LastExitCode,
				StateChangedBy: appData.StateChangedBy,
				StateChangedAt: appData.StateChangedAt,
			}

			// Check if user has admin role for this app - if not, remove env variables
			adminRule := AppAdmin()
			if !adminRule.IsAllowed(vars, user) {
				// User is not admin, remove environment variables
				configCopy := appData.Config
				configCopy.Env = nil // Remove environment variables
				response.Config = configCopy
			}

			authorizedApps[appName] = response
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authorizedApps); err != nil {
		logrus.WithError(err).Error("Failed to encode apps response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGetApp returns details of a specific application
func (s *Server) HandleGetApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]

	// Check authorization - require viewer role to view application details
	if !s.RequireAuthorization(w, r, AppViewer()) {
		return
	}

	app, err := s.manager.GetApp(appName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Create response object
	response := ApplicationResponse{
		Config:         app.Config,
		State:          app.State,
		PID:            app.PID,
		LastExitCode:   app.LastExitCode,
		StateChangedBy: app.StateChangedBy,
		StateChangedAt: app.StateChangedAt,
	}

	// Check if user has admin role for this app - if not, remove env variables
	user := userctx.FromContext(r.Context())
	if user != nil {
		adminRule := AppAdmin()
		if !adminRule.IsAllowed(vars, user) {
			// User is not admin, remove environment variables
			configCopy := app.Config
			configCopy.Env = nil // Remove environment variables
			response.Config = configCopy
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.WithError(err).Error("Failed to encode app response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
