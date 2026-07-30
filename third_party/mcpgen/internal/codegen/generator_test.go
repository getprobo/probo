package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"go.probo.inc/mcpgen/internal/config"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello world", "HelloWorld"},
		{"my_api_key", "MyApiKey"},
		{"user-profile-settings", "UserProfileSettings"},
		{"_leading", "Leading"},
		{"trailing_", "Trailing"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToHandlerName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"create_task", "CreateTask"},
		{"get-user", "GetUser"},
		{"list items", "ListItems"},
		{"simple", "Simple"},
		{"my_custom_handler", "MyCustomHandler"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toHandlerName(tt.input)
			if got != tt.want {
				t.Errorf("toHandlerName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractURIParams(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "single parameter",
			template: "users://{id}",
			want:     []string{"id"},
		},
		{
			name:     "multiple parameters",
			template: "orgs://{orgId}/members/{memberId}",
			want:     []string{"orgId", "memberId"},
		},
		{
			name:     "no parameters",
			template: "static://resource",
			want:     []string{},
		},
		{
			name:     "parameter at start",
			template: "{userId}/profile",
			want:     []string{"userId"},
		},
		{
			name:     "parameter at end",
			template: "users/{id}",
			want:     []string{"id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractURIParams(tt.template)
			if len(got) != len(tt.want) {
				t.Errorf("extractURIParams(%q) returned %d params, want %d", tt.template, len(got), len(tt.want))
				return
			}
			for i, param := range got {
				if paramName, ok := param["Name"].(string); ok {
					if paramName != tt.want[i] {
						t.Errorf("extractURIParams(%q)[%d] = %q, want %q", tt.template, i, paramName, tt.want[i])
					}
				} else {
					t.Errorf("extractURIParams(%q)[%d] missing Name field", tt.template, i)
				}
			}
		})
	}
}

func TestGenerateSchemaCode(t *testing.T) {
	gen := &Generator{}

	tests := []struct {
		name   string
		schema *config.Schema
		want   string
	}{
		{
			name:   "nil schema",
			schema: nil,
			want:   "mustUnmarshalSchema(`null`)", // JSON marshals nil to "null"
		},
		{
			name: "simple schema",
			schema: &config.Schema{
				Type: "string",
			},
			want: "mustUnmarshalSchema(`{\"type\":\"string\"}`)",
		},
		{
			name: "schema with properties",
			schema: &config.Schema{
				Type: "object",
				Properties: map[string]*config.Schema{
					"name": {Type: "string"},
				},
			},
			want: "mustUnmarshalSchema(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gen.generateSchemaCode(tt.schema)
			if !containsString(got, tt.want) {
				t.Errorf("generateSchemaCode() = %q, should contain %q", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: []string{"foo", "bar", "baz"},
			item:  "bar",
			want:  true,
		},
		{
			name:  "item does not exist",
			slice: []string{"foo", "bar", "baz"},
			item:  "qux",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "foo",
			want:  false,
		},
		{
			name:  "nil slice",
			slice: nil,
			item:  "foo",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
			}
		})
	}
}

