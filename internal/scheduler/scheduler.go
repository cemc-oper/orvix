// Package scheduler abstracts the job scheduling backend orvix submits to.
package scheduler

import (
	"fmt"

	"github.com/cemc-oper/orvix/internal/directive"
)

// Generator generates a single scheduler directive line from an orvix directive set.
// It returns the generated line and true if the line should be emitted.
type Generator func(d *directive.Set) (string, bool)

// Scheduler is the contract every backend must satisfy.
type Scheduler interface {
	// Name is the canonical scheduler name (e.g. "slurm", "local").
	Name() string
	// PreambleFor converts orvix directives to the scheduler's directive lines
	// (e.g. #SBATCH lines for SLURM). Returned lines have no trailing newline.
	PreambleFor(d *directive.Set) ([]string, error)
	// Submit hands the script at scriptPath to the scheduler and returns a job id
	// together with the command string that was executed.
	// The caller is responsible for having written scriptPath with executable mode.
	Submit(scriptPath string) (jobID string, command string, err error)
	// Status reports the state of a previously submitted job.
	Status(jobID string) (string, error)
	// Kill terminates a running job.
	Kill(jobID string) error
	// NormalizeState converts a scheduler-specific raw state string to a
	// scheduler-agnostic JobState.
	NormalizeState(raw string) JobState
}

// For picks the scheduler implementation based on parsed directives.
func For(d *directive.Set) (Scheduler, error) {
	return ByName(d.Scheduler())
}

// ByName returns the scheduler with the given canonical name.
func ByName(name string) (Scheduler, error) {
	switch name {
	case "", "local":
		return &Local{}, nil
	case "slurm":
		return &SLURM{}, nil
	case "donau":
		return &Donau{}, nil
	}
	return nil, fmt.Errorf("unknown scheduler %q", name)
}
