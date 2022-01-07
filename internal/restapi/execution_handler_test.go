package restapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/job/mocks"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobStartJobNotFound(t *testing.T) {
	t.Parallel()
	jobStorage := new(mocks.JobStorage)
	jobStorage.On("GetByName", "job").Return(nil, nil)
	executionStorage := new(mocks.ExecutionStorage)

	testWriter := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(jobStorage, executionStorage)
	testRouter := internal.NewTestRouter()
	testRouter.POST("/executions", handler.StartHandle)

	body := `{"job": "job"}`
	req, _ := http.NewRequest("POST", "/executions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 404, testWriter.Code, "%s", testWriter.Body.Bytes())
	jobStorage.AssertExpectations(t)
}

func TestJobStartJobBadRequest(t *testing.T) {
	t.Parallel()
	jobStorage := new(mocks.JobStorage)
	executionStorage := new(mocks.ExecutionStorage)

	testWriter := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(jobStorage, executionStorage)
	testRouter := internal.NewTestRouter()
	testRouter.POST("/executions", handler.StartHandle)

	body := `{"job": "job", "startedAt":"invalid_date"}`
	req, _ := http.NewRequest("POST", "/executions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 400, testWriter.Code, "%s", testWriter.Body.Bytes())
}

func TestStartAllFields(t *testing.T) {
	t.Parallel()
	jobStorage := new(mocks.JobStorage)
	jobStorage.On("GetByName", "job").Return(&job.Job{
		Name:     "job",
		LockMode: job.FreeLockMode,
	}, nil)
	controller := new(mocks.ControllerI)
	controller.On("Start", mock.MatchedBy(func(j *job.Job) bool {
		return j.Name == "job"
	}), mock.MatchedBy(func(args job.StartArguments) bool {
		return *args.Command == "command" &&
			*args.Host == "host1" &&
			*args.Pid == 1 &&
			args.StartedAt.Format(time.RFC3339) == "2021-11-22T11:22:26+03:00"
	})).Return(&job.Execution{}, nil)

	testWriter := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(jobStorage, new(mocks.ExecutionStorage))
	handler.SetController(controller)
	testRouter := internal.NewTestRouter()
	testRouter.POST("/executions", handler.StartHandle)

	body := `{"job":"job","pid":1,"host":"host1","command":"command","startedAt":"2021-11-22T11:22:26+03:00"}`
	req, _ := http.NewRequest("POST", "/executions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 200, testWriter.Code, "%s", testWriter.Body.Bytes())
	controller.AssertExpectations(t)
}

func TestFinishJobNotFound(t *testing.T) {
	t.Parallel()
	executionID := uuid.New()
	controller := new(mocks.ControllerI)
	controller.On("Finish", executionID).Return(nil)

	testWriter := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(new(mocks.JobStorage), new(mocks.ExecutionStorage))
	handler.SetController(controller)
	testRouter := internal.NewTestRouter()
	testRouter.DELETE("/execution/:id", handler.FinishHandle)

	req, _ := http.NewRequest("DELETE", "/execution/"+executionID.String(), nil)
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 200, testWriter.Code, "%s", testWriter.Body.Bytes())
	controller.AssertExpectations(t)
}
