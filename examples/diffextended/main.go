//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/goforj/godump"
)

type FriendlyDuration time.Duration

// String renders the duration as HH:MM:SS.
func (fd FriendlyDuration) String() string {
	td := time.Duration(fd)
	return fmt.Sprintf("%02d:%02d:%02d", int(td.Hours()), int(td.Minutes())%60, int(td.Seconds())%60)
}

type IsZeroer interface {
	IsZero() bool
}

type Inner struct {
	ID    int
	Notes []string
	Blob  []byte
}

type Ref struct {
	Self *Ref
}

type Everything struct {
	String        string
	Bool          bool
	Int           int
	Float         float64
	Time          time.Time
	Duration      time.Duration
	Friendly      FriendlyDuration
	PtrString     *string
	PtrDuration   *time.Duration
	SliceInts     []int
	ArrayStrings  [2]string
	MapValues     map[string]int
	Nested        Inner
	NestedPtr     *Inner
	Interface     any
	InterfaceImpl IsZeroer
	Recursive     *Ref
	privateField  string
	privateStruct Inner
}

// makeEverything builds a populated sample struct.
func makeEverything(now time.Time, label string) Everything {
	ptrStr := "Hello " + label
	dur := time.Minute*20 + time.Second*10

	val := Everything{
		String:       "test " + label,
		Bool:         label == "v2",
		Int:          42,
		Float:        3.1415,
		Time:         now,
		Duration:     dur,
		Friendly:     FriendlyDuration(dur),
		PtrString:    &ptrStr,
		PtrDuration:  &dur,
		SliceInts:    []int{1, 2, 3},
		ArrayStrings: [2]string{"foo", "bar"},
		MapValues:    map[string]int{"a": 1, "b": 2},
		Nested: Inner{
			ID:    10,
			Notes: []string{"alpha", "beta"},
			Blob:  []byte(`{"kind":"test","ok":true}`),
		},
		NestedPtr: &Inner{
			ID:    99,
			Notes: []string{"x", "y"},
			Blob:  []byte(`{"msg":"hi","status":"cool"}`),
		},
		Interface:     map[string]bool{"ok": true},
		InterfaceImpl: time.Time{},
		Recursive:     &Ref{},
		privateField:  "should show",
		privateStruct: Inner{ID: 5, Notes: []string{"private"}},
	}
	val.Recursive.Self = val.Recursive

	return val
}

// main demonstrates diffing two complex structures.
func main() {
	now := time.Date(2025, 12, 18, 16, 34, 37, 0, time.FixedZone("CST", -6*60*60))

	before := makeEverything(now, "v1")
	after := makeEverything(now.Add(time.Hour), "v2")
	after.Int = 99
	after.Duration = time.Minute * 45
	after.Friendly = FriendlyDuration(after.Duration)
	after.SliceInts = []int{1, 3, 4}
	after.MapValues["c"] = 3
	after.Nested.Notes = []string{"alpha", "gamma"}
	after.NestedPtr = nil
	after.Interface = map[string]bool{"ok": false}
	after.privateField = "changed"

	godump.Diff(before, after)

	diff := godump.DiffStr(before, after)
	_ = diff
}
