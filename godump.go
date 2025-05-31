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
func findFirstNonInternalFrame() (string, int) {
	for i := 2; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
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

func callerLocation(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	return file, line
}

func writeDump(tw *tabwriter.Writer, vs ...any) {
	referenceMap = map[uintptr]int{} // reset each time
	visited := map[uintptr]bool{}
	for _, v := range vs {
		printValue(tw, reflect.ValueOf(v), 0, visited)
		fmt.Fprintln(tw)
	}
}

func printValue(tw *tabwriter.Writer, v reflect.Value, indent int, visited map[uintptr]bool) {
	if indent > maxDepth {
		fmt.Fprint(tw, colorize(colorGray, "... (max depth)"))
		return
	}
	if !v.IsValid() {
		fmt.Fprint(tw, colorize(colorGray, "<invalid>"))
		return
	}
	// If value implements fmt.Stringer, use it
	if v.CanInterface() {
		// Skip using String() for reflect.Value so we can dump internals
		if _, skip := v.Interface().(reflect.Value); !skip {
			if s, ok := v.Interface().(fmt.Stringer); ok {
				fmt.Fprint(tw, colorize(colorLime, s.String())+colorize(colorGray, " #"+v.Type().String()))
				return
			}
		}
	}

	// Custom handling for time.Time
	if v.Type().PkgPath() == "time" && v.Type().Name() == "Time" {
		if t, ok := v.Interface().(interface{ String() string }); ok {
			str := t.String()
			fmt.Fprint(tw, colorize(colorLime, str)+colorize(colorGray, " #time.Time"))
			return
		}
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
	case reflect.Chan:
		if v.IsNil() {
			fmt.Fprint(tw, colorize(colorGray, v.Type().String()+"(nil)"))
		} else {
			fmt.Fprintf(tw, "%s(%s)", colorize(colorGray, v.Type().String()), colorize(colorCyan, fmt.Sprintf("%#x", v.Pointer())))
		}
	case reflect.Struct:
		t := v.Type()

		fmt.Fprintf(tw, "%s ", colorize(colorGray, "#"+t.String()))
		fmt.Fprintln(tw)
		for i := range v.NumField() {
			field := t.Field(i)
			fieldVal := v.Field(i)
			symbol := "+"
			if field.PkgPath != "" {
				symbol = "-"
				fieldVal = forceExported(fieldVal)
			}
			indentPrint(tw, indent+1, colorize(colorYellow, symbol)+field.Name)
			fmt.Fprint(tw, "	=> ")
			printValue(tw, fieldVal, indent+1, visited)
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
		fmt.Fprintln(tw, "[")
		for i := range v.Len() {
			if i >= maxItems {
				indentPrint(tw, indent+1, colorize(colorGray, "... (truncated)\n"))
				break
			}
			indentPrint(tw, indent+1, fmt.Sprintf("%s => ", colorize(colorCyan, fmt.Sprintf("%1d", i))))
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
		if v.CanInterface() {
			fmt.Fprint(tw, colorize(colorDefault, fmt.Sprintf("%v", v.Interface())))
		} else {
			fmt.Fprint(tw, colorize(colorGray, "<unreadable>"))
		}
	}
}

func indentPrint(tw *tabwriter.Writer, indent int, text string) {
	fmt.Fprint(tw, strings.Repeat(" ", indent*indentWidth)+text)
}

func forceExported(v reflect.Value) reflect.Value {
	if v.CanInterface() || !v.CanAddr() {
		return v
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

func escapeControl(s string) string {
	replacer := strings.NewReplacer(
		"\n", `\n`,
		"\t", `\t`,
		"\r", `\r`,
		"\v", `\v`,
		"\f", `\f`,
		"\x1b", `\x1b`,
	)
	return replacer.Replace(s)
}

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	return true
}
