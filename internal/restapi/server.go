package restapi

import (
	"log"
	"net/http"

	"github.com/antgubarev/pet/internal/boltdb"
	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

func NewServer(addr string, db *bbolt.DB) *http.Server {
	router := gin.Default()

	jobStorage, err := boltdb.NewJobStorage(db)
	if err != nil {
		log.Fatal(err)
	}
	executionStorage, err := boltdb.NewExecutionStorage(db)
	if err != nil {
		log.Fatal(err)
	}

	jobsHandler := NewJobsHandler(jobStorage)
	router.GET("/jobs", jobsHandler.ListHandle)

	jobHandler := NewJobHandler(jobStorage)
	router.POST("/job", jobHandler.CreateHandle)
	router.DELETE("/job/:name", jobHandler.DeleteHandle)

	executionHandler := NewExecutionHandler(jobStorage, executionStorage)
	router.POST("/executions", executionHandler.StartHandle)
	router.DELETE("/execution/:id", executionHandler.FinishHandle)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return srv
}
