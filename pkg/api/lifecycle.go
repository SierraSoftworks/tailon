package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleStartApp starts an application
func (s *Server) HandleStartApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]

	if err := s.manager.StartApp(appName); err != nil {
		logrus.WithError(err).WithField("app", appName).Error("Failed to start application")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

// HandleStopApp stops an application
func (s *Server) HandleStopApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]

	if err := s.manager.StopApp(appName); err != nil {
		logrus.WithError(err).WithField("app", appName).Error("Failed to stop application")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// HandleRestartApp restarts an application (stop then start)
func (s *Server) HandleRestartApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]

	// Check if app exists first
	_, err := s.manager.GetApp(appName)
	if err != nil {
		logrus.WithError(err).WithField("app", appName).Error("Failed to restart application - app not found")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Try to stop the app if it's running (ignore errors if already stopped)
	if err := s.manager.StopApp(appName); err != nil {
		logrus.WithField("app", appName).Debug("App was not running during restart")
	}

	// Give the app a moment to stop completely
	time.Sleep(100 * time.Millisecond)

	// Start the app
	if err := s.manager.StartApp(appName); err != nil {
		logrus.WithError(err).WithField("app", appName).Error("Failed to start application during restart")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}
