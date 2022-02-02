package restapi

import (
	"log"
	"net/http"

	"github.com/antgubarev/jobs/internal/boltdb"
	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

func NewServer(addr string, boltDB *bbolt.DB) *http.Server {
	router := gin.Default()

	jobStorage, err := boltdb.NewJobStorage(boltDB)
	if err != nil {
		log.Fatal(err)
	}
	executionStorage, err := boltdb.NewExecutionStorage(boltDB)
	if err != nil {
		log.Fatal(err)
	}

	jobsHandler := NewJobsHandler(jobStorage)
	router.GET("/jobs", jobsHandler.ListHandle)

	jobHandler := NewJobHandler(jobStorage, executionStorage)
	router.POST("/job", jobHandler.CreateHandle)
	router.DELETE("/job/:name", jobHandler.DeleteHandle)

	jobStatusHandler := NewJobStatusHandler(jobStorage)
	router.POST("/job/:name/:action", jobStatusHandler.Action)

	executionHandler := NewExecutionHandler(jobStorage, executionStorage)
	router.POST("/executions", executionHandler.StartHandle)
	router.DELETE("/execution/:id", executionHandler.FinishHandle)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return srv
}
