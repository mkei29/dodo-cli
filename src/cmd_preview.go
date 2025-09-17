package main

import (
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

func CreatePreviewCmd() *cobra.Command {
	opts := UploadArgs{}
	cmd := createUploadCommand(
		"preview",
		"upload the project to dodo-doc preview environment",
		"https://api.dodo-doc.com/project/upload/demo",
		&opts,
		func(_ *cobra.Command, _ []string) error {
			env := NewEnvArgs()

			printer := NewErrorPrinter(ErrorLevel)
			err := CheckArgsAndEnv(opts, env)
			if err != nil {
				return printer.HandleError(err)
			}

			printer = NewPrinterFromArgs(&opts)
			jsonWriter := NewJSONWriterFromArgs(opts)
			if err := previewCmdEntrypoint(opts, env, jsonWriter); err != nil {
				jsonWriter.ShowFailedJSONText(err)
				return printer.HandleError(err)
			}
			return nil
		},
	)
	return cmd
}

func previewCmdEntrypoint(args UploadArgs, env EnvArgs, jsonWriter *JSONWriter) error {
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
