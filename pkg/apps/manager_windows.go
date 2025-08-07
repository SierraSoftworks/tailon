//go:build windows

package apps

import (
	"os"

	"github.com/sirupsen/logrus"
)

// gracefulStop attempts to gracefully stop a process on Windows
// Windows has limited signal support compared to Unix systems
func (m *Manager) gracefulStop(process *os.Process, signalName string) error {
	// On Windows, we can't send arbitrary signals like on Unix
	// For graceful stop, we rely on the context cancellation and hope the process
	// handles it properly. If not, we may need to use more aggressive methods.
	
	// For Windows, there's limited signal support. The main options are:
	// 1. Let context cancellation handle it (best case)
	// 2. Use process.Kill() (forceful)
	
	// Since we're trying to be graceful, we'll give the process a chance
	// to clean up through context cancellation, which should already be happening
	// at this point. If the application doesn't respond within a reasonable time,
	// the monitoring goroutine will detect the exit.
	
	// Note: On Windows, we could also try to send a CTRL_C_EVENT or CTRL_BREAK_EVENT
	// but that requires more complex implementation and process group management.
	// For now, we rely on context cancellation which should work for most Go applications
	// and other well-behaved processes.
	
	logrus.WithField("app_signal", signalName).Debug("Graceful stop initiated on Windows")
	return nil
}

// getPlatformStopDetails returns platform-specific details for stop operations
func getPlatformStopDetails(force bool, signalName string) string {
	if force {
		return "Force stopping application"
	}
	return "Gracefully stopping application"
}
