package restapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/jobs/internal"
	"github.com/antgubarev/jobs/internal/job"
	"github.com/antgubarev/jobs/internal/job/mocks"
	"github.com/antgubarev/jobs/internal/restapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActionHandler(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name       string
		request    string
		jobStorage func() *mocks.JobStorage
		status     int
	}{
		{
			name:    "undefinaed action",
			request: "/job/my-job/undefined",
			jobStorage: func() *mocks.JobStorage {
				return &mocks.JobStorage{}
			},
			status: http.StatusBadRequest,
		},
		{
			name:    "pause active job",
			request: "/job/my-job/pause",
			jobStorage: func() *mocks.JobStorage {
				jobStorageMock := &mocks.JobStorage{}

				jobStorageMock.On("GetByName", "my-job").
					Return(job.NewJob("my-job"), nil).
					Once()

				jobStorageMock.On("Store", mock.MatchedBy(func(jobToStore *job.Job) bool {
					return jobToStore.Name == "my-job" && jobToStore.Status == job.JobStatusPaused
				})).Return(nil).Once()

				return jobStorageMock
			},
			status: http.StatusOK,
		},
		{
			name:    "job not found",
			request: "/job/my-job/pause",
			jobStorage: func() *mocks.JobStorage {
				jobStorageMock := &mocks.JobStorage{}

				jobStorageMock.On("GetByName", "my-job").
					Return(nil, nil).
					Once()

				return jobStorageMock
			},
			status: http.StatusNotFound,
		},
		{
			name:    "pause paused job",
			request: "/job/my-job/pause",
			jobStorage: func() *mocks.JobStorage {
				jobStorageMock := &mocks.JobStorage{}

				pausedJob := job.NewJob("my-job")
				pausedJob.Pause()
				jobStorageMock.On("GetByName", "my-job").
					Return(pausedJob, nil).
					Once()

				return jobStorageMock
			},
			status: http.StatusBadRequest,
		},
		{
			name:    "start paused job",
			request: "/job/my-job/start",
			jobStorage: func() *mocks.JobStorage {
				jobStorageMock := &mocks.JobStorage{}

				pausedJob := job.NewJob("my-job")
				pausedJob.Pause()
				jobStorageMock.On("GetByName", "my-job").
					Return(pausedJob, nil).
					Once()

				jobStorageMock.On("Store", mock.MatchedBy(func(jobToStore *job.Job) bool {
					return jobToStore.Name == "my-job" && jobToStore.Status == job.JobStatusActive
				})).Return(nil).Once()

				return jobStorageMock
			},
			status: http.StatusOK,
		},
		{
			name:    "start active job",
			request: "/job/my-job/start",
			jobStorage: func() *mocks.JobStorage {
				jobStorageMock := &mocks.JobStorage{}

				jobStorageMock.On("GetByName", "my-job").
					Return(job.NewJob("my-job"), nil).
					Once()

				return jobStorageMock
			},
			status: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			jobStorageMock := testCase.jobStorage()

			jobsStatusHandler := restapi.NewJobStatusHandler(jobStorageMock)
			testRouter := internal.NewTestRouter()
			testRouter.POST("/job/:name/:action", jobsStatusHandler.Action)

			testWriter := httptest.NewRecorder()

			req, err := http.NewRequest("POST", testCase.request, nil)
			if err != nil {
				t.Fatalf("send request %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			testRouter.ServeHTTP(testWriter, req)

			assert.Equal(t, testCase.status, testWriter.Code, testWriter.Body.String())
			jobStorageMock.AssertExpectations(t)
		})
	}
}
