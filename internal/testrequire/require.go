package testrequire

import (
	"testing"

	assert "github.com/goforj/godump/internal/testassert"
)

// True fails the test immediately if value is false.
//
//nolint:revive // bool flag parameter is intentional for require-style API parity.
func True(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if !assert.True(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

// NoError fails the test immediately if err is non-nil.
func NoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if !assert.NoError(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

// Contains fails the test immediately if s does not contain contains.
func Contains(t *testing.T, s, contains any, msgAndArgs ...any) {
	t.Helper()
	if !assert.Contains(t, s, contains, msgAndArgs...) {
		t.FailNow()
	}
}
