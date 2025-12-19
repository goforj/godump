//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithoutColor disables colorized output for the dumper.

	// Example: disable colors
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithoutColor())
	d.Dump(v)
	// (prints without color)
	// #map[string]int {
	//   a => 1 #int
	// }
}
