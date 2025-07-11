//go:build !windows

package godump

import (
	"os"
	"testing"
)

var isTestEnv = testing.Testing

// isTerminal checks if the given file is a terminal.
// Uses ModeCharDevice on Unix-like systems.
// In test environments, it returns true unless explicitly overridden by environment variables.
func isTerminal(f *os.File) bool {
	if isTestEnv() {
		return true
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}