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
		sb.WriteString(op.text)
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

func (d *Dumper) diffDumps(a, b any) diffDumpPair {
	prevNextRefID := nextRefID
	defer func() { nextRefID = prevNextRefID }()

	nextRefID = 1
	leftDump := d.dumpStrNoHeader(a)
	nextRefID = 1
	rightDump := d.dumpStrNoHeader(b)

	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		leftType := fmt.Sprintf("type: %s", d.typeStringForAny(a))
		rightType := fmt.Sprintf("type: %s", d.typeStringForAny(b))
		leftDump = leftType + "\n" + leftDump
		rightDump = rightType + "\n" + rightDump
	}

	return diffDumpPair{left: leftDump, right: rightDump}
}

func (d *Dumper) dumpStrNoHeader(vs ...any) string {
	d.ensureColorizer()

	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	d.writeDump(tw, vs...)
	tw.Flush()
	return sb.String()
}

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

func (d *Dumper) diffPrefix(kind diffKind) string {
	switch kind {
	case diffDelete:
		return d.colorize(colorYellow, "-") + " "
	case diffInsert:
		return d.colorize(colorLime, "+") + " "
	default:
		return "  "
	}
}

func splitLines(s string) []string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
