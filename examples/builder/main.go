//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

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

// main demonstrates the builder-style API.
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
	//   +Name    => "Alice" #string
	//   +Profile => #main.Profile {
	//     +Age   => 30 #int
	//     +Email => "alice@example.com" #string
	//   }
	// }

	// Dump as string
	strOut := godump.DumpStr(user)
	fmt.Println("DumpStr:", strOut)
	// DumpStr output:
	// #main.User {
	//   +Name    => "Alice" #string
	//   +Profile => #main.Profile {
	//     +Age   => 30 #int
	//     +Email => "alice@example.com" #string
	//   }
	// }

	// Dump as HTML
	htmlOut := godump.DumpHTML(user)
	fmt.Println("DumpHTML:", htmlOut)
	// <pre class="godump">…formatted HTML output…</pre>

	// Dump JSON
	godump.DumpJSON(user)
	// {"Name":"Alice","Profile":{"Age":30,"Email":"alice@example.com"}}

	// Dump to any io.Writer
	godump.Fdump(os.Stderr, user)
	// (same output as Dump but written to stderr)

	// Dump and exit
	// godump.Dd(user)
	// (prints formatted dump then immediately exits)

	// -------------------------------------------------
	// Custom Dumper (Builder API)
	// -------------------------------------------------

	d := godump.NewDumper(
		godump.WithMaxDepth(15),
		godump.WithMaxItems(100),
		godump.WithMaxStringLen(100000),
		godump.WithWriter(os.Stdout),
		godump.WithSkipStackFrames(10),
		godump.WithDisableStringer(false),
	)

	// Using the custom dumper
	d.Dump(user)
	// #main.User {
	//   +Name    => "Alice" #string
	//   +Profile => #main.Profile {
	//     +Age   => 30 #int
	//     +Email => "alice@example.com" #string
	//   }
	// }

	// Dump to string using custom dumper
	out := d.DumpStr(user)
	fmt.Println("Custom DumpStr:\n", out)
	// Custom DumpStr:
	// #main.User {
	//   +Name    => "Alice" #string
	//   +Profile => #main.Profile {
	//     +Age   => 30 #int
	//     +Email => "alice@example.com" #string
	//   }
	// }

	// Dump to HTML
	html := d.DumpHTML(user)
	fmt.Println("Custom DumpHTML:\n", html)
	// <pre class="godump">…formatted HTML output…</pre>

	// JSON as string
	jsonStr := d.DumpJSONStr(user)
	fmt.Println("Custom JSON:\n", jsonStr)
	// {"Name":"Alice","Profile":{"Age":30,"Email":"alice@example.com"}}

	// Print JSON directly
	d.DumpJSON(user)
	// {"Name":"Alice","Profile":{"Age":30,"Email":"alice@example.com"}}

	// Dump to a strings.Builder
	var sb strings.Builder
	custom := godump.NewDumper(godump.WithWriter(&sb))
	custom.Dump(user)
	fmt.Println("Dump to strings.Builder:\n", sb.String())
	// #main.User {
	//   +Name    => "Alice" #string
	//   +Profile => #main.Profile {
	//     +Age   => 30 #int
	//     +Email => "alice@example.com" #string
	//   }
	// }
}
