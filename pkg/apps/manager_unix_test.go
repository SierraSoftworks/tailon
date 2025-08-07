//go:build !windows

package apps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStopSignal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "interrupt"}, // default SIGINT
		{"SIGINT", "interrupt"},
		{"SIGTERM", "terminated"},
		{"SIGQUIT", "quit"},
		{"SIGKILL", "killed"},
		{"SIGHUP", "hangup"},
		{"INVALID", "interrupt"}, // fallback to SIGINT
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			signal := parseStopSignal(test.input)
			assert.Contains(t, signal.String(), test.expected)
		})
	}
}
