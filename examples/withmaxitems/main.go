//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithMaxItems limits how many items from an array, slice, or map can be printed.
	// Param n must be 0 or greater or this will be ignored, and default MaxItems will be 100.

	// Example: limit items
	v := []int{1, 2, 3}
	d := godump.NewDumper(godump.WithMaxItems(2))
	d.Dump(v)
	// #[]int [
	//   0 => 1 #int
	//   1 => 2 #int
	//   ... (truncated)
	// ]
}
