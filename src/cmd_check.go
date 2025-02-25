package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

type CheckArgs struct {
	configPath string // config file path
	debug      bool   // enable debug mode
	noColor    bool   // disable color output
}

func CreateCheckCmd() *cobra.Command {
	opts := CheckArgs{}
	checkCmd := &cobra.Command{
		Use:           "check",
		Short:         "check the configuration file for dodo-doc",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCheckWrapper(opts)
		},
	}
	checkCmd.Flags().StringVarP(&opts.configPath, "config", "c", ".dodo.yaml", "Path to the configuration file")
	checkCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	checkCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return checkCmd
}

func executeCheckWrapper(args CheckArgs) error {
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

	err := CheckArgsAndEnvForCheck(args, env)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	if err := executeCheck(args); err != nil {
		printer.PrintError(err)
		return err
	}
	return nil
}

func executeCheck(args CheckArgs) error {
	// Read config file
	log.Debugf("config file: %s", args.configPath)
	configFile, err := os.Open(args.configPath)
	if err != nil {
		return fmt.Errorf("failed to open the config file: %w", err)
	}
	defer configFile.Close()

	state := NewParseState(args.configPath, "./")
	config, err := ParseConfig(state, configFile)
	if err != nil {
		return err
	}

	// Validate Page structs from config.
	page, merr := convertConfigPageToMetadataPage(".", config)
	if merr != nil {
		return fmt.Errorf("page validation failed: %w", merr)
	}
	log.Debugf("successfully validated config to page. found %d pages", page.Count())

	// Validate Assets struct from config.
	asset, merr := convertConfigAssetToMetadataAsset(".", config.Assets)
	if merr != nil {
		return fmt.Errorf("asset validation failed: %w", merr)
	}
	log.Debugf("successfully validated assets to metadata. found %d assets", len(asset))

	log.Infof("configuration file is valid")
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
		return fmt.Errorf("the API key is empty. Please set the environment variable DODO_API_KEY")
	}
	return nil
}