func TestGetRequiredHandlerNames(t *testing.T) {
	spec := &config.MCPSpec{
		Tools: []config.Tool{
			{Name: "create_task"},
			{Name: "update-task"},
		},
		Resources: []config.Resource{
			{Name: "task_resource"},
		},
		Prompts: []config.Prompt{
			{Name: "help_prompt"},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	got := gen.getRequiredHandlerNames()

	if len(got) != 4 {
		t.Errorf("getRequiredHandlerNames() returned %d handlers, want 4", len(got))
	}

	expected := map[string]bool{
		"CreateTaskTool":       true,
		"UpdateTaskTool":       true,
		"TaskResourceResource": true,
		"HelpPromptPrompt":     true,
	}

	for _, name := range got {
		if !expected[name] {
			t.Errorf("Unexpected handler name: %q", name)
		}
		delete(expected, name)
	}

	if len(expected) > 0 {
		t.Errorf("Missing handler names: %v", expected)
	}
}

func TestLoadSchemas(t *testing.T) {
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
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{},
		},
	}

	gen := New(cfg, spec)

	err = gen.loadSchemas()
	if err != nil {
		t.Fatalf("loadSchemas() error = %v", err)
	}

	componentSchemas := []string{"Timestamp", "UUID", "Decimal", "Metadata", "Duration", "Status", "Task", "OptionalFields", "Project"}
	for _, schemaName := range componentSchemas {
		if _, exists := gen.typeGen.schemas[schemaName]; !exists {
			t.Errorf("Expected schema %q to be loaded", schemaName)
		}
	}

	toolInputs := []string{"CreateTaskInput", "UpdateTaskInput", "CreateProjectInput"}
	for _, inputName := range toolInputs {
		if _, exists := gen.typeGen.schemas[inputName]; !exists {
			t.Errorf("Expected tool input schema %q to be loaded", inputName)
		}
	}

	resourceSchemas := []string{"TaskResourceContent", "ProjectResourceContent"}
	for _, schemaName := range resourceSchemas {
		if _, exists := gen.typeGen.schemas[schemaName]; !exists {
			t.Errorf("Expected resource schema %q to be loaded", schemaName)
		}
	}

	promptArgs := []string{"TaskSummaryArgs"}
	for _, argsName := range promptArgs {
		if _, exists := gen.typeGen.schemas[argsName]; !exists {
			t.Errorf("Expected prompt args schema %q to be loaded", argsName)
		}
	}

	customMappings := []string{"Timestamp", "UUID", "Decimal", "Metadata", "Duration"}
	for _, schemaName := range customMappings {
		if _, exists := gen.typeGen.customMappings[schemaName]; !exists {
			t.Errorf("Expected custom mapping for %q from go.probo.inc/mcpgen/type annotation", schemaName)
		}
	}
}

func TestLoadSchemasWithConfigMappings(t *testing.T) {
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
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{
				"Timestamp": {Model: "time.Time"},
				"UUID":      {Model: "github.com/google/uuid.UUID"},
				"User":      {Model: "github.com/myapp/models.User"},
			},
		},
	}

	gen := New(cfg, spec)

	if _, exists := gen.typeGen.customMappings["Timestamp"]; !exists {
		t.Error("Expected custom mapping for Timestamp from config")
	}
	if _, exists := gen.typeGen.customMappings["UUID"]; !exists {
		t.Error("Expected custom mapping for UUID from config")
	}
	if _, exists := gen.typeGen.customMappings["User"]; !exists {
		t.Error("Expected custom mapping for User from config")
	}

	err = gen.loadSchemas()
	if err != nil {
		t.Fatalf("loadSchemas() error = %v", err)
	}

	expectedSchemas := []string{"Timestamp", "UUID", "User", "Event"}
	for _, schemaName := range expectedSchemas {
		if _, exists := gen.typeGen.schemas[schemaName]; !exists {
			t.Errorf("Expected schema %q to be loaded", schemaName)
		}
	}
}

func TestBuildServerTemplateData(t *testing.T) {
	spec := &config.MCPSpec{
		Info: config.ServerInfo{
			Title:   "test-server",
			Version: "1.0.0",
		},
		Tools: []config.Tool{
			{
				Name:        "create_task",
				Description: "Create a task",
				InputSchema: &config.Schema{
					Type: "object",
					Properties: map[string]*config.Schema{
						"title": {Type: "string"},
					},
				},
			},
		},
		Resources: []config.Resource{
			{
				Name:        "task",
				URITemplate: "task://{id}",
				Description: "Task resource",
			},
		},
		Prompts: []config.Prompt{
			{
				Name:        "help",
				Description: "Get help",
				Arguments: []config.PromptArgument{
					{Name: "topic", Required: true},
				},
			},
		},
	}

	cfg := &config.Config{
		Model: config.ModelConfig{
			Package: "test",
		},
		Resolver: config.ResolverConfig{
			Package: "test",
			Type:    "Resolver",
		},
	}

	gen := New(cfg, spec)
	data := gen.buildServerTemplateData()

	if data["ServerName"] != "test-server" {
		t.Errorf("ServerName = %v, want test-server", data["ServerName"])
	}
	if data["ServerVersion"] != "1.0.0" {
		t.Errorf("ServerVersion = %v, want 1.0.0", data["ServerVersion"])
	}

	tools, ok := data["Tools"].([]map[string]interface{})
	if !ok {
		t.Fatal("Tools should be []map[string]interface{}")
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if tools[0]["Name"] != "create_task" {
		t.Errorf("Tool name = %v, want create_task", tools[0]["Name"])
	}
	if tools[0]["HandlerName"] != "CreateTask" {
		t.Errorf("Handler name = %v, want CreateTask", tools[0]["HandlerName"])
	}

	resources, ok := data["Resources"].([]map[string]interface{})
	if !ok {
		t.Fatal("Resources should be []map[string]interface{}")
	}
	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}
	if resources[0]["Name"] != "task" {
		t.Errorf("Resource name = %v, want task", resources[0]["Name"])
	}

	prompts, ok := data["Prompts"].([]map[string]interface{})
	if !ok {
		t.Fatal("Prompts should be []map[string]interface{}")
	}
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0]["Name"] != "help" {
		t.Errorf("Prompt name = %v, want help", prompts[0]["Name"])
	}

	if data["HasResources"] != true {
		t.Error("HasResources should be true")
	}
	if data["HasPrompts"] != true {
		t.Error("HasPrompts should be true")
	}
}

