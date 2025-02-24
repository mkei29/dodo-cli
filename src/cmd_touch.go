package main

import (
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

// The `touch` command helps the user manage markdown files.
// If the file does not exist, it will create a new file with the given title.
// If the file already exists, it will update the frontmatter fields.

type TouchArgs struct {
	filepath string
	title    string
	path     string
	debug    bool
	noColor  bool
}

func CreateTouchCmd() *cobra.Command {
	opts := TouchArgs{}
	touchCmd := &cobra.Command{
		Use:           "touch [filepath]",
		Short:         "Create a new markdown file. If the file already exists, update the fields",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.filepath = args[0]
			return executeTouchWrapper(opts)
		},
	}
	touchCmd.Flags().StringVarP(&opts.title, "title", "t", "", "the title of newly created file")
	touchCmd.Flags().StringVarP(&opts.path, "path", "p", "", "the URL path of the file")
	touchCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	touchCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return touchCmd
}

func executeTouchWrapper(args TouchArgs) error {
	// Initialize logger and other settings, then execute the main function.
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Running in debug mode")
	}

	printer := NewPrinter(ErrorLevel)
	if args.noColor {
		printer = NewPrinter(NoColor)
	}

	if _, err := os.Stat(args.filepath); os.IsNotExist(err) {
		if err := executeTouchNew(args); err != nil {
			printer.printError(err)
			return err
		}
		return nil
	}
	if err := executeTouchUpdate(args); err != nil {
		printer.printError(err)
		return err
	}
	return nil
}

func executeTouchNew(args TouchArgs) error {
	log.Debug("Creating a new markdown file")
	file, err := os.Create(args.filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	matter := NewFrontMatter(args.title, args.filepath)
	if _, err := file.WriteString(matter.String()); err != nil {
		return fmt.Errorf("failed to write front matter: %w", err)
	}
	return nil
}

func executeTouchUpdate(args TouchArgs) error {
	log.Debug("Updating an existing markdown file")
	matter, err := NewFrontMatterFromMarkdown(args.filepath)
	if err != nil {
		return err
	}
	log.Debug("Successfully read front matter")

	// Update matter
	if args.title != "" {
		matter.Title = args.title
	}
	if args.path != "" {
		matter.Path = args.path
	}
	matter.UpdatedAt = NewSerializableTimeFromTime(time.Now())

	// Rewrite a markdown
	return matter.UpdateMarkdown(args.filepath)
}
