package terminal

import "testing"

// IsTestEnv is a variable that holds the testing.Testing function.
// It's used to determine if the code is running in a test environment,
// and it can be mocked in tests to control behavior.
var IsTestEnv = testing.Testing 