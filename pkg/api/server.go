package api

import (
	"github.com/gorilla/mux"
	"github.com/sierrasoftworks/tail-on/pkg/apps"
	"github.com/sierrasoftworks/tail-on/pkg/userctx"
	"tailscale.com/client/tailscale"
)

type Server struct {
	manager        *apps.Manager
	userMiddleware *userctx.Middleware
}

func NewServer(manager *apps.Manager) *Server {
	return &Server{
		manager:        manager,
		userMiddleware: userctx.NewMiddleware(nil), // Default to no Tailscale client (anonymous users)
	}
}

func NewServerWithTailscale(manager *apps.Manager, localClient *tailscale.LocalClient) *Server {
	return &Server{
		manager:        manager,
		userMiddleware: userctx.NewMiddleware(localClient),
	}
}

func (s *Server) Routes() *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/whoami", s.HandleWhoAmI).Methods("GET")
	api.HandleFunc("/apps", s.HandleGetApps).Methods("GET")
	api.HandleFunc("/apps/{app_name}", s.HandleGetApp).Methods("GET")
	api.HandleFunc("/apps/{app_name}/start", s.HandleStartApp).Methods("POST")
	api.HandleFunc("/apps/{app_name}/stop", s.HandleStopApp).Methods("POST")
	api.HandleFunc("/apps/{app_name}/restart", s.HandleRestartApp).Methods("POST")
	api.HandleFunc("/apps/{app_name}/logs", s.HandleLogs).Methods("GET")

	// Add middleware
	r.Use(s.userMiddleware.Handler) // Add user context middleware first
	r.Use(s.LoggingMiddleware)
	r.Use(s.CORSMiddleware)

	return r
}
