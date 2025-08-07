package main

import (
	"fmt"
	"testing"

	"github.com/sierrasoftworks/tail-on/pkg/apps"
	"github.com/sierrasoftworks/tail-on/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestLastExitCodeInitialization(t *testing.T) {
	configs := []config.ApplicationConfig{
		{
			Name: "test-app",
			Path: "/bin/echo",
			Args: []string{"hello"},
			Env:  []string{},
		},
	}

	manager := apps.NewManager(configs)
	app, err := manager.GetApp("test-app")
	assert.NoError(t, err)

	fmt.Printf("App State: %s\n", app.State)
	fmt.Printf("App PID: %d\n", app.PID)
	fmt.Printf("App LastExitCode: %d\n", app.LastExitCode)

	assert.Equal(t, 0, app.LastExitCode, "LastExitCode should be 0 for a newly created application")
}
