//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithoutHeader disables printing the source location header.

	// Example: disable header
	// Default: false
	d := godump.NewDumper(godump.WithoutHeader())
	d.Dump("hello")
	// "hello" #string
}
