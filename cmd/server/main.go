package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antgubarev/jobs/internal/boltdb"
	"github.com/antgubarev/jobs/internal/restapi"
)

const TIMEOUT = 5

func main() {
	flags := parseFlags()

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT*time.Second)
	boltDB, err := boltdb.NewBoltDB(flags.dbPath)
	if err != nil {
		panic(err)
	}
	defer func() {
		boltDB.Close()
		cancel()
	}()

	srv := restapi.NewServer(flags.listen, boltDB)

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()
	log.Printf("Start listening in %s \n", flags.listen)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %s \n", err.Error())

		return
	}

	log.Println("Server has exited")
}

type runFlags struct {
	listen string
	dbPath string
}

func parseFlags() *runFlags {
	result := runFlags{
		listen: ":8080",
		dbPath: "./data.db",
	}

	flag.StringVar(&result.listen, "listen", ":8080", "listen api host port. default :8080")
	flag.StringVar(&result.dbPath, "dbPath", "./data.db", "data file. default ./data.db")
	flag.Parse()

	return &result
}
