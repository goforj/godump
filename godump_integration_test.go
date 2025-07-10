package godump

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestAnsiInNonTty verifies that no ANSI codes are produced when output is redirected.
func TestAnsiInNonTty(t *testing.T) {
	// ANSI escape character. We expect this to be ABSENT from the output.
	const escape = "\x1b"

	// The source code for the program we're going to run.
	const sourceCode = `
package main
import "github.com/goforj/godump"
func main() {
    s := struct{ Name string }{"test"}
    godump.Dump(s)
}
`

	// Create a temporary directory to avoid package main collision.
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "test_*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tempFile.WriteString(sourceCode); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Run the program using `go run`. By capturing the output, we ensure
	// that the program's stdout is not a TTY.
	cmd := exec.Command("go", "run", tempFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run test program: %v\nOutput:\n%s", err, string(output))
	}

	if strings.Contains(string(output), escape) {
		t.Errorf("expected output to NOT contain ANSI escape codes when not in a TTY, but it did. Output:\n%s", string(output))
	}
}

// TestAnsiInTty verifies that ANSI codes are produced when FORCE_COLOR is set.
func TestAnsiInTty(t *testing.T) {
	// ANSI escape character. We expect this to be PRESENT in the output.
	const escape = "\x1b"

	// The source code for the program we're going to run.
	const sourceCode = `
package main
import "github.com/goforj/godump"
func main() {
    s := struct{ Name string }{"test"}
    godump.Dump(s)
}
`
	// Create a temporary directory to avoid package main collision.
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "test_*.go")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tempFile.WriteString(sourceCode); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Run the program using `go run`. By capturing the output, we ensure
	// that the program's stdout is not a TTY.
	cmd := exec.Command("go", "run", tempFile.Name())

	cmd.Env = append(os.Environ(), "FORCE_COLOR=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run test program: %v\nOutput:\n%s", err, string(output))
	}

	if !strings.Contains(string(output), escape) {
		t.Errorf("expected output to contain ANSI escape codes when FORCE_COLOR is set, but it didn't. Output:\n%s", string(output))
	}
}