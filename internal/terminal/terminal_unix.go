//go:build !windows

package terminal

import (
	"os"
)

// IsTerminal checks if the given file is a terminal.
// Uses ModeCharDevice on Unix-like systems.
// In test environments, it returns true unless explicitly overridden by environment variables.
func IsTerminal(f *os.File) bool {
	if IsTestEnv() {
		return true
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
} 