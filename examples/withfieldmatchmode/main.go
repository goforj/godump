//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithFieldMatchMode sets how field names are matched for WithExcludeFields.

	// Example: use substring matching
	// Default: FieldMatchExact
	type User struct {
		UserID int
	}
	d := godump.NewDumper(
		godump.WithExcludeFields("id"),
		godump.WithFieldMatchMode(godump.FieldMatchContains),
	)
	d.Dump(User{UserID: 10})
	// #godump.User {
	// }
}
