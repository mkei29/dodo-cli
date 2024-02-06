package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func IsUnderRootPath(root string, path string) error {
	absPath, err := cleanAbs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	absRoot, err := cleanAbs(root)
	if err != nil {
		return fmt.Errorf("failed to get absolute root path: %w", err)
	}

	if !strings.HasPrefix(absPath, absRoot) {
		return fmt.Errorf("path is not under root: %s", path)
	}
	return nil
}

func cleanAbs(path string) (string, error) {
	clean := filepath.Clean(path)
	absPath, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}
