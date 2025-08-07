package ui

import (
	"embed"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tail-on/pkg/apps"
	"github.com/sirupsen/logrus"
)

//go:embed static/*
var uiFS embed.FS

type Server struct {
	manager *apps.Manager
}

func NewServer(manager *apps.Manager) *Server {
	return &Server{
		manager: manager,
	}
}

func (s *Server) Routes() *mux.Router {
	r := mux.NewRouter()

	// Serve the SPA index.html for docs root and main routes
	r.HandleFunc("/docs/", s.HandleSPA).Methods("GET")
	r.HandleFunc("/", s.HandleSPA).Methods("GET")

	// Static files
	r.Handle("/docs/openapi.yaml", http.FileServer(http.FS(uiFS)))
	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(uiFS)))

	// Add middleware
	r.Use(s.LoggingMiddleware)

	return r
}

func (s *Server) HandleSPA(w http.ResponseWriter, r *http.Request) {
	// Serve the SPA index.html file
	indexFile, err := uiFS.ReadFile("static/index.html")
	if err != nil {
		logrus.WithError(err).Error("Failed to read index.html")
		http.Error(w, "Failed to load application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(indexFile)
}

func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		logrus.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"status":      wrappedWriter.statusCode,
			"duration":    time.Since(start),
		}).Info("UI request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
