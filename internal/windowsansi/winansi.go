//go:build windows

package windowsansi

import (
	"os"
	"syscall"
	"unsafe"
)

// Enable activates ANSI support on Windows terminals by setting the
// ENABLE_VIRTUAL_TERMINAL_PROCESSING flag.
// It fails silently if the output is not a console.
func Enable() {
	const enableVirtualTerminalProcessing = 0x0004

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
	ret, _, _ := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return
	}

	// Add the virtual terminal processing flag to the current mode.
	newMode := mode | enableVirtualTerminalProcessing

	// Try to set the new console mode.
	// If this call fails, we also silently continue. The result will be
	// that colors are not rendered, which is an acceptable fallback.
	procSetConsoleMode.Call(uintptr(handle), uintptr(newMode))
} 