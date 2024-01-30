package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
)

// List all files to be archived
func collectFiles(configPath string, config *DocumentDefinition) []string {
	pages := config.ListPageHeader()
	fileList := make([]string, len(pages)+1)
	fileList[0] = configPath
	for idx, page := range pages {
		fileList[idx+1] = page.Path
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

	for _, path := range pathList {
		if err := addFile(path, zipWriter); err != nil {
			return fmt.Errorf("failed to add file: %w", err)
		}
	}
	return nil
}

func addFile(filename string, writer *zip.Writer) error {
	targetFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer targetFile.Close()

	log.Debug(fmt.Sprintf("add %s", filename))
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
