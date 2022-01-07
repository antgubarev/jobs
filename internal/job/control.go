package job

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

//go:generate mockery --case underscore --name ControllerI
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

func (e *Controller) Start(lJob *Job, args StartArguments) (*Execution, error) {
	executions, err := e.executionStorage.GetByJobName(lJob.Name)
	if err != nil {
		return nil, fmt.Errorf("controller start: %w", err)
	}
	executionID, err := e.locker.Lock(lJob, LockArguments{
		Pid:       args.Pid,
		Host:      args.Host,
		StartedAt: args.StartedAt,
	}, executions)
	if err != nil {
		return nil, fmt.Errorf("controller start: %w", err)
	}

	if args.StartedAt == nil {
		t := time.Now()
		args.StartedAt = &t
	}

	exec := *NewRunningExecution(lJob.Name)
	if args.Command != nil {
		exec.SetCommand(*args.Command)
	}
	if args.Pid != nil {
		exec.SetPid(*args.Pid)
	}
	if args.Host != nil {
		exec.SetHost(*args.Host)
	}
	exec.SetStartedAt(*args.StartedAt)
	exec.SetID(executionID)
	if err := e.executionStorage.Store(&exec); err != nil {
		return nil, fmt.Errorf("controller start: %w", err)
	}

	return &exec, nil
}

func (e *Controller) Finish(id uuid.UUID) error {
	execution, err := e.executionStorage.GetByID(id)
	if err != nil {
		return fmt.Errorf("finish: %w", err)
	}
	if err := e.executionStorage.Delete(execution.ID); err != nil {
		return fmt.Errorf("finish: %w", err)
	}

	return nil
}
