//go:build ignore
// +build ignore

package main

import (
	"github.com/goforj/godump"
)

type Profile struct {
	Age   int
	Email string
}

type User struct {
	Name    string
	Profile Profile
}

func main() {
	user := User{
		Name: "Alice",
		Profile: Profile{
			Age:   30,
			Email: "alice@example.com",
		},
	}

	// Pretty-print to stdout
	godump.Dump(user)
}
