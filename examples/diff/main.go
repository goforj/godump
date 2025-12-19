//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// Diff prints a diff between two values to the configured writer.

	// Example: print diff with a custom dumper
	d := godump.NewDumper()
	a := map[string]int{"a": 1}
	b := map[string]int{"a": 2}
	d.Diff(a, b)
	// <#diff // path:line
	// - #map[string]int {
	// -   a => 1 #int
	// - }
	// + #map[string]int {
	// +   a => 2 #int
	// + }
}
