//go:build windows

package godump

import "github.com/goforj/godump/internal/windowsansi"

// init activates ANSI support on Windows terminals by calling the Enable
// function from the internal windowsansi package.
func init() {
	windowsansi.Enable()
} 