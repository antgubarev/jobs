package restapi

import (
	"net/http"

	"github.com/antgubarev/pet/internal/job"
	"github.com/gin-gonic/gin"
)

type JobsHandler struct {
	store job.Storage
}

func NewJobsHandler(store job.Storage) *JobsHandler {
	return &JobsHandler{store: store}
}

func (jsh *JobsHandler) ListHandle(ctx *gin.Context) {
	jobs, err := jsh.store.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func (jsh *JobsHandler) ListByNameHandle(ctx *gin.Context) {
	name := ctx.Param("name")

	job, err := jsh.store.GetByName(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"job": job})
}
