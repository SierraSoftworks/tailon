package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetApps returns all configured applications
func (s *Server) HandleGetApps(w http.ResponseWriter, r *http.Request) {
	apps := s.manager.GetApps()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(apps); err != nil {
		logrus.WithError(err).Error("Failed to encode apps response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGetApp returns details of a specific application
func (s *Server) HandleGetApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]

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
