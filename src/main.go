package main

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	shortDescRoot = "The CLI tool for dodo to manage your dodo project"
	longDescRoot  = `
The CLI tool for dodo to manage your dodo project.

Find more information at: https://www.dodo-doc.com/
`
)

func main() {
	rootCmd := &cobra.Command{
		Use:           "dodo",
		Short:         "The CLI tool for dodo to manage your dodo project",
		Long:          longDescRoot,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(CreateInitCmd())
	rootCmd.AddCommand(CreateUploadCmd())
	rootCmd.AddCommand(CreatePreviewCmd())
	rootCmd.AddCommand(CreateVersionCmd())
	rootCmd.AddCommand(CreateTouchCmd())
	rootCmd.AddCommand(CreateCheckCmd())
	rootCmd.AddCommand(CreateSearchCmd())
	rootCmd.AddCommand(CreateDocCmd())
	rootCmd.AddCommand(CreateMCPCmd())
	rootCmd.AddCommand(CreateReadCmd())

	defaultPrinter := NewErrorPrinter(ErrorLevel)
	if err := rootCmd.Execute(); err != nil {
		defaultPrinter.HandleError(err) //nolint:errcheck
		os.Exit(1)
	}
}
