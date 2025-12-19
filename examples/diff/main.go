//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

type Profile struct {
	Age   int
	Email string
}

type User struct {
	Name    string
	Profile Profile
}

// main demonstrates diff output.
func main() {
	before := User{
		Name: "Alice",
		Profile: Profile{
			Age:   30,
			Email: "alice@example.com",
		},
	}

	after := User{
		Name: "Bob",
		Profile: Profile{
			Age:   31,
			Email: "bob@example.com",
		},
	}

	godump.Diff(before, after)

	diff := godump.DiffStr(before, after)
	_ = diff
}
