//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// WithRedactTags redacts struct fields whose selected tag values equal "-".
	// With no keys, or with "*", it checks all tag keys. Calls accumulate selected keys.
	// Matching is case-sensitive and does not treat "-,omitempty" as "-".
	// It applies as text, HTML, and diff output traverse nested and embedded structs.
	// JSON helpers retain encoding/json behavior. Structs rendered through fmt.Stringer
	// are not traversed; use WithDisableStringer(true) to inspect their fields instead.

	// Example: redact fields omitted from JSON
	type User struct {
		Name       string
		Credential string `json:"-"`
	}
	d := godump.NewDumper(godump.WithRedactTags("json"), godump.WithoutHeader())
	d.Dump(User{Name: "Alice", Credential: "secret"})
	// #godump.User {
	//   +Name       => "Alice" #string
	//   +Credential => <redacted> #string
	// }
}
