package main

import (
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

type DemoArgs struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	rootPath string // root path of the project
	noColor  bool   // disable color output
}

func CreateDemoCmd() *cobra.Command {
	opts := DemoArgs{}
	demoCmd := &cobra.Command{
		Use:           "demo",
		Short:         "upload the project to dodo-doc demo environment",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return executeDemoWrapper(opts)
		},
	}
	demoCmd.Flags().StringVarP(&opts.file, "config", "c", ".dodo.yaml", "Path to the configuration file")
	demoCmd.Flags().StringVarP(&opts.rootPath, "workingDir", "w", ".", "Defines the root path of the project for the command's execution context")
	demoCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")

	demoCmd.Flags().StringVarP(&opts.output, "output", "o", "", "archive file path") // Deprecated
	demoCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://api-demo.dodo-doc.com/project/upload", "endpoint to upload")
	demoCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return demoCmd
}

func executeDemoWrapper(args DemoArgs) error {
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

	// Convert DemoArgs to UploadArgs for checking arguments
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
