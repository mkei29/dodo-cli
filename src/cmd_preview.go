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
	rootPath string // root path of the project
	noColor  bool   // disable color output
}

func CreatePreviewCmd() *cobra.Command {
	opts := PreviewArgs{}
	previewCmd := &cobra.Command{
		Use:           "preview",
		Short:         "upload the project to dodo-doc preview environment",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return executePreviewWrapper(opts)
		},
	}
	previewCmd.Flags().StringVarP(&opts.file, "config", "c", ".dodo.yaml", "Path to the configuration file")
	previewCmd.Flags().StringVarP(&opts.rootPath, "workingDir", "w", ".", "Defines the root path of the project for the command's execution context")
	previewCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")

	previewCmd.Flags().StringVarP(&opts.output, "output", "o", "", "archive file path") // Deprecated
	previewCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://api-demo.dodo-doc.com/project/upload", "endpoint to upload")
	previewCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return previewCmd
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
	uploadArgs := UploadArgs{
		file:     args.file,
		output:   args.output,
		endpoint: args.endpoint,
		debug:    args.debug,
		rootPath: args.rootPath,
		noColor:  args.noColor,
	}

	err := CheckArgsAndEnv(uploadArgs, env)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	if err := executeUpload(uploadArgs, env); err != nil {
		printer.PrintError(err)
		return err
	}
	return nil
}
