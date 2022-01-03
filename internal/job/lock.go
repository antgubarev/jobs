package job

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type LockMode string

const (
	FreeLockMode    LockMode = "free"
	HostLockMode    LockMode = "host"
	ClusterLockMode LockMode = "cluster"
)

type Locker struct{}

func NewLocker() *Locker {
	return &Locker{}
}

type LockArguments struct {
	Pid       *int
	Host      *string
	StartedAt *time.Time
}

func (l *Locker) Lock(j *Job, args LockArguments, executions []Execution) (uuid.UUID, error) {
	if j.LockMode == FreeLockMode {
		return uuid.New(), nil
	}
	if err := l.validateHostPidForLockMode(args.Host, args.Pid, j.LockMode); err != nil {
		return uuid.Nil, err
	}
	for _, exec := range executions {
		if exec.Status != StatusRunning {
			continue
		}
		if j.LockMode == ClusterLockMode && exec.Status == StatusRunning {
			return exec.Id, new(Locked)
		}
		if j.LockMode == HostLockMode && *exec.Host == *args.Host {
			return exec.Id, new(Locked)
		}
	}

	if args.StartedAt == nil {
		t := time.Now()
		args.StartedAt = &t
	}

	return uuid.New(), nil
}

// type UnlockArguments struct {
// 	Pid        *int
// 	Host       *string
// 	Status     ExecutionStatus
// 	Msg        *string
// 	FinishedAt *time.Time
// }

// func (l *Locker) UnLock(jb *Job, args UnlockArguments) error {
// 	if args.Status != StatusSuccessed && args.Status != StatusFailed {
// 		return NewInvalidUnlockArgumentsErr("invalid unlock status")
// 	}
// 	executions, err := l.executionStorage.GetByJobName(jb.Name)
// 	if err != nil {
// 		return err
// 	}
// 	if err := l.validateHostPidForLockMode(args.Host, args.Pid, jb.LockMode); err != nil {
// 		return err
// 	}
// 	for _, exec := range executions {
// 		if jb.LockMode == FreeLockMode && *exec.Host == *args.Host && *exec.Pid == *args.Pid {
// 			if err := l.finishAndStore(&exec, &args); err != nil {
// 				return err
// 			}
// 			return nil
// 		}
// 		if jb.LockMode == ClusterLockMode {
// 			if err := l.finishAndStore(&exec, &args); err != nil {
// 				return err
// 			}
// 			return nil
// 		}
// 		if jb.LockMode == HostLockMode && *exec.Host == *args.Host {
// 			if err := l.finishAndStore(&exec, &args); err != nil {
// 				return err
// 			}
// 			return nil
// 		}
// 	}
// 	return errors.New("execution not found")
// }

func (l *Locker) validateHostPidForLockMode(host *string, pid *int, mode LockMode) error {
	if mode == HostLockMode && host == nil {
		return errors.New("host is required for `host` lock mode")
	}
	return nil
}

// func (l *Locker) finishAndStore(execution *Execution, args *UnlockArguments) error {
// 	execution.Status = args.Status
// 	if args.FinishedAt == nil {
// 		now := time.Now()
// 		args.FinishedAt = &now
// 	}
// 	execution.FinishedAt = args.FinishedAt
// 	if execution.FinishedAt.Before(execution.StartedAt) {
// 		return NewInvalidUnlockArgumentsErr("stop job: finish time must be not after start time")
// 	}
// 	execution.Msg = args.Msg
// 	if err := l.executionStorage.Store(execution); err != nil {
// 		return fmt.Errorf("store execution: %v", err)
// 	}
// 	return nil
// }
