package storage

type Repository interface {
	CreateJob(job *Job) error
	GetJob(id int64) (*Job, error)
	ListPendingJobs(limit int) ([]*Job, error)
	ClaimJob(workerID string) (*Job, error)
	UpdateJobStatus(id int64, status JobStatus, retryCount int) error
}
