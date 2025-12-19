package godump_test

import (
	"fmt"
	"github.com/goforj/godump"
	"os"
	"strings"
	"time"
)

func ExampleDd() {
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	d.Dd(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}

func ExampleDiff() {
	d := godump.NewDumper()
	a := map[string]int{"a": 1}
	b := map[string]int{"a": 2}
	d.Diff(a, b)
	// <#diff // path:line
	// - #map[string]int {
	// -   a => 1 #int
	// - }
	// + #map[string]int {
	// +   a => 2 #int
	// + }
}

func ExampleDiffHTML() {
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
}

func ExampleDiffStr() {
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
}

func ExampleDump() {
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}

func ExampleDumpHTML() {
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	html := d.DumpHTML(v)
	_ = html
	fmt.Println(html)
	// <div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">
	// <span style="color:#999"><#dump // examples/dumphtml/main.go:17</span>
	// <span style="color:#999">#map[string]int</span> {
	//   <span style="color:#d087d0">a</span> => <span style="color:#40c0ff">1</span><span style="color:#999"> #int</span>
	// }
	// </pre></div>
}

func ExampleDumpJSON() {
	v := map[string]int{"a": 1}
	godump.DumpJSON(v)
	// {
	//   "a": 1
	// }
}

func ExampleDumpJSONStr() {
	v := map[string]int{"a": 1}
	out := godump.DumpJSONStr(v)
	_ = out
	// {"a":1}
}

func ExampleDumpStr() {
	d := godump.NewDumper()
	v := map[string]int{"a": 1}
	out := d.DumpStr(v)
	_ = out
	// "#map[string]int {\n  a => 1 #int\n}" #string
}

func ExampleFdump() {
	var b strings.Builder
	v := map[string]int{"a": 1}
	godump.Fdump(&b, v)
	// outputs to strings builder
}

func ExampleNewDumper() {
	v := map[string]int{"a": 1}
	d := godump.NewDumper(
		godump.WithMaxDepth(10),
		godump.WithWriter(os.Stdout),
	)
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}

func ExampleWithDisableStringer() {
	// Default: false
	v := time.Duration(3)
	d := godump.NewDumper(godump.WithDisableStringer(true))
	d.Dump(v)
	// 3 #time.Duration
}

func ExampleWithMaxDepth() {
	// Default: 15
	v := map[string]map[string]int{"a": {"b": 1}}
	d := godump.NewDumper(godump.WithMaxDepth(1))
	d.Dump(v)
	// #map[string]int {
	//   a => #map[string]int {
	//     b => ... (max depth)
	//   }
	// }
}

func ExampleWithMaxItems() {
	// Default: 100
	v := []int{1, 2, 3}
	d := godump.NewDumper(godump.WithMaxItems(2))
	d.Dump(v)
	// #[]int [
	//   0 => 1 #int
	//   1 => 2 #int
	//   ... (truncated)
	// ]
}

func ExampleWithMaxStringLen() {
	// Default: 100000
	v := "hello world"
	d := godump.NewDumper(godump.WithMaxStringLen(5))
	d.Dump(v)
	// "hello…" #string
}

func ExampleWithSkipStackFrames() {
	// Default: 0
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithSkipStackFrames(2))
	d.Dump(v)
	// <#dump // ../../../../usr/local/go/src/runtime/asm_arm64.s:1223
	// #map[string]int {
	//   a => 1 #int
	// }
}

func ExampleWithWriter() {
	// Default: stdout
	var b strings.Builder
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithWriter(&b))
	d.Dump(v)
	// #map[string]int {
	//   a => 1 #int
	// }
}

func ExampleWithoutColor() {
	// Default: false
	v := map[string]int{"a": 1}
	d := godump.NewDumper(godump.WithoutColor())
	d.Dump(v)
	// (prints without color)
	// #map[string]int {
	//   a => 1 #int
	// }
}
