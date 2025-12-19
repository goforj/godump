//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithSkipStackFrames skips additional stack frames for header reporting.
	// This is useful when godump is wrapped and the actual call site is deeper.

	// Example: skip wrapper frames
	// Default: 0
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithSkipStackFrames(2))
	d.Dump(v)
	// <#dump // ../../../../usr/local/go/src/runtime/asm_arm64.s:1223
	// #map[string]int {
	//   a => 1 #int
	// }
}
