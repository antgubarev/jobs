package restapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/jobs/internal"
	"github.com/antgubarev/jobs/internal/boltdb"
	"github.com/antgubarev/jobs/internal/job"
	"github.com/antgubarev/jobs/internal/job/mocks"
	"github.com/antgubarev/jobs/internal/restapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const TestJobName = "job"

func TestJobCreate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		jobStorage       func() *mocks.JobStorage
		executionStorage func() *mocks.ExecutionStorage
		request          string
		body             string
		status           int
	}{
		{
			name: "normal create",
			jobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				mockJobStorage.On("Store",
					mock.MatchedBy(
						func(jobModel *job.Job) bool {
							return jobModel.Name == TestJobName &&
								jobModel.LockMode == job.HostLockMode
						})).
					Return(nil).Once()
				mockJobStorage.On("GetByName", TestJobName).Return(nil, nil)

				return mockJobStorage
			},
			executionStorage: func() *mocks.ExecutionStorage {
				return &mocks.ExecutionStorage{}
			},
			request: "/job",
			body:    `{"name":"job","lockMode":"host","status":"active"}`,
			status:  http.StatusCreated,
		},
		{
			name: "minimal arguments create",
			jobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				mockJobStorage.On("Store",
					mock.MatchedBy(
						func(jobModel *job.Job) bool {
							return jobModel.Name == TestJobName &&
								jobModel.LockMode == job.HostLockMode &&
								jobModel.Status == job.JobStatusActive
						})).
					Return(nil).Once()
				mockJobStorage.On("GetByName", TestJobName).Return(nil, nil).Once()

				return mockJobStorage
			},
			executionStorage: func() *mocks.ExecutionStorage {
				return &mocks.ExecutionStorage{}
			},
			request: "/job",
			body:    `{"name":"job"}`,
			status:  http.StatusCreated,
		},
		{
			name: "job exists",
			jobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				mockJobStorage.On("GetByName", "job").Return(job.NewJob("job"), nil)

				return mockJobStorage
			},
			executionStorage: func() *mocks.ExecutionStorage {
				return &mocks.ExecutionStorage{}
			},
			request: "/job",
			body:    `{"name":"job","lockMode":"host","status":"active"}`,
			status:  http.StatusBadRequest,
		},
		{
			name: "validation error",
			jobStorage: func() *mocks.JobStorage {
				return &mocks.JobStorage{}
			},
			executionStorage: func() *mocks.ExecutionStorage {
				return &mocks.ExecutionStorage{}
			},
			request: "/job",
			body:    `{"lockMode":"undefined","status":"undefined"}`,
			status:  http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			mockJobStorage := testCase.jobStorage()
			mockExecutionStorage := testCase.executionStorage()

			testRouter := internal.NewTestRouter()
			jobHandler := restapi.NewJobHandler(mockJobStorage, mockExecutionStorage)
			testRouter.POST(testCase.request, jobHandler.CreateHandle)

			testWriter := httptest.NewRecorder()
			inRequest := &restapi.CreateJobIn{}
			var data []byte
			_ = json.Unmarshal(data, inRequest)

			req, _ := http.NewRequest("POST", testCase.request, bytes.NewReader([]byte(testCase.body)))
			req.Header.Set("Content-Type", "application/json")

			testRouter.ServeHTTP(testWriter, req)

			assert.Equal(t, testCase.status, testWriter.Code, "%s", testWriter.Body.Bytes())
			mockJobStorage.AssertExpectations(t)
			mockExecutionStorage.AssertExpectations(t)
		})
	}
}

func TestJobDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		caseName             string
		jobName              string
		mockJobStorage       func() *mocks.JobStorage
		mockExecutionStorage func() *mocks.ExecutionStorage
		responseStatus       int
		responseBody         func() *string
	}{
		{
			caseName: "golden delete",
			jobName:  TestJobName,
			mockJobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				mockJobStorage.On("GetByName",
					mock.MatchedBy(func(name string) bool {
						return name == TestJobName
					})).
					Return(&job.Job{
						Name: TestJobName,
					}, nil).Once()

				mockJobStorage.On("DeleteByName", mock.MatchedBy(func(name string) bool {
					return name == TestJobName
				})).Return(nil).Once()

				return mockJobStorage
			},
			mockExecutionStorage: func() *mocks.ExecutionStorage {
				mockExecutionStorage := &mocks.ExecutionStorage{}
				mockExecutionStorage.On("GetByJobName", TestJobName).Return(nil, nil)

				return mockExecutionStorage
			},
			responseStatus: http.StatusOK,
			responseBody:   func() *string { return nil },
		},
		{
			caseName: "job not found",
			jobName:  TestJobName,
			mockJobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				mockJobStorage.On("GetByName",
					mock.MatchedBy(func(name string) bool {
						return name == TestJobName
					})).
					Return(nil, fmt.Errorf("%w: %s", boltdb.ErrJobNotFound, TestJobName)).Once()

				return mockJobStorage
			},
			mockExecutionStorage: func() *mocks.ExecutionStorage {
				return nil
			},
			responseStatus: http.StatusNotFound,
			responseBody: func() *string {
				body := `{"msg":"not found"}`

				return &body
			},
		},
		{
			caseName: "job has running executions",
			jobName:  TestJobName,
			mockJobStorage: func() *mocks.JobStorage {
				mockJobStorage := &mocks.JobStorage{}
				mockJobStorage.On("GetByName",
					mock.MatchedBy(func(name string) bool {
						return name == TestJobName
					})).
					Return(job.NewJob(TestJobName), nil).Once()

				return mockJobStorage
			},
			mockExecutionStorage: func() *mocks.ExecutionStorage {
				mockExecutionStorage := &mocks.ExecutionStorage{}
				mockExecutionStorage.On("GetByJobName", TestJobName).Return([]job.Execution{
					*job.NewRunningExecution(TestJobName),
				}, nil)

				return mockExecutionStorage
			},
			responseStatus: http.StatusLocked,
			responseBody: func() *string {
				body := `{"msg":"stop all job's execution and try again"}`

				return &body
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.caseName, func(t *testing.T) {
			t.Parallel()
			mockJobStorage := testCase.mockJobStorage()
			mockExecutionStorage := testCase.mockExecutionStorage()
			testRouter := internal.NewTestRouter()
			jobHandler := restapi.NewJobHandler(mockJobStorage, mockExecutionStorage)
			testRouter.DELETE("/job/:name", jobHandler.DeleteHandle)

			writer := httptest.NewRecorder()
			req, err := http.NewRequest("DELETE", "/job/"+testCase.jobName, nil)
			if err != nil {
				t.Fatalf("send request %v", err)
			}

			testRouter.ServeHTTP(writer, req)
			assert.Equal(t, testCase.responseStatus, writer.Code)
			if mockJobStorage != nil {
				mockJobStorage.AssertExpectations(t)
			}
			if mockExecutionStorage != nil {
				mockExecutionStorage.AssertExpectations(t)
			}
			body := testCase.responseBody()
			if body != nil {
				assert.Equal(t, *body, writer.Body.String())
			}
		})
	}
}
