package restapi

import (
	"net/http"

	"github.com/antgubarev/pet/internal/job"
	"github.com/gin-gonic/gin"
)

type jobsHandler struct {
	store job.JobStorage
}

func NewJobsHandler(store job.JobStorage) *jobsHandler {
	return &jobsHandler{store: store}
}

func (jsh *jobsHandler) ListHandle(c *gin.Context) {
	jobs, err := jsh.store.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(200, gin.H{"jobs": jobs})
}

func (jsh *jobsHandler) ListByNameHandle(c *gin.Context) {
	// @todo need return one job
	name := c.Param("name")

	job, err := jsh.store.GetByName(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(200, gin.H{"job": job})
}
