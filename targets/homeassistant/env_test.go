package homeassistant

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// setEnv and unsetEnv are os.Setenv and os.Unsetenv with the error checked.
// Unlike t.Setenv they do not restore the previous value, which is what
// these tests want: other tests in this package leave SUPERVISOR_API set.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	require.NoError(t, os.Unsetenv(key))
}
