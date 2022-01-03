package restapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJobStartJobNotFound(t *testing.T) {
	jobStorage := new(job.MockJobStorage)
	jobStorage.On("GetByName", "job").Return(nil, nil)
	executionStorage := new(job.MockExecutionStorage)

	w := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(jobStorage, executionStorage)
	r := gin.Default()
	r.POST("/executions", handler.StartHandle)

	body := `{"job": "job"}`
	req, _ := http.NewRequest("POST", "/executions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code, "%s", w.Body.Bytes())
	jobStorage.AssertExpectations(t)
}

func TestJobStartJobBadRequest(t *testing.T) {
	jobStorage := new(job.MockJobStorage)
	executionStorage := new(job.MockExecutionStorage)

	w := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(jobStorage, executionStorage)
	r := gin.Default()
	r.POST("/executions", handler.StartHandle)

	body := `{"job": "job", "started_at":"invalid_date"}`
	req, _ := http.NewRequest("POST", "/executions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code, "%s", w.Body.Bytes())
}

func TestStartAllFields(t *testing.T) {
	jobStorage := new(job.MockJobStorage)
	jobStorage.On("GetByName", "job").Return(&job.Job{
		Name:     "job",
		LockMode: job.FreeLockMode,
	}, nil)
	controller := new(job.MockControllerI)
	controller.On("Start", mock.MatchedBy(func(j *job.Job) bool {
		return j.Name == "job"
	}), mock.MatchedBy(func(args job.StartArguments) bool {
		return *args.Command == "command" &&
			*args.Host == "host1" &&
			*args.Pid == 1 &&
			args.StartedAt.Format(time.RFC3339) == "2021-11-22T11:22:26+03:00"
	})).Return(&job.Execution{}, nil)

	w := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(jobStorage, new(job.MockExecutionStorage))
	handler.SetController(controller)
	r := gin.Default()
	r.POST("/executions", handler.StartHandle)

	body := `{"job":"job","pid":1,"host":"host1","command":"command","started_at":"2021-11-22T11:22:26+03:00"}`
	req, _ := http.NewRequest("POST", "/executions", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "%s", w.Body.Bytes())
	controller.AssertExpectations(t)
}

// Unlock
func TestFinishJobNotFound(t *testing.T) {
	id := uuid.New()
	controller := new(job.MockControllerI)
	controller.On("Finish", id).Return(nil)

	w := httptest.NewRecorder()
	handler := restapi.NewExecutionHandler(new(job.MockJobStorage), new(job.MockExecutionStorage))
	handler.SetController(controller)
	r := gin.Default()
	r.DELETE("/execution/:id", handler.FinishHandle)

	req, _ := http.NewRequest("DELETE", "/execution/"+id.String(), nil)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "%s", w.Body.Bytes())
	controller.AssertExpectations(t)
}
