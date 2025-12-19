//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// DiffStr returns a string diff between two values.

	// Example: diff string with a custom dumper
	d := godump.NewDumper()
	a := map[string]int{"a": 1}
	b := map[string]int{"a": 2}
	out := d.DiffStr(a, b)
	_ = out
	// <#diff // path:line
	// - #map[string]int {
	// -   a => 1 #int
	// - }
	// + #map[string]int {
	// +   a => 2 #int
	// + }
}
