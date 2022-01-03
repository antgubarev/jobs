package restapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antgubarev/pet/internal/job"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateJob(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var in restapi.CreateJobIn
		err := decoder.Decode(&in)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, "job", in.Name)
		assert.Equal(t, "cluster", in.LockMode)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_ = httpClient.JobCreate(&restapi.CreateJobIn{
		Name:     "job",
		LockMode: "cluster",
	})
}

func TestCreateJobInternalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobCreate(&restapi.CreateJobIn{
		Name:     "job",
		LockMode: "cluster",
	})
	assert.Error(t, err)
	assert.Equal(t, "internal server error", err.Error())
}

func TestCreateJobBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"err":"invalig argument"}`))
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobCreate(&restapi.CreateJobIn{
		Name:     "job",
		LockMode: "cluster",
	})
	assert.Error(t, err)
	assert.Equal(t, "send `job create`, bad request invalig argument", err.Error())
}

func TestDeleteJob(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobDelete("job")
	assert.NoError(t, err)
}

func TestDeleteJobNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobDelete("job")
	assert.Error(t, err)
	assert.Equal(t, "job not found", err.Error())
}

func TestDeleteJobInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobDelete("job")
	assert.Error(t, err)
	assert.Equal(t, "internal server error", err.Error())
}

func TestDeleteJobUndefinedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobDelete("job")
	assert.Error(t, err)
	assert.Equal(t, "job delete request: undefined status code 202", err.Error())
}

func TestGetAllJobs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	jobs, err := httpClient.JobsList()
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.JobsList()
	assert.Error(t, err)
	assert.Equal(t, "all job request: undefined status code 202", err.Error())
}

func TestJobByName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jb := &job.Job{
			Name:     "job",
			LockMode: job.ClusterLockMode,
		}
		w.WriteHeader(http.StatusOK)
		jbData, err := json.Marshal(jb)
		if err != nil {
			t.Errorf("marshal job %v", err)
		}
		w.Write(jbData)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	jb, err := httpClient.GetJobByName("job")
	if err != nil {
		t.Errorf("jobbyname %v", err)
	}

	assert.Equal(t, "job", jb.Name)
	assert.Equal(t, job.ClusterLockMode, jb.LockMode)
}

func TestJobByNameNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.GetJobByName("job")
	assert.Error(t, err)
	assert.Equal(t, "send `job by name`, job job not found", err.Error())
}

func TestJobInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.GetJobByName("job")
	assert.Error(t, err)
	assert.Equal(t, "send `job by name` internal server error", err.Error())
}

func TestJobByNameUndefinedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.GetJobByName("job")
	assert.Error(t, err)
	assert.Equal(t, "`job by name` request: undefined status code 202", err.Error())
}

func TestJobStart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.JobStart(&restapi.JobStartIn{Job: "job"})
	assert.NoError(t, err, "job start %v", err)
}

func TestJobStartBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Err string `json:"Err"`
		}{
			Err: "invalid arguments",
		}
		w.WriteHeader(http.StatusBadRequest)
		bodyData, err := json.Marshal(body)
		if err != nil {
			t.Errorf("marshal request data %v", err)
		}
		w.Write(bodyData)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.JobStart(&restapi.JobStartIn{Job: "job"})
	assert.Equal(t, "send `job by name`, bad request invalid arguments", err.Error())
}

func TestJobStartNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.JobStart(&restapi.JobStartIn{Job: "job"})
	assert.Error(t, err)
	assert.Equal(t, "send `job by name`, job job not found", err.Error())
}

func TestJobStartLocked(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusLocked)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	_, err := httpClient.JobStart(&restapi.JobStartIn{Job: "job"})
	assert.Error(t, err)
	assert.Equal(t, "send `job by name`, job job locked", err.Error())
}

func TestJobFinish(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := restapi.NewClientHttp(ts.URL)
	err := httpClient.JobFinish(uuid.New())
	assert.NoError(t, err, "job start %v", err)
}
