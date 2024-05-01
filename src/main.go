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

type argsOpts struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	rootPath string // root path of the project
}

type EnvOpts struct {
	APIKey string
}

func NewEnvOpts() EnvOpts {
	return EnvOpts{
		APIKey: os.Getenv("DODO_API_KEY"),
	}
}

func main() {
	rootOpts := argsOpts{}

	cmd := &cobra.Command{
		Use:   "dodo-client",
		Short: "dodo client which support your documentation",
		Run:   func(cmd *cobra.Command, args []string) { execute(rootOpts) },
	}

	cmd.Flags().StringVarP(&rootOpts.file, "file", "f", ".dodo.yaml", "config file path")
	cmd.Flags().StringVarP(&rootOpts.output, "output", "o", "./output.zip", "archive file path")
	cmd.Flags().StringVar(&rootOpts.endpoint, "endpoint", "http://api.dodo-doc.com/project/upload", "endpoint to upload")
	cmd.Flags().BoolVar(&rootOpts.debug, "debug", false, "run in debug mode")
	cmd.Flags().StringVarP(&rootOpts.rootPath, "root", "r", ".", "root path of the project")

	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to execute command")
	}
}

func execute(args argsOpts) { //nolint: funlen, cyclop
	env := NewEnvOpts()

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

	config, err := ParseDocumentConfig(configFile)
	if err != nil {
		log.Errorf("internal error: failed to parse config file: %w", err)
		return
	}

	// Create Page structs from config.
	page, es := CreatePageTree(*config, args.rootPath)
	if es.HasError() {
		for _, e := range es.Errors() {
			log.Errorf(e.Error())
		}
		return
	}
	log.Debugf("successfully convert config to page. found %d pages", page.Count())
	log.Debugf(page.String())

	es = page.IsValid()
	if es.HasError() {
		for _, e := range es.Errors() {
			log.Errorf(e.Error())
		}
		return
	}

	project := NewMetadataProjectFromConfig(config)

	metadata := Metadata{
		Version: "1",
		Project: project,
		Page:    *page,
	}

	pathList := collectFiles(page)
	err = archive(args.output, pathList)
	if multierr.Errors(err) != nil {
		for _, e := range multierr.Errors(err) {
			log.Errorf("internal error: failed to archive: %w", e)
		}
		return
	}
	log.Infof("successfully archived: %s", args.file)

	if err := uploadFile(args.endpoint, metadata, args.output, env.APIKey); err != nil {
		log.Errorf("internal error: ", err)
		return
	}
	log.Infof("successfully uploaded: %s", args.output)
}

func CheckArgsAndEnv(args argsOpts, env EnvOpts) error { //nolint: cyclop
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

func uploadFile(uri string, metadata Metadata, archivePath string, apiKey string) error {
	req, err := newFileUploadRequest(uri, metadata, archivePath, apiKey)
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

func newFileUploadRequest(uri string, metadata Metadata, zipPath string, apiKey string) (*http.Request, error) {
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
			return nil, fmt.Errorf("failed to write metadata to a multipart section: %w", err)
		}

		// Write archived documents
		filePart, err := writer.CreateFormFile("archive", filepath.Base(zipPath))
		if err != nil {
			return nil, fmt.Errorf("failed to a multipart section: %w", err)
		}
		file, err := os.Open(zipPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open archive file: %w", err)
		}
		defer file.Close()
		_, err = io.Copy(filePart, file)
		if err != nil {
			return nil, fmt.Errorf("failed to copy archive file content to writer: %w", err)
		}
		return writer, nil
	}()
	if err != nil {
		return nil, err
	}

	log.Debugf("contents size %d\n", body.Len())
	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new upload request from body: %w", err)
	}
	bearer := fmt.Sprintf("Bearer %s", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", bearer)
	return req, nil
}
