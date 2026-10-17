package godump

import (
	"reflect"
	"strings"
	"testing"
)

// TestWithRedactTagsSelection checks opt-in behavior, wildcard selection, and independent field-name matching.
func TestWithRedactTagsSelection(t *testing.T) {
	type user struct {
		Username   string `xml:"-"`
		Email      string `json:"email" msgpack:"-"`
		Credential string `json:"-"`
		Named      string
		Visible    string
	}
	value := user{"user-value", "email-value", "credential-value", "named-value", "visible-value"}
	tests := []struct {
		name    string
		options []Option
		hidden  []string
	}{
		{"default", nil, nil},
		{"json", []Option{WithRedactTags("json")}, []string{"credential-value"}},
		{"multiple keys", []Option{WithRedactTags("json", "xml")}, []string{"credential-value", "user-value"}},
		{"any tag", []Option{WithRedactTags()}, []string{"credential-value", "user-value", "email-value"}},
		{"wildcard", []Option{WithRedactTags("*")}, []string{"credential-value", "user-value", "email-value"}},
		{"wildcard after key", []Option{WithRedactTags("missing", "*")}, []string{"credential-value", "user-value", "email-value"}},
		{"repeated calls", []Option{WithRedactTags("json"), WithRedactTags("msgpack")}, []string{"credential-value", "email-value"}},
		{"case sensitive", []Option{WithRedactTags("JSON")}, nil},
		{"unknown key", []Option{WithRedactTags("missing")}, nil},
		{"empty key", []Option{WithRedactTags("")}, nil},
		{"field names", []Option{WithRedactTags("json"), WithRedactFields("Named")}, []string{"credential-value", "named-value"}},
		{"name match mode", []Option{WithRedactTags("json"), WithRedactMatchMode(FieldMatchContains)}, []string{"credential-value"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := append([]Option{WithoutColor(), WithoutHeader()}, tt.options...)
			out := NewDumper(options...).DumpStr(value)
			for _, text := range []string{"user-value", "email-value", "credential-value", "named-value", "visible-value"} {
				hidden := false
				for _, candidate := range tt.hidden {
					if candidate == text {
						hidden = true
					}
				}
				if strings.Contains(out, text) == hidden {
					t.Fatalf("value %q, hidden=%t: %s", text, hidden, out)
				}
			}
			if got := strings.Count(out, "<redacted>"); got != len(tt.hidden) {
				t.Fatalf("redactions = %d, want %d: %s", got, len(tt.hidden), out)
			}
		})
	}
}

// TestWithRedactTagsNested checks traversal through embedded structs, pointers, collections, and private fields.
func TestWithRedactTagsNested(t *testing.T) {
	type Embedded struct {
		Value string `json:"-"`
	}
	type nested struct {
		Value string `json:"-"`
	}
	value := struct {
		Embedded
		Inner     nested
		Pointer   *nested
		Slice     []nested
		Array     [1]nested
		Map       map[string]nested
		Interface any
		Whole     nested `json:"-"`
		private   string `json:"-"` //nolint:revive // GoDump intentionally inspects private fields too.
	}{
		Embedded:  Embedded{"secret-embedded"},
		Inner:     nested{"secret-inner"},
		Pointer:   &nested{"secret-pointer"},
		Slice:     []nested{{"secret-slice"}},
		Array:     [1]nested{{"secret-array"}},
		Map:       map[string]nested{"item": {"secret-map"}},
		Interface: nested{"secret-interface"},
		Whole:     nested{"secret-whole"},
		private:   "secret-private",
	}
	d := NewDumper(WithRedactTags("json"), WithoutColor(), WithoutHeader())
	out := d.DumpStr(value)
	if strings.Contains(out, "secret-") || strings.Count(out, "<redacted>") != 9 {
		t.Fatalf("unexpected nested redaction: %s", out)
	}
	if !strings.Contains(out, "+Whole") || !strings.Contains(out, "-private") {
		t.Fatalf("redaction should retain field names: %s", out)
	}
	filtered := NewDumper(WithRedactTags(), WithOnlyFields("Whole"), WithExcludeFields("Whole"), WithoutHeader(), WithoutColor()).DumpStr(value)
	if strings.Contains(filtered, "+Whole") || strings.Contains(filtered, "<redacted>") {
		t.Fatalf("excluded field should not be rendered: %s", filtered)
	}
}

// TestIgnoredStructTags checks exact values and Go tag syntax without placing malformed tags in compiled structs.
func TestIgnoredStructTags(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{``, false},
		{`   `, false},
		{`json:"-"`, true},
		{`json:"name" xml:"-"`, true},
		{` json:"name" `, false},
		{`json:""`, false},
		{`json:"-,omitempty"`, false},
		{`json:" - "`, false},
		{`json:"\u002d"`, true},
		{`json:"a\"b" xml:"-"`, true},
		{`broken`, false},
		{`:"-"`, false},
		{`json:`, false},
		{`json:-`, false},
		{`json:"unterminated`, false},
		{`json:"\q"`, false},
		{`json:"name", xml:"-"`, false},
		{"json\t:\"-\"", false},
		{"json\x7f:\"-\"", false},
	}
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			if got := hasIgnoredStructTag(reflect.StructTag(tt.tag)); got != tt.want {
				t.Fatalf("hasIgnoredStructTag(%q) = %t, want %t", tt.tag, got, tt.want)
			}
		})
	}
}

// taggedStringer demonstrates the existing Stringer boundary for struct traversal.
type taggedStringer struct {
	Credential string `json:"-"`
}

// String returns a custom representation that bypasses field traversal by default.
func (v taggedStringer) String() string { return v.Credential }

// TestWithRedactTagsOutput checks cloned renderers, hidden types, JSON behavior, and the documented Stringer boundary.
func TestWithRedactTagsOutput(t *testing.T) {
	type user struct {
		Name       string
		Credential string `xml:"-"`
	}
	before := user{"Alice", "secret-before"}
	after := user{"Bob", "secret-after"}
	d := NewDumper(WithRedactTags("xml"), WithoutTypes(), WithoutHeader())
	for _, out := range []string{d.DumpStr(before), d.DumpHTML(before), d.DiffStr(before, after), d.DiffHTML(before, after)} {
		if strings.Contains(out, "secret-") || !strings.Contains(out, "<redacted>") || strings.Contains(out, "#string") {
			t.Fatalf("unexpected rendering: %s", out)
		}
	}
	if got, want := d.DumpJSONStr(before), NewDumper().DumpJSONStr(before); got != want {
		t.Fatalf("JSON behavior changed: %q versus %q", got, want)
	}
	stringer := taggedStringer{"secret-stringer"}
	if got := d.DumpStr(stringer); !strings.Contains(got, "secret-stringer") {
		t.Fatalf("Stringer behavior changed: %s", got)
	}
	traversed := NewDumper(WithRedactTags(), WithDisableStringer(true), WithoutColor(), WithoutHeader()).DumpStr(stringer)
	if strings.Contains(traversed, "secret-stringer") || !strings.Contains(traversed, "<redacted>") {
		t.Fatalf("expected field traversal with Stringer disabled: %s", traversed)
	}
}
