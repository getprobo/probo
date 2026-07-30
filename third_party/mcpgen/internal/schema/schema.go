package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/jsonschema-go/jsonschema"
)

type Schema = jsonschema.Schema

type Loader struct {
	schemas map[string]*Schema
	baseDir string
}

func NewLoader(baseDir string) *Loader {
	return &Loader{
		schemas: make(map[string]*Schema),
		baseDir: baseDir,
	}
}

func (l *Loader) Load(path string) (*Schema, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if schema, ok := l.schemas[absPath]; ok {
		return schema, nil
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", path, err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file %s: %w", path, err)
	}

	l.schemas[absPath] = &schema

	return &schema, nil
}

func GetType(s *Schema) string {
	if s.Type != "" {
		return s.Type
	}
	if len(s.Types) > 0 {
		return s.Types[0]
	}
	return ""
}

func IsRequired(s *Schema, propName string) bool {
	for _, req := range s.Required {
		if req == propName {
			return true
		}
	}
	return false
}

// IsOmittable checks if a schema property has the go.probo.inc/mcpgen/omittable annotation set to true.
// This is used to wrap fields in mcp.Omittable[T] to distinguish between
// "not set", "set to null", and "set to value".
func IsOmittable(s *Schema) bool {
	if s == nil || s.Extra == nil {
		return false
	}

	if omittable, ok := s.Extra["go.probo.inc/mcpgen/omittable"]; ok {
		if omittableBool, ok := omittable.(bool); ok {
			return omittableBool
		}
	}

	return false
}
