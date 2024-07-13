package main

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:generate cp ../version.txt version.txt
//go:embed version.txt
var version string

func CreateVersionCmd() *cobra.Command {
	uploadCmd := &cobra.Command{
		Use:   "version",
		Short: "show the version of the client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeVersion()
		},
	}
	return uploadCmd
}

func executeVersion() error {
	fmt.Printf("%s\n", version) //nolint:forbidigo
	return nil
}
