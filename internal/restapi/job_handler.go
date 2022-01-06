package restapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/antgubarev/pet/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

var errUndefinedLockMode = errors.New("undefined lock mode")

type JobHandler struct {
	jobStorage job.Storage
}

func NewJobHandler(jobStorage job.Storage) *JobHandler {
	return &JobHandler{jobStorage: jobStorage}
}

func (jh *JobHandler) CreateHandle(ctx *gin.Context) {
	var createJobIn CreateJobIn
	if err := ctx.ShouldBindJSON(&createJobIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})

		return
	}

	existJob, err := jh.jobStorage.GetByName(createJobIn.Name)
	if err != nil {
		glog.Errorf("find job by name: %v", err)
		ctx.JSON(http.StatusInternalServerError, nil)

		return
	}
	if existJob != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"err": fmt.Sprintf("%s already exists", createJobIn.Name)})

		return
	}

	testJob := job.NewJob(createJobIn.Name)
	parsedLockMode, err := parseLockMode(createJobIn.LockMode)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})

		return
	}
	testJob.LockMode = parsedLockMode
	if err := jh.jobStorage.Store(testJob); err != nil {
		glog.Errorf("create handle 500: %v", err)
		ctx.JSON(http.StatusInternalServerError, nil)

		return
	}

	ctx.JSON(http.StatusCreated, nil)
}

func (jh *JobHandler) DeleteHandle(ctx *gin.Context) {
	// TODO delete executions
	_, ok := jh.findJobByName(ctx, ctx.Param("name"))
	if !ok {
		ctx.JSON(http.StatusNotFound, nil)

		return
	}

	if err := jh.jobStorage.DeleteByName(ctx.Param("name")); err != nil {
		ctx.JSON(http.StatusOK, nil)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (jh *JobHandler) findJobByName(ctx *gin.Context, name string) (*job.Job, bool) {
	job, err := jh.jobStorage.GetByName(name)
	if err != nil {
		glog.Errorf("find job by name: %v", err)
		ctx.JSON(http.StatusInternalServerError, nil)

		return nil, false
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, nil)

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
		return "", fmt.Errorf("parseLockMode: %w: %s", errUndefinedLockMode, lockMode)
	}
}
