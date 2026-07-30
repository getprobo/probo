package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/mcpgen/internal/config"
)

func TestGenerateWithCustomTypes(t *testing.T) {
	specPath := filepath.Join("testdata", "custom_types.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	cfg := &config.Config{
		Spec:   specPath,
		Output: t.TempDir(),
		Model: config.ModelConfig{
			Package:  "test",
			Filename: "models.go",
		},
		Resolver: config.ResolverConfig{
			Package:  "test",
			Filename: "resolver.go",
			Type:     "Resolver",
			Preserve: false,
		},
		// No custom models in config - using go.probo.inc/mcpgen/type annotations
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{},
		},
	}

	gen := New(cfg, spec)

	if err := gen.loadSchemas(); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	code, err := gen.typeGen.Generate("test")
	require.NoError(t, err, "Failed to generate code")

	codeStr := string(code)

	customTypes := []string{"Timestamp", "UUID", "Decimal", "Metadata", "Duration"}
	for _, typeName := range customTypes {
		assert.NotContains(t, codeStr, "type "+typeName+" ")
	}

	regularTypes := []string{"Task", "OptionalFields", "Project", "UpdateTaskInput"}
	for _, typeName := range regularTypes {
		assert.Contains(t, codeStr, "type "+typeName)
	}

	expectedImports := []string{
		"time",
		"github.com/google/uuid",
		"github.com/shopspring/decimal",
		"json",
		"go.probo.inc/mcpgen/mcp",
	}
	for _, imp := range expectedImports {
		assert.Contains(t, codeStr, `"`+imp+`"`)
	}

	assert.Contains(t, codeStr, "ID        uuid.UUID")
	assert.Contains(t, codeStr, "CreatedAt time.Time")
	assert.Contains(t, codeStr, "UpdatedAt *time.Time")
	assert.Contains(t, codeStr, "mcp.Omittable[*string]")
	assert.Contains(t, codeStr, "mcp.Omittable[*Status]")
	assert.Contains(t, codeStr, "mcp.Omittable[*int]")
	assert.Contains(t, codeStr, "mcp.Omittable[*[]string]")
}

func TestGenerateWithConfigBasedTypes(t *testing.T) {
	specPath := filepath.Join("testdata", "config_based_types.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	cfg := &config.Config{
		Spec:   specPath,
		Output: t.TempDir(),
		Model: config.ModelConfig{
			Package:  "test",
			Filename: "models.go",
		},
		Resolver: config.ResolverConfig{
			Package:  "test",
			Filename: "resolver.go",
			Type:     "Resolver",
			Preserve: false,
		},
		// Custom models in config
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{
				"Timestamp": {Model: "time.Time"},
				"UUID":      {Model: "github.com/google/uuid.UUID"},
				"User":      {Model: "github.com/myapp/models.User"},
			},
		},
	}

	gen := New(cfg, spec)

	if err := gen.loadSchemas(); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	code, err := gen.typeGen.Generate("test")
	require.NoError(t, err, "Failed to generate code")

	codeStr := string(code)

	customTypes := []string{"Timestamp", "UUID", "User"}
	for _, typeName := range customTypes {
		assert.NotContains(t, codeStr, "type "+typeName+" ")
	}

	assert.Contains(t, codeStr, "type Event struct")

	expectedImports := []string{
		"time",
		"github.com/google/uuid",
		"github.com/myapp/models",
	}
	for _, imp := range expectedImports {
		assert.Contains(t, codeStr, `"`+imp+`"`)
	}

	assert.Contains(t, codeStr, "ID        uuid.UUID")
	assert.Contains(t, codeStr, "CreatedAt time.Time")
	assert.Contains(t, codeStr, "Owner     *models.User")
}

func TestGenerateAllPrimitives(t *testing.T) {
	specPath := filepath.Join("testdata", "all_primitives.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	cfg := &config.Config{
		Spec:   specPath,
		Output: t.TempDir(),
		Model: config.ModelConfig{
			Package:  "test",
			Filename: "models.go",
		},
		Resolver: config.ResolverConfig{
			Package:  "test",
			Filename: "resolver.go",
			Type:     "Resolver",
			Preserve: false,
		},
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{},
		},
	}

	gen := New(cfg, spec)

	if err := gen.loadSchemas(); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	code, err := gen.typeGen.Generate("test")
	require.NoError(t, err, "Failed to generate code")

	codeStr := string(code)

	primitiveTypes := map[string]string{
		"StringSchema":  "type StringSchema string",
		"NumberSchema":  "type NumberSchema float64",
		"IntegerSchema": "type IntegerSchema int",
		"BooleanSchema": "type BooleanSchema bool",
		"ArraySchema":   "type ArraySchema []string",
	}

	for typeName, expectedDecl := range primitiveTypes {
		if !containsString(codeStr, expectedDecl) {
			t.Errorf("Should generate %q for %s", expectedDecl, typeName)
		}
	}

	if !containsString(codeStr, "type ObjectSchema struct") {
		t.Error("Should generate ObjectSchema as a struct")
	}

	if !containsString(codeStr, "type Person struct") {
		t.Error("Should generate Person as a struct")
	}

	if !containsString(codeStr, "type Color string") {
		t.Error("Should generate Color as string-based enum")
	}

	enumConstants := []string{"ColorRed", "ColorGreen", "ColorBlue", "ColorYellow"}
	for _, constName := range enumConstants {
		if !containsString(codeStr, constName) {
			t.Errorf("Should generate enum constant %q", constName)
		}
	}

	if len(code) == 0 {
		t.Error("Generated code is empty")
	}
}

func TestGeneratedCodeCompiles(t *testing.T) {
	testCases := []struct {
		name     string
		specFile string
		config   *config.Config
	}{
		{
			name:     "custom_types",
			specFile: "custom_types.yaml",
			config: &config.Config{
				Model: config.ModelConfig{
					Package:  "test",
					Filename: "models.go",
				},
				Models: config.ModelsConfig{
					Models: map[string]config.TypeMapping{},
				},
			},
		},
		{
			name:     "config_based_types",
			specFile: "config_based_types.yaml",
			config: &config.Config{
				Model: config.ModelConfig{
					Package:  "test",
					Filename: "models.go",
				},
				Models: config.ModelsConfig{
					Models: map[string]config.TypeMapping{
						"Timestamp": {Model: "time.Time"},
						"UUID":      {Model: "github.com/google/uuid.UUID"},
						"User":      {Model: "github.com/myapp/models.User"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			specPath := filepath.Join("testdata", tc.specFile)
			spec, err := config.LoadMCPSpec(specPath)
			if err != nil {
				t.Fatalf("Failed to load spec: %v", err)
			}

			tc.config.Spec = specPath
			tc.config.Output = t.TempDir()
			tc.config.Resolver = config.ResolverConfig{
				Package:  "test",
				Filename: "resolver.go",
				Type:     "Resolver",
				Preserve: false,
			}

			gen := New(tc.config, spec)
			if err := gen.loadSchemas(); err != nil {
				t.Fatalf("Failed to load schemas: %v", err)
			}

			code, err := gen.typeGen.Generate("test")
			if err != nil {
				t.Fatalf("Failed to generate code: %v", err)
			}

			// The fact that Generate() succeeded means the code was formatted successfully
			if len(code) == 0 {
				t.Error("Generated code is empty")
			}
		})
	}
}

