package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/toritoritori29/dodo-cli/src/config"
)

var AvailableFormats = []string{ //nolint: gochecknoglobals
	FormatText,
	FormatJSON,
}

type UploadArgs struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	format   string // output style for the command
	rootPath string // root path of the project
	noColor  bool   // disable color output
}

// Implement LoggingConfig and PrinterConfig interface for UploadArgs.
func (opts *UploadArgs) DisableLogging() bool {
	return opts.format == FormatJSON
}

func (opts *UploadArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *UploadArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *UploadArgs) EnablePrinter() bool {
	return true
}

// createUploadCommand creates a cobra command with common flags for upload operations.
func createUploadCommand(use, short, defaultEndpoint string, opts *UploadArgs, runFunc func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runFunc,
	}
	cmd.Flags().StringVarP(&opts.file, "config", "c", ".dodo.yaml", "Path to the configuration file")
	cmd.Flags().StringVarP(&opts.rootPath, "workingDir", "w", ".", "Defines the root path of the project for the command's execution context")
	cmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format for the command. Supported formats: {text, json}")

	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "archive file path") // Deprecated
	cmd.Flags().StringVar(&opts.endpoint, "endpoint", defaultEndpoint, "endpoint to upload")
	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return cmd
}

func CreateUploadCmd() *cobra.Command {
	opts := UploadArgs{}
	return createUploadCommand(
		"upload",
		"upload the project to dodo-doc",
		"https://api.dodo-doc.com/project/upload",
		&opts,
		func(_ *cobra.Command, _ []string) error {
			printer := NewErrorPrinter(ErrorLevel)
			env := NewEnvArgs()
			err := CheckArgsAndEnv(opts, env)
			if err != nil {
				return printer.HandleError(err)
			}

			printer = NewPrinterFromArgs(&opts)
			jsonWriter := NewJSONWriterFromArgs(opts)
			if err := uploadCmdEntrypoint(opts, env, jsonWriter); err != nil {
				jsonWriter.ShowFailedJSONText(err)
				return printer.HandleError(err)
			}
			return nil
		},
	)
}

func uploadCmdEntrypoint(args UploadArgs, env EnvArgs, jsonWriter *JSONWriter) error {
	// Parse the command line arguments and environment variables.
	// Initialize the logging configuration from the command line arguments.
	if err := InitLogger(&args); err != nil {
		return err
	}

	// Execute the upload operation.
	url, err := executeUpload(args, env)
	if err != nil {
		return err
	}
	log.Infof("successfully uploaded")
	log.Infof("please open this link to view the document: %s", url)
	jsonWriter.ShowSucceededJSONText(url)
	return nil
}

func executeUpload(args UploadArgs, env EnvArgs) (string, error) {
	// Read config file
	log.Debugf("config file: %s", args.file)
	configFile, err := os.Open(args.file)
	if err != nil {
		return "", fmt.Errorf("failed to open the config file: %w", err)
	}
	defer configFile.Close()

	state := config.NewParseStateV1(args.file, "./")
	conf, err := config.ParseConfigV1(state, configFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse the config file: %w", err)
	}

	// Transform config to metadata
	metadata, err := NewMetadataFromConfigV1(conf)
	if err != nil {
		return "", err
	}

	// Prepare archive file
	archive, err := NewArchive(args.output)
	if err != nil {
		return "", err
	}
	defer archive.Close()
	if merr := archive.Archive(metadata); merr != nil {
		return "", merr
	}

	// Upload the archive file
	resp, err := archive.Upload(args.endpoint, env.APIKey)
	if err != nil {
		return "", err
	}
	return resp.DocumentURL, nil
}

func CheckArgsAndEnv(args UploadArgs, env EnvArgs) error { //nolint: cyclop
	// Check if `format` and `debug` are compatible
	if args.debug && args.format != "text" {
		return errors.New("debug mode is only supported with text format")
	}

	// Check if `file` is valid
	_, err := os.Stat(args.file)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified `file` argument is invalid. Please check if the file exists. Path: %s", args.file)
	}
	if err != nil {
		return fmt.Errorf("specified `file` argument is invalid. Path: %s", args.file)
	}

	// Check if the output path is valid
	parentDir := filepath.Dir(args.output)
	_, err = os.Stat(parentDir)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified output path is invalid. Please check if the parent directory exists. Path: %s", args.output)
	}
	if err != nil {
		return fmt.Errorf("the provided output path is invalid. Path: %s", args.output)
	}

	// Check if `root` is valid
	_, err = os.Stat(args.rootPath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified `root` argument is invalid. Please check if the directory exists. Path: %s", args.rootPath)
	}
	if err != nil {
		return fmt.Errorf("the provided `root` argument is invalid. Path: %s", args.rootPath)
	}

	// Check if the api key exists
	if env.APIKey == "" {
		return errors.New("the API key is empty. Please set the environment variable DODO_API_KEY")
	}

	// Check if the format is valid
	if !slices.Contains(AvailableFormats, args.format) {
		return fmt.Errorf("invalid format. Supported formats: %v", AvailableFormats)
	}
	return nil
}

func NewJSONWriterFromArgs(args UploadArgs) *JSONWriter {
	if args.format == FormatJSON {
		return NewJSONWriter(JSONLogLevelEnabled)
	}
	return NewJSONWriter(JSONLogLevelDisabled)
}
