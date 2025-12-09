//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/godump"
	"time"
)

type FriendlyDuration time.Duration

func (fd FriendlyDuration) String() string {
	td := time.Duration(fd)
	return fmt.Sprintf("%02d:%02d:%02d", int(td.Hours()), int(td.Minutes())%60, int(td.Seconds())%60)
}

func main() {
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

	now := time.Now()
	ptrStr := "Hello"
	dur := time.Minute * 20

	val := Everything{
		String:       "test",
		Bool:         true,
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
	val.Recursive.Self = val.Recursive // cycle

	godump.Dump(val)
}
