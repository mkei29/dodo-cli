package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	dirPath, err := os.MkdirTemp("", "dodo_test_*")
	require.NoError(t, err)

	args := InitArgs{}
	args.configPath = ".dodo.yaml"
	args.workingDir = dirPath
	args.force = false
	args.debug = false
	args.projectName = "test_project_name"
	args.description = "test_description"

	err = executeInit(args)
	require.NoError(t, err)

	configPath := filepath.Join(dirPath, args.configPath)
	rawContents, err := os.ReadFile(configPath)
	require.NoError(t, err)

	contents := string(rawContents)
	assert.Contains(t, contents, "name: test_project_name", "config file should contain project name definition")
	assert.Contains(t, contents, "description: test_description", "config file should contain description definition")
}
