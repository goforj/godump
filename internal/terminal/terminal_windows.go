//go:build windows

package terminal

import (
	"os"
	"syscall"
)

// IsTerminal checks if the given file is a terminal.
// Uses GetConsoleMode on Windows.
// In test environments, it returns true unless explicitly overridden by environment variables.
func IsTerminal(f *os.File) bool {
	if IsTestEnv() {
		return true
	}

	var mode uint32
	// GetConsoleMode succeeds only for console handles
	// Fails for redirected/piped output
	err := syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode)
	return err == nil
} 