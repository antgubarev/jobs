package restapi

import (
	"fmt"
	"net/http"

	"github.com/antgubarev/pet/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type jobHandler struct {
	jobStorage job.JobStorage
}

func NewJobHandler(jobStorage job.JobStorage) *jobHandler {
	return &jobHandler{jobStorage: jobStorage}
}

func (jh *jobHandler) CreateHandle(c *gin.Context) {
	var createJobIn CreateJobIn
	if err := c.ShouldBindJSON(&createJobIn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	existJob, err := jh.jobStorage.GetByName(createJobIn.Name)
	if err != nil {
		glog.Errorf("find job by name: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	if existJob != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("%s already exists", createJobIn.Name)})
		return
	}

	jb := job.NewJob(createJobIn.Name)
	lm, err := parseLockMode(createJobIn.LockMode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	jb.LockMode = lm
	if err := jh.jobStorage.Store(jb); err != nil {
		glog.Errorf("create handle 500: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, nil)
}

func (jh *jobHandler) DeleteHandle(c *gin.Context) {
	// TODO delete executions
	_, ok := jh.findJobByName(c, c.Param("name"))
	if !ok {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	if err := jh.jobStorage.DeleteByName(c.Param("name")); err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(200, nil)
}

func (jh *jobHandler) findJobByName(c *gin.Context, name string) (*job.Job, bool) {
	job, err := jh.jobStorage.GetByName(name)
	if err != nil {
		glog.Errorf("find job by name: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return nil, false
	}
	if job == nil {
		c.JSON(http.StatusNotFound, nil)
		return nil, false
	}
	return job, true
}

func parseLockMode(lockMode string) (job.LockMode, error) {
	switch lockMode {
	case string(job.FreeLockMode):
		return job.FreeLockMode, nil
	case string(job.ClusterLockMode):
		return job.ClusterLockMode, nil
	case string(job.HostLockMode):
		return job.HostLockMode, nil
	default:
		return "", fmt.Errorf("undefined lock mode")
	}
}
