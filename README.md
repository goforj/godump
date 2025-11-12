<p align="center">
  <img src="./assets/godump.png" width="600" alt="godump logo">
</p>

<p align="center">
    Pretty-print and debug Go structs with a Laravel-inspired developer experience.
</p>

<p align="center">
    <a href="https://pkg.go.dev/github.com/goforj/godump"><img src="https://pkg.go.dev/badge/github.com/goforj/godump.svg" alt="Go Reference"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
    <a href="https://github.com/goforj/godump/actions"><img src="https://github.com/goforj/godump/actions/workflows/test.yml/badge.svg" alt="Go Test"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/go-1.18+-blue?logo=go" alt="Go version"></a>
    <img src="https://img.shields.io/github/v/tag/goforj/godump?label=version&sort=semver" alt="Latest tag">
    <a href="https://goreportcard.com/report/github.com/goforj/godump"><img src="https://goreportcard.com/badge/github.com/goforj/godump" alt="Go Report Card"></a>
    <a href="https://codecov.io/gh/goforj/godump" ><img src="https://codecov.io/gh/goforj/godump/graph/badge.svg?token=ULUTXL03XC"/></a>
    <a href="https://github.com/avelino/awesome-go?tab=readme-ov-file#parsersencodersdecoders"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
</p>

<p align="center">
  <code>godump</code> is a developer-friendly, zero-dependency debug dumper for Go. It provides pretty, colorized terminal output of your structs, slices, maps, and more — complete with cyclic reference detection and control character escaping.
    Inspired by Symfony's VarDumper which is used in Laravel's tools like <code>dump()</code> and <code>dd()</code>.
</p>

<br>

<p align="center">
<strong>Terminal Output Example (Kitchen Sink)</strong><br>
  <img src="./assets/demo-terminal-2.png">
</p>

<p align="center">
<strong>HTML Output Example</strong><br>
  <img src="./assets/demo-html.png">
</p>

## ✨ Features

- 🧠 Struct field inspection with visibility markers (`+`, `-`)
- 🔄 Cycle-safe reference tracking
- 🎨 ANSI color or HTML output
- 🧪 Handles slices, maps, nested structs, pointers, time, etc.
- 🪄 Control character escaping (`\n`, `\t`, etc.)

## 📦 Installation

```bash
go get github.com/goforj/godump
````

## 🚀 Basic Usage

```go
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

	// Get dump as string
	output := godump.DumpStr(user)
	fmt.Println("str", output)

	// HTML for web UI output
	html := godump.DumpHTML(user)
	fmt.Println("html", html)

	// Print JSON directly to stdout
	godump.DumpJSON(user)

	// Write to any io.Writer (e.g. file, buffer, logger)
	godump.Fdump(os.Stderr, user)

	// Dump and exit
	godump.Dd(user) // this will print the dump and exit the program 
}
```

## 🧪 Example Output

```go
<#dump // main.go:26
#main.User
  +Name    => "Alice"
  +Profile => #main.Profile
    +Age   => 30
    +Email => "alice@example.com"
  }
}
```

## 🏗️ Builder Options Usage

```go
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

func main() {
	user := User{
		Name: "Alice",
		Profile: Profile{
			Age:   30,
			Email: "alice@example.com",
		},
	}

	// Custom Dumper with all options set explicitly
	d := godump.NewDumper(
		godump.WithMaxDepth(15),           // default: 15
		godump.WithMaxItems(100),          // default: 100
		godump.WithMaxStringLen(100000),   // default: 100000
		godump.WithWriter(os.Stdout),      // default: os.Stdout
		godump.WithSkipStackFrames(0),     // default: 0
		godump.WithDisableStringer(false), // default: false
		godump.WithHeader(true),           // default: true
	)

	// Use the custom dumper
	d.Dump(user)

	// Dump to string
	out := d.DumpStr(user)
	fmt.Printf("DumpStr output:\n%s\n", out)

	// Dump to HTML string
	html := d.DumpHTML(user)
	fmt.Printf("DumpHTML output:\n%s\n", html)

	// Dump JSON using the Dumper (returns string)
	jsonStr := d.DumpJSONStr(user)
	fmt.Printf("Dumper JSON string:\n%s\n", jsonStr)

	// Print JSON directly from the Dumper
	d.DumpJSON(user)

	// Dump to custom writer (e.g. a string builder)
	var sb strings.Builder
	custom := godump.NewDumper(godump.WithWriter(&sb))
	custom.Dump(user)
	fmt.Printf("Dump to string builder:\n%s\n", sb.String())
}
```

## 📘 How to Read the Output

`godump` output is designed for clarity and traceability. Here's how to interpret its structure:

### 🧭 Location Header

```go
<#dump // main.go:26
````

* The first line shows the **file and line number** where `godump.Dump()` was invoked.
* Helpful for finding where the dump happened during debugging.

### 🔎 Type Names

```go
#main.User
```

* Fully qualified struct name with its package path.

### 🔐 Visibility Markers

```go
  +Name    => "Alice"
  -secret  => "..."
```

* `+` → Exported (public) field
* `-` → Unexported (private) field (accessed reflectively)

### 🔄 Cyclic References

If a pointer has already been printed:

```go
↩︎ &1
```

* Prevents infinite loops in circular structures
* References point back to earlier object instances

### 🔢 Slices and Maps

```go
  0 => "value"
  a => 1
```

* Array/slice indices and map keys are shown with `=>` formatting and indentation
* Slices and maps are truncated if `maxItems` is exceeded

### 🔣 Escaped Characters

```go
"Line1\nLine2\tDone"
```

* Control characters like `\n`, `\t`, `\r`, etc. are safely escaped
* Strings are truncated after `maxStringLen` runes

### 🧩 Supported Types

* ✅ Structs (exported & unexported)
* ✅ Pointers, interfaces
* ✅ Maps, slices, arrays
* ✅ Channels, functions
* ✅ time.Time (nicely formatted)

## 🧩 License

MIT © [goforj](https://github.com/goforj)

## 📇 Author

Created by [Chris Miles](https://github.com/akkadius)  
Maintained as part of the [goforj](https://github.com/goforj) tooling ecosystem.
