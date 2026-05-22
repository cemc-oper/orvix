package scheduler

// JobState is a scheduler-agnostic representation of a job's status.
type JobState string

const (
	StatePending   JobState = "PENDING"
	StateRunning   JobState = "RUNNING"
	StateCompleted JobState = "COMPLETED"
	StateFailed    JobState = "FAILED"
	StateCancelled JobState = "CANCELLED"
	StateTimeout   JobState = "TIMEOUT"
	StateUnknown   JobState = "UNKNOWN"
)

// IsTerminal reports whether the job has reached a final state.
func (s JobState) IsTerminal() bool {
	switch s {
	case StateCompleted, StateFailed, StateCancelled, StateTimeout:
		return true
	}
	return false
}
