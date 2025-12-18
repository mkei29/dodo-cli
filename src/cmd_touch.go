package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/toritoritori29/dodo-cli/src/config"
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
	now      string
}

// Implement LoggingConfig and PrinterConfig interface for TouchArgs.
func (opts *TouchArgs) DisableLogging() bool {
	return false
}

func (opts *TouchArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *TouchArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *TouchArgs) EnablePrinter() bool {
	return true
}

func CreateTouchCmd() *cobra.Command {
	opts := TouchArgs{}
	touchCmd := &cobra.Command{
		Use:           "touch [filepath]",
		Short:         "Create a new markdown file. If the file already exists, update the fields",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.filepath = args[0]
			printer := NewErrorPrinter(ErrorLevel)
			env := NewEnvArgs()
			if err := TouchArgsAndEnv(&opts, env); err != nil {
				return printer.HandleError(err)
			}
			printer = NewPrinterFromArgs(&opts)
			if err := touchCmdEntrypoint(opts); err != nil {
				return printer.HandleError(err)
			}
			return nil
		},
	}
	touchCmd.Flags().StringVarP(&opts.title, "title", "t", "", "the title of newly created file")
	touchCmd.Flags().StringVarP(&opts.path, "path", "p", "", "the URL path of the file")
	touchCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	touchCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	touchCmd.Flags().StringVar(&opts.now, "now", "", "the current time in RFC3339 format")
	return touchCmd
}

func TouchArgsAndEnv(args *TouchArgs, _ EnvArgs) error {
	// Check if the args.now is in the correct format.
	if args.now != "" {
		_, err := time.Parse(time.RFC3339, args.now)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}
	}
	return nil
}

func touchCmdEntrypoint(args TouchArgs) error {
	// Initialize the logging configuration from the command line arguments.
	if err := InitLogger(&args); err != nil {
		return err
	}

	// New file
	if _, err := os.Stat(args.filepath); os.IsNotExist(err) {
		if err := executeTouchNew(args); err != nil {
			return err
		}
		return nil
	}

	// Update existing file
	if err := executeTouchUpdate(args); err != nil {
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

	filepath := args.filepath
	if args.path != "" {
		filepath = args.path
	}

	now, err := parseTime(args.now)
	if err != nil {
		return err
	}

	sanitized := sanitizePath(filepath)
	matter := config.NewFrontMatter(args.title, sanitized, now)

	if _, err := file.WriteString(matter.String()); err != nil {
		return fmt.Errorf("failed to write front matter: %w", err)
	}
	log.Infof("Successfully created a new markdown file: %s", args.filepath)
	return nil
}

func executeTouchUpdate(args TouchArgs) error {
	log.Debug("Updating an existing markdown file")
	matter, err := config.NewFrontMatterFromMarkdown(args.filepath)
	if err != nil {
		return fmt.Errorf("failed to read front matter: %w", err)
	}
	log.Debug("Successfully read front matter")

	// Update matter
	if args.title != "" {
		matter.Title = args.title
	}
	if args.path != "" {
		matter.Path = args.path
	}

	// Update time
	now, err := parseTime(args.now)
	if err != nil {
		return err
	}
	matter.UpdatedAt = config.NewSerializableTimeFromTime(now)

	// Rewrite a markdown
	if err := matter.UpdateMarkdown(args.filepath); err != nil {
		return fmt.Errorf("failed to update markdown file: %w", err)
	}
	log.Infof("Successfully updated the markdown file: %s", args.filepath)
	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Now(), nil
	}
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
	}
	return t, nil
}

// remove the extension and replace / with _.
func sanitizePath(path string) string {
	// NOTE: We can use a-to-z, A-to-Z, 0-to-9, _ and - in the path.
	if path == "" {
		return path
	}
	p := filepath.Clean(path)

	// Remove the extension
	ext := filepath.Ext(p)
	p = strings.TrimSuffix(p, ext)

	// Join path with underscore
	parts := strings.Split(p, string(os.PathSeparator))
	p = strings.Join(parts, "_")

	// Remove leading _
	p = strings.TrimPrefix(p, "_")

	// Remove disallowed characters
	re := regexp.MustCompile(`[^a-zA-Z-0-9_]+`)
	p = re.ReplaceAllString(p, "")
	return p
}
