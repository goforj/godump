//go:build ignore
// +build ignore

package main

import (
	"github.com/goforj/godump"
	"time"
)

func main() {
	// WithDisableStringer disables using the fmt.Stringer output.
	// When enabled, the underlying type is rendered instead of String().

	// Example: show raw types
	v := time.Duration(3)
	d := godump.NewDumper(godump.WithDisableStringer(true))
	d.Dump(v)
	// 3 #time.Duration
}
