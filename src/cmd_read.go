package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

type ReadArgs struct {
	projectID string
	path      string
	debug     bool
	noColor   bool
	endpoint  string
	url       string
}

// Implement LoggingConfig and PrinterConfig interface for PreviewArgs.
func (opts *ReadArgs) DisableLogging() bool {
	return false
}

func (opts *ReadArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *ReadArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *ReadArgs) EnablePrinter() bool {
	return true
}

func CreateReadCmd() *cobra.Command {
	opts := ReadArgs{}
	readCmd := &cobra.Command{
		Use:   "read",
		Short: "Read a document",
		RunE: func(_ *cobra.Command, _ []string) error {
			printer := NewPrinter(ErrorLevel)
			env := NewEnvArgs()
			err := CheckArgsAndEnvForRead(opts, &env)
			if err != nil {
				return printer.HandleError(err)
			}
			printer = NewPrinterFromArgs(&opts)
			if err := readCmdEntrypoint(&opts, &env); err != nil {
				return printer.HandleError(err)
			}
			return nil
		},
	}
	readCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	readCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	readCmd.Flags().StringVarP(&opts.url, "url", "u", "", "The full URL of the document to read (overrides project-id and path if set)")
	readCmd.Flags().StringVarP(&opts.projectID, "project-id", "s", "", "The project ID (slug) to read the document from")
	readCmd.Flags().StringVarP(&opts.path, "path", "p", "", "The path of the document to read")
	readCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/", "Server endpoint for search")
	return readCmd
}

func CheckArgsAndEnvForRead(_ ReadArgs, env *EnvArgs) error {
	if env.APIKey == "" {
		return errors.New("DODO_API_KEY is not set")
	}
	return nil
}

func readCmdEntrypoint(args *ReadArgs, env *EnvArgs) error {
	if err := InitLogger(args); err != nil {
		return err
	}

	endpoint, err := NewEndpoint(args.endpoint)
	if err != nil {
		return err
	}

	projectID := args.projectID
	path := args.path
	if args.url != "" {
		projectID, path, err = parseURL(args.url)
		if err != nil {
			return err
		}
	}
	if projectID == "" {
		return errors.New("project ID is required")
	}
	if path == "" {
		return errors.New("path is required")
	}
	log.Debugf("Reading document from project %s, path %s", projectID, path)

	content, err := sendReadDocumentRequest(env, endpoint, projectID, path)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s\n", content)
	return nil
}
