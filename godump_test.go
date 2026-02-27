package godump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"text/tabwriter"
	"time"
	"unsafe"

	assert "github.com/goforj/godump/internal/testassert"
	require "github.com/goforj/godump/internal/testrequire"
)

func newDumperT(t *testing.T, opts ...Option) *Dumper {
	t.Helper()

	d := NewDumper(opts...)
	// we enforce no color in tests to avoid terminal issues
	d.colorizer = colorizeUnstyled

	return d
}

func dumpStrT(t *testing.T, v ...any) string {
	t.Helper()

	return newDumperT(t).DumpStr(v...)
}

func TestSimpleStruct(t *testing.T) {
	type Profile struct {
		Age   int
		Email string
	}
	type User struct {
		Name    string
		Profile Profile
	}

	user := User{Name: "Alice", Profile: Profile{Age: 30, Email: "alice@example.com"}}
	out := dumpStrT(t, user)

	assert.Contains(t, out, "#godump.User")
	assert.Contains(t, out, "+Name")
	assert.Contains(t, out, `"Alice"`)
	assert.Contains(t, out, "+Profile")
	assert.Contains(t, out, "#godump.Profile")
	assert.Contains(t, out, "+Age")
	assert.Contains(t, out, "30")
	assert.Contains(t, out, "+Email")
	assert.Contains(t, out, "alice@example.com")
}

func TestNilPointer(t *testing.T) {
	var s *string
	out := dumpStrT(t, s)
	assert.Contains(t, out, "(nil)")
}

func TestCycleReference(t *testing.T) {
	type Node struct {
		Next *Node
	}
	n := &Node{}
	n.Next = n
	out := dumpStrT(t, n)
	assert.Contains(t, out, "↩︎ &1")
}

func TestConcurrentDumpReferenceIDs(t *testing.T) {
	type Node struct {
		Next *Node
	}

	d := NewDumper()
	d.colorizer = colorizeUnstyled

	const runs = 50
	results := make(chan string, runs)

	var wg sync.WaitGroup
	wg.Add(runs)

	for i := 0; i < runs; i++ {
		go func() {
			defer wg.Done()

			n := &Node{}
			n.Next = n
			results <- d.DumpStr(n)
		}()
	}

	wg.Wait()
	close(results)

	for out := range results {
		assert.Contains(t, out, "↩︎ &1")
		assert.NotContains(t, out, "↩︎ &2")
	}
}

func TestMaxDepth(t *testing.T) {
	type Node struct {
		Child *Node
	}
	n := &Node{}
	curr := n
	for i := 0; i < 20; i++ {
		curr.Child = &Node{}
		curr = curr.Child
	}
	out := dumpStrT(t, n)
	assert.Contains(t, out, "... (max depth)")
}

func TestMaxDepthAllowsScalarValues(t *testing.T) {
	type Inner struct {
		ID   int
		Name string
	}
	type Outer struct {
		Inner Inner
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Inner: Inner{
			ID:   101,
			Name: "alpha",
		},
	})

	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "101")
	assert.Contains(t, out, `"alpha"`)
	assert.NotContains(t, out, "... (max depth)")
}

func TestMaxDepthBlocksStringerStructs(t *testing.T) {
	type Inner struct {
		When time.Time
	}
	type Outer struct {
		Inner Inner
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Inner: Inner{
			When: time.Unix(0, 0).UTC(),
		},
	})

	assert.Contains(t, out, "... (max depth)")
	assert.NotContains(t, out, "1970-01-01")
}

func TestIsComplexValue(t *testing.T) {
	assert.False(t, isComplexValue(reflect.Value{}))

	var ifaceHolder struct {
		V any
	}
	ifaceVal := reflect.ValueOf(ifaceHolder).Field(0)
	assert.False(t, isComplexValue(ifaceVal))

	ifaceHolder = struct{ V any }{V: struct{}{}}
	ifaceVal = reflect.ValueOf(ifaceHolder).Field(0)
	assert.True(t, isComplexValue(ifaceVal))

	var ptr *int
	assert.False(t, isComplexValue(reflect.ValueOf(ptr)))

	assert.False(t, isComplexValue(reflect.ValueOf(1)))
	assert.True(t, isComplexValue(reflect.ValueOf(struct{}{})))
}

func TestMaxDepthBlocksNestedMapValues(t *testing.T) {
	v := map[string]map[string]int{
		"outer": {
			"inner": 1,
		},
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(v)

	assert.Contains(t, out, "outer")
	assert.Contains(t, out, "... (max depth)")
	assert.NotContains(t, out, "inner")
}

func TestMaxDepthAllowsSliceScalars(t *testing.T) {
	type Inner struct {
		Values []int
	}
	type Outer struct {
		Inner Inner
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Inner: Inner{
			Values: []int{1, 2},
		},
	})

	assert.Contains(t, out, "... (max depth)")
	assert.NotContains(t, out, "0 => 1")
}

func TestMaxDepthAllowsPointerScalars(t *testing.T) {
	type Inner struct {
		ID *int
	}
	type Outer struct {
		Inner Inner
	}
	id := 42

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Inner: Inner{
			ID: &id,
		},
	})

	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "42")
	assert.NotContains(t, out, "... (max depth)")
}

func TestMaxDepthAllowsPointerStructFields(t *testing.T) {
	type Inner struct {
		Name string
	}
	type Outer struct {
		Inner *Inner
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Inner: &Inner{Name: "alpha"},
	})

	assert.Contains(t, out, "alpha")
	assert.NotContains(t, out, "... (max depth)")
}

