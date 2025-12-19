//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// Dump prints the values to stdout with colorized output.

	// Example: print with a custom dumper
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
