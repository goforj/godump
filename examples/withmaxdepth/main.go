//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithMaxDepth limits how deep the structure will be dumped.
	// Param n must be 0 or greater or this will be ignored, and default MaxDepth will be 15.

	// Example: limit depth
	v := map[string]map[string]int{"a": {"b": 1}}
	d := godump.NewDumper(godump.WithMaxDepth(1))
	d.Dump(v)
	// #map[string]int {
	//   a => #map[string]int {
	//     b => ... (max depth)
	//   }
	// }
}
