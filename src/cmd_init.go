package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

//go:embed template.yaml
var configTemplate string

type InitArgs struct {
	configPath string // config file path
	workingDir string // root path of the project
	debug      bool   // server endpoint to upload
}

func CreateInitCmd() *cobra.Command {
	opts := InitArgs{}
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new configuration file for the project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeInit(opts)
		},
	}
	initCmd.Flags().StringVarP(&opts.configPath, "config", "c", ".dodo.yaml", "Path to the configuration file")
	initCmd.Flags().StringVarP(&opts.workingDir, "workingDir", "w", ".", "Defines the root path of the project for the command's execution context")
	initCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	return initCmd
}

func executeInit(args InitArgs) error {
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	configPath := filepath.Join(args.workingDir, args.configPath)
	log.Debugf("config file: %s", configPath)

	_, err := os.Stat(args.workingDir)
	if os.IsExist(err) {
		return fmt.Errorf("configuration file already exists: %s", configPath)
	}

	template, err := generateConfigContent(PlaceHolderForConfig{
		ProjectName: "my-project",
		Version:     "0.1.0",
		Description: "My project description",
	})
	if err != nil {
		return fmt.Errorf("failed to generate configuration file: %w", err)
	}

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}
	_, err = f.WriteString(template)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	f.Sync()
	return nil
}

type PlaceHolderForConfig struct {
	ProjectName string
	Version     string
	Description string
}

func generateConfigContent(placeholder PlaceHolderForConfig) (string, error) {
	template, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	w := &bytes.Buffer{}
	template.Execute(w, placeholder)
	return w.String(), nil
}
