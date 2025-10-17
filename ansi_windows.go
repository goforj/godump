//go:build windows

package godump

import (
	"log"

	"github.com/goforj/godump/internal/ansi"
)

// init activates ANSI support on Windows terminals by calling the Enable
// function from the internal ansi package.
// If enabling ANSI fails (e.g., not running in a real console), it logs
// the error but continues execution, as colors are optional.
func init() {
	if err := ansi.Enable(); err != nil {
		log.Printf("godump: failed to enable ANSI (likely due to output redirection): %v\n", err)
	}
}