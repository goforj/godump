//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithoutTypes omits type annotations while preserving value colors and structure.
	// It applies to text, HTML, and diff output. JSON output is unchanged.

	// Example: hide type information
	type User struct {
		Name string
	}
	d := godump.NewDumper(godump.WithoutTypes(), godump.WithoutHeader())
	d.Dump(User{Name: "Alice"})
	// {
	//   +Name => "Alice"
	// }
}
