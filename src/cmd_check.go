package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/toritoritori29/dodo-cli/src/config"
)

type CheckArgs struct {
	configPath string // config file path
	debug      bool   // enable debug mode
	noColor    bool   // disable color output
}

// Implement LoggingConfig interface for CheckArgs.
func (opts *CheckArgs) DisableLogging() bool {
	return false
}

func (opts *CheckArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *CheckArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *CheckArgs) EnablePrinter() bool {
	return true
}

func CreateCheckCmd() *cobra.Command {
	opts := CheckArgs{}
	checkCmd := &cobra.Command{
		Use:           "check",
		Short:         "check the configuration file for dodo-doc",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			env := NewEnvArgs()
			printer := NewErrorPrinter(ErrorLevel)
			err := CheckArgsAndEnvForCheck(opts, env)
			if err != nil {
				return printer.HandleError(err)
			}
			printer = NewPrinterFromArgs(&opts)
			if err := InitLogger(&opts); err != nil {
				return printer.HandleError(err)
			}

			if err := checkCmdEntrypoint(opts); err != nil {
				return printer.HandleError(err)
			}
			return nil
		},
	}
	checkCmd.Flags().StringVarP(&opts.configPath, "config", "c", ".dodo.yaml", "Path to the configuration file")
	checkCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	checkCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return checkCmd
}

func checkCmdEntrypoint(args CheckArgs) error {
	// Read config file
	log.Debugf("config file: %s", args.configPath)
	configFile, err := os.Open(args.configPath)
	if err != nil {
		return fmt.Errorf("failed to open the config file: %w", err)
	}
	defer configFile.Close()

	state := config.NewParseStateV1(args.configPath, "./")
	conf, err := config.ParseConfigV1(state, configFile)
	if err != nil {
		return fmt.Errorf("failed to parse the config file: %w", err)
	}

	_, err = NewMetadataFromConfigV1(conf)
	if err != nil {
		return fmt.Errorf("failed to convert config to metadata: %w", err)
	}
	return nil
}

func CheckArgsAndEnvForCheck(args CheckArgs, env EnvArgs) error {
	// Check if `configPath` is valid
	_, err := os.Stat(args.configPath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified `configPath` argument is invalid. Please check if the file exists. Path: %s", args.configPath)
	}
	if err != nil {
		return fmt.Errorf("specified `configPath` argument is invalid. Path: %s", args.configPath)
	}

	// Check if the api key exists
	if env.APIKey == "" {
		return errors.New("the API key is empty. Please set the environment variable DODO_API_KEY")
	}
	return nil
}
