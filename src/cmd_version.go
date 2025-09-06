package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//go:generate cp ../version.txt version.txt
//go:embed version.txt
var Version string

func CreateVersionCmd() *cobra.Command {
	uploadCmd := &cobra.Command{
		Use:   "version",
		Short: "show the version of the client",
		RunE: func(_ *cobra.Command, _ []string) error {
			return executeVersion()
		},
	}
	return uploadCmd
}

func executeVersion() error {
	fmt.Fprintf(os.Stdout, "%s\n", Version)
	return nil
}
