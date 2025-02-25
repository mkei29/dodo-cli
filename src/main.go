package main

import (
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
		Use:   "dodo",
		Short: "The CLI tool for dodo to manage your dodo project",
		Long:  longDescRoot,
	}
	rootCmd.AddCommand(CreateInitCmd())
	rootCmd.AddCommand(CreateUploadCmd())
	rootCmd.AddCommand(CreateVersionCmd())
	rootCmd.AddCommand(CreateTouchCmd())
	rootCmd.AddCommand(CreateCheckCmd())
	cobra.CheckErr(rootCmd.Execute())
}
