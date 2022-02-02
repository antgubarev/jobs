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
		writeBadRequestResponse(ctx, err.Error())

		return
	}

	testJob, found := eh.findJobByName(ctx, jobStartIn.Job)
	if !found {
		writeNotFoundResponse(ctx, "job not found")

		return
	}

	if testJob.Status == job.JobStatusPaused {
		writeBadRequestResponse(ctx, "job is paused")

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
		writeInternalServerErrorResponse(ctx, err)

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": execution.ID.String()})
}

func (eh *ExecutionHandler) FinishHandle(ctx *gin.Context) {
	uid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		writeBadRequestResponse(ctx, "invalid id")

		return
	}

	if err := eh.controller.Finish(uid); err != nil {
		writeInternalServerErrorResponse(ctx, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (eh *ExecutionHandler) findJobByName(ctx *gin.Context, name string) (*job.Job, bool) {
	job, err := eh.jobStorage.GetByName(name)
	if err != nil {
		writeInternalServerErrorResponse(ctx, err)

		return nil, false
	}
	if job == nil {
		writeNotFoundResponse(ctx, "job not found")

		return nil, false
	}

	return job, true
}
