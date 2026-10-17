package godump

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// TestWithoutTypesValues checks that hiding annotations preserves values and container delimiters.
func TestWithoutTypesValues(t *testing.T) {
	value := 42
	var pointer *int
	var stringer *time.Time
	var channel chan int
	var function func()
	var iface any
	var raw unsafe.Pointer
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"string", "#string", `"#string"`},
		{"integer", 42, "42"},
		{"unsigned", uint(42), "42"},
		{"float", 1.5, "1.500000"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"complex", complex(1, 2), "(1+2i)"},
		{"pointer", &value, "42"},
		{"struct", struct{ Name string }{"Alice"}, "{\n  +Name => \"Alice\"\n}"},
		{"map", map[string]int{"count": 42}, "{\n   count => 42\n}"},
		{"slice", []int{42}, "[\n  0 => 42\n]"},
		{"array", [1]int{42}, "[\n  0 => 42\n]"},
		{"interface", struct{ Value any }{42}, "{\n  +Value => 42\n}"},
		{"stringer", time.Second, "1s"},
		{"nil pointer", pointer, "(nil)"},
		{"nil stringer", stringer, "(nil)"},
		{"nil map", map[string]int(nil), "(nil)"},
		{"nil slice", []int(nil), "(nil)"},
		{"nil channel", channel, "(nil)"},
		{"nil function", function, "(nil)"},
		{"nil interface", struct{ Value any }{iface}, "{\n  +Value => (nil)\n}"},
		{"nil unsafe pointer", raw, "0x0"},
		{"unsafe pointer", unsafe.Pointer(&value), fmt.Sprintf("%#x", uintptr(unsafe.Pointer(&value)))},
		{"channel", make(chan int), ""},
		{"function", func() {}, "<func>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDumper(WithoutTypes(), WithoutColor(), WithoutHeader())
			want := tt.want
			if tt.name == "channel" {
				want = fmt.Sprintf("(%#x)", reflect.ValueOf(tt.value).Pointer())
			}
			if got := strings.TrimSuffix(d.DumpStr(tt.value), "\n"); got != want {
				t.Fatalf("dump = %q, want %q", got, want)
			}
		})
	}
}

// TestWithoutTypesPresentation verifies value colors and annotations independently across output formats.
func TestWithoutTypesPresentation(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "1")
	value := struct{ Name string }{"Alice"}
	d := NewDumper(WithoutTypes(), WithoutHeader())
	text := d.DumpStr(value)
	if !strings.Contains(text, "\033[") || strings.Contains(text, "#string") {
		t.Fatalf("expected colored values without annotations: %q", text)
	}
	plain := NewDumper(WithoutTypes(), WithoutHeader(), WithoutColor()).DumpStr(value)
	if stripANSI(text) != plain {
		t.Fatalf("color changed displayed values: %q versus %q", stripANSI(text), plain)
	}
	html := d.DumpHTML(value)
	if !strings.Contains(html, "<span") || !strings.Contains(html, "Alice") || strings.Contains(html, "#string") {
		t.Fatalf("unexpected HTML: %s", html)
	}
	for _, diff := range []string{d.DiffStr(1, "two"), d.DiffHTML(1, "two")} {
		if strings.Contains(diff, "type:") || strings.Contains(diff, "#int") || strings.Contains(diff, "#string") || !strings.Contains(diff, "two") {
			t.Fatalf("unexpected diff: %s", diff)
		}
	}
	if got, want := d.DumpJSONStr(value), NewDumper().DumpJSONStr(value); got != want {
		t.Fatalf("JSON changed: %q versus %q", got, want)
	}
	if got := NewDumper(WithoutColor(), WithoutHeader()).DumpStr(42); got != "42 #int\n" {
		t.Fatalf("default output changed: %q", got)
	}
}

// TestWithoutTypesOptions checks redaction, byte dumps, Stringer overrides, and cycle tracking together with the option.
func TestWithoutTypesOptions(t *testing.T) {
	d := NewDumper(WithoutTypes(), WithoutHeader(), WithoutColor(), WithRedactFields("Password"))
	value := struct{ Password string }{"secret"}
	if got, want := d.DumpStr(value), "{\n  +Password => <redacted>\n}\n"; got != want {
		t.Fatalf("redaction = %q, want %q", got, want)
	}
	bytes := d.DumpStr([]byte("hello"))
	if !strings.HasPrefix(bytes, "(len=5 cap=5) {\n") || !strings.Contains(bytes, "hello") || strings.Contains(bytes, "uint8") {
		t.Fatalf("unexpected byte dump: %s", bytes)
	}
	if got := d.DumpStr([]byte{}); got != "(len=0 cap=0) {\n}\n" {
		t.Fatalf("unexpected empty byte dump: %q", got)
	}
	disabled := NewDumper(WithoutTypes(), WithoutHeader(), WithoutColor(), WithDisableStringer(true))
	if got := disabled.DumpStr(time.Second); got != "1000000000\n" {
		t.Fatalf("unexpected duration without Stringer: %q", got)
	}
	type node struct{ Next *node }
	n := &node{}
	n.Next = n
	if got := d.DumpStr(n); !strings.Contains(got, "↩︎ &1") || strings.Contains(got, "#") {
		t.Fatalf("unexpected cyclic dump: %s", got)
	}
}

// TestWithoutTypesNilStringer checks the direct helper path used for nil Stringer implementations.
func TestWithoutTypesNilStringer(t *testing.T) {
	var value *time.Time
	d := NewDumper(WithoutTypes(), WithoutColor())
	if got := d.asStringer(reflect.ValueOf(value)); got != "(nil)" {
		t.Fatalf("nil Stringer = %q", got)
	}
	private := reflect.ValueOf(struct{ hidden string }{"value"}).Field(0)
	if got := d.asStringer(private); got != "" {
		t.Fatalf("inaccessible field should not be treated as a Stringer: %q", got)
	}
}
