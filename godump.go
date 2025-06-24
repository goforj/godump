package godump

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
	"unsafe"
)

const (
	colorReset   = "\033[0m"
	colorGray    = "\033[90m"
	colorYellow  = "\033[33m"
	colorLime    = "\033[38;5;113m"
	colorCyan    = "\033[38;5;38m"
	colorNote    = "\033[38;5;38m"
	colorRef     = "\033[38;5;247m"
	colorMeta    = "\033[38;5;170m"
	colorDefault = "\033[38;5;208m"
	indentWidth  = 2
)

var exitFunc = os.Exit

var (
	maxDepth     = 15
	maxItems     = 100
	maxStringLen = 100000
	enableColor  = detectColor()
	nextRefID    = 1
	referenceMap = map[uintptr]int{}
)

// Colorizer is a function type that takes a color code and a string, returning the colorized string.
type Colorizer func(code, str string) string

// colorize is the default colorizer function.
var colorize Colorizer = ansiColorize // default

// ansiColorize colorizes the string using ANSI escape codes.
func ansiColorize(code, str string) string {
	if !enableColor {
		return str
	}
	return code + str + colorReset
}

// htmlColorMap maps color codes to HTML colors.
var htmlColorMap = map[string]string{
	colorGray:    "#999",
	colorYellow:  "#ffb400",
	colorLime:    "#80ff80",
	colorNote:    "#40c0ff",
	colorRef:     "#aaa",
	colorMeta:    "#d087d0",
	colorDefault: "#ff7f00",
}

// htmlColorize colorizes the string using HTML span tags.
func htmlColorize(code, str string) string {
	return fmt.Sprintf(`<span style="color:%s">%s</span>`, htmlColorMap[code], str)
}

// Dump prints the values to stdout with colorized output.
func Dump(vs ...any) {
	printDumpHeader(os.Stdout, 3)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	writeDump(tw, vs...)
	tw.Flush()
}

// Fdump writes the formatted dump of values to the given io.Writer.
func Fdump(w io.Writer, vs ...any) {
	printDumpHeader(w, 3)
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	writeDump(tw, vs...)
	tw.Flush()
}

// DumpStr dumps the values as a string with colorized output.
func DumpStr(vs ...any) string {
	var sb strings.Builder
	printDumpHeader(&sb, 3)
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	writeDump(tw, vs...)
	tw.Flush()
	return sb.String()
}

// DumpHTML dumps the values as HTML with colorized output.
func DumpHTML(vs ...any) string {
	prevColorize := ansiColorize
	prevEnable := enableColor
	defer func() {
		colorize = prevColorize
		enableColor = prevEnable
	}()

	// Enable HTML coloring
	colorize = htmlColorize
	enableColor = true

	var sb strings.Builder
	sb.WriteString(`<body style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px"></body>` + "\n")

	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	printDumpHeader(&sb, 3)
	writeDump(tw, vs...)
	tw.Flush()

	sb.WriteString("</pre>")
	return sb.String()
}

// Dd is a debug function that prints the values and exits the program.
func Dd(vs ...any) {
	Dump(vs...)
	exitFunc(1)
}

// printDumpHeader prints the header for the dump output, including the file and line number.
func printDumpHeader(out io.Writer, skip int) {
	file, line := findFirstNonInternalFrame()
	if file == "" {
		return
	}

	relPath := file
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil {
			relPath = rel
		}
	}

	header := fmt.Sprintf("<#dump // %s:%d", relPath, line)
	fmt.Fprintln(out, colorize(colorGray, header))
}

// findFirstNonInternalFrame finds the first non-internal frame in the call stack.
var callerFn = runtime.Caller

