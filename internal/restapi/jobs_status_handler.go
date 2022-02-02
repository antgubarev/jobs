package restapi

import (
	"net/http"

	"github.com/antgubarev/jobs/internal/job"
	"github.com/gin-gonic/gin"
)

type JobStatusHandler struct {
	jobStorage job.Storage
}

func NewJobStatusHandler(jobStorage job.Storage) *JobStatusHandler {
	return &JobStatusHandler{
		jobStorage: jobStorage,
	}
}

func (jsh *JobStatusHandler) Action(ctx *gin.Context) {
	action := ctx.Param("action")
	if action != "start" && action != "pause" {
		writeBadRequestResponse(ctx, "undefined action, can be `start` or `pause`")

		return
	}

	jobName := ctx.Param("name")
	jobToAction, err := jsh.jobStorage.GetByName(jobName)
	if err != nil {
		writeInternalServerErrorResponse(ctx, err)

		return
	}

	if jobToAction == nil {
		writeNotFoundResponse(ctx, "job not found")

		return
	}

	if action == "start" {
		jsh.start(ctx, jobToAction)
	}

	if action == "pause" {
		jsh.pause(ctx, jobToAction)
	}
}

func (jsh *JobStatusHandler) start(ctx *gin.Context, jobToStart *job.Job) {
	if jobToStart.Status == job.JobStatusActive {
		writeBadRequestResponse(ctx, "job is already acive")

		return
	}

	jobToStart.Start()
	if err := jsh.jobStorage.Store(jobToStart); err != nil {
		writeInternalServerErrorResponse(ctx, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (jsh *JobStatusHandler) pause(ctx *gin.Context, jobToPause *job.Job) {
	if jobToPause.Status == job.JobStatusPaused {
		writeBadRequestResponse(ctx, "job is already paused")

		return
	}

	jobToPause.Pause()
	if err := jsh.jobStorage.Store(jobToPause); err != nil {
		writeInternalServerErrorResponse(ctx, err)

		return
	}

	ctx.JSON(http.StatusOK, nil)
}
