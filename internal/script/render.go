// Package script rewrites a user script into a runnable form for a specific scheduler.
package script

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/log"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

// Render produces a runnable script for the given scheduler:
//   - Preserves the shebang on the first non-empty line (leading blank lines
//     are tolerated: ecFlow-created job files may start with one)
//   - Inserts the scheduler-specific preamble (e.g. #SBATCH lines)
//   - Strips the original `#ORVIX ...` lines
//   - Keeps the rest of the script intact
func Render(src []byte, d *directive.Set, sched scheduler.Scheduler) ([]byte, error) {
	preamble, err := sched.PreambleFor(d)
	if err != nil {
		return nil, err
	}
	log.Debugf("[script] preamble: %d line(s)", len(preamble))

	var (
		shebang string
		bodyBuf bytes.Buffer
		// shebang may still be the next line while only blanks seen so far
		leading = true
	)

	scanner := bufio.NewScanner(bytes.NewReader(src))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if leading {
			if strings.HasPrefix(line, "#!") {
				shebang = line
				leading = false
				log.Debugf("[script] shebang: %s", shebang)
				continue
			}
			if strings.TrimSpace(line) != "" {
				leading = false
			}
		}
		if directive.IsDirectiveLine(line) {
			continue
		}
		bodyBuf.WriteString(line)
		bodyBuf.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	body := bytes.TrimLeft(bodyBuf.Bytes(), "\n")

	var out bytes.Buffer
	if shebang != "" {
		fmt.Fprintln(&out, shebang)
	}
	for _, p := range preamble {
		fmt.Fprintln(&out, p)
	}
	if len(preamble) > 0 && len(body) > 0 {
		fmt.Fprintln(&out)
	}
	out.Write(body)
	log.Debugf("[script] generated script: %d bytes", out.Len())
	return out.Bytes(), nil
}
