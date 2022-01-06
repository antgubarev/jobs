package create

import (
	"context"
	"fmt"

	"github.com/antgubarev/pet/internal/restapi"
	"github.com/urfave/cli/v2"
)

func GetCreateSubCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name: "job",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Unique job name.",
				},
				&cli.StringFlag{
					Name:    "lock-mode",
					Aliases: []string{"l"},
					Usage:   "Lock mode. Available value: `free`(default), `host`, `cluster`",
					Value:   "free",
				},
			},
			Action: createJobAction,
		},
	}
}

func createJobAction(ctx *cli.Context) error {
	client := restapi.NewClientHTTP(ctx.String("server-url"))
	if err := client.JobCreate(context.Background(), &restapi.CreateJobIn{
		Name:     ctx.String("name"),
		LockMode: ctx.String("lock-mode"),
	}); err != nil {
		return fmt.Errorf("create action: %w", err)
	}

	return nil
}
