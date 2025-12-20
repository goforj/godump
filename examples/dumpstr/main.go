//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// DumpStr returns a string representation of the values with colorized output.

	// Example: get a string dump with a custom dumper
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	out := d.DumpStr(v)
	_ = out
	// "#map[string]int {\n  a => 1 #int\n}" #string
}
