package restapi_test

import (
	"bytes"
	"encoding/json"
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

	testRouter := internal.NewTestRouter()
	jobHandler := restapi.NewJobHandler(mockJobStorage)
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

	testRouter := internal.NewTestRouter()
	jobHandler := restapi.NewJobHandler(mockJobStorage)
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

	testRouter := internal.NewTestRouter()
	jobHandler := restapi.NewJobHandler(mockJobStorage)
	testRouter.DELETE("/job/:name", jobHandler.DeleteHandle)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/job/job", nil)

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "%s", w.Body.Bytes())
	mockJobStorage.AssertExpectations(t)
}
