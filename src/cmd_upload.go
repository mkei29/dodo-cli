package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

type UploadArgs struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	rootPath string // root path of the project
	noColor  bool   // disable color output
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
			return executeUploadWrapper(opts)
		},
	)
}

func executeUploadWrapper(args UploadArgs) error {
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

	err := CheckArgsAndEnv(args, env)
	if err != nil {
		printer.PrintError(err)
		return err
	}

	if err := executeUpload(args, env); err != nil {
		printer.PrintError(err)
		return err
	}
	return nil
}

func executeUpload(args UploadArgs, env EnvArgs) error {
	// Read config file
	log.Debugf("config file: %s", args.file)
	configFile, err := os.Open(args.file)
	if err != nil {
		return fmt.Errorf("failed to open the config file: %w", err)
	}
	defer configFile.Close()

	state := NewParseState(args.file, "./")
	config, err := ParseConfig(state, configFile)
	if err != nil {
		return err
	}

	// Create Project struct from config.
	project := NewMetadataProjectFromConfig(config)

	// Create Page structs from config.
	page, merr := convertConfigPageToMetadataPage(args.rootPath, config)
	if merr != nil {
		return merr
	}
	log.Debugf("successfully convert config to page. found %d pages", page.Count())

	// Create Assets struct from config.
	asset, merr := convertConfigAssetToMetadataAsset(args.rootPath, config.Assets)
	if merr != nil {
		return merr
	}
	log.Debugf("successfully convert assets to metadata. found %d assets", len(asset))

	metadata := Metadata{
		Version: "1",
		Project: project,
		Page:    *page,
		Asset:   asset,
	}

	// Prepare archive file
	archive, err := NewArchive(args.output)
	if err != nil {
		return err
	}
	defer archive.Close()
	if merr := archive.Archive(&metadata); merr != nil {
		return merr
	}

	// Upload the archive file
	resp, err := uploadFile(args.endpoint, metadata, archive.File, env.APIKey)
	if err != nil {
		return err
	}
	log.Infof("successfully uploaded")
	log.Infof("please open this link to view the document: %s", resp.DocumentURL)
	return nil
}

func CheckArgsAndEnv(args UploadArgs, env EnvArgs) error { //nolint: cyclop
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
	return nil
}

func convertConfigPageToMetadataPage(rootDir string, config *Config) (*Page, *MultiError) {
	page, merr := CreatePageTree(config, rootDir)
	if merr != nil {
		return nil, merr
	}
	if merr = page.IsValid(); merr != nil {
		return nil, merr
	}

	return page, nil
}

func convertConfigAssetToMetadataAsset(rootDir string, assets []ConfigAsset) ([]MetadataAsset, *MultiError) {
	// Create Assets struct from config.
	merr := NewMultiError()
	metadataAssets := make([]MetadataAsset, 0, len(assets)*5)
	for _, a := range assets {
		files, err := a.List(rootDir)
		if err != nil {
			merr.Add(err)
		}
		// FIXME: This code could be cause too many allocations.
		for _, f := range files {
			metadataAssets = append(metadataAssets, NewMetadataAsset(f))
		}
	}
	if merr.HasError() {
		return nil, &merr
	}
	return metadataAssets, nil
}

func uploadFile(uri string, metadata Metadata, zipFile *os.File, apiKey string) (*UploadResponse, error) {
	req, err := newFileUploadRequest(uri, metadata, zipFile, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error occurred during communication with the server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to upload file: %d", resp.StatusCode)
	}

	data, err := ParseUploadResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the response: %w", err)
	}
	return data, nil
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
			return nil, fmt.Errorf("failed to write metadata to the multipart section: %w", err)
		}

		// Write archived documents
		filePart, err := writer.CreateFormFile("archive", filepath.Base(zipFile.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to create FormFile: %w", err)
		}

		_, err = zipFile.Seek(0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to seek the archive file: %w", err)
		}
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
	bearer := "Bearer " + apiKey
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", bearer)
	return req, nil
}
