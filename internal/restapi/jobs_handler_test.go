package restapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/pet/internal"
	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/job/mocks"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobsList(t *testing.T) {
	t.Parallel()
	mockJobStorage := &mocks.JobStorage{}
	mockJobStorage.On("GetAll").
		Return([]job.Job{
			{
				Name:     "job1",
				LockMode: job.HostLockMode,
			},
			{
				Name:     "job2",
				LockMode: job.ClusterLockMode,
			},
		}, nil).Once()

	testRouter := internal.NewTestRouter()
	jobsHandler := restapi.NewJobsHandler(mockJobStorage)
	testRouter.GET("/jobs", jobsHandler.ListHandle)

	testWriter := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/jobs", nil)
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 200, testWriter.Code, "%s", testWriter.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}

func TestJobListByName(t *testing.T) {
	t.Parallel()
	mockJobStorage := &mocks.JobStorage{}
	mockJobStorage.On("GetByName", mock.MatchedBy(func(name string) bool {
		return name == "job1"
	})).
		Return(&job.Job{
			Name:     "job1",
			LockMode: job.HostLockMode,
		}, nil).Once()

	testRouter := internal.NewTestRouter()
	jobsHandler := restapi.NewJobsHandler(mockJobStorage)
	testRouter.GET("/jobs/:name", jobsHandler.ListByNameHandle)

	testWriter := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/jobs/job1", nil)
	req.Header.Set("Content-Type", "application/json")

	testRouter.ServeHTTP(testWriter, req)

	assert.Equal(t, 200, testWriter.Code, "%s", testWriter.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}
