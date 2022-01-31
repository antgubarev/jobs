package job_test

import (
	"testing"

	"github.com/antgubarev/jobs/internal"
	"github.com/antgubarev/jobs/internal/job"
	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name       string
		jb         job.Job
		executions []func() *job.Execution
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
			executions: []func() *job.Execution{},
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
			executions: []func() *job.Execution{
				func() *job.Execution {
					exec := job.NewRunningExecution("job1")
					exec.SetPid(1)
					exec.SetHost("host2")

					return exec
				},
			},
			err: new(job.LockedError),
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
			executions: []func() *job.Execution{
				func() *job.Execution {
					exec := job.NewRunningExecution("job1")
					exec.SetPid(2)
					exec.SetHost("host2")

					return exec
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
			executions: []func() *job.Execution{
				func() *job.Execution {
					exec := job.NewRunningExecution("job1")
					exec.SetPid(2)
					exec.SetHost("host1")

					return exec
				},
			},
			err: new(job.LockedError),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			locker := job.NewLocker()
			var executions []job.Execution
			for _, execFunc := range testCase.executions {
				executions = append(executions, *execFunc())
			}
			_, err := locker.Lock(&testCase.jb, testCase.lockArgs, executions)
			if testCase.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, new(job.LockedError))
			}
		})
	}
}
