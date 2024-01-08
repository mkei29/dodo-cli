package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func collectFiles() []string {
	fileList := make([]string, 0, 100)
	filepath.WalkDir("./docs", func(path string, info os.DirEntry, err error) error {
		if !isValidMarkdown(path) {
			return nil
		}
		fileList = append(fileList, path)
		return nil
	})
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

	stat, err := targetFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stat: %w", err)
	}
	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}
	w, err := writer.CreateHeader(header)
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