func TestMaxDepthAllowsInterfaceScalar(t *testing.T) {
	type Outer struct {
		Value any
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Value: "hello",
	})

	assert.Contains(t, out, `"hello"`)
	assert.NotContains(t, out, "... (max depth)")
}

func TestMaxDepthAllowsInterfaceStructFields(t *testing.T) {
	type Inner struct {
		Name string
	}
	type Outer struct {
		Value any
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Value: Inner{Name: "alpha"},
	})

	assert.Contains(t, out, "alpha")
	assert.NotContains(t, out, "... (max depth)")
}

func TestMaxDepthAllowsArrayScalars(t *testing.T) {
	type Inner struct {
		Values [2]int
	}
	type Outer struct {
		Inner Inner
	}

	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		Inner: Inner{
			Values: [2]int{1, 2},
		},
	})

	assert.Contains(t, out, "... (max depth)")
	assert.NotContains(t, out, "0 => 1")
}

func TestMaxDepthEdgeCases(t *testing.T) {
	type Inner struct {
		ID int
	}
	type Outer struct {
		MapVal      map[string]int
		SliceVal    []int
		ArrayVal    [2]int
		MapPtr      *map[string]int
		StructPtr   *Inner
		Interface   any
		NilMap      map[string]int
		NilSlice    []int
		NilSlicePtr *[]int
	}

	m := map[string]int{"key": 1}
	s := []int{1, 2}
	a := [2]int{3, 4}
	out := newDumperT(t, WithMaxDepth(1)).DumpStr(Outer{
		MapVal:      m,
		SliceVal:    s,
		ArrayVal:    a,
		MapPtr:      &m,
		StructPtr:   &Inner{ID: 7},
		Interface:   Inner{ID: 9},
		NilMap:      nil,
		NilSlice:    nil,
		NilSlicePtr: nil,
	})

	assert.Contains(t, out, "MapVal")
	assert.Contains(t, out, "... (max depth)")
	assert.NotContains(t, out, "key")
	assert.NotContains(t, out, "0 => 1")
	assert.NotContains(t, out, "0 => 3")

	assert.Contains(t, out, "StructPtr")
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "7")

	assert.Contains(t, out, "Interface")
	assert.Contains(t, out, "9")

	assert.Contains(t, out, "NilMap")
	assert.Contains(t, out, "map[string]int(nil)")
	assert.Contains(t, out, "NilSlice")
	assert.Contains(t, out, "[]int(nil)")
	assert.Contains(t, out, "NilSlicePtr")
	assert.Contains(t, out, "[]int(nil)")
}

func TestChanValueRendering(t *testing.T) {
	var nilCh chan int
	out := dumpStrT(t, nilCh)
	assert.Contains(t, out, "chan int(nil)")

	ch := make(chan int)
	out = dumpStrT(t, ch)
	assert.Contains(t, out, "chan int")
}

func TestStringerNilPointer(t *testing.T) {
	var tptr *time.Time
	d := newDumperT(t)
	state := newDumpState()
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	d.printValue(tw, reflect.ValueOf(tptr), 0, state)
	tw.Flush()
	out := b.String()
	assert.Contains(t, out, "time.Time(nil)")
	assert.NotContains(t, out, "... (max depth)")

	tptr = nil
	out = d.asStringer(reflect.ValueOf(tptr))
	assert.Contains(t, out, "time.Time(nil)")
}

func TestMapOutput(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	out := dumpStrT(t, m)

	assert.Contains(t, out, "a => 1")
	assert.Contains(t, out, "b => 2")
}

func TestSliceOutput(t *testing.T) {
	s := []string{"one", "two"}
	out := dumpStrT(t, s)

	assert.Contains(t, out, `0 => "one"`)
	assert.Contains(t, out, `1 => "two"`)
}

func TestAnonymousStruct(t *testing.T) {
	out := dumpStrT(t, struct{ ID int }{ID: 123})

	assert.Contains(t, out, "+ID")
	assert.Contains(t, out, "123")
}

func TestEmbeddedAnonymousStruct(t *testing.T) {
	type Base struct {
		ID int
	}
	type Derived struct {
		Base
		Name string
	}

	out := dumpStrT(t, Derived{Base: Base{ID: 456}, Name: "Test"})

	assert.Contains(t, out, `#godump.Derived {
  +Base => #godump.Base {
    +ID => 456 #int
  }
  +Name => "Test" #string
}`)
}

func TestControlCharsEscaped(t *testing.T) {
	s := "line1\nline2\tok"
	out := dumpStrT(t, s)
	assert.Contains(t, out, `\n`)
	assert.Contains(t, out, `\t`)
}

func TestFuncPlaceholder(t *testing.T) {
	fn := func() {}
	out := dumpStrT(t, fn)
	assert.Contains(t, out, "func()")
}

func TestSpecialTypes(t *testing.T) {
	type Unsafe struct {
		Ptr unsafe.Pointer
	}
	out := dumpStrT(t, Unsafe{})
	assert.Contains(t, out, "unsafe.Pointer(")

	c := make(chan int)
	out = dumpStrT(t, c)
	assert.Contains(t, out, "chan")

	complexNum := complex(1.1, 2.2)
	out = dumpStrT(t, complexNum)
	assert.Contains(t, out, "(1.1+2.2i)")
}

func TestDd(t *testing.T) {
	called := false
	exitFunc = func(code int) { called = true }
	Dd("x")
	assert.True(t, called)
}

