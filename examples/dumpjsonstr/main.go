//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// DumpJSONStr dumps the values as a JSON string.

	// Example: JSON string
	v := map[string]int{"a": 1}
	out := godump.DumpJSONStr(v)
	_ = out
	// {"a":1}
}
