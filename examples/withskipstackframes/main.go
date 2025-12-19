//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithSkipStackFrames skips additional stack frames for header reporting.
	// This is useful when godump is wrapped and the actual call site is deeper.

	// Example: skip wrapper frames
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithSkipStackFrames(2))
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}
