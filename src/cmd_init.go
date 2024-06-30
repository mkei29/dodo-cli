package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/caarlos0/log"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

//go:embed template.yaml
var configTemplate string

type InitArgs struct {
	configPath  string // config file path
	workingDir  string // root path of the project
	force       bool   // overwrite the configuration file if it already exists
	debug       bool   // server endpoint to upload
	projectName string
	description string
}

type InitParameter struct {
	ProjectName string
	Version     string
	Description string
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
	initCmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Overwrite the configuration file if it already exists")
	initCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	initCmd.Flags().StringVar(&opts.projectName, "project-name", "", "Project Name")
	initCmd.Flags().StringVar(&opts.description, "description", "", "Project Name")
	return initCmd
}

func executeInit(args InitArgs) error {
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	configPath := filepath.Join(args.workingDir, args.configPath)
	log.Debugf("config file: %s", configPath)

	if !args.force && fileExists(configPath) {
		log.Errorf("configuration file already exists: %s", configPath)
		return fmt.Errorf("configuration file already exists: %s", configPath)
	}

	params, err := receiveUserInput(args.projectName, args.description)
	if err != nil {
		return fmt.Errorf("something went wrong during the user typing value: %w", err)
	}

	content, err := generateConfigContent(*params)
	if err != nil {
		return fmt.Errorf("failed to generate configuration file from the template: %w", err)
	}

	if err := saveConfigContent(configPath, content); err != nil {
		return fmt.Errorf("failed to save configuration file: %w", err)
	}

	if args.force {
		log.Info("Overwrite the configuration file")
	}
	return nil
}

func receiveUserInput(projectNameArgs, descriptionArgs string) (*InitParameter, error) {
	projectName := projectNameArgs
	description := descriptionArgs

	var err error
	if projectName == "" {
		projectNamePrompt := promptui.Prompt{
			Label:   "Project Name",
			Default: "",
		}
		projectName, err = projectNamePrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("prompt failed: %w", err)
		}
	}

	if description == "" {
		descriptionPrompt := promptui.Prompt{
			Label:   "Description",
			Default: "",
		}
		description, err = descriptionPrompt.Run()
		if err != nil {
			return nil, fmt.Errorf("prompt failed: %w", err)
		}
	}

	params := InitParameter{
		ProjectName: projectName,
		Version:     "1",
		Description: description,
	}
	return &params, nil
}

func generateConfigContent(placeholder InitParameter) (string, error) {
	template, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	w := &bytes.Buffer{}
	if err := template.Execute(w, placeholder); err != nil {
		return "", fmt.Errorf("failed to populate variables into the template: %w", err)
	}
	return w.String(), nil
}

func saveConfigContent(configPath, content string) error {
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to flush configuration file: %w", err)
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
