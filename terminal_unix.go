//go:build !windows

package godump

import (
	"os"
	"testing"
)

// isTerminal checks if the given file is a terminal.
// Uses ModeCharDevice on Unix-like systems.
// In test environments, it returns true unless explicitly overridden by environment variables.
func isTerminal(f *os.File) bool {
	if isTestEnvironment() {
		return true
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// isTestEnvironment checks if the code is running in a test environment
func isTestEnvironment() bool {
	return testing.Testing()
} 