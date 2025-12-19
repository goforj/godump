//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/godump"
)

func main() {
	// DumpHTML dumps the values as HTML with colorized output.

	// Example: dump HTML with a custom dumper
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
