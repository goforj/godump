//go:build ignore
// +build ignore

package main

import "github.com/goforj/godump"

func main() {
	// DiffHTML returns an HTML diff between two values.

	// Example: HTML diff with a custom dumper
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
