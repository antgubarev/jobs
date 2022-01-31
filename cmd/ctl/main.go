package main

import (
	"github.com/antgubarev/jobs/cmd/ctl/command"
	"github.com/golang/glog"
)

func main() {
	rootCmd := command.NewBuilder().GetCommand()

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("run command: %v", err)
	}
}
