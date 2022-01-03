package job

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

//go:generate mockery --case underscore --inpackage --name ControllerI
type ControllerI interface {
	Start(j *Job, args StartArguments) (*Execution, error)
	Finish(id uuid.UUID) error
}

type Controller struct {
	executionStorage ExecutionStorage
	locker           *Locker
}

func NewController(executionStorage ExecutionStorage) *Controller {
	return &Controller{
		executionStorage: executionStorage,
		locker:           NewLocker(),
	}
}

type StartArguments struct {
	Command   *string
	Pid       *int
	Host      *string
	StartedAt *time.Time
}

func (e *Controller) Start(j *Job, args StartArguments) (*Execution, error) {
	executions, err := e.executionStorage.GetByJobName(j.Name)
	if err != nil {
		return nil, fmt.Errorf("controller start: %v", err)
	}
	id, err := e.locker.Lock(j, LockArguments{
		Pid:       args.Pid,
		Host:      args.Host,
		StartedAt: args.StartedAt,
	}, executions)
	if err != nil {
		return nil, fmt.Errorf("controller start: %v", err)
	}

	if args.StartedAt == nil {
		t := time.Now()
		args.StartedAt = &t
	}

	exec := &Execution{
		Job:        j.Name,
		Command:    args.Command,
		Pid:        args.Pid,
		Host:       args.Host,
		StartedAt:  *args.StartedAt,
		FinishedAt: nil,
		Status:     StatusRunning,
		Msg:        nil,
	}
	exec.Id = id
	if err := e.executionStorage.Store(exec); err != nil {
		return nil, fmt.Errorf("controller start: %v", err)
	}
	return exec, nil
}

func (e *Controller) Finish(id uuid.UUID) error {
	execution, err := e.executionStorage.GetById(id)
	if err != nil {
		return fmt.Errorf("finish: %v", err)
	}
	if err := e.executionStorage.Delete(&execution); err != nil {
		return fmt.Errorf("finish: %v", err)
	}
	return nil
}
