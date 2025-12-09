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
	// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
	//<span style="color:#999"><#dump // examples/extended/main.go:36</span>
	//<span style="color:#999">#main.User</span> {
	//  <span style="color:#ffb400">+</span>Name    => <span style="color:#ffb400">"</span><span style="color:#80ff80">Alice</span><span style="color:#ffb400">"</span><span style="color:#999"> #string</span>
	//  <span style="color:#ffb400">+</span>Profile => <span style="color:#999">#main.Profile</span> {
	//    <span style="color:#ffb400">+</span>Age   => <span style="color:#40c0ff">30</span><span style="color:#999"> #int</span>
	//    <span style="color:#ffb400">+</span>Email => <span style="color:#ffb400">"</span><span style="color:#80ff80">alice@example.com</span><span style="color:#ffb400">"</span><span style="color:#999"> #string</span>
	//  }
	//}
	//</pre></div>

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
