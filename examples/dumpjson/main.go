//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// DumpJSON dumps the values as a pretty-printed JSON string.
	// If there is more than one value, they are dumped as a JSON array.
	// It returns an error string if marshaling fails.

	// Example: print JSON
	v := map[string]int{"a": 1}
	godump.DumpJSON(v)
	// {
	//   "a": 1
	// }
}
