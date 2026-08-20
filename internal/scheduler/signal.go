package scheduler

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// signalNames maps upper-case signal names (without the "SIG" prefix) to
// their syscall value. It is defined per-platform: signal_unix.go carries
// the standard Linux signal set, signal_windows.go the smaller set that
// syscall defines on Windows.

// ParseSignal converts a user-supplied signal name ("TERM", "sigkill",
// "SIGUSR1") or a signal number ("15") into an os.Signal.
func ParseSignal(s string) (os.Signal, error) {
	name := strings.ToUpper(strings.TrimSpace(s))
	name = strings.TrimPrefix(name, "SIG")
	if sig, ok := signalNames[name]; ok {
		return sig, nil
	}
	n, err := strconv.Atoi(name)
	if err != nil {
		return nil, fmt.Errorf("invalid signal %q", s)
	}
	if n < 1 || n > 64 {
		return nil, fmt.Errorf("signal number %d out of range", n)
	}
	return syscall.Signal(n), nil
}

// signalNumber renders sig as a decimal string for tools that accept numeric
// signals (e.g. scancel --signal).
func signalNumber(sig os.Signal) (string, error) {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return "", fmt.Errorf("unsupported signal type %T", sig)
	}
	return strconv.Itoa(int(s)), nil
}
