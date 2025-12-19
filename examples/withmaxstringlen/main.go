//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithMaxStringLen limits how long printed strings can be.
	// Param n must be 0 or greater or this will be ignored, and default MaxStringLen will be 100000.

	// Example: limit string length
	v := "hello world"
	d := godump.NewDumper(godump.WithMaxStringLen(5))
	d.Dump(v)
	// "hello…" #string
}
