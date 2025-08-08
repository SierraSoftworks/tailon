package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tailon/pkg/apps"
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
	authorizedApps := make(map[string]*apps.Application)

	// Filter applications based on user permissions
	for appName, appData := range allApps {
		// Create a dummy vars map for rule evaluation
		vars := map[string]string{"app_name": appName}

		// Check if user has viewer role or higher for this app
		viewerRule := AppViewer()
		if viewerRule.IsAllowed(vars, user) {
			authorizedApps[appName] = appData
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(app); err != nil {
		logrus.WithError(err).Error("Failed to encode app response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
