package api

import (
	"time"

	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sierrasoftworks/tailon/pkg/userctx"
)

// ApplicationResponseV1 represents the JSON response for application details
type ApplicationResponseV1 struct {
	Config         config.ApplicationConfig `json:"config"`
	State          apps.ApplicationState    `json:"state"`
	PID            int                      `json:"pid,omitempty"`
	LastExitCode   int                      `json:"last_exit_code"`
	StateChangedBy *userctx.User            `json:"state_changed_by,omitempty"`
	StateChangedAt *time.Time               `json:"state_changed_at,omitempty"`
}

func (a *ApplicationResponseV1) Sanitize() {
	a.Config.Env = nil
}
