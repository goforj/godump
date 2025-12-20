//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/godump"
	"os"
)

type Profile struct {
	Age   int
	Email string
}

type User struct {
	Name    string
	Profile Profile
}

// main demonstrates extended output formats.
func main() {
	user := User{
		Name: "Alice",
		Profile: Profile{
			Age:   30,
			Email: "alice@example.com",
		},
	}
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }

	// Pretty-print to stdout
	godump.Dump(user)
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }

	// Get dump as string
	output := godump.DumpStr(user)
	fmt.Println("str", output)

	// HTML for web UI output
	html := godump.DumpHTML(user)
	fmt.Println("html", html)
	// (html output with syntax highlighting)

	// Print JSON directly to stdout
	godump.DumpJSON(user)
	// {
	//  "Name": "Alice",
	//  "Profile": {
	//    "Age": 30,
	//    "Email": "alice@example.com"
	//  }
	// }

	// Write to any io.Writer (e.g. file, buffer, logger)
	godump.Fdump(os.Stderr, user)
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }

	// Dump and exit
	godump.Dd(user) // this will print the dump and exit the program
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }
}
