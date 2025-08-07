//go:build !windows

package apps

import (
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

// gracefulStop attempts to gracefully stop a process using Unix signals
func (m *Manager) gracefulStop(process *os.Process, signalName string) error {
	// On Unix-like systems, use traditional signal handling
	sig := parseStopSignal(signalName)
	return process.Signal(sig)
}

// parseStopSignal converts a string signal name to syscall.Signal
// This is a package-level function only used on Unix systems
func parseStopSignal(signalName string) os.Signal {
	if signalName == "" {
		return syscall.SIGINT // default
	}

	switch signalName {
	case "SIGINT":
		return syscall.SIGINT
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGQUIT":
		return syscall.SIGQUIT
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGHUP":
		return syscall.SIGHUP
	default:
		logrus.WithField("signal", signalName).Warn("Unknown signal, defaulting to SIGINT")
		return syscall.SIGINT
	}
}

// getPlatformStopDetails returns platform-specific details for stop operations
func getPlatformStopDetails(force bool, signalName string) string {
	if force {
		return "Force stopping application"
	}
	
	signal := signalName
	if signal == "" {
		signal = "SIGINT"
	}
	return "Gracefully stopping application (" + signal + ")"
}
