//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithNoHeader disables printing the source location header.

	// Example: disable header
	// Default: false
	d := godump.NewDumper(godump.WithNoHeader())
	d.Dump("hello")
	// "hello" #string
}
