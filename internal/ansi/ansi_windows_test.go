//go:build windows

package ansi

import (
	"os"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// getConsoleMode is a helper to retrieve the current console mode for a given handle.
func getConsoleMode(handle syscall.Handle) (uint32, error) {
	var mode uint32
	ret, _, err := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleMode").Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if ret == SYS_CALL_FAILURE {
		// Note: err may be non-nil even on success, so we must check ret first.
		return 0, err
	}
	return mode, nil
}

// setConsoleMode is a helper to set the console mode for a given handle.
func setConsoleMode(handle syscall.Handle, mode uint32) error {
	ret, _, err := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode").Call(uintptr(handle), uintptr(mode))
	if ret == SYS_CALL_FAILURE {
		return err
	}
	return nil
}

func TestEnable(t *testing.T) {
	// This test requires a real Windows console. If stdout is redirected (e.g., in some CI/CD
	// environments), GetConsoleMode will fail. In that case, we should skip the test.
	handle := syscall.Handle(os.Stdout.Fd())
	originalMode, err := getConsoleMode(handle)
	if err != nil && err.Error() != "The handle is invalid." {
		// "The handle is invalid." is the typical error when not in a console.
		// We skip on this specific error.
		t.Skipf("cannot get console mode, skipping test: %v", err)
	}

	// Defer the restoration of the original console mode.
	// This ensures that we don't mess up the terminal for subsequent tests.
	if err == nil { // Only restore if we successfully got the mode.
		defer func() {
			err := setConsoleMode(handle, originalMode)
			require.NoError(t, err, "failed to restore original console mode")
		}()
	}

	// Run the function we want to test.
	err = Enable()
	require.NoError(t, err, "Enable() should not return an error in a real console")

	// After running Enable(), check the console mode again to see if the flag was set.
	newMode, err := getConsoleMode(handle)
	require.NoError(t, err, "failed to get new console mode after enabling ANSI")

	// Assert that the flag for virtual terminal processing is now set.
	flagIsSet := (newMode & enableVirtualTerminalProcessing) != 0
	require.True(t, flagIsSet, "ENABLE_VIRTUAL_TERMINAL_PROCESSING flag should have been set")
} 