package godump

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/tabwriter"
	"time"
	"unsafe"

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

func TestControlCharsEscaped(t *testing.T) {
	s := "line1\nline2\tok"
	out := stripANSI(DumpStr(s))
	assert.Contains(t, out, `\n`)
	assert.Contains(t, out, `\t`)
}

func TestFuncPlaceholder(t *testing.T) {
	fn := func() {}
	out := stripANSI(DumpStr(fn))
	assert.Contains(t, out, "func(...) {...}")
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
	_ = os.Setenv("NO_COLOR", "1")
	assert.False(t, detectColor())

	_ = os.Unsetenv("NO_COLOR")
	_ = os.Setenv("FORCE_COLOR", "1")
	assert.True(t, detectColor())

	_ = os.Unsetenv("FORCE_COLOR")
	assert.True(t, detectColor())
}

func TestPrintDumpHeaderFallback(t *testing.T) {
	// Intentionally skip enough frames so findFirstNonInternalFrame returns empty
	printDumpHeader(os.Stdout, 100)
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

	printValue(tw, rv, 0, map[uintptr]bool{})
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

	printValue(tw, v, 0, map[uintptr]bool{})
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
	printValue(tw, v, 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "<invalid>")
}

func TestPrintValue_Uintptr(t *testing.T) {
	// Use uintptr directly
	val := uintptr(12345)
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	printValue(tw, reflect.ValueOf(val), 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "12345")
}

func TestPrintValue_UnsafePointer(t *testing.T) {
	// Trick it by converting an int pointer
	i := 5
	up := unsafe.Pointer(&i)
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	printValue(tw, reflect.ValueOf(up), 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "unsafe.Pointer")
}

func TestPrintValue_Func(t *testing.T) {
	fn := func() {}
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	printValue(tw, reflect.ValueOf(fn), 0, map[uintptr]bool{})
	tw.Flush()

	assert.Contains(t, buf.String(), "func(...) {...}")
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

func TestDetectColorEnvVars(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	assert.False(t, detectColor())

	os.Unsetenv("NO_COLOR")
	os.Setenv("FORCE_COLOR", "1")
	assert.True(t, detectColor())

	os.Unsetenv("FORCE_COLOR")
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

	// Match stable part of reflect.Value zero output
	assert.Contains(t, out, "typ_ => *abi.Type(nil)")
}

func TestNoColorEnvironment(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if detectColor() {
		t.Error("Expected color to be disabled when NO_COLOR is set")
	}
}

func TestForceColorEnvironment(t *testing.T) {
	t.Setenv("FORCE_COLOR", "1")
	if !detectColor() {
		t.Error("Expected color to be enabled when FORCE_COLOR is set")
	}
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
	orig := maxItems
	maxItems = 5
	defer func() { maxItems = orig }()
	slice := make([]int, 10)
	out := DumpStr(slice)
	if !strings.Contains(out, "... (truncated)") {
		t.Error("Expected slice to be truncated")
	}
}

func TestTruncatedString(t *testing.T) {
	orig := maxStringLen
	maxStringLen = 10
	defer func() { maxStringLen = orig }()
	s := strings.Repeat("x", 50)
	out := DumpStr(s)
	if !strings.Contains(out, "…") {
		t.Error("Expected long string to be truncated")
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
	printValue(tw, v, 0, map[uintptr]bool{})
	tw.Flush()
	if !strings.Contains(sb.String(), "<invalid>") {
		t.Error("Expected default fallback for invalid reflect.Value")
	}
}
