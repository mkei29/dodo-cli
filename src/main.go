package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "dodo",
		Short: "",
		Long:  "",
		Run:   execute,
	}
	cmd.Flags().StringP("file", "f", ".dodo", "config file path")
	cmd.Flags().StringP("output", "o", "dist/output.zip", "config file path")
	cmd.Flags().StringP("api-key", "k", "", "api key")

	if err := cmd.Execute(); err != nil {
		log.Fatal("failed to execute command")
	}
}

func execute(cmd *cobra.Command, args []string) {
	configPath, err := cmd.Flags().GetString("file")
	if err != nil {
		log.Fatal("internal error: failed to get `file` flag")
	}
	outputPath, err := cmd.Flags().GetString("output")
	if err != nil {
		log.Fatal("internal error: failed to get config path")
	}
	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil || apiKey == "" {
		log.Fatal("internal error: failed to get api key")
	}

	// Read config file
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal("internal error: failed to open config file")
	}
	defer configFile.Close()
	config, err := ParseDocumentDefinition(configFile)
	if err != nil {
		log.Fatal("internal error: failed to parse config file: %w", err)
	}

	pathList := collectFiles(configPath, config)
	err = archive(outputPath, pathList)
	if err != nil {
		log.Fatal("internal error: failed to archive: %w", err)
	}
	fmt.Printf("pathList: %v\n", pathList)
	log.Printf("successfully archived: %s", outputPath)

	if err := uploadFile(outputPath, apiKey); err != nil {
		log.Fatal("internal error: ", err)
	}
	log.Printf("successfully uploaded: %s", outputPath)
}

func uploadFile(path string, apiKey string) error {
	req, err := newFileUploadRequest("http://api.test-doc.com/project/upload", path, apiKey)
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

func newFileUploadRequest(uri string, path string, apiKey string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("archive", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
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
