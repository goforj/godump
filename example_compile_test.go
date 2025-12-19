package godump

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamplesBuild(t *testing.T) {
	examplesDir := "examples"

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("cannot read examples directory: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		// CAPTURE LOOP VARS
		name := e.Name()
		path := filepath.Join(examplesDir, name)

		t.Run(name, func(t *testing.T) {
			t.Parallel() // 🔑 enable concurrency

			if err := buildExampleWithoutTags(path); err != nil {
				t.Fatalf("example %q failed to build:\n%s", name, err)
			}
		})
	}
}

func abs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	return a
}

func buildExampleWithoutTags(exampleDir string) error {
	orig := filepath.Join(exampleDir, "main.go")

	src, err := os.ReadFile(orig)
	if err != nil {
		return fmt.Errorf("read main.go: %w", err)
	}

	clean := stripBuildTags(src)

	tmpDir, err := os.MkdirTemp("", "example-overlay-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, clean, 0644); err != nil {
		return err
	}

	overlay := map[string]any{
		"Replace": map[string]string{
			abs(orig): abs(tmpFile),
		},
	}

	overlayJSON, err := json.Marshal(overlay)
	if err != nil {
		return err
	}

	overlayPath := filepath.Join(tmpDir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlayJSON, 0644); err != nil {
		return err
	}

	cmd := exec.Command(
		"go", "build",
		"-overlay", overlayPath,
		"-o", os.DevNull,
		"./"+exampleDir,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.New(stderr.String())
	}

	return nil
}

func stripBuildTags(src []byte) []byte {
	lines := strings.Split(string(src), "\n")

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "//go:build") ||
			strings.HasPrefix(line, "// +build") ||
			line == "" {
			i++
			continue
		}

		break
	}

	return []byte(strings.Join(lines[i:], "\n"))
}
