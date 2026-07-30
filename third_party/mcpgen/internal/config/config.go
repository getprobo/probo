package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Spec     string         `yaml:"spec" json:"spec"`
	Output   string         `yaml:"output" json:"output"`
	Exec     ExecConfig     `yaml:"exec,omitempty" json:"exec,omitempty"`
	Resolver ResolverConfig `yaml:"resolver" json:"resolver"`
	Model    ModelConfig    `yaml:"model,omitempty" json:"model,omitempty"`
	Models   ModelsConfig   `yaml:"models,omitempty" json:"models,omitempty"`
}

type ExecConfig struct {
	Package  string `yaml:"package,omitempty" json:"package,omitempty"`
	Filename string `yaml:"filename,omitempty" json:"filename,omitempty"`
}

type ResolverConfig struct {
	Package  string `yaml:"package" json:"package"`
	Filename string `yaml:"filename" json:"filename"`
	Type     string `yaml:"type" json:"type"`
	Preserve bool   `yaml:"preserve" json:"preserve"`
}

type ModelConfig struct {
	Package  string `yaml:"package,omitempty" json:"package,omitempty"`
	Filename string `yaml:"filename,omitempty" json:"filename,omitempty"`
}

type ModelsConfig struct {
	// Map schema names to custom Go types
	// Example: User: github.com/myorg/models.User
	Models map[string]TypeMapping `yaml:",inline,omitempty" json:",inline,omitempty"`
}

type TypeMapping struct {
	// Model is the fully qualified Go type to use
	// Example: github.com/google/uuid.UUID
	Model string `yaml:"model" json:"model"`
}

type ServerInfo struct {
	Title       string `yaml:"title" json:"title"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type Components struct {
	Schemas map[string]*jsonschema.Schema `yaml:"schemas,omitempty" json:"schemas,omitempty"`
}

type Schema = jsonschema.Schema

type ToolHints struct {
	Readonly    bool `yaml:"readonly,omitempty" json:"readonly,omitempty"`
	Destructive bool `yaml:"destructive,omitempty" json:"destructive,omitempty"`
	Idempotent  bool `yaml:"idempotent,omitempty" json:"idempotent,omitempty"`
	OpenWorld   bool `yaml:"openWorld,omitempty" json:"openWorld,omitempty"`
}

type Tool struct {
	Name         string            `yaml:"name" json:"name"`
	Title        string            `yaml:"title,omitempty" json:"title,omitempty"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	InputSchema  *Schema           `yaml:"inputSchema" json:"inputSchema"`
	OutputSchema *Schema           `yaml:"outputSchema,omitempty" json:"outputSchema,omitempty"`
	Hints        *ToolHints        `yaml:"hints,omitempty" json:"hints,omitempty"`
	Annotations  map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Handler      string            `yaml:"handler,omitempty" json:"handler,omitempty"`
}

type Resource struct {
	URI         string            `yaml:"uri,omitempty" json:"uri,omitempty"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	MimeType    string            `yaml:"mimeType,omitempty" json:"mimeType,omitempty"`
	URITemplate string            `yaml:"uriTemplate,omitempty" json:"uriTemplate,omitempty"`
	Schema      *Schema           `yaml:"schema,omitempty" json:"schema,omitempty"`
	Readonly    bool              `yaml:"readonly,omitempty" json:"readonly,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Handler     string            `yaml:"handler,omitempty" json:"handler,omitempty"`
}

type Prompt struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Arguments   []PromptArgument  `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Handler     string            `yaml:"handler,omitempty" json:"handler,omitempty"`
}

type PromptArgument struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

func Load(path string) (*Config, *MCPSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{
		Spec:   "schema.yaml",
		Output: "generated",
		Exec: ExecConfig{
			Package:  "server",
			Filename: "server/server.go",
		},
		Resolver: ResolverConfig{
			Package:  "generated",
			Filename: "resolver.go",
			Type:     "Resolver",
			Preserve: true,
		},
		Model: ModelConfig{
			Package:  "generated",
			Filename: "models.go",
		},
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err := config.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Make output path absolute relative to config file directory
	configDir := filepath.Dir(path)
	if !filepath.IsAbs(config.Output) {
		config.Output = filepath.Join(configDir, config.Output)
	}

	specPath := config.Spec
	if !filepath.IsAbs(specPath) {
		configDir := filepath.Dir(path)
		specPath = filepath.Join(configDir, specPath)
	}

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		basePath := specPath
		for _, ext := range []string{".yaml", ".yml", ".json"} {
			tryPath := basePath
			if filepath.Ext(tryPath) == "" {
				tryPath = basePath + ext
			} else {
				tryPath = basePath[:len(basePath)-len(filepath.Ext(basePath))] + ext
			}
			if _, err := os.Stat(tryPath); err == nil {
				specPath = tryPath
				break
			}
		}
	}

	spec, err := LoadMCPSpec(specPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load MCP spec from %s: %w", specPath, err)
	}

	return config, spec, nil
}

func (c *Config) Validate() error {
	if c.Spec == "" {
		return fmt.Errorf("spec path is required")
	}
	if c.Output == "" {
		return fmt.Errorf("output is required")
	}
	if c.Exec.Package == "" {
		return fmt.Errorf("exec.package is required")
	}
	if c.Resolver.Package == "" {
		return fmt.Errorf("resolver.package is required")
	}
	if c.Model.Package == "" {
		return fmt.Errorf("model.package is required")
	}

	return nil
}

func IsSchemaRef(s *Schema) bool {
	return s != nil && s.Ref != ""
}
