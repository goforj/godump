//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithNoColor disables colorized output for the dumper.

	// Example: disable colors
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithNoColor())
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
