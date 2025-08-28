package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type DocsArgs struct {
	debug    bool
	noColor  bool
	endpoint string
}

// Implement LoggingConfig and PrinterConfig interface for TouchArgs.
func (opts *DocsArgs) DisableLogging() bool {
	return false
}

func (opts *DocsArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *DocsArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *DocsArgs) EnablePrinter() bool {
	return true
}

func CreateDocsCmd() *cobra.Command {
	opts := DocsArgs{}
	docsCmd := &cobra.Command{
		Use:           "docs",
		Short:         "List the documentation in your organization",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return docsCmdEntrypoint(&opts)
		},
	}
	docsCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	docsCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	docsCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/projects/v1", "The endpoint of the dodo API server")
	return docsCmd
}

func docsCmdEntrypoint(opts *DocsArgs) error {
	env := NewEnvArgs()
	orgs, err := NewProjectFromAPI(env, opts.endpoint)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		return errors.New("no projects found in your organization")
	}

	return renderProjectsWithJSON(orgs)
}

type DocsJSONOutput struct {
	Slug        string `json:"slug,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	ProjectID   string `json:"project_id"`
	IsPublic    bool   `json:"is_public,omitempty"`
	URL         string `json:"url,omitempty"`
}

func renderProjectsWithJSON(orgs []Project) error {
	outputs := make([]DocsJSONOutput, 0, len(orgs))
	for _, org := range orgs {
		outputs = append(outputs, DocsJSONOutput{
			Slug:        org.Slug,
			ProjectName: org.ProjectName,
			IsPublic:    org.IsPublic,
			ProjectID:   org.ProjectID,
			URL:         org.BaseURL,
		})
	}
	b, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the output: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", b)
	return nil
}
