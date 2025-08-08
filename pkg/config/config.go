package config

import (
	"fmt"
	"os"

	"github.com/sierrasoftworks/tailon/pkg/userctx"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Applications []ApplicationConfig `json:"applications" yaml:"applications"`
	// The address on the local machine to listen on for incoming connections
	Listen string `json:"listen" yaml:"listen"`
	// The address on the Tailscale network to listen on for incoming connections
	Tailscale TailscaleConfig `json:"tailscale" yaml:"tailscale"`
	// Security configuration
	Security SecurityConfig `json:"security" yaml:"security"`
}

type ApplicationConfig struct {
	Name       string   `json:"name" yaml:"name"`
	Path       string   `json:"path" yaml:"path"`
	Args       []string `json:"args" yaml:"args"`
	Env        []string `json:"env" yaml:"env"`
	WorkingDir string   `json:"working_dir" yaml:"working_dir"` // Working directory for the application
	StopSignal string   `json:"stop_signal" yaml:"stop_signal"` // Signal to use for stopping (default: SIGINT)
}

type TailscaleConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Name      string `json:"name" yaml:"name"`
	Ephemeral bool   `json:"ephemeral" yaml:"ephemeral"`
	StateDir  string `json:"state_dir" yaml:"state_dir"`
}

type SecurityConfig struct {
	// Whether to allow anonymous users when Tailscale is disabled
	DefaultRole userctx.Role `json:"default_role,omitempty" yaml:"default_role"`
	// List of allowed IP addresses/CIDR blocks for anonymous access
	AllowedIPs []string `json:"allowed_ips" yaml:"allowed_ips"`
	// Whether to hide environment variables in API responses
	HideEnvVars bool `json:"hide_env_vars" yaml:"hide_env_vars"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
