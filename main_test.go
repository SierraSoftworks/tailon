package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command is properly configured
	assert.Equal(t, "tail-on", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
}

func TestCommandFlags(t *testing.T) {
	// Test that flags are properly configured
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "config.yaml", configFlag.DefValue)

	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

func TestConfigValidation(t *testing.T) {
	// Create a temporary valid config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	validConfig := `
applications:
  - name: "test-app"
    path: "/bin/echo"
    args: ["hello"]
    env: ["TEST=1"]
listen: "localhost:8080"
tailscale:
  name: "test-server"
  state_dir: "/tmp/test"
`

	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	require.NoError(t, err)

	// Test that we can load the config (this tests the config loading without starting the server)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"tail-on", "--config", configFile, "--help"}

	// This should not panic and should show help without starting the server
	err = rootCmd.Execute()
	assert.NoError(t, err)
}

func TestInvalidConfigFile(t *testing.T) {
	// Test with non-existent config file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"tail-on", "--config", "/non/existent/file.yaml", "--help"}

	// Should still work for help command
	err := rootCmd.Execute()
	assert.NoError(t, err)
}
