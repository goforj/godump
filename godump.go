package godump

import (
	"encoding/json"
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
	colorLime    = "\033[1;38;5;113m"
	colorCyan    = "\033[38;5;38m"
	colorNote    = "\033[38;5;38m"
	colorRef     = "\033[38;5;247m"
	colorMeta    = "\033[38;5;170m"
	colorDefault = "\033[38;5;208m"
	indentWidth  = 2
)

// Default configuration values for the Dumper.
const (
	defaultDisableStringer = false
	defaultMaxDepth        = 15
	defaultMaxItems        = 100
	defaultMaxStringLen    = 100000
	defaultMaxStackDepth   = 10
	initialCallerSkip      = 2
)

// defaultDumper is the default Dumper instance used by Dump and DumpStr functions.
var defaultDumper = NewDumper()

// exitFunc is a function that can be overridden for testing purposes.
var exitFunc = os.Exit

var (
	nextRefID    = 1
	referenceMap = map[uintptr]int{}
)

// Colorizer is a function type that takes a color code and a string, returning the colorized string.
type Colorizer func(code, str string) string

// colorizeUnstyled returns the string without any colorization.
//
// It satisfies the [Colorizer] interface.
func colorizeUnstyled(code, str string) string {
	return str // No colorization
}

// colorizeANSI colorizes the string using ANSI escape codes.
//
// It satisfies the [Colorizer] interface.
func colorizeANSI(code, str string) string {
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

// colorizeHTML colorizes the string using HTML span tags.
//
// It satisfies the [Colorizer] interface.
func colorizeHTML(code, str string) string {
	return fmt.Sprintf(`<span style="color:%s">%s</span>`, htmlColorMap[code], str)
}

// Dumper holds configuration for dumping structured data.
// It controls depth, item count, and string length limits.
type Dumper struct {
	maxDepth           int
	maxItems           int
	maxStringLen       int
	writer             io.Writer
	skippedStackFrames int
	disableStringer    bool

	// callerFn is used to get the caller information.
	// It defaults to [runtime.Caller], it is here to be overridden for testing purposes.
	callerFn func(skip int) (uintptr, string, int, bool)

	// colorizer is used to apply color formatting to the output.
	colorizer Colorizer
}

// Option defines a functional option for configuring a Dumper.
type Option func(*Dumper) *Dumper

// WithMaxDepth allows to control how deep the structure will be dumped.
// Param n must be 0 or greater or this will be ignored, and default MaxDepth will be 15
func WithMaxDepth(n int) Option {
	return func(d *Dumper) *Dumper {
		if n >= 0 {
			d.maxDepth = n
		}
		return d
	}
}

// WithMaxItems allows to control how many items from an array, slice or maps can be printed.
// Param n must be 0 or greater or this will be ignored, and default MaxItems will be 100
func WithMaxItems(n int) Option {
	return func(d *Dumper) *Dumper {
		if n >= 0 {
			d.maxItems = n
		}
		return d
	}
}

// WithMaxStringLen allows to control how long can printed strings be.
// Param n must be 0 or greater or this will be ignored, and default MaxStringLen will be 100000
func WithMaxStringLen(n int) Option {
	return func(d *Dumper) *Dumper {
		if n >= 0 {
			d.maxStringLen = n
		}
		return d
	}
}

// WithWriter allows to control the io output.
func WithWriter(w io.Writer) Option {
	return func(d *Dumper) *Dumper {
		d.writer = w
		return d
	}
}

// WithSkipStackFrames allows users to skip additional stack frames
// on top of the frames that godump already skips internally.
// This is useful when godump is wrapped in other functions or utilities,
// and the actual call site is deeper in the stack.
func WithSkipStackFrames(n int) Option {
	return func(d *Dumper) *Dumper {
		if n >= 0 {
			d.skippedStackFrames = n
		}
		return d
	}
}

// WithDisableStringer will determine if the stringer value for types that
// implement the stringer interface should be render instead of the actual type.
func WithDisableStringer(b bool) Option {
	return func(d *Dumper) *Dumper {
		d.disableStringer = b
		return d
	}
}

// NewDumper creates a new Dumper with the given options applied.
// Defaults are used for any setting not overridden.
func NewDumper(opts ...Option) *Dumper {
	d := &Dumper{
		maxDepth:        defaultMaxDepth,
		maxItems:        defaultMaxItems,
		maxStringLen:    defaultMaxStringLen,
		disableStringer: defaultDisableStringer,
		writer:          os.Stdout,
		colorizer:       nil, // ensure no detection is made if we don't need it
		callerFn:        runtime.Caller,
	}
	for _, opt := range opts {
		d = opt(d)
	}
	return d
}

// Dump prints the values to stdout with colorized output.
func Dump(vs ...any) {
	defaultDumper.Dump(vs...)
}

// Dump prints the values to stdout with colorized output.
func (d *Dumper) Dump(vs ...any) {
	fmt.Fprint(d.writer, d.DumpStr(vs...))
}

// Fdump writes the formatted dump of values to the given io.Writer.
func Fdump(w io.Writer, vs ...any) {
	NewDumper(WithWriter(w)).Dump(vs...)
}

// DumpStr returns a string representation of the values with colorized output.
func DumpStr(vs ...any) string {
	return defaultDumper.DumpStr(vs...)
}

// DumpStr returns a string representation of the values with colorized output.
func (d *Dumper) DumpStr(vs ...any) string {
	var sb strings.Builder
	d.printDumpHeader(&sb)
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	d.writeDump(tw, vs...)
	tw.Flush()
	return sb.String()
}

// DumpJSONStr pretty-prints values as JSON and returns it as a string.
func (d *Dumper) DumpJSONStr(vs ...any) string {
	if len(vs) == 0 {
		return `{"error": "DumpJSON called with no arguments"}`
	}

	var data any = vs
	if len(vs) == 1 {
		data = vs[0]
	}

	b, err := json.MarshalIndent(data, "", strings.Repeat(" ", indentWidth))
	if err != nil {
		//nolint:errchkjson // fallback handles this manually below
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(errorJSON)
	}
	return string(b)
}

// DumpJSON prints a pretty-printed JSON string to the configured writer.
func (d *Dumper) DumpJSON(vs ...any) {
	output := d.DumpJSONStr(vs...)
	fmt.Fprintln(d.writer, output)
}

// DumpHTML dumps the values as HTML with colorized output.
func DumpHTML(vs ...any) string {
	return defaultDumper.DumpHTML(vs...)
}

// DumpHTML dumps the values as HTML with colorized output.
func (d *Dumper) DumpHTML(vs ...any) string {
	var sb strings.Builder
	sb.WriteString(`<div style='background-color:black;'><pre style="background-color:black; color:white; padding:5px; border-radius: 5px">` + "\n")

	htmlDumper := d.clone()
	htmlDumper.colorizer = colorizeHTML // use HTML colorizer

	sb.WriteString(htmlDumper.DumpStr(vs...))

	sb.WriteString("</pre></div>")
	return sb.String()
}

// DumpJSON dumps the values as a pretty-printed JSON string.
// If there is more than one value, they are dumped as a JSON array.
// It returns an error string if marshaling fails.
func DumpJSON(vs ...any) {
	defaultDumper.DumpJSON(vs...)
}

// DumpJSONStr dumps the values as a JSON string.
func DumpJSONStr(vs ...any) string {
	return defaultDumper.DumpJSONStr(vs...)
}

// Dd is a debug function that prints the values and exits the program.
func Dd(vs ...any) {
	defaultDumper.Dd(vs...)
}

// Dd is a debug function that prints the values and exits the program.
func (d *Dumper) Dd(vs ...any) {
	d.Dump(vs...)
	exitFunc(1)
}

// clone creates a copy of the [Dumper] with the same configuration.
// This is useful for creating a new dumper with the same settings without modifying the original.
func (d *Dumper) clone() *Dumper {
	var newDumper = *d
	return &newDumper
}

// colorize applies the configured [Colorizer] to the string with the given color code.
func (d *Dumper) colorize(code, str string) string {
	if d.colorizer == nil {
		// this avoids detecting color if not needed
		d.colorizer = newColorizer()
	}
	return d.colorizer(code, str)
}

// printDumpHeader prints the header for the dump output, including the file and line number.
func (d *Dumper) printDumpHeader(out io.Writer) {
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

	header := fmt.Sprintf("<#dump // %s:%d", relPath, line)
	fmt.Fprintln(out, d.colorize(colorGray, header))
}

// findFirstNonInternalFrame iterates through the call stack to find the first non-internal frame.
func (d *Dumper) findFirstNonInternalFrame(skip int) (string, int) {
	for i := initialCallerSkip; i < defaultMaxStackDepth; i++ {
		pc, file, line, ok := d.callerFn(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil || !strings.Contains(fn.Name(), "godump") || strings.HasSuffix(file, "_test.go") {
			if skip > 0 {
				skip--
				continue
			}

			return file, line
		}
	}
	return "", 0
}

// formatByteSliceAsHexDump formats a byte slice as a hex dump with ASCII representation.
func (d *Dumper) formatByteSliceAsHexDump(b []byte, indent int) string {
	var sb strings.Builder

	const lineLen = 16
	const asciiStartCol = 50
	const asciiMaxLen = 16

	fieldIndent := strings.Repeat(" ", indent*indentWidth)
	bodyIndent := fieldIndent

	// Header
	sb.WriteString(fmt.Sprintf("([]uint8) (len=%d cap=%d) {\n", len(b), cap(b)))

	for i := 0; i < len(b); i += lineLen {

		end := i + lineLen
		if end > len(b) {
			end = len(b)
		}
		line := b[i:end]

		visibleLen := 0

		// Offset
		offsetStr := fmt.Sprintf("%08x  ", i)
		sb.WriteString(bodyIndent)
		sb.WriteString(d.colorize(colorMeta, offsetStr))
		visibleLen += len(offsetStr)

		// Hex bytes
		for j := 0; j < lineLen; j++ {
			var hexStr string
			if j < len(line) {
				hexStr = fmt.Sprintf("%02x ", line[j])
			} else {
				hexStr = "   "
			}
			if j == 7 {
				hexStr += " "
			}
			sb.WriteString(d.colorize(colorCyan, hexStr))
			visibleLen += len(hexStr)
		}

		// Padding before ASCII
		padding := asciiStartCol - visibleLen
		if padding < 1 {
			padding = 1
		}
		sb.WriteString(strings.Repeat(" ", padding))

		// ASCII section
		sb.WriteString(d.colorize(colorGray, "| "))
		asciiCount := 0
		for _, c := range line {
			ch := "."
			if c >= 32 && c <= 126 {
				ch = string(c)
			}
			sb.WriteString(d.colorize(colorLime, ch))
			asciiCount++
		}
		if asciiCount < asciiMaxLen {
			sb.WriteString(strings.Repeat(" ", asciiMaxLen-asciiCount))
		}
		sb.WriteString(d.colorize(colorGray, " |") + "\n")
	}

	// Closing
	fieldIndent = fieldIndent[:len(fieldIndent)-indentWidth]
	sb.WriteString(fieldIndent + "}")
	return sb.String()
}

func (d *Dumper) writeDump(w io.Writer, vs ...any) {
	referenceMap = map[uintptr]int{} // reset each time
	visited := map[uintptr]bool{}
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		rv = makeAddressable(rv)
		d.printValue(w, rv, 0, visited)
		fmt.Fprintln(w)
	}
}

func (d *Dumper) printValue(w io.Writer, v reflect.Value, indent int, visited map[uintptr]bool) {
	if indent > d.maxDepth {
		fmt.Fprint(w, d.colorize(colorGray, "... (max depth)"))
		return
	}
	if !v.IsValid() {
		fmt.Fprint(w, d.colorize(colorGray, "<invalid>"))
		return
	}

	if s := d.asStringer(v); s != "" {
		fmt.Fprint(w, s)
		return
	}

	switch v.Kind() {
	case reflect.Chan:
		if v.IsNil() {
			fmt.Fprint(w, d.colorize(colorGray, v.Type().String()+"(nil)"))
		} else {
			fmt.Fprintf(w, "%s(%s)", d.colorize(colorGray, v.Type().String()), d.colorize(colorCyan, fmt.Sprintf("%#x", v.Pointer())))
		}
		return
	}

	if isNil(v) {
		typeStr := v.Type().String()
		fmt.Fprintf(w, d.colorize(colorLime, typeStr)+d.colorize(colorGray, "(nil)"))
		return
	}

	if v.Kind() == reflect.Ptr && v.CanAddr() {
		ptr := v.Pointer()
		if id, ok := referenceMap[ptr]; ok {
			fmt.Fprintf(w, d.colorize(colorRef, "↩︎ &%d"), id)
			return
		} else {
			referenceMap[ptr] = nextRefID
			nextRefID++
		}
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		d.printValue(w, v.Elem(), indent, visited)
	case reflect.Struct:
		t := v.Type()
		fmt.Fprintf(w, "%s {", d.colorize(colorGray, "#"+t.String()))
		fmt.Fprintln(w)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldVal := v.Field(i)

			symbol := "+"
			if field.PkgPath != "" {
				symbol = "-"
				fieldVal = forceExported(fieldVal)
			}
			indentPrint(w, indent+1, d.colorize(colorYellow, symbol)+field.Name)
			fmt.Fprint(w, "	=> ")
			d.printValue(w, fieldVal, indent+1, visited)
			fmt.Fprintln(w)
		}
		indentPrint(w, indent, "")
		fmt.Fprint(w, "}")
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(w, d.colorize(colorCyan, fmt.Sprintf("%v", v.Complex())))
	case reflect.UnsafePointer:
		fmt.Fprint(w, d.colorize(colorGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Map:
		fmt.Fprintln(w, "{")
		keys := v.MapKeys()
		for i, key := range keys {
			if i >= d.maxItems {
				indentPrint(w, indent+1, d.colorize(colorGray, "... (truncated)"))
				break
			}
			keyStr := fmt.Sprintf("%v", key.Interface())
			indentPrint(w, indent+1, fmt.Sprintf(" %s => ", d.colorize(colorMeta, keyStr)))
			d.printValue(w, v.MapIndex(key), indent+1, visited)
			fmt.Fprintln(w)
		}
		indentPrint(w, indent, "")
		fmt.Fprint(w, "}")
	case reflect.Slice, reflect.Array:
		// []byte handling
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if v.CanConvert(reflect.TypeOf([]byte{})) { // Check if it can be converted to []byte
				if data, ok := v.Convert(reflect.TypeOf([]byte{})).Interface().([]byte); ok {
					hexDump := d.formatByteSliceAsHexDump(data, indent+1)
					fmt.Fprint(w, d.colorize(colorLime, hexDump))
					break
				}
			}
		}

		// Default rendering for other slices/arrays
		fmt.Fprintln(w, "[")
		for i := 0; i < v.Len(); i++ {
			if i >= d.maxItems {
				indentPrint(w, indent+1, d.colorize(colorGray, "... (truncated)\n"))
				break
			}
			indentPrint(w, indent+1, fmt.Sprintf("%s => ", d.colorize(colorCyan, fmt.Sprintf("%d", i))))
			d.printValue(w, v.Index(i), indent+1, visited)
			fmt.Fprintln(w)
		}
		indentPrint(w, indent, "")
		fmt.Fprint(w, "]")
	case reflect.String:
		str := escapeControl(v.String())
		if utf8.RuneCountInString(str) > d.maxStringLen {
			runes := []rune(str)
			str = string(runes[:d.maxStringLen]) + "…"
		}
		fmt.Fprint(w, d.colorize(colorYellow, `"`)+d.colorize(colorLime, str)+d.colorize(colorYellow, `"`))
	case reflect.Bool:
		if v.Bool() {
			fmt.Fprint(w, d.colorize(colorYellow, "true"))
		} else {
			fmt.Fprint(w, d.colorize(colorGray, "false"))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(w, d.colorize(colorCyan, fmt.Sprint(v.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprint(w, d.colorize(colorCyan, fmt.Sprint(v.Uint())))
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(w, d.colorize(colorCyan, fmt.Sprintf("%f", v.Float())))
	case reflect.Func:
		fmt.Fprint(w, d.colorize(colorGray, v.Type().String()))
	default:
		// unreachable; all reflect.Kind cases are handled
	}
}

// asStringer checks if the value implements fmt.Stringer and returns its string representation.
func (d *Dumper) asStringer(v reflect.Value) string {
	if d.disableStringer {
		return ""
	}

	val := v
	if !val.CanInterface() {
		val = forceExported(val)
	}
	if val.CanInterface() {
		if s, ok := val.Interface().(fmt.Stringer); ok {
			rv := reflect.ValueOf(s)
			if rv.Kind() == reflect.Ptr && rv.IsNil() {
				return d.colorize(colorGray, val.Type().String()+"(nil)")
			}
			return d.colorize(colorLime, s.String()) + d.colorize(colorGray, " #"+val.Type().String())
		}
	}
	return ""
}

// indentPrint prints indented text to the writer.
func indentPrint(w io.Writer, indent int, text string) {
	fmt.Fprint(w, strings.Repeat(" ", indent*indentWidth)+text)
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

// replacer is used to escape control characters in strings.
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

func newColorizer() Colorizer {
	if detectColor() {
		return colorizeANSI
	}
	return colorizeUnstyled
}
