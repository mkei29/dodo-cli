package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/log"
)

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
