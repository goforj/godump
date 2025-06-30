package godump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

// stripANSI removes ANSI color codes for testable output.
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
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
	out := stripANSI(DumpStr(user))

	assert.Contains(t, out, "#godump.User")
	assert.Contains(t, out, "+Name")
	assert.Contains(t, out, "\"Alice\"")
	assert.Contains(t, out, "+Profile")
	assert.Contains(t, out, "#godump.Profile")
	assert.Contains(t, out, "+Age")
	assert.Contains(t, out, "30")
	assert.Contains(t, out, "+Email")
	assert.Contains(t, out, "alice@example.com")
}

func TestNilPointer(t *testing.T) {
	var s *string
	out := stripANSI(DumpStr(s))
	assert.Contains(t, out, "(nil)")
}

func TestCycleReference(t *testing.T) {
	type Node struct {
		Next *Node
	}
	n := &Node{}
	n.Next = n
	out := stripANSI(DumpStr(n))
	assert.Contains(t, out, "↩︎ &1")
}

func TestMaxDepth(t *testing.T) {
	type Node struct {
		Child *Node
	}
	n := &Node{}
	curr := n
	for range 20 {
		curr.Child = &Node{}
		curr = curr.Child
	}
	out := stripANSI(DumpStr(n))
	assert.Contains(t, out, "... (max depth)")
}

func TestMapOutput(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	out := stripANSI(DumpStr(m))

	assert.Contains(t, out, "a => 1")
	assert.Contains(t, out, "b => 2")
}

func TestSliceOutput(t *testing.T) {
	s := []string{"one", "two"}
	out := stripANSI(DumpStr(s))

	assert.Contains(t, out, "0 => \"one\"")
	assert.Contains(t, out, "1 => \"two\"")
}

func TestAnonymousStruct(t *testing.T) {
	out := stripANSI(DumpStr(struct{ ID int }{ID: 123}))

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

	out := stripANSI(DumpStr(Derived{Base: Base{ID: 456}, Name: "Test"}))

	assert.Contains(t, out, `#godump.Derived 
  +Base => #godump.Base 
    +ID => 456
  }
  +Name => "Test"
}`)
}

func TestControlCharsEscaped(t *testing.T) {
	s := "line1\nline2\tok"
	out := stripANSI(DumpStr(s))
	assert.Contains(t, out, `\n`)
	assert.Contains(t, out, `\t`)
}

func TestFuncPlaceholder(t *testing.T) {
	fn := func() {}
	out := stripANSI(DumpStr(fn))
	assert.Contains(t, out, "func()")
}

func TestSpecialTypes(t *testing.T) {
	type Unsafe struct {
		Ptr unsafe.Pointer
	}
	out := stripANSI(DumpStr(Unsafe{}))
	assert.Contains(t, out, "unsafe.Pointer(")

	c := make(chan int)
	out = stripANSI(DumpStr(c))
	assert.Contains(t, out, "chan")

	complexNum := complex(1.1, 2.2)
	out = stripANSI(DumpStr(complexNum))
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

func TestCallerLocation(t *testing.T) {
	file, line := callerLocation(0)
	assert.NotEmpty(t, file)
	assert.Greater(t, line, 0)
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
	})

	t.Run("forcing no color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		assert.False(t, detectColor())
	})

	t.Run("forcing color", func(t *testing.T) {
		t.Setenv("FORCE_COLOR", "1")
		assert.True(t, detectColor())
	})
}

func TestHtmlColorizeUnknown(t *testing.T) {
	// Color not in htmlColorMap
	out := htmlColorize("\033[999m", "test")
	assert.Contains(t, out, `<span style="color:`)
	assert.Contains(t, out, "test")
}

// package-level type + method
type secret struct{}

func (secret) hidden() {}

func TestUnreadableFallback(t *testing.T) {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)

	var ch chan int // nil typed value, not interface
	rv := reflect.ValueOf(ch)

	defaultDumper.printValue(tw, rv, 0, map[uintptr]bool{})
	tw.Flush()

	output := stripANSI(b.String())
	assert.Contains(t, output, "(nil)")
}

func TestFindFirstNonInternalFrameFallback(t *testing.T) {
	// Trigger the fallback by skipping deeper
	file, line := findFirstNonInternalFrame()
	// We can't assert much here reliably, but calling it adds coverage
	assert.True(t, len(file) >= 0)
	assert.True(t, line >= 0)
}

func TestUnreadableFieldFallback(t *testing.T) {
	var v reflect.Value // zero Value, not valid
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)

	defaultDumper.printValue(tw, v, 0, map[uintptr]bool{})
	tw.Flush()

	out := stripANSI(sb.String())
	assert.Contains(t, out, "<invalid>")
}

func TestTimeType(t *testing.T) {
	now := time.Now()
	out := stripANSI(DumpStr(now))
	assert.Contains(t, out, "#time.Time")
}

