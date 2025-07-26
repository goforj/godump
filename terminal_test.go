package godump

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTerminal(t *testing.T) {
	t.Run("should return false for a regular file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-is-terminal")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		assert.False(t, isTerminal(tmpFile))
	})
} 