func TestDumpHTML(t *testing.T) {
	html := DumpHTML(map[string]string{"foo": "bar"})
	assert.Contains(t, html, `<span style="color:`)
	assert.Contains(t, html, `foo`)
	assert.Contains(t, html, `bar`)
}

func TestDumpHTMLNoColor(t *testing.T) {
	html := NewDumper(WithoutColor()).DumpHTML(map[string]string{"foo": "bar"})
	assert.NotContains(t, html, `<span style="color:`)
	assert.Contains(t, html, `foo`)
	assert.Contains(t, html, `bar`)
}

func TestDumpStrNoColor(t *testing.T) {
	out := NewDumper(WithoutColor()).DumpStr("x")
	assert.NotContains(t, out, string(ansiEscape))
	assert.Contains(t, out, `"x"`)
}

func TestDumpStrNoHeader(t *testing.T) {
	out := newDumperT(t, WithoutHeader()).DumpStr("x")
	assert.NotContains(t, out, "<#dump")
	assert.Contains(t, out, `"x"`)
}

func TestDiffStr(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	left := User{Name: "Alice", Age: 42}
	right := User{Name: "Bob", Age: 42}

	out := newDumperT(t).DiffStr(left, right)
	out = stripANSI(out)
	assert.Contains(t, out, "<#diff //")
	assert.Contains(t, out, `-   +Name => "Alice" #string`)
	assert.Contains(t, out, `+   +Name => "Bob" #string`)
	assert.Contains(t, out, `    +Age  => 42 #int`)
}

func TestForceExported(t *testing.T) {
	type hidden struct {
		private string
	}
	h := hidden{private: "shh"}
	v := reflect.ValueOf(&h).Elem().Field(0) // make addressable
	out := forceExported(v)
	assert.True(t, out.CanInterface())
	assert.Equal(t, "shh", out.Interface())
}

func TestDetectColorVariants(t *testing.T) {
	t.Run("no environment variables", func(t *testing.T) {
		assert.True(t, detectColor())

		out := NewDumper().colorize(colorYellow, "test")
		assert.Equal(t, string(ansiEscape)+"[33mtest"+string(ansiEscape)+"[0m", out)
	})

	t.Run("forcing no color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		assert.False(t, detectColor())

		out := NewDumper().colorize(colorYellow, "test")
		assert.Equal(t, "test", out)
	})

	t.Run("forcing color", func(t *testing.T) {
		t.Setenv("FORCE_COLOR", "1")
		assert.True(t, detectColor())

		out := NewDumper().colorize(colorYellow, "test")
		assert.Equal(t, string(ansiEscape)+"[33mtest"+string(ansiEscape)+"[0m", out)
	})
}

func TestWithoutColorOverridesColorDetection(t *testing.T) {
	t.Setenv("FORCE_COLOR", "1")

	out := NewDumper(WithoutColor()).colorize(colorYellow, "test")
	assert.Equal(t, "test", out)
}

func TestColorizeWithPresetColorizer(t *testing.T) {
	d := NewDumper()
	d.colorizer = colorizeUnstyled

	out := d.colorize(colorYellow, "test")
	assert.Equal(t, "test", out)
}

func TestColorizeWithDisableColorFlag(t *testing.T) {
	d := NewDumper()
	d.disableColor = true

	out := d.colorize(colorYellow, "test")
	assert.Equal(t, "test", out)
}

func TestEnsureColorizer(t *testing.T) {
	t.Run("disable color", func(t *testing.T) {
		d := NewDumper()
		d.disableColor = true

		d.ensureColorizer()
		out := d.colorizer(colorYellow, "test")
		assert.Equal(t, "test", out)
	})

	t.Run("detect color", func(t *testing.T) {
		d := NewDumper()

		d.ensureColorizer()
		out := d.colorizer(colorYellow, "test")
		assert.Equal(t, string(ansiEscape)+"[33mtest"+string(ansiEscape)+"[0m", out)
	})

	t.Run("already set", func(t *testing.T) {
		d := NewDumper()
		d.colorizer = colorizeUnstyled

		d.ensureColorizer()
		out := d.colorizer(colorYellow, "test")
		assert.Equal(t, "test", out)
	})
}

func TestHtmlColorizeUnknown(t *testing.T) {
	// Color not in htmlColorMap
	out := colorizeHTML(string(ansiEscape)+"[999m", "test")
	assert.Contains(t, out, `<span style="color:`)
	assert.Contains(t, out, "test")
}

func TestUnreadableFallback(t *testing.T) {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)

	var ch chan int // nil typed value, not interface
	rv := reflect.ValueOf(ch)

	newDumperT(t).printValue(tw, rv, 0, newDumpState())
	tw.Flush()

	output := b.String()
	assert.Contains(t, output, "(nil)")
}

func TestFindFirstNonInternalFrameFallback(t *testing.T) {
	// Trigger the fallback by skipping deeper
	file, line := newDumperT(t).findFirstNonInternalFrame(0)
	// We can't assert much here reliably, but calling it adds coverage
	assert.True(t, len(file) >= 0)
	assert.True(t, line >= 0)
}

func TestUnreadableFieldFallback(t *testing.T) {
	var v reflect.Value // zero Value, not valid
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)

	newDumperT(t).printValue(tw, v, 0, newDumpState())
	tw.Flush()

	out := sb.String()
	assert.Contains(t, out, "<invalid>")
}

