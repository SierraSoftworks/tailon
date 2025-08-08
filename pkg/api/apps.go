package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
	"github.com/sirupsen/logrus"
)

// HandleGetApps returns all configured applications (filtered by user permissions)
func (s *Server) HandleGetApps(w http.ResponseWriter, r *http.Request) {
	user := userctx.FromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	allApps := s.manager.GetApps()
	authorizedApps := make(map[string]ApplicationResponseV1)

	// Filter applications based on user permissions
	viewerRule := AppViewer()
	for appName, appData := range allApps {
		// Create a dummy vars map for rule evaluation
		vars := map[string]string{"app_name": appName}

		role := viewerRule.GetActiveRole(vars, user)
		if role.IsAllowed() {
			// Create response object
			response := ApplicationResponseV1{
				Config:         appData.Config,
				State:          appData.State,
				PID:            appData.PID,
				LastExitCode:   appData.LastExitCode,
				StateChangedBy: appData.StateChangedBy,
				StateChangedAt: appData.StateChangedAt,
			}

			// Check if user has admin role for this app - if not, remove env variables
			if role != userctx.RoleAdmin {
				response.Sanitize()
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

	role := s.RequireAuthorization(w, r, AppViewer())
	if !role.IsAllowed() {
		return
	}

	app, err := s.manager.GetApp(appName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Create response object
	response := ApplicationResponseV1{
		Config:         app.Config,
		State:          app.State,
		PID:            app.PID,
		LastExitCode:   app.LastExitCode,
		StateChangedBy: app.StateChangedBy,
		StateChangedAt: app.StateChangedAt,
	}

	// Check if user has admin role for this app - if not, remove env variables
	if role != userctx.RoleAdmin {
		response.Sanitize()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.WithError(err).Error("Failed to encode app response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
