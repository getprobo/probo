package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestNewResolverParser(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantErr   bool
		setupFile bool
	}{
		{
			name: "valid resolver file",
			content: `package test

type Resolver struct{}

func (r *Resolver) GetUser(ctx context.Context) error {
	return nil
}`,
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "non-existent file",
			content:   "",
			setupFile: false,
			wantErr:   true,
		},
		{
			name: "invalid Go syntax",
			content: `package test
func ( {  // invalid syntax
}`,
			setupFile: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testFile string
			if tt.setupFile {
				tmpDir := t.TempDir()
				testFile = filepath.Join(tmpDir, "resolver.go")
				if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			} else {
				testFile = filepath.Join(t.TempDir(), "nonexistent.go")
			}

			parser, err := NewResolverParser(testFile)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if parser == nil {
					t.Error("Expected parser but got nil")
				}
			}
		})
	}
}

func TestExtractHandlers(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		resolverType string
		wantHandlers []string
		wantErr      bool
	}{
		{
			name: "extract toolResolver handlers",
			content: `package test

type toolResolver struct{}

func (r *toolResolver) HandleListTasks(ctx context.Context) error {
	return nil
}

func (r *toolResolver) HandleCreateTask(ctx context.Context) error {
	return nil
}

// Not a handler - no receiver
func HelperFunction() {}
`,
			resolverType: "Resolver",
			wantHandlers: []string{"HandleListTasks", "HandleCreateTask"},
			wantErr:      false,
		},
		{
			name: "extract promptResolver handlers",
			content: `package test

type promptResolver struct{}

func (r *promptResolver) HandleGetPrompt(ctx context.Context) error {
	return nil
}
`,
			resolverType: "Resolver",
			wantHandlers: []string{"HandleGetPrompt"},
			wantErr:      false,
		},
		{
			name: "extract resourceResolver handlers",
			content: `package test

type resourceResolver struct{}

func (r *resourceResolver) HandleReadResource(ctx context.Context) error {
	return nil
}
`,
			resolverType: "Resolver",
			wantHandlers: []string{"HandleReadResource"},
			wantErr:      false,
		},
		{
			name: "mixed resolver types",
			content: `package test

type toolResolver struct{}
type promptResolver struct{}
type resourceResolver struct{}

func (r *toolResolver) HandleTool(ctx context.Context) error {
	return nil
}

func (r *promptResolver) HandlePrompt(ctx context.Context) error {
	return nil
}

func (r *resourceResolver) HandleResource(ctx context.Context) error {
	return nil
}

type OtherType struct{}

func (r *OtherType) NotAHandler(ctx context.Context) error {
	return nil
}
`,
			resolverType: "Resolver",
			wantHandlers: []string{"HandleTool", "HandlePrompt", "HandleResource"},
			wantErr:      false,
		},
		{
			name: "no handlers",
			content: `package test

type Resolver struct{}

func HelperFunction() {}
`,
			resolverType: "Resolver",
			wantHandlers: []string{},
			wantErr:      false,
		},
		{
			name: "pointer and value receivers",
			content: `package test

type toolResolver struct{}

func (r toolResolver) HandleNonPointer(ctx context.Context) error {
	return nil
}

func (r *toolResolver) HandlePointer(ctx context.Context) error {
	return nil
}
`,
			resolverType: "Resolver",
			wantHandlers: []string{"HandleNonPointer", "HandlePointer"}, // Both are accepted
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "resolver.go")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			parser, err := NewResolverParser(testFile)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			handlers, err := parser.ExtractHandlers(tt.resolverType)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(handlers) != len(tt.wantHandlers) {
				t.Errorf("Expected %d handlers, got %d", len(tt.wantHandlers), len(handlers))
			}

			for _, wantName := range tt.wantHandlers {
				if _, ok := handlers[wantName]; !ok {
					t.Errorf("Expected handler %q not found", wantName)
				}
			}

			for name, handler := range handlers {
				if handler.Name != name {
					t.Errorf("Handler name mismatch: got %q, want %q", handler.Name, name)
				}
				if handler.SourceCode == "" {
					t.Errorf("Handler %q has empty source code", name)
				}
				if !strings.Contains(handler.RecvType, "Resolver") {
					t.Errorf("Handler %q has unexpected receiver type: %q", name, handler.RecvType)
				}
			}
		})
	}
}

