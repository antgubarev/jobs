package restapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobCreate(t *testing.T) {
	mockJobStorage := &job.MockJobStorage{}
	mockJobStorage.On("Store",
		mock.MatchedBy(
			func(jobModel *job.Job) bool {
				return jobModel.Name == "name" &&
					jobModel.LockMode == job.HostLockMode
			})).
		Return(nil).Once()
	mockJobStorage.On("GetByName", "name").Return(nil, nil)

	r := gin.Default()
	jobHandler := restapi.NewJobHandler(mockJobStorage)
	r.POST("/job", jobHandler.CreateHandle)

	w := httptest.NewRecorder()
	inRequest := &restapi.CreateJobIn{}
	var data []byte
	_ = json.Unmarshal(data, inRequest)

	body := `{"name":"name","lockMode":"host"}`
	req, _ := http.NewRequest("POST", "/job", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code, "%s", w.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}

func TestJobCreateExists(t *testing.T) {
	mockJobStorage := &job.MockJobStorage{}
	mockJobStorage.On("GetByName", "job").Return(&job.Job{}, nil)

	r := gin.Default()
	jobHandler := restapi.NewJobHandler(mockJobStorage)
	r.POST("/job", jobHandler.CreateHandle)

	w := httptest.NewRecorder()

	body := `{"name":"job","lockMode":"host"}`
	req, _ := http.NewRequest("POST", "/job", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code, "%s", w.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}

func TestJobDelete(t *testing.T) {
	mockJobStorage := &job.MockJobStorage{}
	mockJobStorage.On("GetByName",
		mock.MatchedBy(func(name string) bool {
			return name == "name"
		})).
		Return(&job.Job{
			Name: "name",
		}, nil).Once()

	mockJobStorage.On("DeleteByName", mock.MatchedBy(func(name string) bool {
		return name == "name"
	})).Return(nil).Once()

	r := gin.Default()
	jobHandler := restapi.NewJobHandler(mockJobStorage)
	r.DELETE("/job/:name", jobHandler.DeleteHandle)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/job/name", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "%s", w.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}
