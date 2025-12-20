//go:build ignore
// +build ignore

package main

import (
	"github.com/goforj/godump"
	"strings"
)

func main() {
	// WithWriter routes output to the provided writer.

	// Example: write to buffer
	// Default: stdout
	var b strings.Builder
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithWriter(&b))
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
