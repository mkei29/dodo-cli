package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "dodo",
		Short: "",
		Long:  "",
		Run:   execute,
	}
	cmd.Flags().StringP("config", "c", ".dodo.yaml", "config file path")
	cmd.Flags().StringP("output", "o", "dist/output.zip", "config file path")

	if err := cmd.Execute(); err != nil {
		log.Panic("failed to execute command")
	}
}

func execute(cmd *cobra.Command, args []string) {
	fmt.Println("Hello, DODO!")

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatal("internal error: failed to get config path")
	}
	outputPath, err := cmd.Flags().GetString("output")
	if err != nil {
		log.Fatal("internal error: failed to get config path")
	}

	// Read config file
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal("internal error: failed to open config file")
	}
	defer configFile.Close()
	_, err = parseConfig(configFile)
	if err != nil {
		log.Fatal("internal error: failed to parse config file: %w", err)
	}

	pathList := collectFiles()
	err = archive(outputPath, pathList)
	if err != nil {
		log.Fatal("internal error: failed to archive: %w", err)
	}
	log.Printf("successfully archived: %s", outputPath)
}
