package restapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateJob(t *testing.T) {
	t.Parallel()
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		decoder := json.NewDecoder(request.Body)
		var in restapi.CreateJobIn
		err := decoder.Decode(&in)
		if err != nil {
			t.Error(err)

			return
		}
		assert.Equal(t, "job", in.Name)
		assert.Equal(t, "cluster", in.LockMode)
	}))
	defer testServer.Close()

	httpClient := restapi.NewClientHTTP(testServer.URL)
	_ = httpClient.JobCreate(context.Background(), &restapi.CreateJobIn{
		Name:     "job",
		LockMode: "cluster",
	})
}

func TestCreateJobInternalError(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	err := httpClient.JobCreate(context.Background(), &restapi.CreateJobIn{
		Name:     "job",
		LockMode: "cluster",
	})
	assert.Error(t, err)
}

func TestCreateJobBadRequest(t *testing.T) {
	t.Parallel()
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write([]byte(`{"err":"invalig argument"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer testServer.Close()

	httpClient := restapi.NewClientHTTP(testServer.URL)
	err := httpClient.JobCreate(context.Background(), &restapi.CreateJobIn{
		Name:     "job",
		LockMode: "cluster",
	})
	assert.Error(t, err)
}

func TestDeleteJob(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	err := httpClient.JobDelete(context.Background(), "job")
	assert.NoError(t, err)
}

func TestDeleteJobNotFound(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	err := httpClient.JobDelete(context.Background(), "job")
	assert.Error(t, err)
}

func TestDeleteJobInternalServerError(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	err := httpClient.JobDelete(context.Background(), "job")
	assert.Error(t, err)
}

func TestDeleteJobUndefinedStatus(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	err := httpClient.JobDelete(context.Background(), "job")
	assert.Error(t, err)
}

func TestGetAllJobs(t *testing.T) {
	t.Parallel()
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		jobs := struct {
			Jobs []job.Job `json:"jobs"`
		}{
			Jobs: []job.Job{
				{
					Name:     "job1",
					LockMode: job.ClusterLockMode,
				},
				{
					Name:     "job2",
					LockMode: job.HostLockMode,
				},
			},
		}
		data, err := json.Marshal(jobs)
		if err != nil {
			t.Errorf("marshal jobs: %v", err)
		}
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write(data); err != nil {
			t.Fatal(err)
		}
	}))
	defer testServer.Close()

	httpClient := restapi.NewClientHTTP(testServer.URL)
	jobs, err := httpClient.JobsList(context.Background())
	if err != nil {
		t.Errorf("jobs list: %v", err)
	}

	assert.Equal(t, []job.Job{
		{
			Name:     "job1",
			LockMode: job.ClusterLockMode,
		},
		{
			Name:     "job2",
			LockMode: job.HostLockMode,
		},
	}, jobs)
}

func TestGetAllJobsUndefinedStatus(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	_, err := httpClient.JobsList(context.Background())
	assert.Error(t, err)
}

func TestJobByName(t *testing.T) {
	t.Parallel()
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		jb := &job.Job{
			Name:     "job",
			LockMode: job.ClusterLockMode,
		}
		writer.WriteHeader(http.StatusOK)
		jbData, err := json.Marshal(jb)
		if err != nil {
			t.Errorf("marshal job %v", err)
		}
		if _, err := writer.Write(jbData); err != nil {
			t.Error(err)
		}
	}))
	defer testServer.Close()

	httpClient := restapi.NewClientHTTP(testServer.URL)
	testJob, err := httpClient.GetJobByName(context.Background(), "job")
	if err != nil {
		t.Errorf("jobbyname %v", err)
	}

	assert.Equal(t, "job", testJob.Name)
	assert.Equal(t, job.ClusterLockMode, testJob.LockMode)
}

func TestJobByNameNotFound(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	_, err := httpClient.GetJobByName(context.Background(), "job")
	assert.Error(t, err)
}

func TestJobInternalServerError(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	_, err := httpClient.GetJobByName(context.Background(), "job")
	assert.Error(t, err)
}

func TestJobByNameUndefinedStatus(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	_, err := httpClient.GetJobByName(context.Background(), "job")
	assert.Error(t, err)
}

func TestJobStart(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	_, err := httpClient.JobStart(context.Background(), &restapi.JobStartIn{Job: "job"})
	assert.NoError(t, err, "job start %v", err)
}

func TestJobStartBadRequest(t *testing.T) {
	t.Parallel()
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body := struct {
			Err string `json:"err"`
		}{
			Err: "invalid arguments",
		}
		writer.WriteHeader(http.StatusBadRequest)
		bodyData, err := json.Marshal(body)
		if err != nil {
			t.Errorf("marshal request data %v", err)
		}
		if _, err := writer.Write(bodyData); err != nil {
			t.Fatal(err)
		}
	}))
	defer testServer.Close()

	httpClient := restapi.NewClientHTTP(testServer.URL)
	_, err := httpClient.JobStart(context.Background(), &restapi.JobStartIn{Job: "job"})
	assert.Error(t, err)
}

func jobStartWithResponseCode(t *testing.T, responseCode int) error {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseCode)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	_, err := httpClient.JobStart(context.Background(), &restapi.JobStartIn{Job: "job"})

	return fmt.Errorf("start job with reso[ponse code: %w", err)
}

func TestJobStartNotFound(t *testing.T) {
	t.Parallel()
	err := jobStartWithResponseCode(t, http.StatusNotFound)
	assert.Error(t, err)
}

func TestJobStartLocked(t *testing.T) {
	t.Parallel()
	err := jobStartWithResponseCode(t, http.StatusLocked)
	assert.Error(t, err)
}

func TestJobFinish(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHTTP(ts.URL)
	err := httpClient.JobFinish(context.Background(), uuid.New())
	assert.NoError(t, err, "job start %v", err)
}
