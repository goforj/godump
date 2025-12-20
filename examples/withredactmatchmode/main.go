//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithRedactMatchMode sets how field names are matched for WithRedactFields.

	// Example: use substring matching
	// Default: FieldMatchExact
	type User struct {
		APIKey string
	}
	d := godump.NewDumper(
		godump.WithRedactFields("key"),
		godump.WithRedactMatchMode(godump.FieldMatchContains),
	)
	d.Dump(User{APIKey: "abc"})
	// #godump.User {
	//   +APIKey => <redacted> #string
	// }
}
