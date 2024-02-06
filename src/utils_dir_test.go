package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsUnderRootPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		root        string
		path        string
		expectError bool
	}{
		{
			"should not return error valid root and path is given",
			"./",
			"./tmp",
			false,
		},
		{
			"should work with absolute root and relative path",
			"/",
			"./",
			false,
		},
		{
			"should return error when directory traversal is detected",
			"./",
			"../confidential",
			true,
		},
		{
			"should return error when directory traversal is detected",
			"./",
			"./test../../../confidential",
			true,
		},
	}
	for _, tt := range tests {
		c := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := IsUnderRootPath(c.root, c.path)
			if c.expectError {
				require.Error(t, err, c.name)
				return
			}
			require.NoError(t, err, c.name)
		})
	}
}
