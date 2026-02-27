package testrequire

import (
	"testing"

	assert "github.com/goforj/godump/internal/testassert"
)

func True(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if !assert.True(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

func NoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if !assert.NoError(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

func Contains(t *testing.T, s any, contains any, msgAndArgs ...any) {
	t.Helper()
	if !assert.Contains(t, s, contains, msgAndArgs...) {
		t.FailNow()
	}
}
