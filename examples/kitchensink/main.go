//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

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

	// Basic pretty-print
	godump.Dump(user)
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }

	// As string
	strOut := godump.DumpStr(user)
	fmt.Println("DumpStr:", strOut)
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }

	// As HTML
	htmlOut := godump.DumpHTML(user)
	fmt.Println("DumpHTML:", htmlOut)
	// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
	// <span style="color:#999"><#dump // examples/kitchensink/main.go:40</span>
	// <span style="color:#999">#main.User</span> {
	//  <span style="color:#ffb400">+</span>Name    => <span style="color:#ffb400">"</span><span style="color:#80ff80">Alice</span><span style="color:#ffb400">"</span><span style="color:#999"> #string</span>
	//  <span style="color:#ffb400">+</span>Profile => <span style="color:#999">#main.Profile</span> {
	//    <span style="color:#ffb400">+</span>Age   => <span style="color:#40c0ff">30</span><span style="color:#999"> #int</span>
	//    <span style="color:#ffb400">+</span>Email => <span style="color:#ffb400">"</span><span style="color:#80ff80">alice@example.com</span><span style="color:#ffb400">"</span><span style="color:#999"> #string</span>
	//  }
	// }

	// As JSON
	godump.DumpJSON(user)
	// {
	//  "Name": "Alice",
	//  "Profile": {
	//    "Age": 30,
	//    "Email": "alice@example.com"
	//  }
	// }

	// Write to any io.Writer
	godump.Fdump(os.Stderr, user)
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }

	// Dump and exit
	godump.Dd(user)
	// <#dump // examples/kitchensink/main.go:47
	// #main.User {
	//  +Name    => "Alice" #string
	//  +Profile => #main.Profile {
	//    +Age   => 30 #int
	//    +Email => "alice@example.com" #string
	//  }
	// }
}
