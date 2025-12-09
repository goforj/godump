<p align="center">
  <img src="docs/godump.png" width="600" alt="godump logo – Go pretty printer and Laravel-style dump/dd debugging tool">
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
  <code>godump</code> is a developer-friendly, zero-dependency debug dumper for Go. It provides pretty, colorized terminal output of your structs, slices, maps, and more - complete with cyclic reference detection and control character escaping.
    Inspired by Symfony's VarDumper which is used in Laravel's tools like <code>dump()</code> and <code>dd()</code>.
</p>

<p align="center">
<strong>Terminal Output Example (Kitchen Sink)</strong><br>
  <img src="docs/demo-terminal-2.png" alt="Terminal output example kitchen sink">
</p>

<p align="center">
<strong>HTML Output Example</strong><br>
  <img src="docs/demo-html.png" alt="HTML output example">
</p>

<!-- TOC start (generated with https://github.com/derlin/bitdowntoc) -->

## Table of Contents

- [Why `godump`?](#why-godump)
  * [What `godump` provides](#what-godump-provides)
- [📊 Comparison: `godump` vs `go-spew` vs `pp`](#-comparison-godump-vs-go-spew-vs-pp)
- [📦 Installation](#-installation)
- [🚀 Basic Usage](#-basic-usage)
- [🧰 Extended Usage (Snippets)](#-extended-usage-snippets)
- [🏗️ Builder Options Usage](#-builder-options-usage)
- [📁 Full Examples Directory](#-full-examples-directory)
- [🧩 License](#-license)

<!-- TOC end -->

## Why `godump`?

Debugging Go data shouldn’t feel like deciphering noise.

Traditional tools (`fmt.Printf`, `spew`, `pp`) often fall short:

- Hard to read deeply nested structs
- No visibility markers (exported vs unexported)
- No file/line context to know *where* a dump came from
- No cycle detection (infinite recursion danger)
- No HTML output for browser-based debugging
- No `dd()`-style dump-and-stop helper

`godump` gives you a clean, Laravel/Symfony-style debugging experience designed for **clarity**, **traceability**, and **zero configuration**:

### What `godump` provides

- 🧭 **File + line tracing** for every dump
- 🔐 **Visibility markers** (`+` exported, `-` unexported)
- 🔄 **Cycle-safe reference tracking**
- 🧠 **Readable, structured indentation**
- 🎨 **Colorized terminal output** or **full HTML rendering**
- 💥 **`Dd()` dump-and-exit** for emergency debugging
- 🪄 **Control character escaping** (`\n`, `\t`, etc.)
- 🧰 **Zero dependencies**, minimal API surface, and intuitive defaults

## 📊 Comparison: `godump` vs `go-spew` vs `pp`

| Feature                                                                | **godump**  |   **go-spew**    |    **pp**     |
|------------------------------------------------------------------------|:-----------:|:----------------:|:-------------:|
| **Zero dependencies**                                                  |      ✅      |        ❌         |       ❌       |
| **Colorized terminal output**                                          |   ✅ Rich    |     ✅ Basic      |    ✅ Good     |
| **HTML output**                                                        |      ✅      |        ❌         |       ❌       |
| **JSON output helpers** (`DumpJSON`, `DumpJSONStr`)                    |      ✅      |        ❌         |       ❌       |
| **Dump to `io.Writer`**                                                |      ✅      |        ✅         |       ✅       |
| **Shows file + line number of dump call**                              |      ✅      |        ❌         |       ❌       |
| **Cyclic reference detection**                                         | ✅ Advanced  |    ⚠️ Partial    |       ❌       |
| **Handles unexported struct fields**                                   |      ✅      |        ✅         |       ✅       |
| **Visibility markers (`+` / `-`)**                                     |      ✅      |        ❌         |       ❌       |
| **Max depth control**                                                  |      ✅      |        ❌         |       ❌       |
| **Max items (slice/map truncation)**                                   |      ✅      |        ❌         |       ❌       |
| **Max string length truncation**                                       |      ✅      |        ❌         |       ❌       |
| **Dump & Die (`dd()` equivalent)**                                     |      ✅      |        ❌         |       ❌       |
| **Control character escaping**                                         |      ✅      |    ⚠️ Partial    |  ⚠️ Partial   |
| **Supports structs, maps, slices, pointers, interfaces**               |      ✅      |        ✅         |       ✅       |
| **Pretty type name rendering (`#package.Type`)**                       |      ✅      |        ❌         |       ❌       |
| **Builder-style configuration API**                                    |      ✅      |        ❌         |       ❌       |
| **Test-friendly string output** (`DumpStr`, `DumpHTML`, `DumpJSONStr`) |      ✅      |   ✅ (`Sdump`)    | ✅ (`Sprintf`) |
| **HTML / Web UI debugging support**                                    |      ✅      |        ❌         |       ❌       |
| **Output style**                                                       | Human-first | Reflection-first |  Color-first  |

If you'd like to suggest improvements or additional comparisons, feel free to open an issue or PR.

## 📦 Installation

```bash
go get github.com/goforj/godump
````

## 🚀 Basic Usage

<p> <a href="./examples/basic/main.go"><strong>View Full Runnable Example →</strong></a> </p>

```go
type User struct { Name string }
godump.Dump(User{Name: "Alice"})
// #main.User {
//    +Name => "Alice" #string
// }	
```

## 🧰 Extended Usage (Snippets)

```go
godump.DumpStr(v)  // return as string
godump.DumpHTML(v) // return HTML output
godump.DumpJSON(v) // print JSON directly
godump.Fdump(w, v) // write to io.Writer
godump.Dd(v)       // dump + exit
````

## 🏗️ Builder Options Usage

`godump` aims for simple usage with sensible defaults out of the box, but also provides a flexible builder-style API for customization.

If you want to heavily customize the dumper behavior, you can create a `Dumper` instance with specific options:

<p> <a href="./examples/builder/main.go"><strong>View Full Runnable Example →</strong></a> </p>

```go
godump.NewDumper(
    godump.WithMaxDepth(15),           // default: 15
    godump.WithMaxItems(100),          // default: 100
    godump.WithMaxStringLen(100000),   // default: 100000
    godump.WithWriter(os.Stdout),      // default: os.Stdout
    godump.WithSkipStackFrames(10),    // default: 10
    godump.WithDisableStringer(false), // default: false
).Dump(v)
```

## 📁 Full Examples Directory

All runnable examples can be found under [`./examples`](./examples):

- **Basic usage** → [`examples/basic/main.go`](./examples/basic/main.go)
- **Extended usage** → [`examples/extended/main.go`](./examples/extended/main.go)
- **Kitchen sink** → [`examples/kitchensink/main.go`](./examples/kitchensink/main.go)
- **Builder API** → [`examples/builder/main.go`](./examples/builder/main.go)

<details>
<summary><strong>📘 How to Read the Output</strong></summary>

<br>

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
  +Name => "Alice"
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

</details>

## 🧩 License

MIT © [goforj](https://github.com/goforj)