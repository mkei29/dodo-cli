package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
	"go.uber.org/multierr"
)

// List all files to be archived
func collectFiles(p *Page) []string {
	pages := p.ListPageHeader()
	fileList := make([]string, len(pages))
	for idx, page := range pages {
		if page.IsDirectory {
			continue
		}
		fileList[idx] = page.Filepath
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

	var merr error
	for _, path := range pathList {
		if err := addFile(path, zipWriter); err != nil {
			multierr.Append(merr, fmt.Errorf("failed to add %s to archive", path))
		}
	}
	return merr
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

func isValidMarkdown(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".md" {
		return false
	}

	_, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return true
}
