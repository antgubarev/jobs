package restapi

import (
	"net/http"

	"github.com/antgubarev/jobs/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExecutionHandler struct {
	jobStorage       job.Storage
	executionStorage job.ExecutionStorage
	controller       job.ControllerI
}

func NewExecutionHandler(jobStorage job.Storage, executionStorage job.ExecutionStorage) *ExecutionHandler {
	return &ExecutionHandler{
		jobStorage:       jobStorage,
		executionStorage: executionStorage,
		controller:       job.NewController(executionStorage),
	}
}

func (eh *ExecutionHandler) SetController(controller job.ControllerI) {
	eh.controller = controller
}

func (eh *ExecutionHandler) StartHandle(ctx *gin.Context) {
	var jobStartIn JobStartIn
	if err := ctx.ShouldBindJSON(&jobStartIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	testJob, found := eh.findJobByName(ctx, jobStartIn.Job)
	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{"err": "job not found"})

		return
	}

	execution, err := eh.controller.Start(testJob, job.StartArguments{
		Command:   jobStartIn.Command,
		Pid:       jobStartIn.Pid,
		Host:      jobStartIn.Host,
		StartedAt: jobStartIn.StartedAt,
	})
	if err != nil {
		if _, ok := err.(*job.LockedError); ok {
			ctx.JSON(http.StatusLocked, nil)

			return
		}
		ctx.JSON(http.StatusInternalServerError, nil)

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": execution.ID.String()})
}

func (eh *ExecutionHandler) FinishHandle(ctx *gin.Context) {
	uid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"err": "invalid id"})

		return
	}

	if err := eh.controller.Finish(uid); err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (eh *ExecutionHandler) findJobByName(ctx *gin.Context, name string) (*job.Job, bool) {
	job, err := eh.jobStorage.GetByName(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)

		return nil, false
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, nil)

		return nil, false
	}

	return job, true
}
