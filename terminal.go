package godump

import (
	"os"

	"golang.org/x/term"
)

// isTerminal checks if the given file is a terminal using the Go standard library.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
} 