package main

import (
	"github.com/spf13/cobra"
)

var shortDescRoot = "The CLI tool for dodo to manage your dodo project"
var longDescRoot = `
The CLI tool for dodo to manage your dodo project.

Find more information at: https://www.dodo-doc.com/
`

func main() {
	rootCmd := &cobra.Command{
		Use:   "dodo",
		Short: "The CLI tool for dodo to manage your dodo project",
		Long:  longDescRoot,
	}
	rootCmd.AddCommand(initCmd)
	cobra.CheckErr(rootCmd.Execute())
}
