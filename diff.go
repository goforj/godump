package godump

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/tabwriter"
)

// Diff prints a diff between two values to stdout.
func Diff(a, b any) {
	defaultDumper.Diff(a, b)
}

// Diff prints a diff between two values to the configured writer.
func (d *Dumper) Diff(a, b any) {
	fmt.Fprint(d.writer, d.DiffStr(a, b))
}

// DiffStr returns a string diff between two values.
func DiffStr(a, b any) string {
	return defaultDumper.DiffStr(a, b)
}

// DiffStr returns a string diff between two values.
func (d *Dumper) DiffStr(a, b any) string {
	var sb strings.Builder
	d.printDiffHeader(&sb)
	d.ensureColorizer()

	dumps := d.diffDumps(a, b)
	leftLines := splitLines(dumps.left)
	rightLines := splitLines(dumps.right)
	ops := diffLines(leftLines, rightLines)

	for _, op := range ops {
		sb.WriteString(d.diffPrefix(op.kind))
		sb.WriteString(d.diffTintLine(op.text, op.kind))
		sb.WriteString("\n")
	}

	return sb.String()
}

// DiffHTML returns an HTML diff between two values.
func DiffHTML(a, b any) string {
	return defaultDumper.DiffHTML(a, b)
}

// DiffHTML returns an HTML diff between two values.
func (d *Dumper) DiffHTML(a, b any) string {
	var sb strings.Builder
	sb.WriteString(`<div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">` + "\n")

	htmlDumper := d.clone()
	htmlDumper.colorizer = colorizeHTML

	sb.WriteString(htmlDumper.DiffStr(a, b))
	sb.WriteString("</pre></div>")
	return sb.String()
}

type diffDumpPair struct {
	left  string
	right string
}

// diffDumps builds the left and right dump strings, aligning reference ids.
func (d *Dumper) diffDumps(a, b any) diffDumpPair {
	leftDump := d.dumpStrNoHeader(a)
	rightDump := d.dumpStrNoHeader(b)

	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		leftType := fmt.Sprintf("type: %s", d.typeStringForAny(a))
		rightType := fmt.Sprintf("type: %s", d.typeStringForAny(b))
		leftDump = leftType + "\n" + leftDump
		rightDump = rightType + "\n" + rightDump
	}

	return diffDumpPair{left: leftDump, right: rightDump}
}

// dumpStrNoHeader renders a dump without the header line.
func (d *Dumper) dumpStrNoHeader(vs ...any) string {
	d.ensureColorizer()
	state := newDumpState()

	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	d.writeDump(tw, state, vs...)
	tw.Flush()
	return sb.String()
}

// printDiffHeader writes the diff header line when a caller frame is available.
func (d *Dumper) printDiffHeader(out io.Writer) {
	file, line := d.findFirstNonInternalFrame(d.skippedStackFrames)
	if file == "" {
		return
	}

	relPath := file
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil {
			relPath = rel
		}
	}

	header := fmt.Sprintf("<#diff // %s:%d", relPath, line)
	fmt.Fprintln(out, d.colorize(colorGray, header))
}

// typeStringForAny returns a displayable type for a value.
func (d *Dumper) typeStringForAny(v any) string {
	if v == nil {
		return "<nil>"
	}
	return d.getTypeString(reflect.TypeOf(v))
}

type diffKind int

const (
	diffEqual diffKind = iota
	diffDelete
	diffInsert
)

type diffLine struct {
	kind diffKind
	text string
}

// diffLines computes a line-level diff with insert/delete operations.
func diffLines(a, b []string) []diffLine {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}

	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := 0; i <= n; i++ {
		dp[i] = make([]int, m+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				dp[i][j] = dp[i+1][j+1] + 1
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	out := make([]diffLine, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, diffLine{kind: diffEqual, text: a[i]})
			i++
			j++
			continue
		case dp[i+1][j] >= dp[i][j+1]:
			out = append(out, diffLine{kind: diffDelete, text: a[i]})
			i++
		default:
			out = append(out, diffLine{kind: diffInsert, text: b[j]})
			j++
		}
	}

	for i < n {
		out = append(out, diffLine{kind: diffDelete, text: a[i]})
		i++
	}
	for j < m {
		out = append(out, diffLine{kind: diffInsert, text: b[j]})
		j++
	}

	return out
}

// diffPrefix returns the colored diff marker prefix.
func (d *Dumper) diffPrefix(kind diffKind) string {
	switch kind {
	case diffDelete:
		return d.colorize(colorRed, "-") + " "
	case diffInsert:
		return d.colorize(colorGreen, "+") + " "
	default:
		return "  "
	}
}

// diffTintLine tints a full diff line based on change type.
func (d *Dumper) diffTintLine(line string, kind diffKind) string {
	switch kind {
	case diffDelete:
		return d.tintLine(line, colorRed)
	case diffInsert:
		return d.tintLine(line, colorGreen)
	default:
		return line
	}
}

// tintLine strips any existing styling and applies a full-line color.
func (d *Dumper) tintLine(line, colorCode string) string {
	if isHTMLLine(line) {
		return d.colorize(colorCode, stripHTMLSpans(line))
	}
	if strings.Contains(line, "\x1b[") {
		return d.colorize(colorCode, stripANSI(line))
	}
	return d.colorize(colorCode, line)
}

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\x1b' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
	}
	return b.String()
}

// stripHTMLSpans removes color span tags while preserving content.
func stripHTMLSpans(s string) string {
	for {
		start := strings.Index(s, `<span style="color:`)
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], `">`)
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+2:]
	}
	return strings.ReplaceAll(s, "</span>", "")
}

// isHTMLLine reports whether the line contains HTML color spans.
func isHTMLLine(line string) bool {
	return strings.Contains(line, `<span style="color:`)
}

// splitLines splits a string into lines while normalizing CRLF and trimming a trailing newline.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
