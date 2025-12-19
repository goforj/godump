//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("✔ Examples generated in ./examples/")
}

func run() error {
	root, err := findRoot()
	if err != nil {
		return err
	}

	examplesDir := filepath.Join(root, "examples")
	if err := os.MkdirAll(examplesDir, 0o755); err != nil {
		return err
	}

	modPath, err := modulePath(root)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, root, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName, err := selectPackage(pkgs)
	if err != nil {
		return err
	}

	pkg, ok := pkgs[pkgName]
	if !ok {
		return fmt.Errorf(`package %q not found in %s`, pkgName, root)
	}

	funcs := map[string]*FuncDoc{}

	for filename, file := range pkg.Files {
		if strings.Contains(filename, "_test.go") {
			continue
		}

		for name, fd := range extractFuncDocs(fset, filename, file) {
			if existing, ok := funcs[name]; ok {
				existing.Examples = append(existing.Examples, fd.Examples...)
			} else {
				funcs[name] = fd
			}
		}
	}

	for _, fd := range funcs {
		sort.Slice(fd.Examples, func(i, j int) bool {
			return fd.Examples[i].Line < fd.Examples[j].Line
		})

		if err := writeMain(examplesDir, fd, modPath); err != nil {
			return err
		}

		// Debug / inspection hook (optional)
		//env.Dump(fd)
	}

	return nil
}

func findRoot() (string, error) {
	wd, _ := os.Getwd()
	if fileExists(filepath.Join(wd, "go.mod")) {
		return wd, nil
	}
	parent := filepath.Join(wd, "..")
	if fileExists(filepath.Join(parent, "go.mod")) {
		return filepath.Clean(parent), nil
	}
	return "", fmt.Errorf("could not find project root")
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func modulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}

//
// ------------------------------------------------------------
// Data models
// ------------------------------------------------------------
//

type FuncDoc struct {
	Name        string
	Group       string
	Description string
	Examples    []Example
}

type Example struct {
	FuncName string
	File     string
	Label    string
	Line     int
	Code     string
}

//
// ------------------------------------------------------------
// Example extraction
// ------------------------------------------------------------
//

var exampleHeader = regexp.MustCompile(`(?i)^\s*Example:\s*(.*)$`)
var groupHeader = regexp.MustCompile(`(?i)^\s*@group\s+(.+)$`)

type docLine struct {
	text string
	pos  token.Pos
}

func extractFuncDocs(
	fset *token.FileSet,
	filename string,
	file *ast.File,
) map[string]*FuncDoc {

	out := map[string]*FuncDoc{}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Doc == nil {
			continue
		}

		name := fn.Name.Name
		if !ast.IsExported(name) {
			continue
		}

		out[name] = &FuncDoc{
			Name:        name,
			Group:       extractGroup(fn.Doc),
			Description: extractFuncDescription(fn.Doc),
			Examples:    extractBlocks(fset, filename, name, fn),
		}
	}

	return out
}

func extractGroup(group *ast.CommentGroup) string {
	lines := docLines(group)

	for _, dl := range lines {
		trimmed := strings.TrimSpace(dl.text)
		if m := groupHeader.FindStringSubmatch(trimmed); m != nil {
			return strings.TrimSpace(m[1])
		}
	}

	return "Other"
}

func extractFuncDescription(group *ast.CommentGroup) string {
	lines := docLines(group)
	var desc []string

	for _, dl := range lines {
		trimmed := strings.TrimSpace(dl.text)

		// Stop before Example or @group
		if exampleHeader.MatchString(trimmed) || groupHeader.MatchString(trimmed) {
			break
		}

		if len(desc) == 0 && trimmed == "" {
			continue
		}

		desc = append(desc, dl.text)
	}

	for len(desc) > 0 && strings.TrimSpace(desc[len(desc)-1]) == "" {
		desc = desc[:len(desc)-1]
	}

	return strings.Join(desc, "\n")
}

func docLines(group *ast.CommentGroup) []docLine {
	var lines []docLine

	for _, c := range group.List {
		text := c.Text

		if strings.HasPrefix(text, "//") {
			line := strings.TrimPrefix(text, "//")
			if strings.HasPrefix(line, " ") {
				line = line[1:]
			}
			if strings.HasPrefix(line, "\t") {
				line = line[1:]
			}
			lines = append(lines, docLine{
				text: line,
				pos:  c.Slash,
			})
		}
	}

	return lines
}

