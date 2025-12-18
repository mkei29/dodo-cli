package config

import (
	"bytes"
	"fmt"
	"io"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// DetectConfigVersion parses the YAML and returns the top-level version number.
// The reader is fully consumed by this function.
func DetectConfigVersion(reader io.Reader) (int, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return 0, fmt.Errorf("failed to read config: %w", err)
	}

	root, err := parser.ParseBytes(buf.Bytes(), parser.Mode(0))
	if err != nil {
		return 0, fmt.Errorf("failed to parse config: %w", err)
	}
	if len(root.Docs) != 1 {
		return 0, fmt.Errorf("there should be only one document. Got %d", len(root.Docs))
	}

	body, ok := root.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		return 0, fmt.Errorf("the root node must be of mapping type")
	}

	for _, mapping := range body.Values {
		if mapping.Key.String() != "version" {
			continue
		}
		intNode, ok := mapping.Value.(*ast.IntegerNode)
		if !ok {
			return 0, fmt.Errorf("`version` must have an integer value")
		}
		switch v := intNode.Value.(type) {
		case int:
			return v, nil
		case uint64:
			return int(v), nil
		default:
			return 0, fmt.Errorf("internal error: `version` has an unexpected type")
		}
	}

	return 0, fmt.Errorf("the `version` field is required")
}
