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
<!-- test-count:embed:start -->
    <img src="https://img.shields.io/badge/tests-126-brightgreen" alt="Tests">
<!-- test-count:embed:end -->
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


<p align="center">
<strong>godump.Diff(a,b) Output Example</strong><br>
  <img src="docs/demo-diff.png" alt="Diff output example">
</p>

## Feature Comparison: `godump` vs `go-spew` vs `pp`

| **Feature**                                                            | **godump** | **go-spew** | **pp** |
|-----------------------------------------------------------------------:|:----------:|:-----------:|:------:|
| **Zero dependencies**                                                   | ✓          | -           | -      |
| **Colorized terminal output**                                           | ✓          | ✓           | ✓      |
| **HTML output**                                                         | ✓          | -           | -      |
| **JSON output helpers** (`DumpJSON`, `DumpJSONStr`)                     | ✓          | -           | -      |
| **Diff output helpers** (`Diff`, `DiffStr`)                             | ✓          | -           | -      |
| **Diff HTML output** (`DiffHTML`)                                       | ✓          | -           | -      |
| **Dump to `io.Writer`**                                                 | ✓          | ✓           | ✓      |
| **Shows file + line number of dump call**                               | ✓          | -           | -      |
| **Cyclic reference detection**                                          | ✓          | ~           | -      |
| **Handles unexported struct fields**                                    | ✓          | ✓           | ✓      |
| **Visibility markers** (`+` / `-`)                                      | ✓          | -           | -      |
| **Max depth control**                                                   | ✓          | -           | -      |
| **Max items (slice/map truncation)**                                    | ✓          | -           | -      |
| **Max string length truncation**                                        | ✓          | -           | -      |
| **Dump & Die** (`dd()` equivalent)                                      | ✓          | -           | -      |
| **Control character escaping**                                          | ✓          | ~           | ~      |
| **Supports structs, maps, slices, pointers, interfaces**                | ✓          | ✓           | ✓      |
| **Pretty type name rendering** (`#package.Type`)                        | ✓          | -           | -      |
| **Builder-style configuration API**                                     | ✓          | -           | -      |
| **Test-friendly string output** (`DumpStr`, `DiffStr`, `DumpJSONStr`) | ✓          | ✓           | ✓      |
| **HTML / Web UI debugging support**                                     | ✓          | -           | -      |

If you'd like to suggest improvements or additional comparisons, feel free to open an issue or PR.

## Installation

```bash
go get github.com/goforj/godump
````

## Basic Usage

<p> <a href="./examples/basic/main.go"><strong>View Full Runnable Example →</strong></a> </p>

```go
type User struct { Name string }
godump.Dump(User{Name: "Alice"})
// #main.User {
//    +Name => "Alice" #string
// }	
```

## Extended Usage (Snippets)

```go
godump.DumpStr(v)     // return as string
godump.DumpHTML(v)    // return HTML output
godump.DumpJSON(v)    // print JSON directly
godump.Fdump(w, v)    // write to io.Writer
godump.Dd(v)          // dump + exit
godump.Diff(a, b)     // diff two values
godump.DiffStr(a, b)  // diff two values as string
godump.DiffHTML(a, b) // diff two values as HTML
````

## Diff Usage

<p> <a href="./examples/diff/main.go"><strong>View Diff Example →</strong></a> </p>

```go
type User struct {
    Name string
}
before := User{Name: "Alice"}
after := User{Name: "Bob"}
godump.Diff(before, after)
//   #main.User {
// -   +Name => "Alice" #string
// +   +Name => "Bob" #string
//   }
```

<p> <a href="./examples/diffextended/main.go"><strong>View Diff Extended Example →</strong></a> </p>

