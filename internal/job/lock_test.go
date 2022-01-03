package job_test

import (
	"testing"
	"time"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/job"
	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	cases := []struct {
		name       string
		jb         job.Job
		executions []job.Execution
		lockArgs   job.LockArguments
		err        error
	}{
		{},
		{
			name: "Once at cluster mode and exec first time",
			jb: job.Job{
				Name:     "job1",
				LockMode: job.ClusterLockMode,
			},
			lockArgs: job.LockArguments{
				Pid:       internal.NewPointerOfInt(1),
				Host:      internal.NewPointerOfString("host"),
				StartedAt: nil,
			},
			executions: []job.Execution{},
			err:        nil,
		},
		{
			name: "Once at cluster mode and exec second time",
			jb: job.Job{
				Name:     "job1",
				LockMode: job.ClusterLockMode,
			},
			lockArgs: job.LockArguments{
				Pid:       internal.NewPointerOfInt(1),
				Host:      internal.NewPointerOfString("host1"),
				StartedAt: nil,
			},
			executions: []job.Execution{
				job.Execution{
					Job:       "job1",
					Pid:       internal.NewPointerOfInt(1),
					Host:      internal.NewPointerOfString("host2"),
					StartedAt: time.Now(),
					Status:    job.StatusRunning,
				},
			},
			err: new(job.Locked),
		},
		{
			name: "Once at host mode and exec at another host",
			jb: job.Job{
				Name:     "job1",
				LockMode: job.HostLockMode,
			},
			lockArgs: job.LockArguments{
				Pid:       internal.NewPointerOfInt(1),
				Host:      internal.NewPointerOfString("host1"),
				StartedAt: nil,
			},
			executions: []job.Execution{
				job.Execution{
					Job:       "job1",
					Pid:       internal.NewPointerOfInt(2),
					Host:      internal.NewPointerOfString("host2"),
					StartedAt: time.Now(),
					Status:    job.StatusRunning,
				},
			},
			err: nil,
		},
		{
			name: "Once at host mode and exec at same host",
			jb: job.Job{
				Name:     "job1",
				LockMode: job.HostLockMode,
			},
			lockArgs: job.LockArguments{
				Pid:       internal.NewPointerOfInt(1),
				Host:      internal.NewPointerOfString("host1"),
				StartedAt: nil,
			},
			executions: []job.Execution{
				job.Execution{
					Job:       "job1",
					Pid:       internal.NewPointerOfInt(2),
					Host:      internal.NewPointerOfString("host1"),
					StartedAt: time.Now(),
					Status:    job.StatusRunning,
				},
			},
			err: new(job.Locked),
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			locker := job.NewLocker()
			_, err := locker.Lock(&cs.jb, cs.lockArgs, cs.executions)
			if cs.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, new(job.Locked))
			}
		})
	}
}
