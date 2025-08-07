package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expected    *Config
		expectError bool
	}{
		{
			name: "valid config",
			configYAML: `
applications:
  - name: "test-app"
    path: "/bin/echo"
    args: ["hello"]
    env: ["TEST=1"]
listen: "localhost:8080"
tailscale:
  name: "test-server"
  state_dir: "/tmp/test"
`,
			expected: &Config{
				Applications: []ApplicationConfig{
					{
						Name: "test-app",
						Path: "/bin/echo",
						Args: []string{"hello"},
						Env:  []string{"TEST=1"},
					},
				},
				Listen: "localhost:8080",
				Tailscale: TailscaleConfig{
					Name:     "test-server",
					StateDir: "/tmp/test",
				},
			},
			expectError: false,
		},
		{
			name: "empty config",
			configYAML: `
applications: []
listen: ""
tailscale:
  name: ""
  state_dir: ""
`,
			expected: &Config{
				Applications: []ApplicationConfig{},
				Listen:       "",
				Tailscale: TailscaleConfig{
					Name:     "",
					StateDir: "",
				},
			},
			expectError: false,
		},
		{
			name: "invalid yaml",
			configYAML: `
invalid: yaml: content
  - malformed
`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configFile, []byte(tt.configYAML), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := Load(configFile)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, cfg)
			}
		})
	}
}

func TestConfigLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/non/existent/file.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}
