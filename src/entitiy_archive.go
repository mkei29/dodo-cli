package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
	appErrors "github.com/toritoritori29/dodo-cli/src/errors"
)

const (
	DefaultArchivePath = "dodo.zip"
	DocsDir            = "docs"
	BlobsDir           = "blobs"
)

type Archive struct {
	File          *os.File
	Metadata      *Metadata
	shouldCleanUp bool
}

func NewArchive(path string) (*Archive, error) {
	// Prepare archive file
	if path == "" {
		zipFile, err := os.CreateTemp("", DefaultArchivePath)
		if err != nil {
			log.Error("failed to create a temporary file")
			return nil, fmt.Errorf("failed to create a temporary file: %w", err)
		}
		return &Archive{
			File:          zipFile,
			shouldCleanUp: true,
		}, nil
	}

	zipFile, err := os.Create(path)
	if err != nil {
		log.Errorf("failed to create an archive file at '%s'", path)
		return nil, fmt.Errorf("failed to create a file. Path: %s: %w", path, err)
	}
	return &Archive{
		File:          zipFile,
		shouldCleanUp: false,
	}, nil
}

func (a *Archive) Close() error {
	err := a.File.Close()
	if err != nil {
		return fmt.Errorf("failed to close the archive file: %w", err)
	}
	if a.shouldCleanUp {
		if err = os.Remove(a.File.Name()); err != nil {
			return fmt.Errorf("failed to remove the archive file: %w", err)
		}
	}
	return nil
}

func (a *Archive) Archive(metadata *Metadata) *appErrors.MultiError {
	// Archive documents
	zipWriter := zip.NewWriter(a.File)
	defer zipWriter.Close()

	// New docs logics
	merr := appErrors.NewMultiError()
	for _, page := range metadata.Page.ListPageHeader() {
		if page.Type != PageTypeLeafNode {
			continue
		}

		for _, lang := range page.Language {
			from := lang.Filepath
			to := filepath.Join(BlobsDir, lang.Hash)
			if err := addFile(from, to, zipWriter); err != nil {
				merr.Add(err)
			}
		}
	}

	// Archive assets
	// Add assets with the hash name under the `blobs` directory.
	for _, asset := range metadata.Asset {
		from := asset.Path
		to := filepath.Join(BlobsDir, filepath.Base(asset.Hash))
		if err := addFile(from, to, zipWriter); err != nil {
			merr.Add(err)
		}
	}

	// Add metadata
	a.Metadata = metadata
	err := addMetadata(metadata, zipWriter)
	if err != nil {
		merr.Add(err)
	}
	if merr.HasError() {
		return &merr
	}
	return nil
}

func (a *Archive) Upload(url string, apiKey string) (*UploadResponse, error) {
	if a.Metadata == nil {
		return nil, errors.New("metadata is not set. Please call Archive() before Upload()")
	}
	req, err := newFileUploadRequest(url, a.Metadata, a.File, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error occurred during communication with the server: %w", err)
	}
	defer resp.Body.Close()

	data, err := ParseUploadResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to upload file: Status %d,  %s", resp.StatusCode, data.Message)
	}
	return data, nil
}

func addFile(from, to string, writer *zip.Writer) error {
	targetFile, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("failed to open the file. File: %s: %w", from, err)
	}
	defer targetFile.Close()

	log.Debug(fmt.Sprintf("add %s to archive", from))
	w, err := writer.Create(to)
	if err != nil {
		return fmt.Errorf("failed to get zip writer: %w", err)
	}
	_, err = io.Copy(w, targetFile)
	if err != nil {
		return fmt.Errorf("failed to write file into zip archive: %w", err)
	}
	return nil
}

func addMetadata(metadata *Metadata, writer *zip.Writer) error {
	metadataJSON, err := metadata.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}
	log.Debug("add metadata.json to archive")

	w, err := writer.Create("metadata.json")
	if err != nil {
		return fmt.Errorf("failed to get zip writer: %w", err)
	}
	_, err = io.Copy(w, bytes.NewReader(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to write metadata into zip archive: %w", err)
	}
	return nil
}

func newFileUploadRequest(uri string, metadata *Metadata, zipFile *os.File, apiKey string) (*http.Request, error) {
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
