package scheduler

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSignalNames(t *testing.T) {
	cases := map[string]syscall.Signal{
		"TERM":    syscall.SIGTERM,
		"term":    syscall.SIGTERM,
		"SIGTERM": syscall.SIGTERM,
		"KILL":    syscall.SIGKILL,
		"sigkill": syscall.SIGKILL,
		"HUP":     syscall.SIGHUP,
		"USR1":    syscall.SIGUSR1,
		"USR2":    syscall.SIGUSR2,
	}
	for input, want := range cases {
		got, err := ParseSignal(input)
		require.NoError(t, err, input)
		assert.Equal(t, want, got, input)
	}
}

func TestParseSignalNumbers(t *testing.T) {
	got, err := ParseSignal("15")
	require.NoError(t, err)
	assert.Equal(t, syscall.SIGTERM, got)

	got, err = ParseSignal("9")
	require.NoError(t, err)
	assert.Equal(t, syscall.SIGKILL, got)
}

func TestParseSignalInvalid(t *testing.T) {
	for _, input := range []string{"", "BOGUS", "SIG", "0", "65", "-1", "15x"} {
		_, err := ParseSignal(input)
		assert.Error(t, err, input)
	}
}

func TestSignalNumber(t *testing.T) {
	num, err := signalNumber(syscall.SIGTERM)
	require.NoError(t, err)
	assert.Equal(t, "15", num)

	num, err = signalNumber(syscall.SIGKILL)
	require.NoError(t, err)
	assert.Equal(t, "9", num)
}
