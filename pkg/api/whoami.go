package api

import (
	"encoding/json"
	"net/http"

	"github.com/sierrasoftworks/tailon/pkg/userctx"
)

// HandleWhoAmI returns information about the current user
func (s *Server) HandleWhoAmI(w http.ResponseWriter, r *http.Request) {
	user := userctx.FromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
