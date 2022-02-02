package job

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	JobStatusActive = "active"
	JobStatusPaused = "paused"
)

type Job struct {
	Name      string    `json:"name"`
	LockMode  LockMode  `json:"lockMode"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewJob(name string) *Job {
	return &Job{
		Name:      name,
		LockMode:  HostLockMode,
		Status:    JobStatusActive,
		CreatedAt: time.Now(),
	}
}

func (j *Job) Start() {
	j.Status = JobStatusActive
}

func (j *Job) Pause() {
	j.Status = JobStatusPaused
}

type ExecutionStatus string

const (
	StatusRunning   ExecutionStatus = "running"
	StatusSuccessed ExecutionStatus = "successed"
	StatusFailed    ExecutionStatus = "failed"
)

func NewRunningExecution(job string) *Execution {
	return &Execution{
		ID:         uuid.New(),
		Job:        job,
		Command:    nil,
		Pid:        nil,
		Host:       nil,
		StartedAt:  time.Now(),
		FinishedAt: nil,
		Status:     StatusRunning,
		Msg:        nil,
	}
}

type Execution struct {
	ID         uuid.UUID       `json:"id"`
	Job        string          `json:"job"`
	Command    *string         `json:"command"`
	Pid        *int            `json:"pid"`
	Host       *string         `json:"host"`
	StartedAt  time.Time       `json:"startedAt"`
	FinishedAt *time.Time      `json:"finishedAt"`
	Status     ExecutionStatus `json:"status"`
	Msg        *string         `json:"msg"`
}

func (e *Execution) SetID(id uuid.UUID) {
	e.ID = id
}

func (e *Execution) SetCommand(cmd string) {
	e.Command = &cmd
}

func (e *Execution) SetPid(pid int) {
	e.Pid = &pid
}

func (e *Execution) SetHost(host string) {
	e.Host = &host
}

func (e *Execution) SetStartedAt(at time.Time) {
	e.StartedAt = at
}

func (e *Execution) Finish(status ExecutionStatus, timeAt time.Time, msg string) {
	e.Status = status
	e.FinishedAt = &timeAt
	e.Msg = &msg
}
