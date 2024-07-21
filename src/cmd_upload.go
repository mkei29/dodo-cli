package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

type UploadArgs struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	rootPath string // root path of the project
}

func CreateUploadCmd() *cobra.Command {
	opts := UploadArgs{}
	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "upload the project to dodo-doc",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeUpload(opts)
		},
	}
	uploadCmd.Flags().StringVarP(&opts.file, "config", "c", ".dodo.yaml", "Path to the configuration file")
	uploadCmd.Flags().StringVarP(&opts.rootPath, "workingDir", "w", ".", "Defines the root path of the project for the command's execution context")
	uploadCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")

	uploadCmd.Flags().StringVarP(&opts.output, "output", "o", "", "archive file path") // Deprecated
	uploadCmd.Flags().StringVar(&opts.endpoint, "endpoint", "http://api.dodo-doc.com/project/upload", "endpoint to upload")
	return uploadCmd
}

func executeUpload(args UploadArgs) error { //nolint: funlen, cyclop
	env := NewEnvArgs()

	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	err := CheckArgsAndEnv(args, env)
	if err != nil {
		log.Fatalf("%w", err)
	}

	// Read config file
	log.Debugf("config file: %s", args.file)
	configFile, err := os.Open(args.file)
	if err != nil {
		log.Fatal("internal error: failed to open config file")
	}
	defer configFile.Close()

	config, err := ParseConfig(configFile)
	if err != nil {
		log.Errorf("internal error: failed to parse config file: %w", err)
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Create Page structs from config.
	page, es := CreatePageTree(*config, args.rootPath)
	if es.HasError() {
		for _, e := range es.Errors() {
			log.Errorf(e.Error())
		}
		return fmt.Errorf("failed to create page tree")
	}
	log.Debugf("successfully convert config to page. found %d pages", page.Count())

	es = page.IsValid()
	if es.HasError() {
		for _, e := range es.Errors() {
			log.Errorf(e.Error())
		}
		return fmt.Errorf("failed to validate page tree")
	}

	project := NewMetadataProjectFromConfig(config)

	metadata := Metadata{
		Version: "1",
		Project: project,
		Page:    *page,
	}

	var zipFile *os.File
	if args.output == "" {
		zipFile, err = os.CreateTemp("", "dodo.zip")
		if err != nil {
			log.Error("failed to create a temporary file")
			return fmt.Errorf("failed to create a temporary file: %w", err)
		}
		defer func() {
			log.Debugf("clean up a temporary file: %s", zipFile.Name())
			os.Remove(zipFile.Name())
		}()
	} else {
		zipFile, err = os.Create(args.output)
		if err != nil {
			log.Errorf("failed to create an archive file at '%s'", args.output)
			return fmt.Errorf("failed to create a file. Path: %s: %w", args.output, err)
		}

	}
	defer zipFile.Close()
	log.Debugf("prepare an archive file on %s", zipFile.Name())

	pathList := collectFiles(page)
	err = archive(zipFile, pathList)

	if multierr.Errors(err) != nil {
		for _, e := range multierr.Errors(err) {
			log.Errorf("internal error: failed to archive: %w", e)
		}
		return fmt.Errorf("failed to archive documents")
	}
	log.Infof("successfully archived: %s", args.file)

	if err := uploadFile(args.endpoint, metadata, zipFile, env.APIKey); err != nil {
		log.Errorf("internal error: ", err)
		return fmt.Errorf("failed to upload zip: %w", err)
	}
	log.Infof("successfully uploaded: %s", args.output)
	return nil
}

func CheckArgsAndEnv(args UploadArgs, env EnvArgs) error { //nolint: cyclop
	// Check if `file` is valid
	_, err := os.Stat(args.file)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified `file` argument is invalid. please check the file exist. path: %s", args.file)
	}
	if err != nil {
		return fmt.Errorf("specified `file` argument is invalid. path: %s", args.file)
	}

	// Check if the output path is valid
	parentDir := filepath.Dir(args.output)
	_, err = os.Stat(parentDir)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified output path is invalid. please check parent directory exist. path: %s", args.output)
	}
	if err != nil {
		return fmt.Errorf("the provided output path is invalid. path: %s", args.output)
	}

	// Check if `root` is valid
	_, err = os.Stat(args.rootPath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("specified `root` argument is invalid. please check the directory exist. path: %s", args.rootPath)
	}
	if err != nil {
		return fmt.Errorf("the provided `root` argument is invalid. path: %s", args.rootPath)
	}

	// Check if the api key exists
	if env.APIKey == "" {
		return fmt.Errorf("the API key is empty. Please set the environment variable DODO_API_KEY")
	}
	return nil
}

func uploadFile(uri string, metadata Metadata, zipFile *os.File, apiKey string) error {
	req, err := newFileUploadRequest(uri, metadata, zipFile, apiKey)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error raised during communicating to the server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %d", resp.StatusCode)
	}
	return nil
}

func newFileUploadRequest(uri string, metadata Metadata, zipFile *os.File, apiKey string) (*http.Request, error) {
	body := &bytes.Buffer{}
	// Try to create a new multipart writer in a closure.
	// This is to ensure that the multipart writer is closed properly.
	// writer.Close() must be called before pass it to http.NewRequest.
	// If we break this rule, the request will not be sent properly.
	writer, err := func() (*multipart.Writer, error) {
		writer := multipart.NewWriter(body)
		defer writer.Close()

		// Write metadata
		serialized, err := metadata.Serialize()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize metadata: %w", err)
		}
		metadataPart, err := writer.CreateFormField("metadata")
		if err != nil {
			return nil, fmt.Errorf("failed to create a multipart section: %w", err)
		}
		_, err = metadataPart.Write(serialized)
		if err != nil {
			return nil, fmt.Errorf("failed to write the metadata to the multipart section: %w", err)
		}

		// Write archived documents
		filePart, err := writer.CreateFormFile("archive", filepath.Base(zipFile.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to create FormFile: %w", err)
		}

		zipFile.Seek(0, 0)
		_, err = io.Copy(filePart, zipFile)
		if err != nil {
			return nil, fmt.Errorf("failed to copy archive file content to writer: %w", err)
		}
		return writer, nil
	}()
	if err != nil {
		return nil, err
	}

	log.Debugf("contents size %d", body.Len())
	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new upload request from body: %w", err)
	}
	bearer := fmt.Sprintf("Bearer %s", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", bearer)
	return req, nil
}
