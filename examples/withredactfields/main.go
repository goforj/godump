//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithRedactFields replaces matching struct fields with a redacted placeholder.

	// Example: redact fields
	// Default: none
	type User struct {
		ID       int
		Password string
	}
	d := godump.NewDumper(
		godump.WithRedactFields("Password"),
	)
	d.Dump(User{ID: 1, Password: "secret"})
	// #godump.User {
	//   +ID       => 1 #int
	//   +Password => <redacted> #string
	// }
}
