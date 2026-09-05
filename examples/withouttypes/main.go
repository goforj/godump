//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithoutTypes hides type information when the structure will be dumped.

	// Example: hide type information
	v := map[string]map[string]int{"a": {"b": 1}}
	d := godump.NewDumper(godump.WithoutTypes())
	d.Dump(v)
	// {
	//   a => {
	//     b => 1
	//   }
	// }
}