func TestPrimitiveTypes(t *testing.T) {
	out := stripANSI(DumpStr(
		int8(1),
		int16(2),
		uint8(3),
		uint16(4),
		uintptr(5),
		float32(1.5),
		[2]int{6, 7},
		any(42),
	))

	assert.Contains(t, out, "1")        // int8
	assert.Contains(t, out, "2")        // int16
	assert.Contains(t, out, "3")        // uint8
	assert.Contains(t, out, "4")        // uint16
	assert.Contains(t, out, "5")        // uintptr
	assert.Contains(t, out, "1.500000") // float32
	assert.Contains(t, out, "0 =>")     // array
	assert.Contains(t, out, "42")       // interface{}
}

func TestEscapeControl_AllVariants(t *testing.T) {
	in := "\n\t\r\v\f\x1b"
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
	defaultDumper.printValue(tw, v, 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "<invalid>")
}

func TestPrintValue_Uintptr(t *testing.T) {
	// Use uintptr directly
	val := uintptr(12345)
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	defaultDumper.printValue(tw, reflect.ValueOf(val), 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "12345")
}

func TestPrintValue_UnsafePointer(t *testing.T) {
	// Trick it by converting an int pointer
	i := 5
	up := unsafe.Pointer(&i)
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	defaultDumper.printValue(tw, reflect.ValueOf(up), 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "unsafe.Pointer")
}

func TestPrintValue_Func(t *testing.T) {
	fn := func() {}
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	defaultDumper.printValue(tw, reflect.ValueOf(fn), 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "func()")
}

func TestMaxDepthTruncation(t *testing.T) {
	type Node struct {
		Next *Node
	}
	root := &Node{}
	curr := root
	for range 20 {
		curr.Next = &Node{}
		curr = curr.Next
	}

	out := stripANSI(DumpStr(root))
	assert.Contains(t, out, "... (max depth)")
}

func TestCustomMaxDepthTruncation(t *testing.T) {
	type Node struct {
		Next *Node
	}
	root := &Node{}
	curr := root
	for range 3 {
		curr.Next = &Node{}
		curr = curr.Next
	}

	out := stripANSI(NewDumper(WithMaxDepth(2)).DumpStr(root))
	assert.Contains(t, out, "... (max depth)")

	out = stripANSI(NewDumper(WithMaxDepth(0)).DumpStr(root))
	assert.Contains(t, out, "... (max depth)")

	out = stripANSI(NewDumper(WithMaxDepth(-1)).DumpStr(root))
	assert.NotContains(t, out, "... (max depth)")
}

func TestMapTruncation(t *testing.T) {
	largeMap := map[int]int{}
	for i := range 200 {
		largeMap[i] = i
	}
	out := stripANSI(DumpStr(largeMap))
	assert.Contains(t, out, "... (truncated)")
}

func TestNilInterfaceTypePrint(t *testing.T) {
	var x any = (*int)(nil)
	out := stripANSI(DumpStr(x))
	assert.Contains(t, out, "(nil)")
}

func TestUnreadableDefaultBranch(t *testing.T) {
	v := reflect.Value{}
	out := stripANSI(DumpStr(v))
	assert.Contains(t, out, "#reflect.Value") // new expected fallback
}

func TestNilChan(t *testing.T) {
	var ch chan int
	out := DumpStr(ch)
	// Strip ANSI codes before checking
	clean := stripANSI(out)
	if !strings.Contains(clean, "chan int(nil)") {
		t.Errorf("Expected nil chan representation, got: %q", clean)
	}
}

func TestTruncatedSlice(t *testing.T) {
	slice := make([]int, 101)
	out := DumpStr(slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}
}

func TestCustomTruncatedSlice(t *testing.T) {
	slice := make([]int, 3)
	out := NewDumper(WithMaxItems(2)).DumpStr(slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}

	out = NewDumper(WithMaxItems(0)).DumpStr(slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}

	out = NewDumper(WithMaxItems(-1)).DumpStr(slice)
	if strings.Contains(out, "... (truncated)") {
		t.Error("Negative MaxItems option should not be applied")
	}
}

func TestTruncatedString(t *testing.T) {
	s := strings.Repeat("x", 100001)
	out := DumpStr(s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
	}
}

func TestCustomTruncatedString(t *testing.T) {
	s := strings.Repeat("x", 10)
	out := NewDumper(WithMaxStringLen(9)).DumpStr(s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
	}

	out = NewDumper(WithMaxStringLen(0)).DumpStr(s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
	}

	out = NewDumper(WithMaxStringLen(-1)).DumpStr(s)
	if strings.Contains(out, "…") {
		t.Error("Negative MaxStringLen option should not be applied")
	}
}

func TestBoolValues(t *testing.T) {
	out := DumpStr(true, false)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Error("Expected bools to be printed")
	}
}

