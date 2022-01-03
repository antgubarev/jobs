package job_test

import (
	"testing"
	"time"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/job"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStart(t *testing.T) {
	executionStorage := new(job.MockExecutionStorage)
	executionStorage.On("GetByJobName", mock.MatchedBy(func(name string) bool {
		return name == "job"
	})).Return([]job.Execution{}, nil)
	executionStorage.On("Store", mock.MatchedBy(func(execution *job.Execution) bool {
		return execution.Job == "job" &&
			*execution.Command == "command" &&
			*execution.Host == "host" &&
			*execution.Pid == 1
	})).Return(nil)

	controller := job.NewController(executionStorage)
	execution, err := controller.Start(&job.Job{
		Name:     "job",
		LockMode: job.FreeLockMode,
	}, job.StartArguments{
		Command: internal.NewPointerOfString("command"),
		Pid:     internal.NewPointerOfInt(1),
		Host:    internal.NewPointerOfString("host"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "job", execution.Job)
	assert.Equal(t, "command", *execution.Command)
	assert.Equal(t, "host", *execution.Host)
	assert.Equal(t, 1, *execution.Pid)
}

func TestFinish(t *testing.T) {
	executionStorage := new(job.MockExecutionStorage)
	id := uuid.New()
	executionStorage.On("GetById", id).Return(job.Execution{
		Id:        id,
		Job:       "job",
		StartedAt: time.Now(),
		Status:    job.StatusRunning,
	}, nil)
	executionStorage.On("Delete", mock.MatchedBy(func(execution *job.Execution) bool {
		return execution.Job == "job" &&
			execution.Id == id
	})).Return(nil)
	controller := job.NewController(executionStorage)
	err := controller.Finish(id)
	assert.NoError(t, err)
}
