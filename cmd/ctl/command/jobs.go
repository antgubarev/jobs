package command

import (
	"context"
	"os"
	"time"

	"github.com/antgubarev/jobs/internal/restapi"
	"github.com/golang/glog"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func (b *CmdBuilder) jobsCommand() *cobra.Command {
	jobsCmd := &cobra.Command{
		Use:     "job",
		Short:   "Jobs CRUD",
		Aliases: []string{"j"},
	}

	jobsCmd.AddCommand(b.jobsCreateCommand())
	jobsCmd.AddCommand(b.jobsListCommand())
	jobsCmd.AddCommand(b.jobsDeleteCommand())

	return jobsCmd
}

func (b *CmdBuilder) jobsCreateCommand() *cobra.Command {
	var (
		jobName  string
		lockMode string
	)

	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new job",
		Aliases: []string{"c"},
		Run: func(cmd *cobra.Command, args []string) {
			client := restapi.NewClientHTTP(b.globalFlags.serverURL)
			if err := client.JobCreate(context.Background(), &restapi.CreateJobIn{
				Name:     jobName,
				LockMode: lockMode,
			}); err != nil {
				glog.Errorf("create action: %v", err)
			}

			glog.Infof("job `%s` created \n", jobName)
		},
	}

	createCmd.Flags().StringVarP(&jobName, "name", "n", "", "Unique job name")
	createCmd.Flags().StringVarP(&lockMode, "lock-mode", "l", "free",
		"Lock mode. Available value: `free`(default), `host`, `cluster`")
	if err := createCmd.MarkFlagRequired("name"); err != nil {
		glog.Fatalf("config required flag `name`: %v", err)
	}

	return createCmd
}

func (b *CmdBuilder) jobsListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Jobs list",
		Aliases: []string{"l", "ls"},
		Run: func(cmd *cobra.Command, args []string) {
			client := restapi.NewClientHTTP(b.globalFlags.serverURL)
			jobs, err := client.JobsList(context.Background())
			if err != nil {
				glog.Errorf("job list action: %v", err)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Lock mode", "Created"})

			for _, jb := range jobs {
				table.Append([]string{jb.Name, string(jb.LockMode), jb.CreatedAt.Format(time.RFC3339)})
			}

			table.Render()
		},
	}

	return listCmd
}

func (b *CmdBuilder) jobsDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete job",
		Aliases: []string{"d", "del"},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	var jobName string

	deleteCmd.Flags().StringVarP(&jobName, "name", "n", "", "Unique job name")
	if err := deleteCmd.MarkFlagRequired("name"); err != nil {
		glog.Fatalf("config required flag `name`: %v", err)
	}

	return deleteCmd
}
