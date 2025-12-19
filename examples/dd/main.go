//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// Dd is a debug function that prints the values and exits the program.

	// Example: dump and exit with a custom dumper
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	d.Dd(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