## Builder Options Usage

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
    godump.WithNoColor(),              // default: false
).Dump(v)
```

## Contributing

Ensure that all tests pass, and you run ./docs/generate.sh to update the API index in the README before submitting a PR.

Ensure all public functions have documentation blocks with examples, as these are used to generate runnable examples and the API index.

## Runnable Examples Directory

Every function has a corresponding runnable example under [`./examples`](./examples).

These examples are **generated directly from the documentation blocks** of each function, ensuring the docs and code never drift. These are the same examples you see here in the README and GoDoc.

An automated test executes **every example** to verify it builds and runs successfully.

This guarantees all examples are valid, up-to-date, and remain functional as the API evolves.

<details>
<summary><strong>📘 How to Read the Output</strong></summary>

<br>

`godump` output is designed for clarity and traceability. Here's how to interpret its structure:

### Location Header

```go
<#dump // main.go:26
````

* The first line shows the **file and line number** where `godump.Dump()` was invoked.
* Helpful for finding where the dump happened during debugging.

### Type Names

```go
#main.User
```

* Fully qualified struct name with its package path.

### Visibility Markers

```go
  +Name => "Alice"
  -secret  => "..."
```

* `+` → Exported (public) field
* `-` → Unexported (private) field (accessed reflectively)

### Cyclic References

If a pointer has already been printed:

```go
↩︎ &1
```

* Prevents infinite loops in circular structures
* References point back to earlier object instances

### Slices and Maps

```go
  0 => "value"
  a => 1
```

* Array/slice indices and map keys are shown with `=>` formatting and indentation
* Slices and maps are truncated if `maxItems` is exceeded

### Escaped Characters

```go
"Line1\nLine2\tDone"
```

* Control characters like `\n`, `\t`, `\r`, etc. are safely escaped
* Strings are truncated after `maxStringLen` runes

### Supported Types

* ✅ Structs (exported & unexported)
* ✅ Pointers, interfaces
* ✅ Maps, slices, arrays
* ✅ Channels, functions
* ✅ time.Time (nicely formatted)

</details>

<!-- api:embed:start -->

## API Index

