package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/log"
	"go.uber.org/multierr"
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
		if page.Type == PageTypeRootNode {
			fileList = append(fileList, page.Filepath)
		}
	}
	return fileList
}

func archive(output string, pathList []string) error {
	zipFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// TODO: replace mutierr with error set
	var merr error
	for _, path := range pathList {
		if err := addFile(path, zipWriter); err != nil {
			merr = multierr.Append(merr, fmt.Errorf("failed to add %s to archive", path))
		}
	}
	return merr //nolint: wrapcheck
}

func addFile(filename string, writer *zip.Writer) error {
	targetFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer targetFile.Close()

	log.Debug(fmt.Sprintf("add %s to archive", filename))
	w, err := writer.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to get zip writer: %w", err)
	}
	_, err = io.Copy(w, targetFile)
	if err != nil {
		return fmt.Errorf("failed to write file into zip archive: %w", err)
	}
	return nil
}
