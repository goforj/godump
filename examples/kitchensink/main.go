//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/godump"
	"time"
)

type FriendlyDuration time.Duration

// String renders the duration as HH:MM:SS.
func (fd FriendlyDuration) String() string {
	td := time.Duration(fd)
	return fmt.Sprintf("%02d:%02d:%02d", int(td.Hours()), int(td.Minutes())%60, int(td.Seconds())%60)
}

// main demonstrates a kitchen-sink dump.
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
	// #main.Everything {
	//  +String      => "test" #string
	//  +Bool        => true #bool
	//  +Int         => 42 #int
	//  +Float       => 3.141500 #float64
	//  +Time        => 2025-12-09 17:57:25.585793 -0600 CST m=+0.000045251 #time.Time
	//  +Duration    => 20m0s #time.Duration
	//  +Friendly    => 00:20:00 #main.FriendlyDuration
	//  +PtrString   => "Hello" #*string
	//  +PtrDuration => 20m0s #*time.Duration
	//  +SliceInts   => #[]int [
	//    0 => 1 #int
	//    1 => 2 #int
	//    2 => 3 #int
	//  ]
	//  +ArrayStrings => #[2]string [
	//    0 => "foo" #string
	//    1 => "bar" #string
	//  ]
	//  +MapValues => #map[string]int {
	//     a => 1 #int
	//     b => 2 #int
	//  }
	//  +Nested  => #main.Inner {
	//    +ID    => 10 #int
	//    +Notes => #[]string [
	//      0 => "alpha" #string
	//      1 => "beta" #string
	//    ]
	//    +Blob => ([]uint8) (len=25 cap=25) {
	//      00000000  7b 22 6b 69 6e 64 22 3a  22 74 65 73 74 22 2c 22  | {"kind":"test"," |
	//      00000010  6f 6b 22 3a 74 72 75 65  7d                       | ok":true}        |
	//    }
	//  }
	//  +NestedPtr => #*main.Inner {
	//    +ID      => 99 #int
	//    +Notes   => #[]string [
	//      0 => "x" #string
	//      1 => "y" #string
	//    ]
	//    +Blob => ([]uint8) (len=28 cap=28) {
	//      00000000  7b 22 6d 73 67 22 3a 22  68 69 22 2c 22 73 74 61  | {"msg":"hi","sta |
	//      00000010  74 75 73 22 3a 22 63 6f  6f 6c 22 7d              | tus":"cool"}     |
	//    }
	//  }
	//  +Interface => #map[string]bool {
	//     ok => true #bool
	//  }
	//  +InterfaceImpl => 0001-01-01 00:00:00 +0000 UTC #main.IsZeroer
	//  +Recursive     => #*main.Ref {
	//    +Self        => ↩︎ &3
	//  }
	//  -privateField  => "should show" #string
	//  -privateStruct => #main.Inner {
	//    +ID          => 5 #int
	//    +Notes       => #[]string [
	//      0 => "private" #string
	//    ]
	//    +Blob => []uint8(nil)
	//  }
	// }
}
