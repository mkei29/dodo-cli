package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new configuration file for the project",
	RunE:  executeInit,
}

func executeInit(cmd *cobra.Command, args []string) error {
	fmt.Println("initializing the project")
	return nil
}
