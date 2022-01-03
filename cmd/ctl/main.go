package main

import (
	"log"
	"os"

	"github.com/antgubarev/pet/cmd/ctl/create"
	"github.com/antgubarev/pet/cmd/ctl/get"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage: "CLI fo job control",
		Name:  "ctl",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server-url",
				Aliases: []string{"s"},
				Value:   "http://localhost:8080",
				Usage:   "Address of api server. Default `http://localhost:8080`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "get",
				Subcommands: get.GetGetSubcommands(),
			},
			{
				Name:        "create",
				Subcommands: create.GetCreateSubCommands(),
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
