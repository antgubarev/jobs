package create

import (
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

func createJobAction(c *cli.Context) error {
	client := restapi.NewClientHttp(c.String("server-url"))
	if err := client.JobCreate(&restapi.CreateJobIn{
		Name:     c.String("name"),
		LockMode: c.String("lock-mode"),
	}); err != nil {
		return fmt.Errorf("create action: %v", err)
	}
	return nil
}
