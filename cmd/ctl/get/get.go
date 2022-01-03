package get

import (
	"fmt"
	"os"

	"github.com/antgubarev/pet/internal/restapi"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

func GetGetSubcommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "jobs",
			Aliases: []string{"j"},
			Action:  jobsListAction,
		},
	}
}

func jobsListAction(c *cli.Context) error {
	client := restapi.NewClientHttp(c.String("server-url"))
	jobs, err := client.JobsList()
	if err != nil {
		return fmt.Errorf("job list action: %v", err)
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
