package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// a `touch` command aims to help the user to manage the markdown files.
// If the file does not exist, it will create a new file with the given title.
// If the file already exists, it will update the frontmatter fields.

type TouchArgs struct {
	filepath string
	title    string
	path     string
}

func CreateTouchCmd() *cobra.Command {
	opts := TouchArgs{}
	touchCmd := &cobra.Command{
		Use:           "touch [filepath]",
		Short:         "create a new markdown file. if the file already exists, update the fields",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.filepath = args[0]
			return executeTouch(opts)
		},
	}
	touchCmd.Flags().StringVarP(&opts.title, "title", "t", "", "the title of newly created file")
	touchCmd.Flags().StringVarP(&opts.path, "path", "p", "", "the URL path of the file")
	return touchCmd
}

func executeTouch(args TouchArgs) error {
	if _, err := os.Stat(args.filepath); os.IsNotExist(err) {
		return executeTouchNew(args)
	}
	return executeTouchUpdate(args)
}

func executeTouchNew(args TouchArgs) error {
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
	matter, err := NewFrontMatterFromMarkdown(args.filepath)
	if err != nil {
		return err
	}

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
