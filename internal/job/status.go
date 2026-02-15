package job

type JobStatus string

const (
	StatusPending    JobStatus = "PENDING"
	StatusInProgress JobStatus = "IN_PROGRESS"
	StatusCompleted  JobStatus = "COMPLETED"
	StatusRetrying   JobStatus = "RETRYING"
	StatusFailed     JobStatus = "FAILED"
)

// state machine: valid transitions
// PENDING -> IN_PROGRESS
// IN_PROGRESS -> COMPLETED
// IN_PROGRESS -> RETRYING
// RETRYING -> IN_PROGRESS
// RETRYING -> FAILED

func IsValidTransition(from, to JobStatus) bool {
	switch from {
	case StatusPending:
		return to == StatusInProgress
	case StatusInProgress:
		return to == StatusCompleted || to == StatusRetrying
	case StatusRetrying:
		return to == StatusInProgress || to == StatusFailed
	default:
		return false
	}
}