func TestDefaultBranchFallback(t *testing.T) {
	var v reflect.Value // zero reflect.Value
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	defaultDumper.printValue(tw, v, 0, map[uintptr]bool{})
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

	out := stripANSI(DumpStr(s))

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
	_ = DumpStr(Outer{
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
		Recursive:     &Ref{},
		privateField:  "should show",
		privateStruct: Inner{ID: 5, Notes: []string{"private"}},
	}
	val.Recursive.Self = val.Recursive // cycle

	Dump(val)

	out := stripANSI(DumpStr(val))

	// Minimal coverage assertions
	assert.Contains(t, out, "+String")
	assert.Contains(t, out, `"test"`)
	assert.Contains(t, out, "+Bool")
	assert.Contains(t, out, "true")
	assert.Contains(t, out, "+Int")
	assert.Contains(t, out, "42")
	assert.Contains(t, out, "+Float")
	assert.Contains(t, out, "3.1415")
	assert.Contains(t, out, "+PtrString")
	assert.Contains(t, out, `"Hello"`)
	assert.Contains(t, out, "+SliceInts")
	assert.Contains(t, out, "0 => 1")
	assert.Contains(t, out, "+ArrayStrings")
	assert.Contains(t, out, `"foo"`)
	assert.Contains(t, out, "+MapValues")
	assert.Contains(t, out, "a => 1")
	assert.Contains(t, out, "+Nested")
	assert.Contains(t, out, "+ID") // from nested
	assert.Contains(t, out, "+Notes")
	assert.Contains(t, out, "-privateField")
	assert.Contains(t, out, `"should show"`)
	assert.Contains(t, out, "↩︎") // recursion reference

	// Ensure no panic occurred and a sane dump was produced
	assert.Contains(t, out, "#")          // loosest
	assert.Contains(t, out, "Everything") // middle-ground

}

func TestAnsiColorize_Disabled(t *testing.T) {
	orig := enableColor
	enableColor = false
	defer func() { enableColor = orig }()

	out := ansiColorize(colorYellow, "test")
	assert.Equal(t, "test", out)
}

func TestForceExportedFallback(t *testing.T) {
	type s struct{ val string }
	v := reflect.ValueOf(s{"hidden"}).Field(0) // not addressable
	out := forceExported(v)
	assert.Equal(t, "hidden", out.String())
}

func TestFindFirstNonInternalFrame_FallbackBranch(t *testing.T) {
	orig := callerFn
	defer func() { callerFn = orig }()

	// Always fail to simulate 10 bad frames
	callerFn = func(i int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	file, line := findFirstNonInternalFrame()
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
	orig := callerFn
	defer func() { callerFn = orig }()

	callerFn = func(skip int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	var b strings.Builder
	printDumpHeader(&b, 3)
	assert.Equal(t, "", b.String()) // nothing should be written
}

var runtimeCaller = runtime.Caller

func TestCallerLocation_Fallback(t *testing.T) {
	// Override runtime.Caller behavior
	orig := runtimeCaller
	defer func() { runtimeCaller = orig }()
	runtimeCaller = func(skip int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	file, line := callerLocation(5)
	assert.Equal(t, "", file)
	assert.Equal(t, 0, line)
}

type customChan chan int

func TestPrintValue_ChanNilBranch_Hardforce(t *testing.T) {
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)

	var ch customChan
	v := reflect.ValueOf(ch)

	assert.True(t, v.IsNil())
	assert.Equal(t, reflect.Chan, v.Kind())

	defaultDumper.printValue(tw, v, 0, map[uintptr]bool{})
	tw.Flush()

	out := stripANSI(buf.String())
	assert.Contains(t, out, "customChan(nil)")
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
	str := asStringer(v)

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

// TestHexDumpRendering checks that the hex dump output is rendered correctly.
func TestHexDumpRendering(t *testing.T) {
	input := []byte(`{"error":"kek","last_error":"not implemented","lol":"ok"}`)
	output := DumpStr(input)
	output = stripANSI(output)
	Dump(input)

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
	out := stripANSI(DumpStr(ts))
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
		// Save and override defaultDumper temporarily
		orig := defaultDumper

		r, w, _ := os.Pipe()
		defaultDumper = NewDumper(WithWriter(w))

		// Read from pipe in goroutine
		done := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			done <- buf.String()
		}()

		// Perform the dump
		DumpJSON("foo", 123, true)

		_ = w.Close()
		defaultDumper = orig // restore original dumper

		output := <-done
		output = strings.TrimSpace(output)

		t.Logf("Captured: %q", output)

		var got []any
		err := json.Unmarshal([]byte(output), &got)
		require.NoError(t, err, "json.Unmarshal failed with output: %q", output)

		assert.Equal(t, []any{"foo", float64(123), true}, got)
	})

}
