//go:build windows

package ansi

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	SYS_CALL_FAILURE                   = 0
	enableVirtualTerminalProcessing     = 0x0004
)

// Enable activates ANSI support on Windows terminals by setting the
// ENABLE_VIRTUAL_TERMINAL_PROCESSING flag.
// Returns an error if the output is not a console or if setting the mode fails.
func Enable() error {

	// Load kernel32.dll and the necessary procedures dynamically.
	// This avoids a hard dependency and allows the program to run on non-Windows
	// systems, although this file is guarded by a build tag.
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode := kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode := kernel32.NewProc("SetConsoleMode")

	// Get the handle for standard output.
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32

	// GetConsoleMode fails if not in a real console.
	ret, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if ret == SYS_CALL_FAILURE {
		return err
	}

	// Add the virtual terminal processing flag to the current mode.
	newMode := mode | enableVirtualTerminalProcessing

	// Try to set the new console mode.
	ret, _, err = procSetConsoleMode.Call(uintptr(handle), uintptr(newMode))
	if ret == SYS_CALL_FAILURE {
		return err
	}

	return nil
} 