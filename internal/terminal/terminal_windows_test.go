//go:build windows

package terminal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTerminal_Windows(t *testing.T) {
	// Mock IsTestEnv to bypass the test environment check
	originalIsTestEnv := IsTestEnv
	IsTestEnv = func() bool { return false }
	defer func() { IsTestEnv = originalIsTestEnv }()

	t.Run("should return false for a regular file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-is-terminal-windows")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// On Windows, GetConsoleMode should fail for a regular file, so isTerminal returns false.
		assert.False(t, IsTerminal(tmpFile))
	})

	t.Run("should return true when in test environment", func(t *testing.T) {
		IsTestEnv = func() bool { return true }
		defer func() { IsTestEnv = originalIsTestEnv }()

		assert.True(t, IsTerminal(os.Stdout))
	})
} 