func TestIdentifyOrphanedHandlers(t *testing.T) {
	tests := []struct {
		name             string
		existingHandlers map[string]*HandlerInfo
		requiredHandlers []string
		wantOrphaned     []string
	}{
		{
			name: "no orphaned handlers",
			existingHandlers: map[string]*HandlerInfo{
				"HandleA": {Name: "HandleA"},
				"HandleB": {Name: "HandleB"},
			},
			requiredHandlers: []string{"HandleA", "HandleB"},
			wantOrphaned:     []string{},
		},
		{
			name: "one orphaned handler",
			existingHandlers: map[string]*HandlerInfo{
				"HandleA": {Name: "HandleA"},
				"HandleB": {Name: "HandleB"},
				"HandleC": {Name: "HandleC"},
			},
			requiredHandlers: []string{"HandleA", "HandleB"},
			wantOrphaned:     []string{"HandleC"},
		},
		{
			name: "all orphaned",
			existingHandlers: map[string]*HandlerInfo{
				"HandleA": {Name: "HandleA"},
				"HandleB": {Name: "HandleB"},
			},
			requiredHandlers: []string{},
			wantOrphaned:     []string{"HandleA", "HandleB"},
		},
		{
			name: "new handlers required",
			existingHandlers: map[string]*HandlerInfo{
				"HandleA": {Name: "HandleA"},
			},
			requiredHandlers: []string{"HandleA", "HandleB", "HandleC"},
			wantOrphaned:     []string{},
		},
		{
			name:             "empty existing handlers",
			existingHandlers: map[string]*HandlerInfo{},
			requiredHandlers: []string{"HandleA"},
			wantOrphaned:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			IdentifyOrphanedHandlers(tt.existingHandlers, tt.requiredHandlers)

			var gotOrphaned []string
			for name, handler := range tt.existingHandlers {
				if handler.IsOrphaned {
					gotOrphaned = append(gotOrphaned, name)
				}
			}

			if len(gotOrphaned) != len(tt.wantOrphaned) {
				t.Errorf("Expected %d orphaned handlers, got %d", len(tt.wantOrphaned), len(gotOrphaned))
			}

			orphanedSet := make(map[string]bool)
			for _, name := range gotOrphaned {
				orphanedSet[name] = true
			}

			for _, wantName := range tt.wantOrphaned {
				if !orphanedSet[wantName] {
					t.Errorf("Expected %q to be orphaned but it wasn't", wantName)
				}
			}
		})
	}
}

func TestFormatOrphanedHandlers(t *testing.T) {
	tests := []struct {
		name     string
		handlers map[string]*HandlerInfo
		wantContains []string
		isEmpty  bool
	}{
		{
			name: "single orphaned handler",
			handlers: map[string]*HandlerInfo{
				"HandleOldTask": {
					Name:       "HandleOldTask",
					IsOrphaned: true,
					SourceCode: `func (r *Resolver) HandleOldTask(ctx context.Context) error {
	return nil
}`,
				},
			},
			wantContains: []string{
				"Orphaned Handlers",
				"Orphaned: HandleOldTask",
				"// func (r *Resolver) HandleOldTask",
			},
			isEmpty: false,
		},
		{
			name: "multiple orphaned handlers",
			handlers: map[string]*HandlerInfo{
				"HandleA": {
					Name:       "HandleA",
					IsOrphaned: true,
					SourceCode: "func (r *Resolver) HandleA() {}",
				},
				"HandleB": {
					Name:       "HandleB",
					IsOrphaned: true,
					SourceCode: "func (r *Resolver) HandleB() {}",
				},
			},
			wantContains: []string{
				"Orphaned: HandleA",
				"Orphaned: HandleB",
			},
			isEmpty: false,
		},
		{
			name: "no orphaned handlers",
			handlers: map[string]*HandlerInfo{
				"HandleActive": {
					Name:       "HandleActive",
					IsOrphaned: false,
					SourceCode: "func (r *Resolver) HandleActive() {}",
				},
			},
			wantContains: []string{},
			isEmpty:      true,
		},
		{
			name:         "empty handlers map",
			handlers:     map[string]*HandlerInfo{},
			wantContains: []string{},
			isEmpty:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOrphanedHandlers(tt.handlers)

			if tt.isEmpty {
				if result != "" {
					t.Errorf("Expected empty result, got: %q", result)
				}
				return
			}

			if result == "" {
				t.Error("Expected non-empty result but got empty string")
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Result should contain %q but doesn't.\nGot: %s", want, result)
				}
			}

		assert.Contains(t, result, "Orphaned Handlers", "Result should contain orphaned handlers header")
		})
	}
}

func TestGetReceiverType(t *testing.T) {
	// but we can add a direct test using a sample AST if needed
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

type toolResolver struct{}

func (r *toolResolver) Method() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser, err := NewResolverParser(testFile)
	require.NoError(t, err, "Failed to create parser")

	handlers, err := parser.ExtractHandlers("Resolver")
	require.NoError(t, err, "Failed to extract handlers")

	if len(handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(handlers))
	}

	handler := handlers["Method"]
	// After transformation, the receiver type should be *Resolver
	if handler.RecvType != "*Resolver" {
		t.Errorf("Expected receiver type '*Resolver' (after transformation), got %q", handler.RecvType)
	}
}

func TestExtractFunctionSource(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

type toolResolver struct{}

func (r *toolResolver) HandleTest(ctx context.Context) error {
	x := 42
	return nil
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser, err := NewResolverParser(testFile)
	require.NoError(t, err, "Failed to create parser")

	handlers, err := parser.ExtractHandlers("Resolver")
	require.NoError(t, err, "Failed to extract handlers")

	handler := handlers["HandleTest"]
	// After transformation, toolResolver should be changed to Resolver
	if !strings.Contains(handler.SourceCode, "func (r *Resolver) HandleTest") {
		t.Errorf("Source code should contain transformed function signature, got: %s", handler.SourceCode)
	}
	assert.Contains(t, handler.SourceCode, "return nil", "Source code should contain function body")
	assert.Contains(t, handler.SourceCode, "x := 42", "Source code should contain function body statements")
}
