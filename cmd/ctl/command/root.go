package command

import "github.com/spf13/cobra"

type CmdBuilder struct {
	globalFlags struct {
		serverURL string
	}
}

func NewBuilder() *CmdBuilder {
	return &CmdBuilder{}
}

func (b *CmdBuilder) GetCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "jobsctl",
		Short: "Jobs CLI",
	}

	rootCommand.PersistentFlags().StringVarP(&b.globalFlags.serverURL, "server-url", "s",
		"localhost:8080", "Api server addr. Default `http://localhost:8080`")

	rootCommand.AddCommand(b.jobsCommand())

	return rootCommand
}
