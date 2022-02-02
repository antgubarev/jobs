package restapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/antgubarev/jobs/internal"
	"github.com/antgubarev/jobs/internal/job"
	"github.com/antgubarev/jobs/internal/job/mocks"
	"github.com/antgubarev/jobs/internal/restapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartJob(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		jobStorage       func() *mocks.JobStorage
		executionStorage func() *mocks.ExecutionStorage
		controller       func() *mocks.ControllerI
		body             string
		request          string
		status           int
	}{
		{
			name: "job not found",
			jobStorage: func() *mocks.JobStorage {
				jobStorage := new(mocks.JobStorage)
				jobStorage.On("GetByName", "job").Return(nil, nil)

				return jobStorage
			},
			body:    `{"job": "job"}`,
			request: "/executions",
			status:  http.StatusNotFound,
		},
		{
			name:    "validation error",
			body:    `{"job": "job", "startedAt":"invalid_date"}`,
			request: "/executions",
			status:  http.StatusBadRequest,
		},
		{
			name: "all fields in request",
			jobStorage: func() *mocks.JobStorage {
				jobStorage := new(mocks.JobStorage)
				jobStorage.On("GetByName", "job").Return(&job.Job{
					Name:     "job",
					LockMode: job.FreeLockMode,
				}, nil)

				return jobStorage
			},
			controller: func() *mocks.ControllerI {
				controller := new(mocks.ControllerI)
				controller.On("Start", mock.MatchedBy(func(j *job.Job) bool {
					return j.Name == "job"
				}), mock.MatchedBy(func(args job.StartArguments) bool {
					return *args.Command == "command" &&
						*args.Host == "host1" &&
						*args.Pid == 1 &&
						args.StartedAt.Format(time.RFC3339) == "2021-11-22T11:22:26+03:00"
				})).Return(&job.Execution{}, nil)

				return controller
			},
			body:    `{"job":"job","pid":1,"host":"host1","command":"command","startedAt":"2021-11-22T11:22:26+03:00"}`,
			request: "/executions",
			status:  http.StatusOK,
		},
		{
			name: "job is paused",
			jobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				jobToStart := job.NewJob("job")
				jobToStart.Pause()
				mockJobStorage.On("GetByName", "job").Return(jobToStart, nil).Once()

				return mockJobStorage
			},
			body:    `{"job": "job"}`,
			request: "/executions",
			status:  http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockJonStorage := &mocks.JobStorage{}
			if testCase.jobStorage != nil {
				mockJonStorage = testCase.jobStorage()
			}
			mockExecutionStorage := &mocks.ExecutionStorage{}
			if testCase.executionStorage != nil {
				mockExecutionStorage = testCase.executionStorage()
			}

			testWriter := httptest.NewRecorder()
			handler := restapi.NewExecutionHandler(mockJonStorage, mockExecutionStorage)
			mockController := &mocks.ControllerI{}
			if testCase.controller != nil {
				mockController = testCase.controller()
				handler.SetController(mockController)
			}

			testRouter := internal.NewTestRouter()
			testRouter.POST("/executions", handler.StartHandle)

			req, _ := http.NewRequest("POST", testCase.request, bytes.NewReader([]byte(testCase.body)))
			req.Header.Set("Content-Type", "application/json")

			testRouter.ServeHTTP(testWriter, req)

			assert.Equal(t, testCase.status, testWriter.Code, "%s", testWriter.Body.Bytes())
			mockJonStorage.AssertExpectations(t)
			mockExecutionStorage.AssertExpectations(t)
			if testCase.controller != nil {
				mockController.AssertExpectations(t)
			}
		})
	}
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
