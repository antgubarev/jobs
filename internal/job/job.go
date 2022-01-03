package job

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Name     string   `json:"name"`
	LockMode LockMode `json:"lock_mode"`
}

func NewJob(name string) *Job {
	return &Job{
		Name:     name,
		LockMode: HostLockMode,
	}
}

type ExecutionStatus string

const (
	StatusRunning   ExecutionStatus = "running"
	StatusSuccessed ExecutionStatus = "successed"
	StatusFailed    ExecutionStatus = "failed"
)

type Execution struct {
	Id         uuid.UUID       `json:"id"`
	Job        string          `json:"job"`
	Command    *string         `json:"command"`
	Pid        *int            `json:"pid"`
	Host       *string         `json:"host"`
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at"`
	Status     ExecutionStatus `json:"status"`
	Msg        *string         `json:"msg"`
}
