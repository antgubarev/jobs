package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/antgubarev/pet/internal/executor"
	"github.com/antgubarev/pet/internal/restapi"
	"github.com/urfave/cli/v2"
)

const usageText = "job-exec [global options] -- [command] [args]"

var errInvalidArgument = errors.New("invalid argument")

func action(ctx *cli.Context) error {
	var commandArgs []string
	for i, arg := range os.Args {
		if arg == "--" {
			commandArgs = os.Args[i+1:]
		}
	}
	if len(commandArgs) == 0 {
		return fmt.Errorf("%w: `command` is required, usage: %s", errInvalidArgument, usageText)
	}

	client := restapi.NewClientHTTP(ctx.String("server-url"))
	exectr := executor.NewExecutor(client, executor.WithOutFile(os.Stdout), executor.WithErrFile(os.Stderr))

	code, err := exectr.StartAndWatch(context.Background(), ctx.String("job-name"), commandArgs)
	if err != nil {
		return fmt.Errorf("process hasn't started: %w", err)
	}
	if code != executor.ExitOK {
		return fmt.Errorf("process is corrupted: %w", err)
	}

	return nil
}

func main() {
	app := &cli.App{
		Usage:     "Starts new process (command after `--`) and register to the server.",
		Name:      "job-exec",
		UsageText: usageText,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "job-name",
				Usage:    "Unique name of job to start (require argument)",
				Aliases:  []string{"j"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "server-url",
				Aliases: []string{"s"},
				Value:   "http://localhost:8080",
				Usage:   "Address of api server. Default `http://localhost:8080`",
			},
		},
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
