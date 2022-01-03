package restapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobsList(t *testing.T) {
	mockJobStorage := &job.MockJobStorage{}
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

	r := gin.Default()
	jobsHandler := restapi.NewJobsHandler(mockJobStorage)
	r.GET("/jobs", jobsHandler.ListHandle)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/jobs", nil)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "%s", w.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}

func TestJobListByName(t *testing.T) {
	mockJobStorage := &job.MockJobStorage{}
	mockJobStorage.On("GetByName", mock.MatchedBy(func(name string) bool {
		return name == "job1"
	})).
		Return(&job.Job{
			Name:     "job1",
			LockMode: job.HostLockMode,
		}, nil).Once()

	r := gin.Default()
	jobsHandler := restapi.NewJobsHandler(mockJobStorage)
	r.GET("/jobs/:name", jobsHandler.ListByNameHandle)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/jobs/job1", nil)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "%s", w.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}
