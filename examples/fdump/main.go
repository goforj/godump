//go:build ignore
// +build ignore

package main

import (
	"github.com/goforj/godump"
	"strings"
)

func main() {
	// Fdump writes the formatted dump of values to the given io.Writer.

	// Example: dump to writer
	var b strings.Builder
	v := map[string]int{"a": 1}
	godump.Fdump(&b, v)
	// outputs to strings builder
}
