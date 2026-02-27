package testassert

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func message(msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 {
		return ""
	}

	if format, ok := msgAndArgs[0].(string); ok {
		if len(msgAndArgs) > 1 {
			return fmt.Sprintf(format, msgAndArgs[1:]...)
		}
		return format
	}

	return fmt.Sprint(msgAndArgs...)
}

func fail(t *testing.T, msg string, msgAndArgs ...any) bool {
	t.Helper()
	if extra := message(msgAndArgs...); extra != "" {
		t.Errorf("%s: %s", msg, extra)
	} else {
		t.Error(msg)
	}
	return false
}

func containsString(container string, item any) bool {
	needle, ok := item.(string)
	if !ok {
		return false
	}
	return strings.Contains(container, needle)
}

// Contains fails the test if s does not contain contains.
func Contains(t *testing.T, s, contains any, msgAndArgs ...any) bool {
	t.Helper()
	str, ok := s.(string)
	if !ok {
		return fail(t, fmt.Sprintf("Contains expects string, got %T", s), msgAndArgs...)
	}
	if !containsString(str, contains) {
		return fail(t, fmt.Sprintf("%q does not contain %q", str, contains), msgAndArgs...)
	}
	return true
}

// NotContains fails the test if s contains contains.
func NotContains(t *testing.T, s, contains any, msgAndArgs ...any) bool {
	t.Helper()
	str, ok := s.(string)
	if !ok {
		return fail(t, fmt.Sprintf("NotContains expects string, got %T", s), msgAndArgs...)
	}
	if containsString(str, contains) {
		return fail(t, fmt.Sprintf("%q unexpectedly contains %q", str, contains), msgAndArgs...)
	}
	return true
}

// Equal fails the test if expected and actual are not deeply equal.
func Equal(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		return fail(t, fmt.Sprintf("not equal\nexpected: %#v\nactual:   %#v", expected, actual), msgAndArgs...)
	}
	return true
}

// True fails the test if value is false.
//
//nolint:revive // bool flag parameter is intentional for assert-style API parity.
func True(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if !value {
		return fail(t, "expected true", msgAndArgs...)
	}
	return true
}

// False fails the test if value is true.
//
//nolint:revive // bool flag parameter is intentional for assert-style API parity.
func False(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if value {
		return fail(t, "expected false", msgAndArgs...)
	}
	return true
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// Nil fails the test if v is not nil.
func Nil(t *testing.T, v any, msgAndArgs ...any) bool {
	t.Helper()
	if !isNil(v) {
		return fail(t, fmt.Sprintf("expected nil, got %#v", v), msgAndArgs...)
	}
	return true
}

// JSONEq fails the test if expected and actual are not equivalent JSON values.
func JSONEq(t *testing.T, expected, actual string, msgAndArgs ...any) bool {
	t.Helper()
	var e any
	var a any
	if err := json.Unmarshal([]byte(expected), &e); err != nil {
		return fail(t, fmt.Sprintf("invalid expected JSON: %v", err), msgAndArgs...)
	}
	if err := json.Unmarshal([]byte(actual), &a); err != nil {
		return fail(t, fmt.Sprintf("invalid actual JSON: %v", err), msgAndArgs...)
	}
	if !reflect.DeepEqual(e, a) {
		return fail(t, fmt.Sprintf("JSON not equal\nexpected: %s\nactual:   %s", expected, actual), msgAndArgs...)
	}
	return true
}

// NoError fails the test if err is non-nil.
func NoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err != nil {
		return fail(t, fmt.Sprintf("unexpected error: %v", err), msgAndArgs...)
	}
	return true
}
