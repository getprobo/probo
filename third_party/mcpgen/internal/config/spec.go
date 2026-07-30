package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type MCPSpec struct {
	Info       ServerInfo `yaml:"info" json:"info"`
	Components Components `yaml:"components,omitempty" json:"components,omitempty"`
	Tools      []Tool     `yaml:"tools,omitempty" json:"tools,omitempty"`
	Resources  []Resource `yaml:"resources,omitempty" json:"resources,omitempty"`
	Prompts    []Prompt   `yaml:"prompts,omitempty" json:"prompts,omitempty"`
}

func LoadMCPSpec(path string) (*MCPSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP spec file: %w", err)
	}

	spec := &MCPSpec{}

	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		var intermediate interface{}
		if err := yaml.Unmarshal(data, &intermediate); err != nil {
			return nil, fmt.Errorf("failed to parse YAML spec: %w", err)
		}
		jsonData, err := json.Marshal(intermediate)
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
		}
		if err := json.Unmarshal(jsonData, spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, spec); err != nil {
			return nil, fmt.Errorf("failed to parse JSON spec: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported spec file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid MCP specification: %w", err)
	}

	return spec, nil
}

func (s *MCPSpec) Validate() error {
	if s.Info.Title == "" {
		return fmt.Errorf("info.title is required")
	}
	if s.Info.Version == "" {
		return fmt.Errorf("info.version is required")
	}

	for i, tool := range s.Tools {
		if tool.Name == "" {
			return fmt.Errorf("tools[%d].name is required", i)
		}
		if tool.InputSchema == nil {
			return fmt.Errorf("tools[%d].inputSchema is required", i)
		}
	}

	for i, resource := range s.Resources {
		if resource.Name == "" {
			return fmt.Errorf("resources[%d].name is required", i)
		}
		if resource.URI == "" && resource.URITemplate == "" {
			return fmt.Errorf("resources[%d] must have either uri or uriTemplate", i)
		}
		if resource.URI != "" && resource.URITemplate != "" {
			return fmt.Errorf("resources[%d] cannot have both uri and uriTemplate", i)
		}
	}

	for i, prompt := range s.Prompts {
		if prompt.Name == "" {
			return fmt.Errorf("prompts[%d].name is required", i)
		}
	}

	return nil
}

func (s *MCPSpec) ResolveSchemaRef(ref string) (*Schema, error) {
	if len(ref) > 0 && ref[0] == '#' {
		if ref == "#/components/schemas" {
			return nil, fmt.Errorf("incomplete schema reference: %s", ref)
		}

		const prefix = "#/components/schemas/"
		if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
			schemaName := ref[len(prefix):]
			if schema, ok := s.Components.Schemas[schemaName]; ok {
				return schema, nil
			}
			return nil, fmt.Errorf("schema not found: %s", schemaName)
		}

		return nil, fmt.Errorf("unsupported reference format: %s", ref)
	}

	return nil, nil
}
