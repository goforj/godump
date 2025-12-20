package godump

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffStrTypeMismatch(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	out := NewDumper().DiffStr(10, "ten")
	out = stripANSI(out)
	assert.Contains(t, out, "<#diff //")
	assert.Contains(t, out, "- type: int")
	assert.Contains(t, out, "+ type: string")
}

func TestDiffHelpers(t *testing.T) {
	assert.Nil(t, splitLines(""))
	assert.Nil(t, diffLines(nil, nil))

	assert.Equal(t, []string{"", ""}, splitLines("\r\n\r\n"))
	assert.Equal(t, []string{"a", "b"}, splitLines("a\r\nb\r\n"))

	lines := diffLines([]string{"A", "B"}, []string{"A", "C"})
	assert.Equal(t, []diffLine{
		{kind: diffEqual, text: "A"},
		{kind: diffDelete, text: "B"},
		{kind: diffInsert, text: "C"},
	}, lines)

	lines = diffLines([]string{"A", "B"}, []string{"A"})
	assert.Equal(t, []diffLine{
		{kind: diffEqual, text: "A"},
		{kind: diffDelete, text: "B"},
	}, lines)

	lines = diffLines([]string{"A"}, []string{"A", "B"})
	assert.Equal(t, []diffLine{
		{kind: diffEqual, text: "A"},
		{kind: diffInsert, text: "B"},
	}, lines)

	d := NewDumper()
	d.colorizer = colorizeUnstyled

	assert.Equal(t, "  ", d.diffPrefix(diffEqual))
	assert.Equal(t, "- ", d.diffPrefix(diffDelete))
	assert.Equal(t, "+ ", d.diffPrefix(diffInsert))

	assert.Equal(t, "<nil>", d.typeStringForAny(nil))
}

func TestDiffWriterOutput(t *testing.T) {
	var buf bytes.Buffer
	d := NewDumper(WithWriter(&buf))
	d.colorizer = colorizeUnstyled

	d.Diff(map[string]int{"a": 1}, map[string]int{"a": 2})

	out := buf.String()
	out = stripANSI(out)
	assert.Contains(t, out, "<#diff //")
	assert.Contains(t, out, "-    a => 1 #int")
	assert.Contains(t, out, "+    a => 2 #int")
}

func TestDiffTopLevelHelpers(t *testing.T) {
	var buf bytes.Buffer
	oldDefault := defaultDumper
	defaultDumper = NewDumper(WithWriter(&buf))
	defaultDumper.colorizer = colorizeUnstyled
	defer func() { defaultDumper = oldDefault }()

	Diff(1, 2)
	out := buf.String()
	out = stripANSI(out)
	assert.Contains(t, out, "<#diff //")
	assert.Contains(t, out, "- 1 #int")
	assert.Contains(t, out, "+ 2 #int")

	out = DiffStr("a", "b")
	out = stripANSI(out)
	assert.Contains(t, out, "<#diff //")
	assert.Contains(t, out, `- "a" #string`)
	assert.Contains(t, out, `+ "b" #string`)
}

func TestDiffHTML(t *testing.T) {
	html := DiffHTML(map[string]int{"a": 1}, map[string]int{"a": 2})
	assert.Contains(t, html, `<span style="color:`)
	assert.Contains(t, html, "<#diff //")
	assert.Contains(t, html, "-")
	assert.Contains(t, html, "+")
}

func TestDiffStrNoColor(t *testing.T) {
	out := NewDumper(WithoutColor()).DiffStr("a", "b")
	assert.NotContains(t, out, string(ansiEscape))
	assert.Contains(t, out, `- "a" #string`)
	assert.Contains(t, out, `+ "b" #string`)
}

func TestDiffStrNoHeader(t *testing.T) {
	d := NewDumper(WithoutHeader())
	d.colorizer = colorizeUnstyled
	out := d.DiffStr(1, 2)
	out = stripANSI(out)
	assert.NotContains(t, out, "<#diff")
	assert.Contains(t, out, "- 1")
	assert.Contains(t, out, "+ 2")
}

func TestDiffHTMLNoColor(t *testing.T) {
	out := NewDumper(WithoutColor()).DiffHTML("a", "b")
	assert.NotContains(t, out, `<span style="color:`)
	assert.NotContains(t, out, `<span style="background-color:`)
	assert.Contains(t, out, `"a"`)
	assert.Contains(t, out, `"b"`)
}

func TestDiffStrNoHeaderWhenNoCaller(t *testing.T) {
	d := NewDumper()
	d.colorizer = colorizeUnstyled
	d.callerFn = func(skip int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}

	out := d.DiffStr("a", "b")
	assert.False(t, strings.Contains(out, "<#diff //"))
}

func TestEnsureColorizerNoHeader(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	d := NewDumper()
	d.colorizer = nil
	out := d.dumpStrNoHeader("x")
	assert.Contains(t, out, `"x"`)
}

func TestDiffTintHelpers(t *testing.T) {
	d := NewDumper()
	d.colorizer = colorizeUnstyled

	line := d.tintBackgroundLine(string(ansiEscape)+"[33mabc"+string(ansiEscape)+"[0m", colorRedBg, "#3a0d0d")
	assert.Contains(t, line, "abc")

	assert.Equal(t, "abc", stripANSI(string(ansiEscape)+"[31mabc"+string(ansiEscape)+"[0m"))
	assert.Equal(t, "abc", stripANSI("abc"))
	assert.Equal(t, "foo", stripANSI("foo"+string(ansiEscape)))
	assert.Equal(t, "foo", stripANSI("foo"+string(ansiEscape)+"["))
	assert.Equal(t, "foo", stripANSI("foo"+string(ansiEscape)+"[31"))
	assert.Equal(t, "foo", stripANSI("foo"+ansiEraseLine))

	html := `<span style="color:#999">x</span>`
	assert.Equal(t, "x", stripHTMLSpans(html))
	assert.True(t, isHTMLLine(html))

	htmlBroken := `<span style="color:#999"broken`
	assert.Equal(t, htmlBroken, stripHTMLSpans(htmlBroken))

	line = d.tintBackgroundLine(html, colorRedBg, "#3a0d0d")
	assert.Contains(t, line, "x")
}
