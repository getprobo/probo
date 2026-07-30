package codegen

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/mcpgen/internal/config"
)

func TestParseTypeMapping(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantGoType     string
		wantImportPath string
	}{
		{
			name:           "built-in type",
			input:          "string",
			wantGoType:     "string",
			wantImportPath: "",
		},
		{
			name:           "standard library type",
			input:          "time.Time",
			wantGoType:     "time.Time",
			wantImportPath: "time",
		},
		{
			name:           "external package with full path",
			input:          "github.com/google/uuid.UUID",
			wantGoType:     "uuid.UUID",
			wantImportPath: "github.com/google/uuid",
		},
		{
			name:           "external package with nested path",
			input:          "github.com/shopspring/decimal.Decimal",
			wantGoType:     "decimal.Decimal",
			wantImportPath: "github.com/shopspring/decimal",
		},
		{
			name:           "json.RawMessage",
			input:          "json.RawMessage",
			wantGoType:     "json.RawMessage",
			wantImportPath: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTypeMapping(tt.input)
			assert.Equal(t, tt.wantGoType, got.GoType)
			assert.Equal(t, tt.wantImportPath, got.ImportPath)
		})
	}
}

func TestExtractGoTypeAnnotation(t *testing.T) {
	tests := []struct {
		name   string
		schema *config.Schema
		want   string
	}{
		{
			name:   "nil schema",
			schema: nil,
			want:   "",
		},
		{
			name: "schema without extra",
			schema: &config.Schema{
				Type: "string",
			},
			want: "",
		},
		{
			name: "schema with go.probo.inc/mcpgen/type annotation",
			schema: &config.Schema{
				Type: "string",
				Extra: map[string]any{
					"go.probo.inc/mcpgen/type": "time.Time",
				},
			},
			want: "time.Time",
		},
		{
			name: "schema with other annotations",
			schema: &config.Schema{
				Type: "string",
				Extra: map[string]any{
					"x-custom": "value",
				},
			},
			want: "",
		},
		{
			name: "schema with non-string go.probo.inc/mcpgen/type",
			schema: &config.Schema{
				Type: "string",
				Extra: map[string]any{
					"go.probo.inc/mcpgen/type": 123,
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractGoTypeAnnotation(tt.schema)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeGeneratorCustomMappings(t *testing.T) {
	tests := []struct {
		name           string
		setupGen       func(*TypeGenerator)
		schema         *config.Schema
		hint           string
		wantType       string
		wantImports    []string
		wantErr        bool
	}{
		{
			name: "custom mapping for ref",
			setupGen: func(g *TypeGenerator) {
				g.AddCustomMapping("Timestamp", &CustomTypeMapping{
					GoType:     "time.Time",
					ImportPath: "time",
				})
			},
			schema: &config.Schema{
				Ref: "#/components/schemas/Timestamp",
			},
			hint:        "CreatedAt",
			wantType:    "time.Time",
			wantImports: []string{"time"},
			wantErr:     false,
		},
		{
			name: "custom mapping for UUID",
			setupGen: func(g *TypeGenerator) {
				g.AddCustomMapping("UUID", &CustomTypeMapping{
					GoType:     "uuid.UUID",
					ImportPath: "github.com/google/uuid",
				})
			},
			schema: &config.Schema{
				Ref: "#/components/schemas/UUID",
			},
			hint:        "ID",
			wantType:    "uuid.UUID",
			wantImports: []string{"github.com/google/uuid"},
			wantErr:     false,
		},
		{
			name: "nullable custom type",
			setupGen: func(g *TypeGenerator) {
				g.AddCustomMapping("Timestamp", &CustomTypeMapping{
					GoType:     "time.Time",
					ImportPath: "time",
				})
			},
			schema: &config.Schema{
				AnyOf: []*config.Schema{
					{Ref: "#/components/schemas/Timestamp"},
					{Type: "null"},
				},
			},
			hint:        "UpdatedAt",
			wantType:    "*time.Time",
			wantImports: []string{"time"},
			wantErr:     false,
		},
		{
			name: "regular ref without custom mapping",
			setupGen: func(g *TypeGenerator) {
				// No custom mapping added
			},
			schema: &config.Schema{
				Ref: "#/components/schemas/User",
			},
			hint:        "Owner",
			wantType:    "*User",
			wantImports: []string{},
			wantErr:     false,
		},
		{
			name: "string type",
			setupGen: func(g *TypeGenerator) {
				// No custom mapping needed
			},
			schema: &config.Schema{
				Type: "string",
			},
			hint:        "Name",
			wantType:    "string",
			wantImports: []string{},
			wantErr:     false,
		},
		{
			name: "time.Time from format",
			setupGen: func(g *TypeGenerator) {
				// No custom mapping, should use format
			},
			schema: &config.Schema{
				Type:   "string",
				Format: "date-time",
			},
			hint:        "CreatedAt",
			wantType:    "time.Time",
			wantImports: []string{"time"},
			wantErr:     false,
		},
		{
			name: "nullable time.Time from format with types array",
			setupGen: func(g *TypeGenerator) {
				// No custom mapping, should use format
			},
			schema: &config.Schema{
				Types:  []string{"string", "null"},
				Format: "date-time",
			},
			hint:        "ContractStartDate",
			wantType:    "*time.Time",
			wantImports: []string{"time"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTypeGenerator()
			tt.setupGen(gen)

			gotType, err := gen.goType(tt.schema, tt.hint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantType, gotType)

			for _, wantImport := range tt.wantImports {
				assert.True(t, gen.imports[wantImport], "expected import %q not found", wantImport)
			}
		})
	}
}

func TestTypeGeneratorSkipsCustomMappedSchemas(t *testing.T) {
	gen := NewTypeGenerator()

	gen.AddCustomMapping("Timestamp", &CustomTypeMapping{
		GoType:     "time.Time",
		ImportPath: "time",
	})

	gen.AddSchema("Timestamp", &config.Schema{
		Type:   "string",
		Format: "date-time",
	})

	// Add a schema that references Timestamp (so the import gets added)
	gen.AddSchema("Event", &config.Schema{
		Type: "object",
		Properties: map[string]*config.Schema{
			"name": {Type: "string"},
			"createdAt": {
				Ref: "#/components/schemas/Timestamp",
			},
		},
	})

	gen.AddSchema("User", &config.Schema{
		Type: "object",
		Properties: map[string]*config.Schema{
			"name": {Type: "string"},
		},
	})

	code, err := gen.Generate("test")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	codeStr := string(code)

	if contains := containsTypeDefinition(codeStr, "type Timestamp"); contains {
		t.Error("Generated code should not contain Timestamp type definition (it's custom mapped)")
	}

	if !containsTypeDefinition(codeStr, "type User struct") {
		t.Error("Generated code should contain User type definition")
	}

	if !containsTypeDefinition(codeStr, "type Event struct") {
		t.Error("Generated code should contain Event type definition")
	}

	if !containsImport(codeStr, "time") {
		t.Error("Generated code should import time package")
	}
}

func TestIsNullableType(t *testing.T) {
	tests := []struct {
		name       string
		schema     *config.Schema
		wantNullable bool
		wantSchema *config.Schema
	}{
		{
			name: "anyOf with null and type",
			schema: &config.Schema{
				AnyOf: []*config.Schema{
					{Type: "string"},
					{Type: "null"},
				},
			},
			wantNullable: true,
			wantSchema:   &config.Schema{Type: "string"},
		},
		{
			name: "anyOf with null and ref",
			schema: &config.Schema{
				AnyOf: []*config.Schema{
					{Ref: "#/components/schemas/User"},
					{Type: "null"},
				},
			},
			wantNullable: true,
			wantSchema:   &config.Schema{Ref: "#/components/schemas/User"},
		},
		{
			name: "anyOf with multiple types (not nullable)",
			schema: &config.Schema{
				AnyOf: []*config.Schema{
					{Type: "string"},
					{Type: "number"},
				},
			},
			wantNullable: false,
			wantSchema:   nil,
		},
		{
			name: "regular type (not nullable)",
			schema: &config.Schema{
				Type: "string",
			},
			wantNullable: false,
			wantSchema:   nil,
		},
		{
			name: "types array with null",
			schema: &config.Schema{
				Types: []string{"string", "null"},
			},
			wantNullable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNullable, gotSchema := isNullableType(tt.schema)
			if gotNullable != tt.wantNullable {
				t.Errorf("isNullableType() nullable = %v, want %v", gotNullable, tt.wantNullable)
			}
			if tt.wantSchema != nil && gotSchema != nil {
				if gotSchema.Type != tt.wantSchema.Type || gotSchema.Ref != tt.wantSchema.Ref {
					t.Errorf("isNullableType() schema = %+v, want %+v", gotSchema, tt.wantSchema)
				}
			}
		})
	}
}

func TestEnumGeneration(t *testing.T) {
	gen := NewTypeGenerator()

	enumSchema := &config.Schema{
		Type: "string",
		Enum: []any{"pending", "in_progress", "completed"},
		Description: "Task status",
	}

	code, err := gen.generateEnum("Status", enumSchema)
	if err != nil {
		t.Fatalf("generateEnum() error = %v", err)
	}

	if !containsTypeDefinition(code, "type Status string") {
		t.Error("Generated enum should contain type definition")
	}

	expectedConstants := []string{"StatusPending", "StatusInProgress", "StatusCompleted"}
	for _, constant := range expectedConstants {
		if !containsString(code, constant) {
			t.Errorf("Generated enum should contain constant %q", constant)
		}
	}

	if !containsString(code, "func (e Status) IsValid() bool") {
		t.Error("Generated enum should contain IsValid method")
	}

	if !containsString(code, "func (e *Status) UnmarshalJSON") {
		t.Error("Generated enum should contain UnmarshalJSON method")
	}
	if !containsString(code, "func (e Status) MarshalJSON") {
		t.Error("Generated enum should contain MarshalJSON method")
	}
}

func containsTypeDefinition(code, typeDef string) bool {
	return containsString(code, typeDef)
}

func containsImport(code, importPath string) bool {
	return containsString(code, `"`+importPath+`"`)
}

func containsString(haystack, needle string) bool {
	// Use jsonschema import to avoid unused import error
	_ = jsonschema.Schema{}
	// Simple contains check
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func TestGenerateArrayType(t *testing.T) {
	tests := []struct {
		name         string
		schema       *config.Schema
		typeName     string
		depth        int
		want         string
		wantContains []string
		wantErr      bool
	}{
		{
			name: "array of strings (top-level)",
			schema: &config.Schema{
				Type:        "array",
				Description: "An array of tags",
				Items: &config.Schema{
					Type: "string",
				},
			},
			typeName:     "Tags",
			depth:        0,
			wantContains: []string{"type Tags []string", "An array of tags"},
			wantErr:      false,
		},
		{
			name: "array of strings (nested)",
			schema: &config.Schema{
				Type: "array",
				Items: &config.Schema{
					Type: "string",
				},
			},
			typeName: "Tags",
			depth:    1,
			want:     "[]string",
			wantErr:  false,
		},
		{
			name: "array of numbers (top-level)",
			schema: &config.Schema{
				Type: "array",
				Items: &config.Schema{
					Type: "number",
				},
			},
			typeName:     "Scores",
			depth:        0,
			wantContains: []string{"type Scores []float64"},
			wantErr:      false,
		},
		{
			name: "array of numbers (nested)",
			schema: &config.Schema{
				Type: "array",
				Items: &config.Schema{
					Type: "number",
				},
			},
			typeName: "Scores",
			depth:    1,
			want:     "[]float64",
			wantErr:  false,
		},
		{
			name: "array of objects (ref, top-level)",
			schema: &config.Schema{
				Type: "array",
				Items: &config.Schema{
					Ref: "#/components/schemas/User",
				},
			},
			typeName:     "Users",
			depth:        0,
			wantContains: []string{"type Users []*User"},
			wantErr:      false,
		},
		{
			name: "array of objects (ref, nested)",
			schema: &config.Schema{
				Type: "array",
				Items: &config.Schema{
					Ref: "#/components/schemas/User",
				},
			},
			typeName: "Users",
			depth:    1,
			want:     "[]*User",
			wantErr:  false,
		},
		{
			name: "array without items (top-level)",
			schema: &config.Schema{
				Type: "array",
			},
			typeName:     "Unknown",
			depth:        0,
			wantContains: []string{"type Unknown []any"},
			wantErr:      false,
		},
		{
			name: "array without items (nested)",
			schema: &config.Schema{
				Type: "array",
			},
			typeName: "Unknown",
			depth:    1,
			want:     "[]any",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTypeGenerator()
			got, err := gen.generateArrayType(tt.typeName, tt.schema, tt.depth)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.want != "" {
				// Exact match test
				if got != tt.want {
					t.Errorf("generateArrayType() = %q, want %q", got, tt.want)
				}
			} else {
				// Contains test
				for _, want := range tt.wantContains {
					if !containsString(got, want) {
						t.Errorf("generateArrayType() should contain %q\nGot: %s", want, got)
					}
				}
			}
		})
	}
}

func TestGeneratePrimitiveTypeAlias(t *testing.T) {
	tests := []struct {
		name         string
		typeName     string
		schema       *config.Schema
		goType       string
		wantContains []string
	}{
		{
			name:     "string with description",
			typeName: "Username",
			schema: &config.Schema{
				Description: "A username string",
			},
			goType:       "string",
			wantContains: []string{"type Username string", "A username string"},
		},
		{
			name:     "integer without description",
			typeName: "Count",
			schema:   &config.Schema{},
			goType:   "int",
			wantContains: []string{"type Count int", "Count represents a int schema"},
		},
		{
			name:     "float64 type",
			typeName: "Score",
			schema:   &config.Schema{},
			goType:   "float64",
			wantContains: []string{"type Score float64"},
		},
		{
			name:     "boolean type",
			typeName: "IsActive",
			schema:   &config.Schema{},
			goType:   "bool",
			wantContains: []string{"type IsActive bool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTypeGenerator()
			got, err := gen.generatePrimitiveTypeAlias(tt.typeName, tt.schema, tt.goType)
			if err != nil {
				t.Errorf("generatePrimitiveTypeAlias() error = %v", err)
				return
			}

			for _, want := range tt.wantContains {
				if !containsString(got, want) {
					t.Errorf("generatePrimitiveTypeAlias() should contain %q\nGot: %s", want, got)
				}
			}
		})
	}
}

func TestGoTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setupGen func(*TypeGenerator)
		schema   *config.Schema
		hint     string
		want     string
		wantErr  bool
	}{
		{
			name: "object with title",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type:  "object",
				Title: "CustomObject",
				Properties: map[string]*config.Schema{
					"name": {Type: "string"},
				},
			},
			hint:    "Field",
			want:    "CustomObject",
			wantErr: false,
		},
		{
			name: "object with properties but no title",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type: "object",
				Properties: map[string]*config.Schema{
					"snapshot_id": {Type: "string"},
				},
			},
			hint:    "Filter",
			want:    "Filter",
			wantErr: false,
		},
		{
			name: "object without title or properties",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type: "object",
			},
			hint:    "Metadata",
			want:    "map[string]any",
			wantErr: false,
		},
		{
			name: "null type",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type: "null",
			},
			hint:    "Value",
			want:    "any",
			wantErr: false,
		},
		{
			name: "schema with only properties (no type)",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Properties: map[string]*config.Schema{
					"id":   {Type: "string"},
					"name": {Type: "string"},
				},
			},
			hint:    "User",
			want:    "User",
			wantErr: false,
		},
		{
			name: "integer type",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type: "integer",
			},
			hint:    "Count",
			want:    "int",
			wantErr: false,
		},
		{
			name: "number type",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type: "number",
			},
			hint:    "Price",
			want:    "float64",
			wantErr: false,
		},
		{
			name: "boolean type",
			setupGen: func(g *TypeGenerator) {
			},
			schema: &config.Schema{
				Type: "boolean",
			},
			hint:    "Active",
			want:    "bool",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTypeGenerator()
			tt.setupGen(gen)

			got, err := gen.goType(tt.schema, tt.hint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGoStringType(t *testing.T) {
	tests := []struct {
		name   string
		schema *config.Schema
		want   string
	}{
		{
			name:   "date-time format",
			schema: &config.Schema{Format: "date-time"},
			want:   "time.Time",
		},
		{
			name:   "date format",
			schema: &config.Schema{Format: "date"},
			want:   "string",
		},
		{
			name:   "email format",
			schema: &config.Schema{Format: "email"},
			want:   "string",
		},
		{
			name:   "uuid format",
			schema: &config.Schema{Format: "uuid"},
			want:   "string",
		},
		{
			name:   "uri format",
			schema: &config.Schema{Format: "uri"},
			want:   "string",
		},
		{
			name:   "no format",
			schema: &config.Schema{},
			want:   "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTypeGenerator()
			got := gen.goStringType(tt.schema)
			assert.Equal(t, tt.want, got)

			if tt.want == "time.Time" && !gen.imports["time"] {
				t.Error("time.Time should add time import")
			}
		})
	}
}

func TestGenerateStruct(t *testing.T) {
	tests := []struct {
		name    string
		schema  *config.Schema
		want    []string // Strings that should be in the output
		wantErr bool
	}{
		{
			name: "struct with description",
			schema: &config.Schema{
				Description: "A user object",
				Type:        "object",
				Properties: map[string]*config.Schema{
					"name": {
						Type:        "string",
						Description: "User name",
					},
					"age": {
						Type: "integer",
					},
				},
				Required: []string{"name"},
			},
			want: []string{
				"// A user object",
				"type User struct",
				"Name string",
				"Age *int", // Age is not required, so it's a pointer
				"`json:\"name\"`",
				"`json:\"age,omitempty\"`",
			},
			wantErr: false,
		},
		{
			name: "struct with title",
			schema: &config.Schema{
				Title: "Person",
				Type:  "object",
				Properties: map[string]*config.Schema{
					"id": {Type: "string"},
				},
			},
			want: []string{
				"// Person",
				"type Person struct",
			},
			wantErr: false,
		},
		{
			name: "struct with no description or title",
			schema: &config.Schema{
				Type: "object",
				Properties: map[string]*config.Schema{
					"value": {Type: "string"},
				},
			},
			want: []string{
				"// Anonymous represents the schema",
				"type Anonymous struct",
			},
			wantErr: false,
		},
		{
			name: "struct with omittable on non-nullable field - should error",
			schema: &config.Schema{
				Type: "object",
				Properties: map[string]*config.Schema{
					"name": {
						Type: "string",
						Extra: map[string]any{
							"go.probo.inc/mcpgen/omittable": true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "struct with omittable on nullable field - should succeed",
			schema: &config.Schema{
				Type: "object",
				Properties: map[string]*config.Schema{
					"description": {
						AnyOf: []*config.Schema{
							{Type: "string"},
							{Type: "null"},
						},
						Extra: map[string]any{
							"go.probo.inc/mcpgen/omittable": true,
						},
					},
				},
			},
			want: []string{
				"type UpdateInput struct",
				"Description mcp.Omittable[*string]",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTypeGenerator()
			typeName := "User"
			if tt.schema.Title != "" {
				typeName = tt.schema.Title
			} else if tt.name == "struct with no description or title" {
				typeName = "Anonymous"
			} else if tt.name == "struct with omittable on nullable field - should succeed" {
				typeName = "UpdateInput"
			}

			got, err := gen.generateStruct(typeName, tt.schema, 0)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, wantStr := range tt.want {
				assert.Contains(t, got, wantStr)
			}
		})
	}
}

func TestToGoTypeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"user-settings", "UserSettings"},
		{"user.data", "UserData"},
		{"user_input.json", "User"}, // .json is stripped first, then _input
		{"task_input_schema", "TaskInput"}, // only _schema is stripped as a suffix
		{"task_schema", "Task"}, // _schema is stripped
		{"my-cool-type", "MyCoolType"},
		{"id", "Id"},
		{"url", "Url"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toGoTypeName(tt.input)
			if got != tt.want {
				t.Errorf("toGoTypeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToGoFieldName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"id", "ID"},
		{"url", "URL"},
		{"uri", "URI"},
		{"api", "API"},
		{"json", "JSON"},
		{"xml", "XML"},
		{"html", "HTML"},
		{"http", "HTTP"},
		{"https", "HTTPS"},
		{"sql", "SQL"},
		{"ssh", "SSH"},
		{"tcp", "TCP"},
		{"udp", "UDP"},
		{"ip", "IP"},
		{"ui", "UI"},
		{"uuid", "UUID"},
		{"jwt", "JWT"},
		{"oauth", "OAuth"},
		{"snapshot_id", "SnapshotID"},
		{"user_id", "UserID"},
		{"account_id", "AccountID"},
		{"resource_url", "ResourceURL"},
		{"redirect_uri", "RedirectURI"},
		{"object_uuid", "ObjectUUID"},
		{"session_jwt", "SessionJWT"},
		{"api_key", "APIKey"},
		{"api_secret", "APISecret"},
		{"http_status", "HTTPStatus"},
		{"https_enabled", "HTTPSEnabled"},
		{"json_data", "JSONData"},
		{"xml_content", "XMLContent"},
		{"html_body", "HTMLBody"},
		{"sql_query", "SQLQuery"},
		{"ip_address", "IPAddress"},
		{"ui_state", "UIState"},
		{"user_api_key", "UserAPIKey"},
		{"get_json_data", "GetJSONData"},
		{"parse_xml_content", "ParseXMLContent"},
		{"api_url", "APIURL"},
		{"http_api_key", "HTTPAPIKey"},
		{"json_api_url", "JSONAPIURL"},
		{"user_name", "UserName"},
		{"first-name", "FirstName"},
		{"created_at", "CreatedAt"},
		{"is_active", "IsActive"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toGoFieldName(tt.input)
			if got != tt.want {
				t.Errorf("toGoFieldName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToEnumConstName(t *testing.T) {
	tests := []struct {
		enumType string
		value    string
		want     string
	}{
		{"Status", "pending", "StatusPending"},
		{"Status", "in_progress", "StatusInProgress"},
		{"ColorType", "red", "ColorRed"},
		{"Priority", "high-priority", "PriorityHighPriority"},
	}

	for _, tt := range tests {
		t.Run(tt.enumType+"_"+tt.value, func(t *testing.T) {
			got := toEnumConstName(tt.enumType, tt.value)
			if got != tt.want {
				t.Errorf("toEnumConstName(%q, %q) = %q, want %q", tt.enumType, tt.value, got, tt.want)
			}
		})
	}
}