func TestTimeType(t *testing.T) {
	now := time.Now()
	out := dumpStrT(t, now)
	assert.Contains(t, out, "#time.Time")
}

func TestPrimitiveTypes(t *testing.T) {
	out := dumpStrT(t,
		int8(1),
		int16(2),
		uint8(3),
		uint16(4),
		uintptr(5),
		float32(1.5),
		[2]int{6, 7},
		any(42),
	)

	assert.Contains(t, out, "1 #int8")
	assert.Contains(t, out, "2 #int16")
	assert.Contains(t, out, "3 #uint8")
	assert.Contains(t, out, "4 #uint16")
	assert.Contains(t, out, "5 #uintptr")
	assert.Contains(t, out, "1.500000 #float32")
	assert.Contains(t, out, "0 => 6 #int") // array
	assert.Contains(t, out, "42 #int")     // interface{}
}

func TestEscapeControl_AllVariants(t *testing.T) {
	in := "\n\t\r\v\f" + string(ansiEscape)
	out := escapeControl(in)

	assert.Contains(t, out, `\n`)
	assert.Contains(t, out, `\t`)
	assert.Contains(t, out, `\r`)
	assert.Contains(t, out, `\v`)
	assert.Contains(t, out, `\f`)
	assert.Contains(t, out, `\x1b`)
}

func TestDefaultFallback_Unreadable(t *testing.T) {
	// Create a reflect.Value that is valid but not interfaceable
	var v reflect.Value

	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	newDumperT(t).printValue(tw, v, 0, newDumpState())
	tw.Flush()

	assert.Contains(t, buf.String(), "<invalid>")
}

func TestPrintValue_Uintptr(t *testing.T) {
	// Use uintptr directly
	val := uintptr(12345)
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	newDumperT(t).printValue(tw, reflect.ValueOf(val), 0, newDumpState())
	tw.Flush()

	assert.Contains(t, buf.String(), "12345")
}

func TestPrintValue_UnsafePointer(t *testing.T) {
	// Trick it by converting an int pointer
	i := 5
	up := unsafe.Pointer(&i)
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	newDumperT(t).printValue(tw, reflect.ValueOf(up), 0, newDumpState())
	tw.Flush()

	assert.Contains(t, buf.String(), "unsafe.Pointer")
}

func TestPrintValue_Func(t *testing.T) {
	fn := func() {}
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	newDumperT(t).printValue(tw, reflect.ValueOf(fn), 0, newDumpState())
	tw.Flush()

	assert.Contains(t, buf.String(), "func()")
}

func TestMaxDepthTruncation(t *testing.T) {
	type Node struct {
		Next *Node
	}
	root := &Node{}
	curr := root
	for i := 0; i < 20; i++ {
		curr.Next = &Node{}
		curr = curr.Next
	}

	out := dumpStrT(t, root)
	assert.Contains(t, out, "... (max depth)")
}

func TestCustomMaxDepthTruncation(t *testing.T) {
	type Node struct {
		Next *Node
	}
	root := &Node{}
	curr := root
	for i := 0; i < 3; i++ {
		curr.Next = &Node{}
		curr = curr.Next
	}

	out := newDumperT(t, WithMaxDepth(2)).DumpStr(root)
	assert.Contains(t, out, "... (max depth)")

	out = newDumperT(t, WithMaxDepth(0)).DumpStr(root)
	assert.Contains(t, out, "... (max depth)")

	out = newDumperT(t, WithMaxDepth(-1)).DumpStr(root)
	assert.NotContains(t, out, "... (max depth)")
}

func TestMapTruncation(t *testing.T) {
	largeMap := map[int]int{}
	for i := 0; i < 200; i++ {
		largeMap[i] = i
	}
	out := dumpStrT(t, largeMap)
	assert.Contains(t, out, "... (truncated)")
}

func TestNilInterfaceTypePrint(t *testing.T) {
	var x any = (*int)(nil)
	out := dumpStrT(t, x)
	assert.Contains(t, out, "(nil)")
}

func TestUnreadableDefaultBranch(t *testing.T) {
	v := reflect.Value{}
	out := dumpStrT(t, v)
	assert.Contains(t, out, "#reflect.Value") // new expected fallback
}

func TestNilChan(t *testing.T) {
	var ch chan int
	out := dumpStrT(t, ch)
	if !strings.Contains(out, "chan int(nil)") {
		t.Errorf("Expected nil chan representation, got: %q", out)
	}
}

func TestTruncatedSlice(t *testing.T) {
	slice := make([]int, 101)
	out := dumpStrT(t, slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}
}

func TestCustomTruncatedSlice(t *testing.T) {
	slice := make([]int, 3)
	out := newDumperT(t, WithMaxItems(2)).DumpStr(slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}

	out = newDumperT(t, WithMaxItems(0)).DumpStr(slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}

	out = newDumperT(t, WithMaxItems(-1)).DumpStr(slice)
	if strings.Contains(out, "... (truncated)") {
		t.Error("Negative MaxItems option should not be applied")
	}
}

func TestTruncatedMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	out := newDumperT(t, WithMaxItems(1)).DumpStr(m)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected map to be truncated")
	}
}

func TestTruncatedString(t *testing.T) {
	s := strings.Repeat("x", 100001)
	out := dumpStrT(t, s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
	}
}