// findFirstNonInternalFrame iterates through the call stack to find the first non-internal frame.
func findFirstNonInternalFrame() (string, int) {
	for i := 2; i < 10; i++ {
		pc, file, line, ok := callerFn(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil || !strings.Contains(fn.Name(), "godump") {
			return file, line
		}
	}
	return "", 0
}

// formatByteSliceAsHexDump formats a byte slice as a hex dump with ASCII representation.
func formatByteSliceAsHexDump(b []byte, indent int) string {
	var sb strings.Builder

	const lineLen = 16
	const asciiStartCol = 50
	const asciiMaxLen = 16

	fieldIndent := strings.Repeat(" ", indent*indentWidth)
	bodyIndent := fieldIndent

	// Header
	sb.WriteString(fmt.Sprintf("([]uint8) (len=%d cap=%d) {\n", len(b), cap(b)))

	for i := 0; i < len(b); i += lineLen {
		end := min(i+lineLen, len(b))
		line := b[i:end]

		visibleLen := 0

		// Offset
		offsetStr := fmt.Sprintf("%08x  ", i)
		sb.WriteString(bodyIndent)
		sb.WriteString(colorize(colorMeta, offsetStr))
		visibleLen += len(offsetStr)

		// Hex bytes
		for j := range lineLen {
			var hexStr string
			if j < len(line) {
				hexStr = fmt.Sprintf("%02x ", line[j])
			} else {
				hexStr = "   "
			}
			if j == 7 {
				hexStr += " "
			}
			sb.WriteString(colorize(colorCyan, hexStr))
			visibleLen += len(hexStr)
		}

		// Padding before ASCII
		padding := max(1, asciiStartCol-visibleLen)
		sb.WriteString(strings.Repeat(" ", padding))

		// ASCII section
		sb.WriteString(colorize(colorGray, "| "))
		asciiCount := 0
		for _, c := range line {
			ch := "."
			if c >= 32 && c <= 126 {
				ch = string(c)
			}
			sb.WriteString(colorize(colorLime, ch))
			asciiCount++
		}
		if asciiCount < asciiMaxLen {
			sb.WriteString(strings.Repeat(" ", asciiMaxLen-asciiCount))
		}
		sb.WriteString(colorize(colorGray, " |") + "\n")
	}

	// Closing
	fieldIndent = fieldIndent[:len(fieldIndent)-indentWidth]
	sb.WriteString(fieldIndent + "}")
	return sb.String()
}

// callerLocation returns the file and line number of the caller at the specified skip level.
func callerLocation(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	return file, line
}

// writeDump writes the values to the tabwriter, handling references and indentation.
func writeDump(tw *tabwriter.Writer, vs ...any) {
	referenceMap = map[uintptr]int{} // reset each time
	visited := map[uintptr]bool{}
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		rv = makeAddressable(rv)
		printValue(tw, rv, 0, visited)
		fmt.Fprintln(tw)
	}
}

