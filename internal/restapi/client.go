package restapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/antgubarev/pet/internal/job"
	"github.com/google/uuid"
)

type CreateJobIn struct {
	Name     string `json:"name"`
	LockMode string `json:"lockMode"`
}

type JobStartIn struct {
	Job       string     `json:"job"`
	StartedAt *time.Time `json:"started_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Command   *string    `json:"command"`
	Pid       *int       `json:"pid"`
	Host      *string    `json:"host"`
}

//go:generate mockery --case underscore --inpackage --name Client
type Client interface {
	JobCreate(in *CreateJobIn) error
	JobDelete(name string) error
	JobsList() ([]job.Job, error)
	GetJobByName(name string) (*job.Job, error)
	JobStart(in *JobStartIn) (uuid.UUID, error)
	JobFinish(id uuid.UUID) error
}

type ClientHttp struct {
	baseUrl string
	client  http.Client
}

func NewClientHttp(baseUrl string) Client {
	return &ClientHttp{
		baseUrl: baseUrl,
		client:  http.Client{},
	}
}

func (c *ClientHttp) JobCreate(in *CreateJobIn) error {
	jsonStr, err := json.Marshal(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseUrl+"/job", bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Errorf("create job create request %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send job create request %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		msg, err := parseResponseBodyErr(resp)
		if err != nil {
			return err
		}
		return fmt.Errorf("send `job create`, bad request %s", msg)
	}

	return errors.New("internal server error")
}

func (c *ClientHttp) JobDelete(name string) error {
	req, err := http.NewRequest("DELETE", c.baseUrl+"/job/"+name, nil)
	if err != nil {
		return fmt.Errorf("create job delete request %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send job delete request %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return errors.New("job not found")
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return errors.New("internal server error")
	}

	return fmt.Errorf("job delete request: undefined status code %d", resp.StatusCode)
}

func (c *ClientHttp) JobsList() ([]job.Job, error) {
	req, err := http.NewRequest("GET", c.baseUrl+"/jobs", nil)
	if err != nil {
		return nil, fmt.Errorf("create all jobs request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send all jobs request: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		responseData := struct {
			Jobs []job.Job `json:"jobs"`
		}{}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("parse response body: %v", err)
		}

		if err := json.Unmarshal(body, &responseData); err != nil {
			return nil, fmt.Errorf("all jobs request unmarshal response %v", err)
		}

		return responseData.Jobs, nil
	}

	return nil, fmt.Errorf("all job request: undefined status code %d", resp.StatusCode)
}

func (c *ClientHttp) GetJobByName(name string) (*job.Job, error) {
	req, err := http.NewRequest("GET", c.baseUrl+"/job/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("create `job by name` request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send `jobs by name` request: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("`job by name request` parse response body: %v", err)
		}
		jobRes := &job.Job{}
		if err := json.Unmarshal(respData, jobRes); err != nil {
			return nil, fmt.Errorf("`job by name` request unmarshal response %v", err)
		}

		return jobRes, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("send `job by name`, job %s not found", name)
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("send `job by name` internal server error")
	}

	return nil, fmt.Errorf("`job by name` request: undefined status code %d", resp.StatusCode)
}

func (c *ClientHttp) JobStart(in *JobStartIn) (uuid.UUID, error) {
	inData, err := json.Marshal(in)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("marshal job start arguments: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseUrl+"/executions", bytes.NewBuffer(inData))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create `job start` request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("send `jobs start` request: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		return uuid.UUID{}, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return uuid.UUID{}, fmt.Errorf("send `job by name`, job %s not found", in.Job)
	}

	if resp.StatusCode == http.StatusLocked {
		return uuid.UUID{}, fmt.Errorf("send `job by name`, job %s locked", in.Job)
	}

	if resp.StatusCode == http.StatusBadRequest {
		msg, err := parseResponseBodyErr(resp)
		if err != nil {
			return uuid.UUID{}, err
		}
		return uuid.UUID{}, fmt.Errorf("send `job by name`, bad request %s", msg)
	}

	return uuid.UUID{}, fmt.Errorf("`job by name` request: undefined status code %d", resp.StatusCode)
}

func (c *ClientHttp) JobFinish(id uuid.UUID) error {
	req, err := http.NewRequest("DELETE", c.baseUrl+"/executions/"+id.String(), nil)
	if err != nil {
		return fmt.Errorf("create `job finish` request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send `jobs finish` request: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusInternalServerError {
		msg, err := parseResponseBodyErr(resp)
		if err != nil {
			return err
		}
		return fmt.Errorf("send `job finish`, internal error %s", msg)
	}

	return nil
}

func parseResponseBodyErr(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse response bosy: %v", err)
	}
	response := struct {
		Err string `json:"err"`
	}{Err: ""}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("job create: parse response body: %v", err)
	}

	return response.Err, nil
}