func extractBlocks(
	fset *token.FileSet,
	filename, funcName string,
	fn *ast.FuncDecl,
) []Example {

	var out []Example
	lines := docLines(fn.Doc)

	var label string
	var collected []string
	var startLine int
	inExample := false

	flush := func() {
		if len(collected) == 0 {
			return
		}

		out = append(out, Example{
			FuncName: funcName,
			File:     filename,
			Label:    label,
			Line:     startLine,
			Code:     strings.Join(collected, "\n"),
		})

		collected = nil
		label = ""
		inExample = false
	}

	for _, dl := range lines {
		raw := dl.text
		trimmed := strings.TrimSpace(raw)

		if m := exampleHeader.FindStringSubmatch(trimmed); m != nil {
			flush()
			inExample = true
			label = strings.TrimSpace(m[1])
			startLine = fset.Position(dl.pos).Line
			continue
		}

		if !inExample {
			continue
		}

		collected = append(collected, raw)
	}

	flush()
	return out
}

// selectPackage picks the primary package to document.
// Strategy:
//  1. If only one package exists, use it.
//  2. Prefer the non-"main" package with the most files.
//  3. Fall back to the first package alphabetically.
func selectPackage(pkgs map[string]*ast.Package) (string, error) {
	if len(pkgs) == 0 {
		return "", fmt.Errorf("no packages found")
	}

	if len(pkgs) == 1 {
		for name := range pkgs {
			return name, nil
		}
	}

	type candidate struct {
		name  string
		count int
	}

	candidates := make([]candidate, 0, len(pkgs))
	for name, pkg := range pkgs {
		candidates = append(candidates, candidate{
			name:  name,
			count: len(pkg.Files),
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].count == candidates[j].count {
			return candidates[i].name < candidates[j].name
		}
		return candidates[i].count > candidates[j].count
	})

	for _, cand := range candidates {
		if cand.name != "main" {
			return cand.name, nil
		}
	}

	return candidates[0].name, nil
}

//
// ------------------------------------------------------------
// Write ./examples/<func>/main.go
// ------------------------------------------------------------
//

func writeMain(base string, fd *FuncDoc, importPath string) error {
	if len(fd.Examples) == 0 {
		return nil
	}

	if importPath == "" {
		return fmt.Errorf("import path cannot be empty")
	}

	dir := filepath.Join(base, strings.ToLower(fd.Name))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer

	// Build tag
	buf.WriteString("//go:build ignore\n")
	buf.WriteString("// +build ignore\n\n")

	buf.WriteString("package main\n\n")

	imports := map[string]bool{
		importPath: true,
	}

	for _, ex := range fd.Examples {
		if strings.Contains(ex.Code, "fmt.") {
			imports["fmt"] = true
		}
		if strings.Contains(ex.Code, "strings.") {
			imports["strings"] = true
		}
		if strings.Contains(ex.Code, "os.") {
			imports["os"] = true
		}
		if strings.Contains(ex.Code, "context.") {
			imports["context"] = true
		}
		if strings.Contains(ex.Code, "regexp.") {
			imports["regexp"] = true
		}
		if strings.Contains(ex.Code, "redis.") {
			imports["github.com/redis/go-redis/v9"] = true
		}
		if strings.Contains(ex.Code, "time.") {
			imports["time"] = true
		}
		if strings.Contains(ex.Code, "gocron") {
			imports["github.com/go-co-op/gocron/v2"] = true
		}
		if strings.Contains(ex.Code, "scheduler") {
			imports["github.com/goforj/scheduler"] = true
		}
		if strings.Contains(ex.Code, "filepath.") {
			imports["path/filepath"] = true
		}
		if strings.Contains(ex.Code, "godump.") {
			imports["github.com/goforj/godump"] = true
		}
		if strings.Contains(ex.Code, "rand.") {
			imports["crypto/rand"] = true
		}
		if strings.Contains(ex.Code, "base64.") {
			imports["encoding/base64"] = true
		}
	}

	if len(imports) == 1 {
		buf.WriteString("import ")
		for imp := range imports {
			buf.WriteString(fmt.Sprintf("%q", imp))
		}
		buf.WriteString("\n\n")
	} else {
		buf.WriteString("import (\n")
		keys := make([]string, 0, len(imports))
		for k := range imports {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, imp := range keys {
			buf.WriteString("\t\"" + imp + "\"\n")
		}
		buf.WriteString(")\n\n")
	}

	buf.WriteString("func main() {\n")

	// Description
	if fd.Description != "" {
		for _, line := range strings.Split(fd.Description, "\n") {
			buf.WriteString("\t// " + line + "\n")
		}
		buf.WriteString("\n")
	}

	// Examples
	for _, ex := range fd.Examples {
		if ex.Label != "" {
			buf.WriteString("\t// Example: " + ex.Label + "\n")
		}

		ex.Code = strings.TrimLeft(ex.Code, "\n")

		for _, line := range strings.Split(ex.Code, "\n") {
			if strings.TrimSpace(line) == "" {
				buf.WriteString("\n")
			} else {
				buf.WriteString("\t" + line + "\n")
			}
		}
	}

	buf.WriteString("}\n")

	return os.WriteFile(filepath.Join(dir, "main.go"), buf.Bytes(), 0o644)
}
