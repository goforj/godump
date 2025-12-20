//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/godump"
)

func main() {
	// DumpHTML dumps the values as HTML with colorized output.

	// Example: dump HTML with a custom dumper
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	html := d.DumpHTML(v)
	_ = html
	fmt.Println(html)
	// (html output)
}
