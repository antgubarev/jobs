package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/antgubarev/jobs/internal/job"
	"github.com/google/uuid"
)

type CreateJobIn struct {
	Name     string `json:"name" binding:"required"`
	LockMode string `json:"lockMode" binding:"omitempty,oneof=free host cluster"`
	Status   string `json:"status" binding:"omitempty,oneof=active paused"`
}

type JobStartIn struct {
	Job       string     `json:"job"`
	StartedAt *time.Time `json:"startedAt" time_format:"2006-01-02T15:04:05Z07:00"`
	Command   *string    `json:"command"`
	Pid       *int       `json:"pid"`
	Host      *string    `json:"host"`
}

var (
	errWrongResponse       = errors.New("wrong response")
	errJobNotFound         = errors.New("job not found")
	errInternalServerError = errors.New("internal server error")
	errLocked              = errors.New("locked")
)

//go:generate mockery --case underscore --name Client
type Client interface {
	JobCreate(ctx context.Context, in *CreateJobIn) error
	JobDelete(ctx context.Context, name string) error
	JobsList(ctx context.Context) ([]job.Job, error)
	GetJobByName(ctx context.Context, name string) (*job.Job, error)
	JobStart(ctx context.Context, in *JobStartIn) (uuid.UUID, error)
	JobFinish(ctx context.Context, id uuid.UUID) error
}

type ClientHTTP struct {
	baseURL string
	client  http.Client
}

func NewClientHTTP(baseURL string) *ClientHTTP {
	return &ClientHTTP{
		baseURL: baseURL,
		client:  http.Client{},
	}
}

func (c *ClientHTTP) JobCreate(ctx context.Context, in *CreateJobIn) error {
	jsonStr, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("JobCreate marshal in: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/job", bytes.NewBuffer(jsonStr))
	if err != nil {
		return fmt.Errorf("JobCreate create request %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("JobCreate send request: %w", err)
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

		return fmt.Errorf("JobCreate %w: %s", errWrongResponse, msg)
	}

	return fmt.Errorf("internal server error: %w", err)
}

func (c *ClientHTTP) JobDelete(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/job/"+name, nil)
	if err != nil {
		return fmt.Errorf("create job delete request %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send job delete request %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("JobDelete: %w", errJobNotFound)
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return fmt.Errorf("JobDelete: %w", errInternalServerError)
	}

	return fmt.Errorf("JobDelete status %d: %w", resp.StatusCode, errWrongResponse)
}

func (c *ClientHTTP) JobsList(ctx context.Context) ([]job.Job, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/jobs", nil)
	if err != nil {
		return nil, fmt.Errorf("JobList create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JobList send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		responseData := struct {
			Jobs []job.Job `json:"jobs"`
		}{}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("JobList parse response body: %w", err)
		}

		if err := json.Unmarshal(body, &responseData); err != nil {
			return nil, fmt.Errorf("JobList unmarshal response %w", err)
		}

		return responseData.Jobs, nil
	}

	return nil, fmt.Errorf("JobList status %d: %w", resp.StatusCode, errWrongResponse)
}

func (c *ClientHTTP) GetJobByName(ctx context.Context, name string) (*job.Job, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/job/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("GetJobByName create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetJobByName send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("GetJobByName parse response body: %w", err)
		}
		jobRes := &job.Job{}
		if err := json.Unmarshal(respData, jobRes); err != nil {
			return nil, fmt.Errorf("GetJobByName unmarshal response %w", err)
		}

		return jobRes, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("GetJobByName %w", errJobNotFound)
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("GetJobByName %w", errInternalServerError)
	}

	return nil, fmt.Errorf("GetJobByName status %d: %w", resp.StatusCode, errWrongResponse)
}

func (c *ClientHTTP) JobStart(ctx context.Context, in *JobStartIn) (uuid.UUID, error) {
	inData, err := json.Marshal(in)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("marshal job start arguments: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/executions", bytes.NewBuffer(inData))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("JobStart create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("JobStart send request: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		return uuid.UUID{}, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return uuid.UUID{}, fmt.Errorf("JobStart %w", errJobNotFound)
	}

	if resp.StatusCode == http.StatusLocked {
		return uuid.UUID{}, fmt.Errorf("JobStart %w", errLocked)
	}

	if resp.StatusCode == http.StatusBadRequest {
		msg, err := parseResponseBodyErr(resp)
		if err != nil {
			return uuid.UUID{}, err
		}

		return uuid.UUID{}, fmt.Errorf("JobStart %w: %s", errWrongResponse, msg)
	}

	return uuid.UUID{}, fmt.Errorf("JobStart code %d: %w", resp.StatusCode, errWrongResponse)
}

func (c *ClientHTTP) JobFinish(ctx context.Context, id uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/executions/"+id.String(), nil)
	if err != nil {
		return fmt.Errorf("JobFinish create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("JobFinish send request: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusInternalServerError {
		msg, err := parseResponseBodyErr(resp)
		if err != nil {
			return err
		}

		return fmt.Errorf("JobFinish %w: %s", errInternalServerError, msg)
	}

	return nil
}

func parseResponseBodyErr(resp *http.Response) (string, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parseResponseBodyErr: %w", err)
	}
	response := struct {
		Err string `json:"err"`
	}{Err: ""}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("parseResponseBodyErr: %w", err)
	}

	return response.Err, nil
}
