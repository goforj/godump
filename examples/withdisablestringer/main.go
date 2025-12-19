//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithDisableStringer disables using the fmt.Stringer output.
	// When enabled, the underlying type is rendered instead of String().

	// Example: show raw types
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithDisableStringer(true))
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
