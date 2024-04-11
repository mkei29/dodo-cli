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
)

type argsOpts struct {
	file     string // config file path
	output   string // deprecated: the path to locate the archive file
	endpoint string // server endpoint to upload
	debug    bool   // server endpoint to upload
	rootPath string // root path of the project
}

type envOpts struct {
	apiKey string
}

func NewEnvOpts() envOpts {
	return envOpts{
		apiKey: os.Getenv("DODO_API_KEY"),
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
	cmd.Flags().StringVarP(&rootOpts.output, "output", "o", "dist/output.zip", "config file path")
	cmd.Flags().StringVar(&rootOpts.endpoint, "endpoint", "http://api.test-doc.com/project/upload", "endpoint to upload")
	cmd.Flags().BoolVar(&rootOpts.debug, "debug", false, "run in debug mode")
	cmd.Flags().StringVarP(&rootOpts.rootPath, "root", "r", ".", "root path of the project")

	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to execute command")
	}
}

func execute(args argsOpts) {
	env := NewEnvOpts()

	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
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
		log.Fatalf("internal error: failed to parse config file: %w", err)
	}

	// Convert config to
	page, err := CreatePageTree(*config, args.rootPath)
	if err != nil {
		log.Fatalf("internal error: failed to convert the config page to the page: %w", err)
	}
	log.Debugf("successfully convert config to page. found %d pages", page.Count())

	if !page.IsValid() {
		log.Fatal("invalid page was found")
	}

	project := NewMetadataProjectFromConfig(config)

	metadata := Metadata{
		Version: "1",
		Project: project,
		Page:    *page,
	}

	pathList := collectFiles(args.file, page)
	err = archive(args.output, pathList)
	if err != nil {
		log.Fatalf("internal error: failed to archive: %w", err)
	}
	log.Infof("successfully archived: %s", args.file)

	if err := uploadFile(args.endpoint, metadata, args.output, env.apiKey); err != nil {
		log.Fatalf("internal error: ", err)
	}
	log.Infof("successfully uploaded: %s", args.output)
}

func uploadFile(uri string, metadata Metadata, archivePath string, apiKey string) error {
	log.Infof("uploading endpoint: %s", uri)
	req, err := newFileUploadRequest(uri, metadata, archivePath, apiKey)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error raised during communicating to the server: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %d", resp.StatusCode)
	}
	return nil
}

func newFileUploadRequest(uri string, metadata Metadata, path string, apiKey string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Write metadata
	serialized, err := metadata.Serialize()
	if err != nil {
		return nil, err
	}
	metadataPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}
	metadataPart.Write(serialized)

	// Write archived documents
	filePart, err := writer.CreateFormFile("archive", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(filePart, file)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	bearer := fmt.Sprintf("Bearer %s", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", bearer)
	return req, err
}
