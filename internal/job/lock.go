package job

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LockMode string

const (
	FreeLockMode    LockMode = "free"
	HostLockMode    LockMode = "host"
	ClusterLockMode LockMode = "cluster"
)

var errValidationLockArguments = errors.New("lock argument is invalid")

type Locker struct{}

func NewLocker() *Locker {
	return &Locker{}
}

type LockArguments struct {
	Pid       *int
	Host      *string
	StartedAt *time.Time
}

func (l *Locker) Lock(lJob *Job, args LockArguments, executions []Execution) (uuid.UUID, error) {
	if lJob.LockMode == FreeLockMode {
		return uuid.New(), nil
	}
	if err := l.validateHostPidForLockMode(args.Host, lJob.LockMode); err != nil {
		return uuid.Nil, err
	}
	for _, exec := range executions {
		if exec.Status != StatusRunning {
			continue
		}
		if lJob.LockMode == ClusterLockMode && exec.Status == StatusRunning {
			return exec.ID, new(LockedError)
		}
		if lJob.LockMode == HostLockMode && *exec.Host == *args.Host {
			return exec.ID, new(LockedError)
		}
	}

	if args.StartedAt == nil {
		t := time.Now()
		args.StartedAt = &t
	}

	return uuid.New(), nil
}

func (l *Locker) validateHostPidForLockMode(host *string, mode LockMode) error {
	if mode == HostLockMode && host == nil {
		return fmt.Errorf("%w: host is required for `host` lock mode", errValidationLockArguments)
	}

	return nil
}
