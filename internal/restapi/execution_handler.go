package restapi

import (
	"net/http"

	"github.com/antgubarev/pet/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type executionHandler struct {
	jobStorage       job.JobStorage
	executionStorage job.ExecutionStorage
	controller       job.ControllerI
}

func NewExecutionHandler(jobStorage job.JobStorage, executionStorage job.ExecutionStorage) *executionHandler {
	return &executionHandler{
		jobStorage:       jobStorage,
		executionStorage: executionStorage,
		controller:       job.NewController(executionStorage),
	}
}

func (eh *executionHandler) SetController(controller job.ControllerI) {
	eh.controller = controller
}

func (eh *executionHandler) StartHandle(c *gin.Context) {
	var jobStartIn JobStartIn
	if err := c.ShouldBindJSON(&jobStartIn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jb, found := eh.findJobByName(c, jobStartIn.Job)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"err": "job not found"})
		return
	}

	execution, err := eh.controller.Start(jb, job.StartArguments{
		Command:   jobStartIn.Command,
		Pid:       jobStartIn.Pid,
		Host:      jobStartIn.Host,
		StartedAt: jobStartIn.StartedAt,
	})
	if err != nil {
		if _, ok := err.(*job.Locked); ok {
			c.JSON(http.StatusLocked, nil)
			return
		}
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": execution.Id.String()})
}

func (eh *executionHandler) FinishHandle(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "invalid id"})
		return
	}

	if err := eh.controller.Finish(uid); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (eh *executionHandler) findJobByName(c *gin.Context, name string) (*job.Job, bool) {
	job, err := eh.jobStorage.GetByName(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return nil, false
	}
	if job == nil {
		c.JSON(http.StatusNotFound, nil)
		return nil, false
	}
	return job, true
}
