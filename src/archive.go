package main

import (
	"archive/zip"
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

func archive(zipWriter *zip.Writer, pathList []string) ErrorSet {
	es := NewErrorSet()
	for _, path := range pathList {
		if err := addFile(path, zipWriter); err != nil {
			es.Add(fmt.Errorf("failed to add a file to the archive. File: '%s': %w", path, err))
		}
	}
	return es
}

func addFile(filename string, writer *zip.Writer) error {
	targetFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open the file. File: %s: %w", filename, err)
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
