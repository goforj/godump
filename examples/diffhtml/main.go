//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// DiffHTML returns an HTML diff between two values.

	// Example: HTML diff with a custom dumper
	d := godump.NewDumper()
	a := map[string]int{"a": 1}
	b := map[string]int{"a": 2}
	html := d.DiffHTML(a, b)
	_ = html
	// (html diff)
}