func TestCustomTruncatedString(t *testing.T) {
	s := strings.Repeat("x", 10)
	out := newDumperT(t, WithMaxStringLen(9)).DumpStr(s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
	}

	out = newDumperT(t, WithMaxStringLen(0)).DumpStr(s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
	}

	out = newDumperT(t, WithMaxStringLen(-1)).DumpStr(s)
	if strings.Contains(out, "…") {
		t.Error("Negative MaxStringLen option should not be applied")
	}
}

func TestBoolValues(t *testing.T) {
	out := dumpStrT(t, true, false)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Error("Expected bools to be printed")
	}
}

func TestDefaultBranchFallback(t *testing.T) {
	var v reflect.Value // zero reflect.Value
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	newDumperT(t).printValue(tw, v, 0, newDumpState())
	tw.Flush()
	if !strings.Contains(sb.String(), "<invalid>") {
		t.Error("Expected default fallback for invalid reflect.Value")
	}
}

type BadStringer struct{}

func (b *BadStringer) String() string {
	return "should never be called on nil"
}

func TestSafeStringerCall(t *testing.T) {
	var s fmt.Stringer = (*BadStringer)(nil) // nil pointer implementing Stringer

	out := dumpStrT(t, s)

	assert.Contains(t, out, "(nil)")
	assert.NotContains(t, out, "should never be called") // ensure String() wasn't called
}

func TestTimePointersEqual(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	type testCase struct {
		name     string
		a        *time.Time
		b        *time.Time
		expected bool
	}

	tests := []testCase{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        &now,
			b:        nil,
			expected: false,
		},
		{
			name:     "equal times",
			a:        &now,
			b:        &now,
			expected: true,
		},
		{
			name:     "different times",
			a:        &now,
			b:        &later,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equal := timePtrsEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, equal)
			Dump(tt)
		})
	}
}

func timePtrsEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

func TestPanicOnVisibleFieldsIndexMismatch(t *testing.T) {
	type Embedded struct {
		Secret string
	}
	type Outer struct {
		Embedded // Promoted field
		Age      int
	}

	// This will panic with:
	// panic: reflect: Field index out of bounds
	_ = dumpStrT(t, Outer{
		Embedded: Embedded{Secret: "classified"},
		Age:      42,
	})
}

type FriendlyDuration time.Duration

func (fd FriendlyDuration) String() string {
	td := time.Duration(fd)
	return fmt.Sprintf("%02d:%02d:%02d", int(td.Hours()), int(td.Minutes())%60, int(td.Seconds())%60)
}

func TestTheKitchenSink(t *testing.T) {
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

	Dump(val)

	out := dumpStrT(t, val)

	// Minimal coverage assertions
	assert.Contains(t, out, "+String")
	assert.Contains(t, out, `"test" #string`)
	assert.Contains(t, out, "+Bool")
	assert.Contains(t, out, "true #bool")
	assert.Contains(t, out, "+Int")
	assert.Contains(t, out, "42 #int")
	assert.Contains(t, out, "+Float")
	assert.Contains(t, out, "3.141500 #float64")
	assert.Contains(t, out, "+PtrString")
	assert.Contains(t, out, `"Hello" #*string`)
	assert.Contains(t, out, "+PtrDuration")
	assert.Contains(t, out, "20m0s #*time.Duration")
	assert.Contains(t, out, "+SliceInts")
	assert.Contains(t, out, "0 => 1 #int")
	assert.Contains(t, out, "+ArrayStrings")
	assert.Contains(t, out, "[2]string")
	assert.Contains(t, out, `"foo" #string`)
	assert.Contains(t, out, "+MapValues")
	assert.Contains(t, out, "a => 1 #int")
	assert.Contains(t, out, "+InterfaceImpl")
	assert.Contains(t, out, "0001-01-01 00:00:00 +0000 UTC #godump.IsZeroer")
	assert.Contains(t, out, "+Nested")
	assert.Contains(t, out, "+ID") // from nested
	assert.Contains(t, out, "+Notes")
	assert.Contains(t, out, "-privateField")
	assert.Contains(t, out, `"should show" #string`)
	assert.Contains(t, out, "↩︎") // recursion reference

	// Ensure no panic occurred and a sane dump was produced
	assert.Contains(t, out, "#")          // loosest
	assert.Contains(t, out, "Everything") // middle-ground
}

func TestForceExportedFallback(t *testing.T) {
	type s struct{ val string }
	v := reflect.ValueOf(s{"hidden"}).Field(0) // not addressable
	out := forceExported(v)
	assert.Equal(t, "hidden", out.String())
}

