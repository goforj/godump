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

// main demonstrates basic dumping.
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
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }
}
