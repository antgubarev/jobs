package restapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/boltdb"
	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/job/mocks"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const TestJobName = "job"

func TestJobCreate(t *testing.T) {
	t.Parallel()
	mockJobStorage := &mocks.JobStorage{}
	mockJobStorage.On("Store",
		mock.MatchedBy(
			func(jobModel *job.Job) bool {
				return jobModel.Name == TestJobName &&
					jobModel.LockMode == job.HostLockMode
			})).
		Return(nil).Once()
	mockJobStorage.On("GetByName", TestJobName).Return(nil, nil)

	mockExecutuonStorage := &mocks.ExecutionStorage{}

	testRouter := internal.NewTestRouter()
	jobHandler := restapi.NewJobHandler(mockJobStorage, mockExecutuonStorage)
	testRouter.POST("/job", jobHandler.CreateHandle)

	testWriter := httptest.NewRecorder()
	inRequest := &restapi.CreateJobIn{}
	var data []byte
	_ = json.Unmarshal(data, inRequest)

	body := `{"name":"job","lockMode":"host"}`
	req, _ := http.NewRequest("POST", "/job", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 201, testWriter.Code, "%s", testWriter.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}

func TestJobCreateExists(t *testing.T) {
	t.Parallel()
	mockJobStorage := &mocks.JobStorage{}
	mockJobStorage.On("GetByName", "job").Return(&job.Job{}, nil)

	mockExecutuonStorage := &mocks.ExecutionStorage{}

	testRouter := internal.NewTestRouter()
	jobHandler := restapi.NewJobHandler(mockJobStorage, mockExecutuonStorage)
	testRouter.POST("/job", jobHandler.CreateHandle)

	testWriter := httptest.NewRecorder()

	body := `{"name":"job","lockMode":"host"}`
	req, _ := http.NewRequest("POST", "/job", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 400, testWriter.Code, "%s", testWriter.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
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
				body := `{"msg":"job not found"}`

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

			w := httptest.NewRecorder()
			req, err := http.NewRequest("DELETE", "/job/"+testCase.jobName, nil)
			if err != nil {
				t.Fatalf("send request %v", err)
			}

			testRouter.ServeHTTP(w, req)
			assert.Equal(t, testCase.responseStatus, w.Code)
			if mockJobStorage != nil {
				mockJobStorage.AssertExpectations(t)
			}
			if mockExecutionStorage != nil {
				mockExecutionStorage.AssertExpectations(t)
			}
			body := testCase.responseBody()
			if body != nil {
				assert.Equal(t, *body, w.Body.String())
			}
		})
	}
}