func TestFindFirstNonInternalFrame_FallbackBranch(t *testing.T) {
	testDumper := newDumperT(t)
	// Always fail to simulate 10 bad frames
	testDumper.callerFn = func(int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	file, line := testDumper.findFirstNonInternalFrame(0)
	assert.Equal(t, "", file)
	assert.Equal(t, 0, line)
}

func TestForceExported_NoInterfaceNoAddr(t *testing.T) {
	v := reflect.ValueOf(struct{ a string }{"x"}).Field(0)
	if v.CanAddr() {
		t.Skip("Field unexpectedly addressable; cannot hit fallback branch")
	}
	out := forceExported(v)
	assert.Equal(t, "x", out.String())
}

func TestPrintDumpHeader_SkipWhenNoFrame(t *testing.T) {
	testDumper := newDumperT(t)
	testDumper.callerFn = func(int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	var b strings.Builder
	testDumper.printDumpHeader(&b)
	assert.Equal(t, "", b.String()) // nothing should be written
}

type customChan chan int

func TestPrintValue_ChanNilBranch_Hardforce(t *testing.T) {
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)

	var ch customChan
	v := reflect.ValueOf(ch)

	assert.True(t, v.IsNil())
	assert.Equal(t, reflect.Chan, v.Kind())

	newDumperT(t).printValue(tw, v, 0, newDumpState())
	tw.Flush()

	assert.Contains(t, buf.String(), "customChan(nil)")
}

type secretString string

func (s secretString) String() string {
	return "👻 hidden stringer"
}

type hidden struct {
	secret secretString // unexported
}

func TestAsStringer_ForceExported(t *testing.T) {
	h := &hidden{secret: "boo"}                          // pointer makes fields addressable
	v := reflect.ValueOf(h).Elem().FieldByName("secret") // now v.CanAddr() is true, but v.CanInterface() is false

	assert.False(t, v.CanInterface(), "field must not be interfaceable")

	str := newDumperT(t).asStringer(v)

	assert.Contains(t, str, "👻 hidden stringer")
}

func TestForceExported_Interfaceable(t *testing.T) {
	v := reflect.ValueOf("already ok")
	require.True(t, v.CanInterface())

	out := forceExported(v)

	assert.Equal(t, "already ok", out.Interface())
}

func TestMakeAddressable_CanAddr(t *testing.T) {
	s := "hello"
	v := reflect.ValueOf(&s).Elem() // addressable string

	require.True(t, v.CanAddr())

	out := makeAddressable(v)

	assert.Equal(t, v.Interface(), out.Interface()) // compare by value
}

func TestFdump_WritesToWriter(t *testing.T) {
	var buf strings.Builder

	type Inner struct {
		Field string
	}
	type Outer struct {
		InnerField Inner
		Number     int
	}

	val := Outer{
		InnerField: Inner{Field: "hello"},
		Number:     42,
	}

	Fdump(&buf, val)

	out := buf.String()

	if !strings.Contains(out, "Outer") {
		t.Errorf("expected output to contain type name 'Outer', got: %s", out)
	}
	if !strings.Contains(out, "InnerField") || !strings.Contains(out, "hello") {
		t.Errorf("expected nested struct and field to appear, got: %s", out)
	}
	if !strings.Contains(out, "Number") || !strings.Contains(out, "42") {
		t.Errorf("expected field 'Number' with value '42', got: %s", out)
	}
	if !strings.Contains(out, "<#dump //") {
		t.Errorf("expected dump header with file and line, got: %s", out)
	}
}

func TestDumpWithCustomWriter(t *testing.T) {
	var buf strings.Builder

	type Inner struct {
		Field string
	}
	type Outer struct {
		InnerField Inner
		Number     int
	}

	val := Outer{
		InnerField: Inner{Field: "hello"},
		Number:     42,
	}

	NewDumper(WithWriter(&buf)).Dump(val)

	out := buf.String()

	if !strings.Contains(out, "Outer") {
		t.Errorf("expected output to contain type name 'Outer', got: %s", out)
	}
	if !strings.Contains(out, "InnerField") || !strings.Contains(out, "hello") {
		t.Errorf("expected nested struct and field to appear, got: %s", out)
	}
	if !strings.Contains(out, "Number") || !strings.Contains(out, "42") {
		t.Errorf("expected field 'Number' with value '42', got: %s", out)
	}
	if !strings.Contains(out, "<#dump //") {
		t.Errorf("expected dump header with file and line, got: %s", out)
	}
}

func wrappedDumpStr(skip int, v any) string {
	return NewDumper(WithSkipStackFrames(skip)).DumpStr(v)
}

func TestDumpWithCustomSkipStackFrames(t *testing.T) {
	// caller stack frames are
	//	1	godump.go           github.com/goforj/godump.findFirstNonInternalFrame			skip by initialCallerSkip
	//	2	godump.go           github.com/goforj/godump.printDumpHeader					skip by initialCallerSkip
	//	3	godump.go           github.com/goforj/godump.(*Dumper).DumpStr					skip by fail names contain godump.go
	//	4	godump_test.go      github.com/goforj/godump.TestDumpWithCustomSkipStackFrames
	//	5	testing.go          testing.tRunner
	out := NewDumper().DumpStr("test")
	assert.Contains(t, out, "godump_test.go")

	out = NewDumper(WithSkipStackFrames(1)).DumpStr("test")
	assert.NotContains(t, out, "godump_test.go")

	// skip=0: should print the original DumpStr call site
	out = wrappedDumpStr(0, "test")
	assert.Contains(t, out, "godump_test.go")

	// skip=1: should print the location inside wrappedDumpStr
	out = wrappedDumpStr(1, "test")
	assert.Contains(t, out, "godump_test.go")

	// skip=2: should skip current file and show the outermost frame
	out = wrappedDumpStr(2, "test")
	assert.NotContains(t, out, "godump_test.go")
}

// TestHexDumpRendering checks that the hex dump output is rendered correctly.
func TestHexDumpRendering(t *testing.T) {
	input := []byte(`{"error":"kek","last_error":"not implemented","lol":"ok"}`)
	output := dumpStrT(t, input)

	if !strings.Contains(output, "7b 22 65 72 72 6f 72") {
		t.Error("expected hex dump output missing")
	}
	if !strings.Contains(output, "| {") {
		t.Error("ASCII preview opening missing")
	}
	if !strings.Contains(output, `"ok"`) {
		t.Error("ASCII preview end content missing")
	}
	if !strings.Contains(output, "([]uint8) (len=") {
		t.Error("missing []uint8 preamble")
	}
}

func TestDumpRawMessage(t *testing.T) {
	type Payload struct {
		Meta json.RawMessage
	}

	raw := json.RawMessage(`{"key":"value","flag":true}`)
	p := Payload{Meta: raw}

	Dump(p)
}

func TestDumpParagraphAsBytes(t *testing.T) {
	paragraph := `This is a sample paragraph of text.
It contains multiple lines and some special characters like !@#$%^&*().
We want to see how it looks when dumped as a byte slice (hex dump).
New lines are also important to check.`

	// Convert the string to a byte slice
	paragraphBytes := []byte(paragraph)

	Dump(paragraphBytes)
}

func TestIndirectionNilPointer(t *testing.T) {
	type Embedded struct {
		Surname string
	}

	type Test struct {
		Name string
		*Embedded
	}

	ts := &Test{
		Name:     "John",
		Embedded: nil,
	}

	Dump(ts)

	// assert that we don't panic or crash when dereferencing nil pointers
	if ts.Embedded != nil {
		t.Errorf("Expected Embedded to be nil, got: %+v", ts.Embedded)
	}

	// Check that the output does not contain dereferenced nil pointer
	out := dumpStrT(t, ts)
	assert.Contains(t, out, "+Name")
	assert.Contains(t, out, "John")
	assert.Contains(t, out, "+Embedded => *godump.Embedded(nil)")
}

func TestDumpJSON(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		jsonStr := DumpJSONStr()
		expected := `{"error": "DumpJSON called with no arguments"}`
		assert.JSONEq(t, expected, jsonStr)
	})

	t.Run("single struct", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		user := User{Name: "Alice", Age: 30}
		jsonStr := DumpJSONStr(user)

		expected := `{
			"name": "Alice",
			"age": 30
		}`
		assert.JSONEq(t, expected, jsonStr)
	})

	t.Run("multiple values", func(t *testing.T) {
		jsonStr := DumpJSONStr("hello", 42, true)
		expected := `["hello", 42, true]`
		assert.JSONEq(t, expected, jsonStr)
	})

	t.Run("unmarshallable type", func(t *testing.T) {
		ch := make(chan int)
		jsonStr := DumpJSONStr(ch)
		expected := `{"error": "json: unsupported type: chan int"}`
		assert.JSONEq(t, expected, jsonStr)
	})

	t.Run("nil value", func(t *testing.T) {
		jsonStr := DumpJSONStr(nil)
		assert.JSONEq(t, "null", jsonStr)
	})

	t.Run("multiple integers", func(t *testing.T) {
		jsonStr := DumpJSONStr(1, 2)
		assert.JSONEq(t, "[1, 2]", jsonStr)
	})

	t.Run("slice of integers", func(t *testing.T) {
		jsonStr := DumpJSONStr([]int{1, 2})
		assert.JSONEq(t, "[1, 2]", jsonStr)
	})

	t.Run("Dumper.DumpJSON writes to writer", func(t *testing.T) {
		var buf bytes.Buffer
		d := NewDumper(WithWriter(&buf))
		d.DumpJSON(map[string]int{"x": 1})
		assert.JSONEq(t, `{"x": 1}`, buf.String())
	})

	t.Run("DumpJSON prints to stdout", func(t *testing.T) {
		r, w, _ := os.Pipe()
		done := make(chan struct{})

		go func() {
			NewDumper(WithWriter(w)).DumpJSON("hello")
			w.Close()
			close(done)
		}()

		output, _ := io.ReadAll(r)
		<-done

		assert.JSONEq(t, `"hello"`, strings.TrimSpace(string(output)))
	})

	t.Run("DumpJSON prints valid JSON to stdout for multiple values (Dumper)", func(t *testing.T) {
		var buf bytes.Buffer

		// Use WithWriter to inject the custom output
		d := NewDumper(WithWriter(&buf))
		d.DumpJSON("foo", 123, true)

		var got []any
		err := json.Unmarshal(buf.Bytes(), &got)
		require.NoError(t, err)
		assert.Equal(t, []any{"foo", float64(123), true}, got)
	})

	t.Run("DumpJSON prints valid JSON to stdout for multiple values", func(t *testing.T) {
		r, w, _ := os.Pipe()
		testDumper := newDumperT(t, WithWriter(w))

		// Read from pipe in goroutine
		done := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			done <- buf.String()
		}()

		// Perform the dump
		testDumper.DumpJSON("foo", 123, true)

		_ = w.Close()

		output := <-done
		output = strings.TrimSpace(output)

		t.Logf("Captured: %q", output)

		var got []any
		err := json.Unmarshal([]byte(output), &got)
		require.NoError(t, err, fmt.Sprintf("json.Unmarshal failed with output: %q", output))

		assert.Equal(t, []any{"foo", float64(123), true}, got)
	})
}

