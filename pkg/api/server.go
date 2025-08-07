package api

import (
	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tail-on/pkg/apps"
)

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

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/apps", s.HandleGetApps).Methods("GET")
	api.HandleFunc("/apps/{app_name}", s.HandleGetApp).Methods("GET")
	api.HandleFunc("/apps/{app_name}/start", s.HandleStartApp).Methods("POST")
	api.HandleFunc("/apps/{app_name}/stop", s.HandleStopApp).Methods("POST")
	api.HandleFunc("/apps/{app_name}/restart", s.HandleRestartApp).Methods("POST")
	api.HandleFunc("/apps/{app_name}/logs", s.HandleLogs).Methods("GET")

	// Documentation routes
	docs := r.PathPrefix("/docs").Subrouter()
	docs.HandleFunc("/openapi.yaml", s.HandleOpenAPISpec).Methods("GET")
	docs.HandleFunc("/swagger/", s.HandleSwaggerUI).Methods("GET")
	docs.HandleFunc("/swagger", s.HandleSwaggerUI).Methods("GET")
	docs.HandleFunc("/", s.HandleAPIDocs).Methods("GET")
	docs.HandleFunc("", s.HandleAPIDocs).Methods("GET")

	// Add middleware
	r.Use(s.LoggingMiddleware)
	r.Use(s.CORSMiddleware)

	return r
}
