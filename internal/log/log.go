// Package log provides a minimal debug logger for orvix.
// All debug output goes to stderr by default.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	mu     sync.Mutex
	enabled bool
	out     io.Writer = os.Stderr
)

// SetDebug enables or disables debug logging.
func SetDebug(v bool) {
	mu.Lock()
	enabled = v
	mu.Unlock()
}

// Debug prints a debug message if debugging is enabled.
func Debug(msg string) {
	mu.Lock()
	e := enabled
	mu.Unlock()
	if e {
		fmt.Fprintln(out, msg)
	}
}

// Debugf prints a formatted debug message if debugging is enabled.
func Debugf(format string, args ...any) {
	mu.Lock()
	e := enabled
	mu.Unlock()
	if e {
		fmt.Fprintf(out, format+"\n", args...)
	}
}