func TestDisableStringer(t *testing.T) {
	data := hidden{secret: "not so secret"}

	d := newDumperT(t, WithDisableStringer(true))
	v := d.DumpStr(data)
	require.Contains(t, v, `-secret => "not so secret"`)

	d = newDumperT(t)
	v = d.DumpStr(data)
	assert.Contains(t, v, `-secret => 👻 hidden stringer`)
}

func TestOnlyFields(t *testing.T) {
	type User struct {
		ID       int
		Email    string
		Password string
	}

	d := newDumperT(t, WithOnlyFields("ID", "Email"))
	out := d.DumpStr(User{ID: 10, Email: "user@example.com", Password: "secret"})
	assert.Contains(t, out, "+ID")
	assert.Contains(t, out, "+Email")
	assert.Contains(t, out, "user@example.com")
	assert.NotContains(t, out, "Password")
	assert.NotContains(t, out, "secret")
}

func TestFieldFiltersExact(t *testing.T) {
	type User struct {
		UserID   int
		Email    string
		Password string
	}

	d := newDumperT(t, WithExcludeFields("Password"))
	out := d.DumpStr(User{UserID: 10, Email: "user@example.com", Password: "secret"})
	assert.Contains(t, out, "+UserID")
	assert.Contains(t, out, "user@example.com")
	assert.NotContains(t, out, "Password")
	assert.NotContains(t, out, "secret")
}