func TestExtractOrphanedHandlerNames(t *testing.T) {
	// Create a temporary file with orphaned handlers section
	content := `package test

// Some code here

// ==============================================================================
// Orphaned Handlers
// ==============================================================================
// The following handlers were found in the resolver file but are no longer
// defined in the MCP specification. They have been preserved here as comments
// in case you need to reference or restore them.
// ==============================================================================

// Orphaned: OldHandler
func (r *Resolver) OldHandler() {
}

// Orphaned: AnotherOldHandler
func (r *Resolver) AnotherOldHandler() {
}
`

	tmpFile := filepath.Join(t.TempDir(), "resolver.go")
	if err := writeFile(tmpFile, []byte(content)); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	got := extractOrphanedHandlerNames(tmpFile)

	expected := []string{"OldHandler", "AnotherOldHandler"}
	if len(got) != len(expected) {
		t.Errorf("extractOrphanedHandlerNames() returned %d names, want %d", len(got), len(expected))
		return
	}

	for i, name := range expected {
		if got[i] != name {
			t.Errorf("extractOrphanedHandlerNames()[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestExtractOrphanedHandlerNamesNoSection(t *testing.T) {
	// Create a file without orphaned section
	content := `package test

func (r *Resolver) NormalHandler() {
}
`

	tmpFile := filepath.Join(t.TempDir(), "resolver.go")
	if err := writeFile(tmpFile, []byte(content)); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	got := extractOrphanedHandlerNames(tmpFile)

	if len(got) != 0 {
		t.Errorf("extractOrphanedHandlerNames() returned %d names, want 0", len(got))
	}
}

func writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

func TestGenerateModels(t *testing.T) {
	specPath := filepath.Join("testdata", "custom_types.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	outputDir := t.TempDir()
	cfg := &config.Config{
		Spec:   specPath,
		Output: outputDir,
		Model: config.ModelConfig{
			Package:  "test",
			Filename: "models.go",
		},
		Resolver: config.ResolverConfig{
			Package:  "test",
			Type:     "Resolver",
		},
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{},
		},
	}

	gen := New(cfg, spec)

	// Load schemas first
	if err := gen.loadSchemas(); err != nil {
		t.Fatalf("loadSchemas() error = %v", err)
	}

	if err := gen.generateModels(); err != nil {
		t.Fatalf("generateModels() error = %v", err)
	}

	modelsPath := filepath.Join(outputDir, "models.go")
	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		t.Errorf("models.go should be created at %s", modelsPath)
	}

	content, err := os.ReadFile(modelsPath)
	require.NoError(t, err, "Failed to read models.go")

	codeStr := string(content)

	if !containsString(codeStr, "package test") {
		t.Error("Generated models should have correct package declaration")
	}

	if containsString(codeStr, "type Timestamp") {
		t.Error("Should not generate Timestamp type (has go.probo.inc/mcpgen/type)")
	}
	if containsString(codeStr, "type UUID") {
		t.Error("Should not generate UUID type (has go.probo.inc/mcpgen/type)")
	}

	if !containsString(codeStr, "type Task struct") {
		t.Error("Should generate Task type")
	}

	if !containsString(codeStr, `"time"`) {
		t.Error("Should import time package")
	}
	if !containsString(codeStr, `"github.com/google/uuid"`) {
		t.Error("Should import uuid package")
	}
}

func TestFullGenerateWorkflow(t *testing.T) {
	specPath := filepath.Join("testdata", "config_based_types.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	outputDir := t.TempDir()
	cfg := &config.Config{
		Spec:   specPath,
		Output: outputDir,
		Exec: config.ExecConfig{
			Package:  "test",
			Filename: "server.go",
		},
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
			Models: map[string]config.TypeMapping{
				"Timestamp": {Model: "time.Time"},
				"UUID":      {Model: "github.com/google/uuid.UUID"},
			},
		},
	}

	gen := New(cfg, spec)

	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expectedFiles := []string{
		"models.go",
		"server.go",
		"resolver.go",
		"schema.resolvers.go",
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(outputDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s should be created", filename)
		}
	}

	modelsContent, err := os.ReadFile(filepath.Join(outputDir, "models.go"))
	require.NoError(t, err, "Failed to read models.go")

	modelsStr := string(modelsContent)
	if !containsString(modelsStr, "type Event struct") {
		t.Error("models.go should contain Event type")
	}

	serverContent, err := os.ReadFile(filepath.Join(outputDir, "server.go"))
	require.NoError(t, err, "Failed to read server.go")

	serverStr := string(serverContent)
	if !containsString(serverStr, "func New(") {
		t.Error("server.go should contain New function")
	}
	if !containsString(serverStr, "mcp.NewServer") {
		t.Error("server.go should use MCP SDK")
	}

	resolverContent, err := os.ReadFile(filepath.Join(outputDir, "resolver.go"))
	require.NoError(t, err, "Failed to read resolver.go")

	resolverStr := string(resolverContent)
	if !containsString(resolverStr, "type Resolver struct") {
		t.Error("resolver.go should contain Resolver struct")
	}

	resolversContent, err := os.ReadFile(filepath.Join(outputDir, "schema.resolvers.go"))
	require.NoError(t, err, "Failed to read schema.resolvers.go")

	resolversStr := string(resolversContent)
	if !containsString(resolversStr, "func (r *Resolver) CreateEvent") {
		t.Error("schema.resolvers.go should contain CreateEvent handler")
	}
}

func TestGenerateWithDifferentPackages(t *testing.T) {
	specPath := filepath.Join("testdata", "custom_types.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	outputDir := t.TempDir()
	cfg := &config.Config{
		Spec:   specPath,
		Output: outputDir,
		Exec: config.ExecConfig{
			Package:  "server",
			Filename: "server.go",
		},
		Model: config.ModelConfig{
			Package:  "types",
			Filename: "types/models.go",
		},
		Resolver: config.ResolverConfig{
			Package:  "mcp_v1",
			Filename: "resolver.go",
			Type:     "Resolver",
			Preserve: false,
		},
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{},
		},
	}

	gen := New(cfg, spec)

	// Load schemas
	if err := gen.loadSchemas(); err != nil {
		t.Fatalf("loadSchemas() error = %v", err)
	}

	data := gen.buildServerTemplateData()

	// When model package differs from exec package, should have imports
	imports, ok := data["Imports"]
	assert.True(t, ok, "Should have Imports when packages differ")

	if imports != nil {
		importList, ok := imports.([]map[string]string)
		require.True(t, ok, "Imports should be []map[string]string")
		assert.NotEmpty(t, importList, "Imports list should not be empty when packages differ")
	}
}

func TestCountOrphanedHandlers(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		want   int
	}{
		{
			name: "no orphaned handlers",
			code: "func (r *Resolver) Handler() {}",
			want: 0,
		},
		{
			name: "one orphaned handler",
			code: "// Orphaned: OldHandler\nfunc (r *Resolver) OldHandler() {}",
			want: 1,
		},
		{
			name: "multiple orphaned handlers",
			code: "// Orphaned: Handler1\nfunc (r *Resolver) Handler1() {}\n// Orphaned: Handler2\nfunc (r *Resolver) Handler2() {}",
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countOrphanedHandlers(tt.code)
			if got != tt.want {
				t.Errorf("countOrphanedHandlers() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGenerateNewHandlersCode(t *testing.T) {
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
		Models: config.ModelsConfig{
			Models: map[string]config.TypeMapping{},
		},
	}

	gen := New(cfg, spec)
	if err := gen.loadSchemas(); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	tests := []struct {
		name         string
		handlerNames []string
		wantContains []string
		isEmpty      bool
	}{
		{
			name:         "empty handler list",
			handlerNames: []string{},
			isEmpty:      true,
		},
		{
			name:         "single handler",
			handlerNames: []string{"CreateEventTool"},
			wantContains: []string{"CreateEventTool", "func (r *Resolver) CreateEventTool"},
			isEmpty:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := gen.generateNewHandlersCode(tt.handlerNames)
			if err != nil {
				t.Errorf("generateNewHandlersCode() error = %v", err)
				return
			}

			if tt.isEmpty {
				if code != "" {
					t.Errorf("Expected empty code, got: %q", code)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !containsString(code, want) {
					t.Errorf("Generated code should contain %q\nGot: %s", want, code)
				}
			}
		})
	}
}

func TestResolveAllRefs(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{
				"SimpleType": {
					Type:   "string",
					Format: "uuid",
				},
				"ComplexType": {
					Type: "object",
					Properties: map[string]*config.Schema{
						"id": {
							Ref: "#/components/schemas/SimpleType",
						},
						"name": {
							Type: "string",
						},
					},
				},
				"WithExtensions": {
					Type:   "string",
					Format: "date-time",
					Extra: map[string]interface{}{
						"go.probo.inc/mcpgen/type":      "time.Time",
						"x-custom-field": "value",
					},
				},
				"NestedRef": {
					Type: "object",
					Properties: map[string]*config.Schema{
						"complex": {
							Ref: "#/components/schemas/ComplexType",
						},
					},
				},
			},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	tests := []struct {
		name           string
		schema         *config.Schema
		wantType       string
		wantNoRef      bool
		wantNoExtras   bool
		checkProperty  string
		propertyType   string
	}{
		{
			name:         "nil schema",
			schema:       nil,
			wantType:     "",
			wantNoRef:    true,
		},
		{
			name: "simple ref resolution",
			schema: &config.Schema{
				Ref: "#/components/schemas/SimpleType",
			},
			wantType:  "string",
			wantNoRef: true,
		},
		{
			name: "nested ref in properties",
			schema: &config.Schema{
				Ref: "#/components/schemas/ComplexType",
			},
			wantType:      "object",
			wantNoRef:     true,
			checkProperty: "id",
			propertyType:  "string",
		},
		{
			name: "deep nested refs",
			schema: &config.Schema{
				Ref: "#/components/schemas/NestedRef",
			},
			wantType:      "object",
			wantNoRef:     true,
			checkProperty: "complex",
			propertyType:  "object",
		},
		{
			name: "ref in array items",
			schema: &config.Schema{
				Type: "array",
				Items: &config.Schema{
					Ref: "#/components/schemas/SimpleType",
				},
			},
			wantType:  "array",
			wantNoRef: true,
		},
		{
			name: "ref in anyOf",
			schema: &config.Schema{
				AnyOf: []*config.Schema{
					{
						Ref: "#/components/schemas/SimpleType",
					},
					{
						Type: "null",
					},
				},
			},
			wantNoRef: true,
		},
		{
			name: "removes x-* extensions",
			schema: &config.Schema{
				Ref: "#/components/schemas/WithExtensions",
			},
			wantType:     "string",
			wantNoRef:    true,
			wantNoExtras: true,
		},
		{
			name: "no ref to resolve",
			schema: &config.Schema{
				Type: "string",
			},
			wantType:  "string",
			wantNoRef: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gen.resolveAllRefs(tt.schema)
			if err != nil {
				t.Fatalf("resolveAllRefs() error = %v", err)
			}

			if tt.schema == nil {
				if got != nil {
					t.Errorf("resolveAllRefs(nil) should return nil")
				}
				return
			}

			if tt.wantType != "" && got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}

			if tt.wantNoRef && got.Ref != "" {
				t.Errorf("Ref should be empty, got %q", got.Ref)
			}

			if tt.wantNoExtras && got.Extra != nil && len(got.Extra) > 0 {
				t.Errorf("Extra fields should be removed, got %v", got.Extra)
			}

			if tt.checkProperty != "" {
				if got.Properties == nil {
					t.Error("Properties should not be nil")
					return
				}
				prop, ok := got.Properties[tt.checkProperty]
				if !ok {
					t.Errorf("Property %q should exist", tt.checkProperty)
					return
				}
				if prop.Type != tt.propertyType {
					t.Errorf("Property %q type = %q, want %q", tt.checkProperty, prop.Type, tt.propertyType)
				}
				if prop.Ref != "" {
					t.Errorf("Property %q should have ref resolved, got %q", tt.checkProperty, prop.Ref)
				}
			}

			if tt.schema.Type == "array" && got.Items != nil {
				if got.Items.Ref != "" {
					t.Errorf("Array items should have ref resolved, got %q", got.Items.Ref)
				}
			}

			if len(tt.schema.AnyOf) > 0 {
				for i, schema := range got.AnyOf {
					if schema.Ref != "" {
						t.Errorf("anyOf[%d] should have ref resolved, got %q", i, schema.Ref)
					}
				}
			}
		})
	}
}

func TestResolveAllRefsInAllOf(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{
				"Base": {
					Type: "object",
					Properties: map[string]*config.Schema{
						"id": {Type: "string"},
					},
				},
				"Extended": {
					Type: "object",
					Properties: map[string]*config.Schema{
						"name": {Type: "string"},
					},
				},
			},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	schema := &config.Schema{
		AllOf: []*config.Schema{
			{Ref: "#/components/schemas/Base"},
			{Ref: "#/components/schemas/Extended"},
		},
	}

	got, err := gen.resolveAllRefs(schema)
	if err != nil {
		t.Fatalf("resolveAllRefs() error = %v", err)
	}

	if len(got.AllOf) != 2 {
		t.Errorf("AllOf length = %d, want 2", len(got.AllOf))
	}

	for i, s := range got.AllOf {
		if s.Ref != "" {
			t.Errorf("allOf[%d] should have ref resolved, got %q", i, s.Ref)
		}
		if s.Type != "object" {
			t.Errorf("allOf[%d] type = %q, want object", i, s.Type)
		}
	}
}

func TestResolveAllRefsInOneOf(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{
				"Option1": {Type: "string"},
				"Option2": {Type: "number"},
			},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	schema := &config.Schema{
		OneOf: []*config.Schema{
			{Ref: "#/components/schemas/Option1"},
			{Ref: "#/components/schemas/Option2"},
		},
	}

	got, err := gen.resolveAllRefs(schema)
	if err != nil {
		t.Fatalf("resolveAllRefs() error = %v", err)
	}

	if len(got.OneOf) != 2 {
		t.Errorf("OneOf length = %d, want 2", len(got.OneOf))
	}

	for i, s := range got.OneOf {
		if s.Ref != "" {
			t.Errorf("oneOf[%d] should have ref resolved, got %q", i, s.Ref)
		}
	}
}

func TestResolveAllRefsInAdditionalProperties(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{
				"Value": {Type: "string"},
			},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	schema := &config.Schema{
		Type: "object",
		AdditionalProperties: &config.Schema{
			Ref: "#/components/schemas/Value",
		},
	}

	got, err := gen.resolveAllRefs(schema)
	if err != nil {
		t.Fatalf("resolveAllRefs() error = %v", err)
	}

	if got.AdditionalProperties == nil {
		t.Fatal("AdditionalProperties should not be nil")
	}

	if got.AdditionalProperties.Ref != "" {
		t.Errorf("AdditionalProperties should have ref resolved, got %q", got.AdditionalProperties.Ref)
	}

	if got.AdditionalProperties.Type != "string" {
		t.Errorf("AdditionalProperties type = %q, want string", got.AdditionalProperties.Type)
	}
}

func TestResolveAllRefsInPatternProperties(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{
				"Pattern": {Type: "number"},
			},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	schema := &config.Schema{
		Type: "object",
		PatternProperties: map[string]*config.Schema{
			"^[a-z]+$": {
				Ref: "#/components/schemas/Pattern",
			},
		},
	}

	got, err := gen.resolveAllRefs(schema)
	if err != nil {
		t.Fatalf("resolveAllRefs() error = %v", err)
	}

	if got.PatternProperties == nil {
		t.Fatal("PatternProperties should not be nil")
	}

	pattern, ok := got.PatternProperties["^[a-z]+$"]
	if !ok {
		t.Fatal("Pattern should exist")
	}

	if pattern.Ref != "" {
		t.Errorf("Pattern should have ref resolved, got %q", pattern.Ref)
	}

	if pattern.Type != "number" {
		t.Errorf("Pattern type = %q, want number", pattern.Type)
	}
}

func TestResolveAllRefsError(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	schema := &config.Schema{
		Ref: "#/components/schemas/NonExistent",
	}

	_, err := gen.resolveAllRefs(schema)
	if err == nil {
		t.Error("resolveAllRefs() should return error for non-existent ref")
	}
}

func TestResolveAllRefsRemovesExtensions(t *testing.T) {
	spec := &config.MCPSpec{
		Components: config.Components{
			Schemas: map[string]*config.Schema{
				"TypeWithExtensions": {
					Type:   "object",
					Extra: map[string]interface{}{
						"go.probo.inc/mcpgen/type": "custom.Type",
						"x-custom":  "value",
					},
					Properties: map[string]*config.Schema{
						"field": {
							Type: "string",
							Extra: map[string]interface{}{
								"x-validation": "required",
							},
						},
					},
				},
			},
		},
	}

	cfg := &config.Config{}
	gen := New(cfg, spec)

	schema := &config.Schema{
		Ref: "#/components/schemas/TypeWithExtensions",
	}

	got, err := gen.resolveAllRefs(schema)
	if err != nil {
		t.Fatalf("resolveAllRefs() error = %v", err)
	}

	if got.Extra != nil && len(got.Extra) > 0 {
		t.Errorf("Extra should be removed from resolved schema, got %v", got.Extra)
	}

	if got.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	field := got.Properties["field"]
	if field.Extra != nil && len(field.Extra) > 0 {
		t.Errorf("Extra should be removed from property, got %v", field.Extra)
	}
}

func TestUpdateResolverIncremental(t *testing.T) {
	specPath := filepath.Join("testdata", "config_based_types.yaml")
	spec, err := config.LoadMCPSpec(specPath)
	require.NoError(t, err, "Failed to load spec")

	tests := []struct {
		name             string
		existingResolver string
		wantContains     []string
		wantNotContains  []string
		expectUpdate     bool
	}{
		{
			name: "add new handler",
			existingResolver: `package test

import (
	"context"
	"fmt"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

type Resolver struct{}

type toolResolver struct {
	*Resolver
}

type promptResolver struct {
	*Resolver
}

type resourceResolver struct {
	*Resolver
}

func (r *toolResolver) OldHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, map[string]any, error) {
	return nil, nil, nil
}
`,
			wantContains: []string{
				"func (r *Resolver) CreateEventTool",
				"Orphaned: OldHandler",
			},
			expectUpdate: true,
		},
		{
			name: "no changes needed",
			existingResolver: `package test

import (
	"context"
	"fmt"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

type Resolver struct{}

type toolResolver struct {
	*Resolver
}

type promptResolver struct {
	*Resolver
}

type resourceResolver struct {
	*Resolver
}

func (r *Resolver) CreateEventTool(ctx context.Context, req *mcp.CallToolRequest, input *Event) (*mcp.CallToolResult, map[string]any, error) {
	return nil, nil, fmt.Errorf("create-event not implemented")
}
`,
			wantContains: []string{
				"func (r *Resolver) CreateEventTool",
			},
			wantNotContains: []string{
				"Orphaned Handlers",
			},
			expectUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			resolverFile := filepath.Join(tmpDir, "schema.resolvers.go")

			if err := os.WriteFile(resolverFile, []byte(tt.existingResolver), 0644); err != nil {
				t.Fatalf("Failed to write resolver file: %v", err)
			}

			cfg := &config.Config{
				Spec:   specPath,
				Output: tmpDir,
				Model: config.ModelConfig{
					Package:  "test",
					Filename: "models.go",
				},
				Resolver: config.ResolverConfig{
					Package:  "test",
					Filename: "schema.resolvers.go",
					Type:     "Resolver",
					Preserve: true,
				},
				Models: config.ModelsConfig{
					Models: map[string]config.TypeMapping{},
				},
			}

			gen := New(cfg, spec)
			if err := gen.loadSchemas(); err != nil {
				t.Fatalf("Failed to load schemas: %v", err)
			}

			err := gen.updateResolverIncremental(resolverFile)
			if err != nil {
				t.Fatalf("updateResolverIncremental() error = %v", err)
			}

			content, err := os.ReadFile(resolverFile)
			if err != nil {
				t.Fatalf("Failed to read updated resolver: %v", err)
			}

			contentStr := string(content)

			for _, want := range tt.wantContains {
				if !containsString(contentStr, want) {
					t.Errorf("Updated resolver should contain %q\nGot: %s", want, contentStr)
				}
			}

			for _, wantNot := range tt.wantNotContains {
				if containsString(contentStr, wantNot) {
					t.Errorf("Updated resolver should NOT contain %q", wantNot)
				}
			}
		})
	}
}

