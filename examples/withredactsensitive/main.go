//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithRedactSensitive enables default redaction for common sensitive fields.

	// Example: redact common sensitive fields
	// Default: disabled
	type User struct {
		Password string
		Token    string
	}
	d := godump.NewDumper(
		godump.WithRedactSensitive(),
	)
	d.Dump(User{Password: "secret", Token: "abc"})
	// #godump.User {
	//   +Password => <redacted> #string
	//   +Token    => <redacted> #string
	// }
}