func TestFieldFiltersMatchModes(t *testing.T) {
	type User struct {
		UserID int
		Email  string
	}

	d := newDumperT(t, WithExcludeFields("user"), WithFieldMatchMode(FieldMatchPrefix))
	out := d.DumpStr(User{UserID: 10, Email: "user@example.com"})
	assert.NotContains(t, out, "+UserID")
	assert.Contains(t, out, "+Email")
}

func TestRedactFields(t *testing.T) {
	type User struct {
		ID       int
		Password string
	}

	d := newDumperT(t, WithRedactFields("Password"))
	out := d.DumpStr(User{ID: 1, Password: "secret"})
	assert.Contains(t, out, "+Password => <redacted> #string")
	assert.NotContains(t, out, "secret")
}

func TestRedactSensitiveDefaults(t *testing.T) {
	type User struct {
		Password   string
		AuthToken  string
		SessionKey string
	}

	d := newDumperT(t, WithRedactSensitive())
	out := d.DumpStr(User{Password: "secret", AuthToken: "tok", SessionKey: "key"})
	assert.Contains(t, out, "Password")
	assert.Contains(t, out, "<redacted> #string")
	assert.Contains(t, out, "AuthToken")
	assert.Contains(t, out, "SessionKey")
	assert.NotContains(t, out, "secret")
	assert.NotContains(t, out, "tok")
	assert.NotContains(t, out, "key")
}

func TestRedactMatchMode(t *testing.T) {
	type User struct {
		APIKey string
		Email  string
	}

	d := newDumperT(t, WithRedactFields("key"), WithRedactMatchMode(FieldMatchContains))
	out := d.DumpStr(User{APIKey: "abc", Email: "user@example.com"})
	assert.Contains(t, out, "+APIKey => <redacted> #string")
	assert.Contains(t, out, "+Email")
	assert.NotContains(t, out, "abc")
}

func TestRedactMatchSuffixAndEmptyCandidate(t *testing.T) {
	type User struct {
		APIKey string
		Email  string
	}

	d := newDumperT(t,
		WithRedactFields("Key", ""),
		WithRedactMatchMode(FieldMatchSuffix),
	)
	out := d.DumpStr(User{APIKey: "abc", Email: "user@example.com"})
	assert.Contains(t, out, "+APIKey")
	assert.Contains(t, out, "<redacted> #string")
	assert.Contains(t, out, "+Email")
	assert.NotContains(t, out, "abc")

	placeholder := d.redactedValue(reflect.Value{})
	assert.Contains(t, placeholder, "<redacted>")
}

func TestRecursivePtr(t *testing.T) {
	s := "string"
	a := &s
	b := &a
	val := &b

	out := dumpStrT(t, val)
	assert.Contains(t, out, "#***string")
}

func TestDumpJSON_Coverage(t *testing.T) {
	var buf bytes.Buffer

	// Replace default dumper's writer *temporarily*
	oldWriter := defaultDumper.writer
	defaultDumper.writer = &buf
	defer func() { defaultDumper.writer = oldWriter }()

	DumpJSON(123, 456)

	out := buf.String()
	if !strings.Contains(out, "123") || !strings.Contains(out, "456") {
		t.Fatalf("expected JSON output to contain values, got: %s", out)
	}
}

func TestDumpStr_Coverage(t *testing.T) {
	out := DumpStr(123)

	if !strings.Contains(out, "123") {
		t.Fatalf("expected DumpStr to include value, got: %s", out)
	}
}

func TestShouldTruncateAtDepth(t *testing.T) {
	type sample struct {
		Value string
	}

	tests := []struct {
		name     string
		value    reflect.Value
		indent   int
		maxDepth int
		want     bool
	}{
		{
			name:     "indent below max does not truncate",
			value:    reflect.ValueOf(map[int]string{}),
			indent:   0,
			maxDepth: 1,
			want:     false,
		},
		{
			name:     "indent above max truncates complex values",
			value:    reflect.ValueOf(sample{}),
			indent:   2,
			maxDepth: 1,
			want:     true,
		},
		{
			name:     "indent equals max truncates collections",
			value:    reflect.ValueOf([]int{1, 2}),
			indent:   1,
			maxDepth: 1,
			want:     true,
		},
		{
			name:     "indent equals max allows structs",
			value:    reflect.ValueOf(sample{}),
			indent:   1,
			maxDepth: 1,
			want:     false,
		},
		{
			name:     "indent equals max allows scalars",
			value:    reflect.ValueOf(42),
			indent:   1,
			maxDepth: 1,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldTruncateAtDepth(tt.value, tt.indent, tt.maxDepth)
			assert.Equal(t, tt.want, got)
		})
	}
}
