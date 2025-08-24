package main

import (
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

type PreviewArgs struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	format   string // output style for the command
	rootPath string // root path of the project
	noColor  bool   // disable color output
}

// Implement LoggingConfig and PrinterConfig interface for PreviewArgs.
func (opts *PreviewArgs) DisableLogging() bool {
	return opts.format == FormatJSON
}

func (opts *PreviewArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *PreviewArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *PreviewArgs) EnablePrinter() bool {
	return opts.format == FormatText
}

func CreatePreviewCmd() *cobra.Command {
	opts := PreviewArgs{}
	uploadOpts := UploadArgs(opts)
	cmd := createUploadCommand(
		"preview",
		"upload the project to dodo-doc preview environment",
		"https://api.dodo-doc.com/project/upload/demo",
		&uploadOpts,
		func(_ *cobra.Command, _ []string) error {
			// Convert back to PreviewArgs
			opts = PreviewArgs(uploadOpts)
			return executePreviewWrapper(opts)
		},
	)
	return cmd
}

func executePreviewWrapper(args PreviewArgs) error {
	// Initialize logger and so on, then execute the main function.
	env := NewEnvArgs()
	uploadArgs := UploadArgs(args)

	// Parse the command line arguments and environment variables.
	printer := NewPrinter(ErrorLevel)
	err := CheckArgsAndEnv(uploadArgs, env)
	if err != nil {
		printer.PrintError(err)
		return err
	}
	printer = NewPrinterFromArgs(&args)
	jsonWriter := NewJSONWriterFromArgs(uploadArgs)

	// Initialize the logging configuration from the command line arguments.
	if err := InitLogger(&args); err != nil {
		printer.PrintError(err)
		jsonWriter.ShowFailedJSONText(err)
	}

	// Execute the upload operation.
	url, err := executeUpload(uploadArgs, env)
	if err != nil {
		printer.PrintError(err)
		jsonWriter.ShowFailedJSONText(err)
		return err
	}
	log.Infof("successfully uploaded")
	log.Infof("please open this link to view the document: %s", url)
	jsonWriter.ShowSucceededJSONText(url)
	return nil
}