| Group | Functions |
|------:|-----------|
| **Builder** | [NewDumper](#newdumper) |
| **Debug** | [Dd](#dd) |
| **Diff** | [Diff](#diff) [DiffHTML](#diffhtml) [DiffStr](#diffstr) |
| **Dump** | [Dump](#dump) [DumpStr](#dumpstr) [Fdump](#fdump) |
| **HTML** | [DumpHTML](#dumphtml) |
| **JSON** | [DumpJSON](#dumpjson) [DumpJSONStr](#dumpjsonstr) |
| **Options** | [WithDisableStringer](#withdisablestringer) [WithMaxDepth](#withmaxdepth) [WithMaxItems](#withmaxitems) [WithMaxStringLen](#withmaxstringlen) [WithNoColor](#withnocolor) [WithSkipStackFrames](#withskipstackframes) [WithWriter](#withwriter) |


## Builder

### <a id="newdumper"></a>NewDumper

NewDumper creates a new Dumper with the given options applied.
Defaults are used for any setting not overridden.

```go
v := map[string]int{"a": 1}
d := godump.NewDumper(
	godump.WithMaxDepth(10),
	godump.WithWriter(os.Stdout),
)
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

## Debug

### <a id="dd"></a>Dd

Dd is a debug function that prints the values and exits the program.

_Example: dump and exit_

```go
v := map[string]int{"a": 1}
godump.Dd(v)
// #map[string]int {
//   a => 1 #int
// }
```

_Example: dump and exit with a custom dumper_

```go
d := godump.NewDumper()
v := map[string]int{"a": 1}
d.Dd(v)
// #map[string]int {
//   a => 1 #int
// }
```

## Diff

### <a id="diff"></a>Diff

Diff prints a diff between two values to stdout.

_Example: print diff_

```go
a := map[string]int{"a": 1}
b := map[string]int{"a": 2}
godump.Diff(a, b)
// <#diff // path:line
// - #map[string]int {
// -   a => 1 #int
// - }
// + #map[string]int {
// +   a => 2 #int
// + }
```

_Example: print diff with a custom dumper_

```go
d := godump.NewDumper()
a := map[string]int{"a": 1}
b := map[string]int{"a": 2}
d.Diff(a, b)
```

### <a id="diffhtml"></a>DiffHTML

DiffHTML returns an HTML diff between two values.

_Example: HTML diff_

```go
a := map[string]int{"a": 1}
b := map[string]int{"a": 2}
html := godump.DiffHTML(a, b)
_ = html
// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
// <span style="color:#999"><#diff // path:line</span>
// <span style="background-color:#221010; display:block; width:100%;">- <span style="color:#999">#map[string]int</span> {</span>
// <span style="background-color:#221010; display:block; width:100%;">-   <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">1</span><span style="color:#999"> #int</span></span>
// <span style="background-color:#221010; display:block; width:100%;">- }</span>
// <span style="background-color:#102216; display:block; width:100%;">+ <span style="color:#999">#map[string]int</span> {</span>
// <span style="background-color:#102216; display:block; width:100%;">+   <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">2</span><span style="color:#999"> #int</span></span>
// <span style="background-color:#102216; display:block; width:100%;">+ }</span>
// </pre></div>
```

_Example: HTML diff with a custom dumper_

```go
d := godump.NewDumper()
a := map[string]int{"a": 1}
b := map[string]int{"a": 2}
html := d.DiffHTML(a, b)
_ = html
// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
// <span style="color:#999"><#diff // path:line</span>
// <span style="background-color:#221010; display:block; width:100%;">- <span style="color:#999">#map[string]int</span> {</span>
// <span style="background-color:#221010; display:block; width:100%;">-   <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">1</span><span style="color:#999"> #int</span></span>
// <span style="background-color:#221010; display:block; width:100%;">- }</span>
// <span style="background-color:#102216; display:block; width:100%;">+ <span style="color:#999">#map[string]int</span> {</span>
// <span style="background-color:#102216; display:block; width:100%;">+   <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">2</span><span style="color:#999"> #int</span></span>
// <span style="background-color:#102216; display:block; width:100%;">+ }</span>
// </pre></div>
```

### <a id="diffstr"></a>DiffStr

DiffStr returns a string diff between two values.

_Example: diff string_

```go
a := map[string]int{"a": 1}
b := map[string]int{"a": 2}
out := godump.DiffStr(a, b)
_ = out
// <#diff // path:line
// - #map[string]int {
// -   a => 1 #int
// - }
// + #map[string]int {
// +   a => 2 #int
// + }
```

_Example: diff string with a custom dumper_

```go
d := godump.NewDumper()
a := map[string]int{"a": 1}
b := map[string]int{"a": 2}
out := d.DiffStr(a, b)
_ = out
// <#diff // path:line
// - #map[string]int {
// -   a => 1 #int
// - }
// + #map[string]int {
// +   a => 2 #int
// + }
```

## Dump

### <a id="dump"></a>Dump

Dump prints the values to stdout with colorized output.

_Example: print to stdout_

```go
v := map[string]int{"a": 1}
godump.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

_Example: print with a custom dumper_

```go
d := godump.NewDumper()
v := map[string]int{"a": 1}
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

### <a id="dumpstr"></a>DumpStr

DumpStr returns a string representation of the values with colorized output.

_Example: get a string dump_

```go
v := map[string]int{"a": 1}
out := godump.DumpStr(v)
godump.Dump(out)
// "#map[string]int {\n  a => 1 #int\n}" #string
```

_Example: get a string dump with a custom dumper_

```go
d := godump.NewDumper()
v := map[string]int{"a": 1}
out := d.DumpStr(v)
_ = out
```

### <a id="fdump"></a>Fdump

Fdump writes the formatted dump of values to the given io.Writer.

```go
var b strings.Builder
v := map[string]int{"a": 1}
godump.Fdump(&b, v)
// #map[string]int {
//   a => 1 #int
// }
```

## HTML

### <a id="dumphtml"></a>DumpHTML

DumpHTML dumps the values as HTML with colorized output.

_Example: dump HTML_

```go
v := map[string]int{"a": 1}
html := godump.DumpHTML(v)
_ = html
// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
// <span style="color:#999"><#dump // path:line</span>
// <span style="color:#999">#map[string]int</span> {
//  <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">1</span><span style="color:#999"> #int</span>
// }
// </pre></div>
```

_Example: dump HTML with a custom dumper_

```go
v := map[string]int{"a": 1}
html := d.DumpHTML(v)
_ = html
// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
// <span style="color:#999"><#dump // path:line</span>
// <span style="color:#999">#map[string]int</span> {
//  <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">1</span><span style="color:#999"> #int</span>
// }
// </pre></div>
```

## JSON

### <a id="dumpjson"></a>DumpJSON

DumpJSON prints a pretty-printed JSON string to the configured writer.

_Example: print JSON_

```go
v := map[string]int{"a": 1}
d.DumpJSON(v)
// {
//   "a": 1
// }
```

_Example: print JSON_

```go
v := map[string]int{"a": 1}
godump.DumpJSON(v)
// {
//   "a": 1
// }
```

### <a id="dumpjsonstr"></a>DumpJSONStr

DumpJSONStr pretty-prints values as JSON and returns it as a string.

_Example: dump JSON string_

```go
v := map[string]int{"a": 1}
out := d.DumpJSONStr(v)
_ = out
// {"a":1}
```

_Example: JSON string_

```go
v := map[string]int{"a": 1}
out := godump.DumpJSONStr(v)
_ = out
// {"a":1}
```

## Options

### <a id="withdisablestringer"></a>WithDisableStringer

WithDisableStringer disables using the fmt.Stringer output.
When enabled, the underlying type is rendered instead of String().

```go
v := map[string]int{"a": 1}
d := godump.NewDumper(godump.WithDisableStringer(true))
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

### <a id="withmaxdepth"></a>WithMaxDepth

WithMaxDepth limits how deep the structure will be dumped.
Param n must be 0 or greater or this will be ignored, and default MaxDepth will be 15.

```go
v := map[string]int{"a": 1}
d := godump.NewDumper(godump.WithMaxDepth(3))
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

### <a id="withmaxitems"></a>WithMaxItems

WithMaxItems limits how many items from an array, slice, or map can be printed.
Param n must be 0 or greater or this will be ignored, and default MaxItems will be 100.

```go
v := []int{1, 2, 3}
d := godump.NewDumper(godump.WithMaxItems(2))
d.Dump(v)
// #[]int [
//   0 => 1 #int
//   1 => 2 #int
//   ... (truncated)
// ]
```

### <a id="withmaxstringlen"></a>WithMaxStringLen

WithMaxStringLen limits how long printed strings can be.
Param n must be 0 or greater or this will be ignored, and default MaxStringLen will be 100000.

```go
v := "hello world"
d := godump.NewDumper(godump.WithMaxStringLen(5))
d.Dump(v)
// "hello…" #string
```

### <a id="withnocolor"></a>WithNoColor

WithNoColor disables colorized output for the dumper.

```go
v := map[string]int{"a": 1}
d := godump.NewDumper(godump.WithNoColor())
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

### <a id="withskipstackframes"></a>WithSkipStackFrames

WithSkipStackFrames skips additional stack frames for header reporting.
This is useful when godump is wrapped and the actual call site is deeper.

```go
v := map[string]int{"a": 1}
d := godump.NewDumper(godump.WithSkipStackFrames(2))
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```

### <a id="withwriter"></a>WithWriter

WithWriter routes output to the provided writer.

```go
var b strings.Builder
v := map[string]int{"a": 1}
d := godump.NewDumper(godump.WithWriter(&b))
d.Dump(v)
// #map[string]int {
//   a => 1 #int
// }
```
<!-- api:embed:end -->
