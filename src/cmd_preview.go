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

func CreatePreviewCmd() *cobra.Command {
	opts := PreviewArgs{}
	uploadOpts := UploadArgs(opts)
	cmd := createUploadCommand(
		"preview",
		"upload the project to dodo-doc preview environment",
		"https://api-demo.dodo-doc.com/project/upload",
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
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	printer := NewPrinter(ErrorLevel)
	if args.noColor {
		printer = NewPrinter(NoColor)
	}

	// Convert PreviewArgs to UploadArgs for checking arguments
	uploadArgs := UploadArgs(args)

	err := CheckArgsAndEnv(uploadArgs, env)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	if _, err := executeUpload(uploadArgs, env); err != nil {
		printer.PrintError(err)
		return err
	}
	return nil
}
