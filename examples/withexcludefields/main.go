//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithExcludeFields omits struct fields that match the provided names.

	// Example: exclude fields
	// Default: none
	type User struct {
		ID       int
		Email    string
		Password string
	}
	d := godump.NewDumper(
		godump.WithExcludeFields("Password"),
	)
	d.Dump(User{ID: 1, Email: "user@example.com", Password: "secret"})
	// #godump.User {
	//   +ID    => 1 #int
	//   +Email => "user@example.com" #string
	// }
}
