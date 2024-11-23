package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
)

const DEFAULT_ARCHIVE_PATH = "dodo.zip"

type Archive struct {
	File          *os.File
	shouldCleanUp bool
}

func NewArchive(path string) (*Archive, error) {
	// Prepare archive file
	if path == "" {
		zipFile, err := os.CreateTemp("", DEFAULT_ARCHIVE_PATH)
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
		return os.Remove(a.File.Name())
	}
	return nil
}

func (a *Archive) Archive(metadata *Metadata) ErrorSet {
	// Archive documents
	zipWriter := zip.NewWriter(a.File)
	defer zipWriter.Close()

	es := NewErrorSet()
	pathList := collectFiles(&metadata.Page)
	for _, from := range pathList {
		to := filepath.Join("docs", from)
		if err := addFile(from, to, zipWriter); err != nil {
			es.Add(err)
		}
	}

	// Archive assets
	// Add assets with the hash name under the `blobs` directory.
	for _, asset := range metadata.Asset {
		from := string(asset.Path)
		to := filepath.Join("blobs", filepath.Base(asset.Hash))
		if err := addFile(from, to, zipWriter); err != nil {
			es.Add(err)
		}
	}

	// Add metadata
	err := addMetadata(metadata, zipWriter)
	if err != nil {
		es.Add(err)
	}
	return es
}

// List all files to be archived.
func collectFiles(p *Page) []string {
	log.Debug("enter collectFiles")
	pages := p.ListPageHeader()
	fileList := make([]string, 0, len(pages))
	for _, page := range pages {
		if page.Type == PageTypeLeafNode {
			fileList = append(fileList, page.Filepath)
		}
	}
	return fileList
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
