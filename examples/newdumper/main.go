//go:build ignore
// +build ignore

package main

import (
	"github.com/goforj/godump"
	"os"
)

func main() {
	// NewDumper creates a new Dumper with the given options applied.
	// Defaults are used for any setting not overridden.

	// Example: build a custom dumper
	v := map[string]int{"a": 1}
	d := godump.NewDumper(
		godump.WithMaxDepth(10),
		godump.WithWriter(os.Stdout),
	)
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
