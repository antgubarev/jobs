package get

import (
	"context"
	"fmt"
	"os"

	"github.com/antgubarev/pet/internal/restapi"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

func Subcommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "jobs",
			Aliases: []string{"j"},
			Action:  jobsListAction,
		},
	}
}

func jobsListAction(ctx *cli.Context) error {
	client := restapi.NewClientHTTP(ctx.String("server-url"))
	jobs, err := client.JobsList(context.Background())
	if err != nil {
		return fmt.Errorf("job list action: %w", err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeader([]string{"Name", "Lock mode"})

	for _, jb := range jobs {
		table.Append([]string{jb.Name, string(jb.LockMode)})
	}

	table.Render()

	return nil
}