// printValue recursively prints the value with indentation and handles references.
func printValue(tw *tabwriter.Writer, v reflect.Value, indent int, visited map[uintptr]bool) {
	if indent > maxDepth {
		fmt.Fprint(tw, colorize(colorGray, "... (max depth)"))
		return
	}
	if !v.IsValid() {
		fmt.Fprint(tw, colorize(colorGray, "<invalid>"))
		return
	}

	if s := asStringer(v); s != "" {
		fmt.Fprint(tw, s)
		return
	}

	switch v.Kind() {
	case reflect.Chan:
		if v.IsNil() {
			fmt.Fprint(tw, colorize(colorGray, v.Type().String()+"(nil)"))
		} else {
			fmt.Fprintf(tw, "%s(%s)", colorize(colorGray, v.Type().String()), colorize(colorCyan, fmt.Sprintf("%#x", v.Pointer())))
		}
		return
	}

	if isNil(v) {
		typeStr := v.Type().String()
		fmt.Fprintf(tw, colorize(colorLime, typeStr)+colorize(colorGray, "(nil)"))
		return
	}

	if v.Kind() == reflect.Ptr && v.CanAddr() {
		ptr := v.Pointer()
		if id, ok := referenceMap[ptr]; ok {
			fmt.Fprintf(tw, colorize(colorRef, "↩︎ &%d"), id)
			return
		} else {
			referenceMap[ptr] = nextRefID
			nextRefID++
		}
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		printValue(tw, v.Elem(), indent, visited)
	case reflect.Struct:
		t := v.Type()
		fmt.Fprintf(tw, "%s ", colorize(colorGray, "#"+t.String()))
		fmt.Fprintln(tw)

		for i := range t.NumField() {
			field := t.Field(i)
			fieldVal := v.Field(i)

			symbol := "+"
			if field.PkgPath != "" {
				symbol = "-"
				fieldVal = forceExported(fieldVal)
			}
			indentPrint(tw, indent+1, colorize(colorYellow, symbol)+field.Name)
			fmt.Fprint(tw, "	=> ")
			if s := asStringer(fieldVal); s != "" {
				fmt.Fprint(tw, s)
			} else {
				printValue(tw, fieldVal, indent+1, visited)
			}
			fmt.Fprintln(tw)
		}
		indentPrint(tw, indent, "")
		fmt.Fprint(tw, "}")
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(tw, colorize(colorCyan, fmt.Sprintf("%v", v.Complex())))
	case reflect.UnsafePointer:
		fmt.Fprint(tw, colorize(colorGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Map:
		fmt.Fprintln(tw, "{")
		keys := v.MapKeys()
		for i, key := range keys {
			if i >= maxItems {
				indentPrint(tw, indent+1, colorize(colorGray, "... (truncated)"))
				break
			}
			keyStr := fmt.Sprintf("%v", key.Interface())
			indentPrint(tw, indent+1, fmt.Sprintf(" %s => ", colorize(colorMeta, keyStr)))
			printValue(tw, v.MapIndex(key), indent+1, visited)
			fmt.Fprintln(tw)
		}
		indentPrint(tw, indent, "")
		fmt.Fprint(tw, "}")
	case reflect.Slice, reflect.Array:
		// []byte handling
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if v.CanConvert(reflect.TypeOf([]byte{})) { // Check if it can be converted to []byte
				if data, ok := v.Convert(reflect.TypeOf([]byte{})).Interface().([]byte); ok {
					hexDump := formatByteSliceAsHexDump(data, indent+1)
					fmt.Fprint(tw, colorize(colorLime, hexDump))
					break
				}
			}
		}

		// Default rendering for other slices/arrays
		fmt.Fprintln(tw, "[")
		for i := range v.Len() {
			if i >= maxItems {
				indentPrint(tw, indent+1, colorize(colorGray, "... (truncated)\n"))
				break
			}
			indentPrint(tw, indent+1, fmt.Sprintf("%s => ", colorize(colorCyan, fmt.Sprintf("%d", i))))
			printValue(tw, v.Index(i), indent+1, visited)
			fmt.Fprintln(tw)
		}
		indentPrint(tw, indent, "")
		fmt.Fprint(tw, "]")

	case reflect.String:
		str := escapeControl(v.String())
		if utf8.RuneCountInString(str) > maxStringLen {
			runes := []rune(str)
			str = string(runes[:maxStringLen]) + "…"
		}
		fmt.Fprint(tw, colorize(colorYellow, `"`)+colorize(colorLime, str)+colorize(colorYellow, `"`))
	case reflect.Bool:
		if v.Bool() {
			fmt.Fprint(tw, colorize(colorYellow, "true"))
		} else {
			fmt.Fprint(tw, colorize(colorGray, "false"))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(tw, colorize(colorCyan, fmt.Sprint(v.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprint(tw, colorize(colorCyan, fmt.Sprint(v.Uint())))
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(tw, colorize(colorCyan, fmt.Sprintf("%f", v.Float())))
	case reflect.Func:
		fmt.Fprint(tw, colorize(colorGray, "func(...) {...}"))
	default:
		// unreachable; all reflect.Kind cases are handled
	}
}

// asStringer checks if the value implements fmt.Stringer and returns its string representation.
func asStringer(v reflect.Value) string {
	val := v
	if !val.CanInterface() {
		val = forceExported(val)
	}
	if val.CanInterface() {
		if s, ok := val.Interface().(fmt.Stringer); ok {
			rv := reflect.ValueOf(s)
			if rv.Kind() == reflect.Ptr && rv.IsNil() {
				return colorize(colorGray, val.Type().String()+"(nil)")
			}
			return colorize(colorLime, s.String()) + colorize(colorGray, " #"+val.Type().String())
		}
	}
	return ""
}

// indentPrint prints indented text to the tabwriter.
func indentPrint(tw *tabwriter.Writer, indent int, text string) {
	fmt.Fprint(tw, strings.Repeat(" ", indent*indentWidth)+text)
}

// forceExported returns a value that is guaranteed to be exported, even if it is unexported.
func forceExported(v reflect.Value) reflect.Value {
	if v.CanInterface() {
		return v
	}
	if v.CanAddr() {
		return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	// Final fallback: return original value, even if unexported
	return v
}

// makeAddressable ensures the value is addressable, wrapping structs in pointers if necessary.
func makeAddressable(v reflect.Value) reflect.Value {
	// Already addressable? Do nothing
	if v.CanAddr() {
		return v
	}

	// If it's a struct and not addressable, wrap it in a pointer
	if v.Kind() == reflect.Struct {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		return ptr.Elem()
	}

	return v
}

// isNil checks if the value is nil based on its kind.
func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

var replacer = strings.NewReplacer(
	"\n", `\n`,
	"\t", `\t`,
	"\r", `\r`,
	"\v", `\v`,
	"\f", `\f`,
	"\x1b", `\x1b`,
)

// escapeControl escapes control characters in a string for safe display.
func escapeControl(s string) string {
	return replacer.Replace(s)
}

// detectColor checks environment variables to determine if color output should be enabled.
func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	return true
}
