//go:build windows

package godump

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTerminal_Windows(t *testing.T) {
	// Mock isTestEnv to bypass the test environment check
	originalIsTestEnv := isTestEnv
	isTestEnv = func() bool { return false }
	defer func() { isTestEnv = originalIsTestEnv }()

	t.Run("should return false for a regular file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-is-terminal-windows")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		// On Windows, GetConsoleMode should fail for a regular file, so isTerminal returns false.
		assert.False(t, isTerminal(tmpFile))
	})

	t.Run("should return true when in test environment", func(t *testing.T) {
		isTestEnv = func() bool { return true }
		defer func() { isTestEnv = originalIsTestEnv }()
		
		assert.True(t, isTerminal(os.Stdout))
	})
} 