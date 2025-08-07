package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleLogs returns application logs (JSON or Server-Sent Events)
func (s *Server) HandleLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]

	// Check if this is a Server-Sent Events request
	if r.Header.Get("Accept") == "text/event-stream" || r.URL.Query().Get("stream") == "true" {
		s.handleLogsSSE(w, r, appName)
		return
	}

	// Regular JSON response
	logs, err := s.manager.GetLogs(appName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		logrus.WithError(err).Error("Failed to encode logs response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleLogsSSE handles Server-Sent Events streaming for logs
func (s *Server) handleLogsSSE(w http.ResponseWriter, r *http.Request, appName string) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send existing logs first
	logs, err := s.manager.GetLogs(appName)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	for _, log := range logs {
		data, _ := json.Marshal(log)
		fmt.Fprintf(w, "data: %s\n\n", data)
	}
	flusher.Flush()

	// Keep connection alive and send new logs
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastLogCount := len(logs)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			currentLogs, err := s.manager.GetLogs(appName)
			if err != nil {
				return
			}

			// Send only new logs
			if len(currentLogs) > lastLogCount {
				newLogs := currentLogs[lastLogCount:]
				for _, log := range newLogs {
					data, _ := json.Marshal(log)
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
				flusher.Flush()
				lastLogCount = len(currentLogs)
			}
		}
	}
}
