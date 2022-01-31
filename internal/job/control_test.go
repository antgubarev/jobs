package job_test

import (
	"testing"

	"github.com/antgubarev/jobs/internal"
	"github.com/antgubarev/jobs/internal/job"
	"github.com/antgubarev/jobs/internal/job/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const TestJobName string = "job"

func TestStart(t *testing.T) {
	t.Parallel()
	executionStorage := new(mocks.ExecutionStorage)
	executionStorage.On("GetByJobName", mock.MatchedBy(func(name string) bool {
		return name == TestJobName
	})).Return([]job.Execution{}, nil)
	executionStorage.On("Store", mock.MatchedBy(func(execution *job.Execution) bool {
		return execution.Job == TestJobName &&
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
	t.Parallel()
	executionStorage := new(mocks.ExecutionStorage)
	executionID := uuid.New()
	executionStorage.On("GetByID", executionID).Return(func(uuid.UUID) *job.Execution {
		exec := job.NewRunningExecution(TestJobName)
		exec.SetID(executionID)

		return exec
	}, nil)
	executionStorage.On("Delete", mock.MatchedBy(func(passedExecutionID uuid.UUID) bool {
		return passedExecutionID == executionID
	})).Return(nil)
	controller := job.NewController(executionStorage)
	err := controller.Finish(executionID)
	assert.NoError(t, err)
}
