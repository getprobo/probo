package codegen

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"go.probo.inc/mcpgen/internal/config"
	"go.probo.inc/mcpgen/internal/schema"
	"golang.org/x/mod/modfile"
)

//go:embed templates/*.gotpl
var templates embed.FS

type Generator struct {
	config       *config.Config
	spec         *config.MCPSpec
	schemaLoader *schema.Loader
	typeGen      *TypeGenerator
}

func New(cfg *config.Config, spec *config.MCPSpec) *Generator {
	typeGen := NewTypeGenerator()

	// Sort schema names for deterministic output
	schemaNames := make([]string, 0, len(cfg.Models.Models))
	for schemaName := range cfg.Models.Models {
		schemaNames = append(schemaNames, schemaName)
	}
	sort.Strings(schemaNames)

	for _, schemaName := range schemaNames {
		typeMapping := cfg.Models.Models[schemaName]
		customMapping := parseTypeMapping(typeMapping.Model)
		typeGen.AddCustomMapping(schemaName, customMapping)
	}

	return &Generator{
		config:       cfg,
		spec:         spec,
		schemaLoader: schema.NewLoader("."),
		typeGen:      typeGen,
	}
}

func (g *Generator) Generate() error {
	if err := g.loadSchemas(); err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	if err := g.generateModels(); err != nil {
		return fmt.Errorf("failed to generate models: %w", err)
	}

	if err := g.generateServer(); err != nil {
		return fmt.Errorf("failed to generate server: %w", err)
	}

	if err := g.generateResolverStruct(); err != nil {
		return fmt.Errorf("failed to generate resolver struct: %w", err)
	}

	if err := g.generateResolverImplementations(); err != nil {
		return fmt.Errorf("failed to generate resolver implementations: %w", err)
	}

	return nil
}

func (g *Generator) loadSchemas() error {
	// Sort schema names for deterministic output
	schemaNames := make([]string, 0, len(g.spec.Components.Schemas))
	for name := range g.spec.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, name := range schemaNames {
		schema := g.spec.Components.Schemas[name]
		if config.IsSchemaRef(schema) {
			s, err := g.schemaLoader.Load(schema.Ref)
			if err != nil {
				return fmt.Errorf("failed to load schema %s: %w", name, err)
			}
			if goType := extractGoTypeAnnotation(s); goType != "" {
				customMapping := parseTypeMapping(goType)
				g.typeGen.AddCustomMapping(name, customMapping)
			}
			g.typeGen.AddSchema(name, s)
		} else {
			if goType := extractGoTypeAnnotation(schema); goType != "" {
				customMapping := parseTypeMapping(goType)
				g.typeGen.AddCustomMapping(name, customMapping)
			}
			g.typeGen.AddSchema(name, schema)
		}
	}

	for _, tool := range g.spec.Tools {
		if tool.InputSchema != nil {
			typeName := toPascalCase(tool.Name) + "Input"
			handlerName := toHandlerName(tool.Name)
			schemaVarName := handlerName + "ToolInputSchema"

			var resolvedSchema *config.Schema
			if config.IsSchemaRef(tool.InputSchema) {
				if len(tool.InputSchema.Ref) > 0 && tool.InputSchema.Ref[0] == '#' {
					resolved, err := g.spec.ResolveSchemaRef(tool.InputSchema.Ref)
					if err != nil {
						return fmt.Errorf("failed to resolve input schema ref for tool %s: %w", tool.Name, err)
					}
					resolvedSchema = resolved
					g.typeGen.AddSchema(typeName, resolvedSchema)
				} else {
					s, err := g.schemaLoader.Load(tool.InputSchema.Ref)
					if err != nil {
						return fmt.Errorf("failed to load input schema for tool %s: %w", tool.Name, err)
					}
					resolvedSchema = s
					g.typeGen.AddSchema(typeName, s)
				}
			} else {
				resolvedSchema = tool.InputSchema
				g.typeGen.AddSchema(typeName, tool.InputSchema)
			}

			if resolvedSchema != nil {
				fullyResolvedSchema, err := g.resolveAllRefs(resolvedSchema)
				if err != nil {
					return fmt.Errorf("failed to fully resolve schema for tool %s: %w", tool.Name, err)
				}
				schemaJSON, err := json.Marshal(fullyResolvedSchema)
				if err == nil {
					g.typeGen.AddSchemaVar(schemaVarName, string(schemaJSON))
				}
			}
		}

		// Process output schema if present
		if tool.OutputSchema != nil {
			typeName := toPascalCase(tool.Name) + "Output"
			handlerName := toHandlerName(tool.Name)
			schemaVarName := handlerName + "ToolOutputSchema"

			var resolvedSchema *config.Schema
			if config.IsSchemaRef(tool.OutputSchema) {
				if len(tool.OutputSchema.Ref) > 0 && tool.OutputSchema.Ref[0] == '#' {
					resolved, err := g.spec.ResolveSchemaRef(tool.OutputSchema.Ref)
					if err != nil {
						return fmt.Errorf("failed to resolve output schema ref for tool %s: %w", tool.Name, err)
					}
					resolvedSchema = resolved
					g.typeGen.AddSchema(typeName, resolvedSchema)
				} else {
					s, err := g.schemaLoader.Load(tool.OutputSchema.Ref)
					if err != nil {
						return fmt.Errorf("failed to load output schema for tool %s: %w", tool.Name, err)
					}
					resolvedSchema = s
					g.typeGen.AddSchema(typeName, s)
				}
			} else {
				resolvedSchema = tool.OutputSchema
				g.typeGen.AddSchema(typeName, tool.OutputSchema)
			}

			if resolvedSchema != nil {
				fullyResolvedSchema, err := g.resolveAllRefs(resolvedSchema)
				if err != nil {
					return fmt.Errorf("failed to fully resolve schema for tool %s: %w", tool.Name, err)
				}
				schemaJSON, err := json.Marshal(fullyResolvedSchema)
				if err == nil {
					g.typeGen.AddSchemaVar(schemaVarName, string(schemaJSON))
				}
			}
		}
	}

	for _, resource := range g.spec.Resources {
		if resource.Schema != nil {
			typeName := toPascalCase(resource.Name) + "Content"
			if config.IsSchemaRef(resource.Schema) {
				if len(resource.Schema.Ref) > 0 && resource.Schema.Ref[0] == '#' {
					resolvedSchema, err := g.spec.ResolveSchemaRef(resource.Schema.Ref)
					if err != nil {
						return fmt.Errorf("failed to resolve schema ref for resource %s: %w", resource.Name, err)
					}
					g.typeGen.AddSchema(typeName, resolvedSchema)
					continue
				}
				s, err := g.schemaLoader.Load(resource.Schema.Ref)
				if err != nil {
					return fmt.Errorf("failed to load schema for resource %s: %w", resource.Name, err)
				}
				g.typeGen.AddSchema(typeName, s)
			} else {
				g.typeGen.AddSchema(typeName, resource.Schema)
			}
		}
	}

	// Generate typed argument structs for prompts
	for _, prompt := range g.spec.Prompts {
		if len(prompt.Arguments) > 0 {
			typeName := toPascalCase(prompt.Name) + "Args"

			// Create a schema from the prompt arguments
			argSchema := &config.Schema{
				Type:       "object",
				Properties: make(map[string]*config.Schema),
				Required:   []string{},
			}

			for _, arg := range prompt.Arguments {
				argSchema.Properties[arg.Name] = &config.Schema{
					Type:        "string",
					Description: arg.Description,
				}
				if arg.Required {
					argSchema.Required = append(argSchema.Required, arg.Name)
				}
			}

			g.typeGen.AddSchema(typeName, argSchema)
		}
	}

	return nil
}

func (g *Generator) resolveAllRefs(s *config.Schema) (*config.Schema, error) {
	if s == nil {
		return nil, nil
	}

	if config.IsSchemaRef(s) {
		if len(s.Ref) > 0 && s.Ref[0] == '#' {
			resolved, err := g.spec.ResolveSchemaRef(s.Ref)
			if err != nil {
				return nil, err
			}
			return g.resolveAllRefs(resolved)
		}
		return s, nil
	}

	result := &config.Schema{
		Type:             s.Type,
		Types:            s.Types,
		Format:           s.Format,
		Description:      s.Description,
		Default:          s.Default,
		Enum:             s.Enum,
		Title:            s.Title,
		Required:         s.Required,
		ReadOnly:         s.ReadOnly,
		WriteOnly:        s.WriteOnly,
		Deprecated:       s.Deprecated,
		Minimum:          s.Minimum,
		Maximum:          s.Maximum,
		ExclusiveMinimum: s.ExclusiveMinimum,
		ExclusiveMaximum: s.ExclusiveMaximum,
		MinLength:        s.MinLength,
		MaxLength:        s.MaxLength,
		Pattern:          s.Pattern,
		MinItems:         s.MinItems,
		MaxItems:         s.MaxItems,
		UniqueItems:      s.UniqueItems,
		MinProperties:    s.MinProperties,
		MaxProperties:    s.MaxProperties,
	}

	if len(s.Properties) > 0 {
		result.Properties = make(map[string]*config.Schema)
		// Sort property names for deterministic output
		propNames := make([]string, 0, len(s.Properties))
		for key := range s.Properties {
			propNames = append(propNames, key)
		}
		sort.Strings(propNames)
		for _, key := range propNames {
			propSchema := s.Properties[key]
			resolvedProp, err := g.resolveAllRefs(propSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve property %s: %w", key, err)
			}
			result.Properties[key] = resolvedProp
		}
	}

	if s.Items != nil {
		resolvedItems, err := g.resolveAllRefs(s.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve items: %w", err)
		}
		result.Items = resolvedItems
	}

	if len(s.AnyOf) > 0 {
		result.AnyOf = make([]*config.Schema, len(s.AnyOf))
		for i, schema := range s.AnyOf {
			resolvedSchema, err := g.resolveAllRefs(schema)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve anyOf[%d]: %w", i, err)
			}
			result.AnyOf[i] = resolvedSchema
		}
	}

	if len(s.AllOf) > 0 {
		result.AllOf = make([]*config.Schema, len(s.AllOf))
		for i, schema := range s.AllOf {
			resolvedSchema, err := g.resolveAllRefs(schema)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve allOf[%d]: %w", i, err)
			}
			result.AllOf[i] = resolvedSchema
		}
	}

	if len(s.OneOf) > 0 {
		result.OneOf = make([]*config.Schema, len(s.OneOf))
		for i, schema := range s.OneOf {
			resolvedSchema, err := g.resolveAllRefs(schema)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve oneOf[%d]: %w", i, err)
			}
			result.OneOf[i] = resolvedSchema
		}
	}

	if s.Not != nil {
		resolvedNot, err := g.resolveAllRefs(s.Not)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve not: %w", err)
		}
		result.Not = resolvedNot
	}

	if s.AdditionalProperties != nil {
		resolvedAdditional, err := g.resolveAllRefs(s.AdditionalProperties)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve additionalProperties: %w", err)
		}
		result.AdditionalProperties = resolvedAdditional
	}

	if len(s.PatternProperties) > 0 {
		result.PatternProperties = make(map[string]*config.Schema)
		// Sort pattern names for deterministic output
		patterns := make([]string, 0, len(s.PatternProperties))
		for pattern := range s.PatternProperties {
			patterns = append(patterns, pattern)
		}
		sort.Strings(patterns)
		for _, pattern := range patterns {
			patternSchema := s.PatternProperties[pattern]
			resolvedPattern, err := g.resolveAllRefs(patternSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve patternProperties[%s]: %w", pattern, err)
			}
			result.PatternProperties[pattern] = resolvedPattern
		}
	}

	return result, nil
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func (g *Generator) generateModels() error {
	code, err := g.typeGen.Generate(g.config.Model.Package)
	if err != nil {
		return err
	}

	modelsFile := "models.go"
	if g.config.Model.Filename != "" {
		modelsFile = g.config.Model.Filename
	}
	modelsPath := filepath.Join(g.config.Output, modelsFile)

	dir := filepath.Dir(modelsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(modelsPath, code, 0644); err != nil {
		return fmt.Errorf("failed to write models file: %w", err)
	}

	fmt.Printf("Generated models: %s\n", modelsPath)
	return nil
}

func (g *Generator) generateServer() error {
	tmpl, err := template.ParseFS(templates, "templates/server.gotpl")
	if err != nil {
		return fmt.Errorf("failed to parse server template: %w", err)
	}

	data := g.buildServerTemplateData()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute server template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format server code: %w\n%s", err, buf.String())
	}

	serverFile := "server.go"
	if g.config.Exec.Filename != "" {
		serverFile = g.config.Exec.Filename
	}
	serverPath := filepath.Join(g.config.Output, serverFile)

	dir := filepath.Dir(serverPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(serverPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write server file: %w", err)
	}

	fmt.Printf("Generated server: %s\n", serverPath)
	return nil
}

// generateResolverStruct creates the main resolver.go file with the Resolver struct
// This file is only generated once and users can edit it freely
func (g *Generator) generateResolverStruct() error {
	resolverFile := filepath.Join(g.config.Output, "resolver.go")

	// Only generate if file doesn't exist
	if _, err := os.Stat(resolverFile); err == nil {
		fmt.Printf("Resolver struct already exists, skipping: %s\n", resolverFile)
		return nil
	}

	tmpl, err := template.ParseFS(templates, "templates/resolver_struct.gotpl")
	if err != nil {
		return fmt.Errorf("failed to parse resolver_struct template: %w", err)
	}

	data := map[string]interface{}{
		"Package":      g.config.Resolver.Package,
		"ResolverType": g.config.Resolver.Type,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute resolver_struct template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format resolver struct code: %w\n%s", err, buf.String())
	}

	if err := os.WriteFile(resolverFile, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write resolver struct file: %w", err)
	}

	fmt.Printf("Generated resolver struct: %s\n", resolverFile)
	return nil
}

// generateResolverImplementations creates/updates schema.resolvers.go with tool/prompt/resource implementations
func (g *Generator) generateResolverImplementations() error {
	resolverFile := filepath.Join(g.config.Output, "schema.resolvers.go")

	fileExists := false
	if _, err := os.Stat(resolverFile); err == nil {
		fileExists = true
	}

	// If file doesn't exist, generate from template (initial generation)
	if !fileExists {
		return g.generateResolverFromTemplate(resolverFile)
	}

	// File exists and preserve is enabled - do incremental update (gqlgen-style)
	if g.config.Resolver.Preserve {
		return g.updateResolverIncremental(resolverFile)
	}

	// File exists but preserve is disabled - regenerate completely
	return g.generateResolverFromTemplate(resolverFile)
}

func (g *Generator) generateResolverFromTemplate(resolverFile string) error {
	tmpl, err := template.ParseFS(templates, "templates/resolver.gotpl")
	if err != nil {
		return fmt.Errorf("failed to parse resolver template: %w", err)
	}

	data := g.buildResolverTemplateData()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute resolver template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format resolver code: %w\n%s", err, buf.String())
	}

	if err := os.WriteFile(resolverFile, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write resolver file: %w", err)
	}

	fmt.Printf("Generated resolver implementations: %s\n", resolverFile)
	return nil
}

func (g *Generator) updateResolverIncremental(resolverFile string) error {
	parser, err := NewResolverParser(resolverFile)
	if err != nil {
		return fmt.Errorf("failed to parse existing resolver: %w", err)
	}

	existingHandlers, err := parser.ExtractHandlers(g.config.Resolver.Type)
	if err != nil {
		return fmt.Errorf("failed to extract handlers: %w", err)
	}

	requiredHandlers := g.getRequiredHandlerNames()

	// Identify which handlers are new
	newHandlers := []string{}
	for _, required := range requiredHandlers {
		if _, exists := existingHandlers[required]; !exists {
			newHandlers = append(newHandlers, required)
		}
	}

	// Identify orphaned handlers (exist in file but not in spec, excluding already orphaned ones)
	// First, get the list of handlers that were already in the orphaned section
	previouslyOrphanedHandlers := extractOrphanedHandlerNames(resolverFile)

	// Identify which handlers are orphaned (not in required list and not already in orphaned section)
	currentlyOrphanedHandlers := []string{}
	for handlerName := range existingHandlers {
		isRequired := false
		for _, required := range requiredHandlers {
			if required == handlerName {
				isRequired = true
				break
			}
		}
		// Only mark as newly orphaned if not required and not already orphaned
		if !isRequired && !contains(previouslyOrphanedHandlers, handlerName) {
			currentlyOrphanedHandlers = append(currentlyOrphanedHandlers, handlerName)
		}
	}

	// Check if any previously orphaned handlers are now required (should be removed from orphaned)
	orphanedHandlersRemoved := []string{}
	for _, orphanedName := range previouslyOrphanedHandlers {
		if contains(requiredHandlers, orphanedName) {
			orphanedHandlersRemoved = append(orphanedHandlersRemoved, orphanedName)
		}
	}

	// Mark handlers as orphaned for formatting
	IdentifyOrphanedHandlers(existingHandlers, requiredHandlers)
	orphanedHandlers := FormatOrphanedHandlers(existingHandlers)

	// If nothing changed, skip update
	if len(newHandlers) == 0 && len(currentlyOrphanedHandlers) == 0 && len(orphanedHandlersRemoved) == 0 {
		fmt.Printf("Resolver is up to date, skipping: %s\n", resolverFile)
		return nil
	}

	// Generate code for new handlers only
	newHandlersCode, err := g.generateNewHandlersCode(newHandlers)
	if err != nil {
		return fmt.Errorf("failed to generate new handlers: %w", err)
	}

	// Read the existing file
	content, err := os.ReadFile(resolverFile)
	if err != nil {
		return fmt.Errorf("failed to read resolver file: %w", err)
	}

	// Remove any existing orphaned handlers section
	contentStr := string(content)
	if idx := strings.Index(contentStr, "\n// ==============================================================================\n// Orphaned Handlers\n"); idx != -1 {
		contentStr = contentStr[:idx]
	}

	// Build final content: existing code + new handlers + orphaned section
	var buf bytes.Buffer
	buf.WriteString(contentStr)

	if newHandlersCode != "" {
		buf.WriteString("\n")
		buf.WriteString(newHandlersCode)
	}

	if orphanedHandlers != "" {
		buf.WriteString(orphanedHandlers)
	}

	// Format the final code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format resolver code: %w\n%s", err, buf.String())
	}

	if err := os.WriteFile(resolverFile, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write resolver file: %w", err)
	}

	// Build status message
	var updates []string
	if len(newHandlers) > 0 {
		updates = append(updates, fmt.Sprintf("added %d new", len(newHandlers)))
	}
	if len(currentlyOrphanedHandlers) > 0 {
		updates = append(updates, fmt.Sprintf("orphaned %d", len(currentlyOrphanedHandlers)))
	}
	if len(orphanedHandlersRemoved) > 0 {
		updates = append(updates, fmt.Sprintf("restored %d from orphaned", len(orphanedHandlersRemoved)))
	}

	fmt.Printf("Updated resolver: %s: %s\n", strings.Join(updates, ", "), resolverFile)

	return nil
}

func countOrphanedHandlers(orphanedCode string) int {
	return strings.Count(orphanedCode, "// Orphaned:")
}

// extractOrphanedHandlerNames reads the orphaned section and returns list of handler names
func extractOrphanedHandlerNames(resolverFile string) []string {
	content, err := os.ReadFile(resolverFile)
	if err != nil {
		return nil
	}

	contentStr := string(content)
	orphanedSectionStart := strings.Index(contentStr, "\n// ==============================================================================\n// Orphaned Handlers\n")
	if orphanedSectionStart == -1 {
		return nil
	}

	orphanedSection := contentStr[orphanedSectionStart:]
	var names []string

	// Find all "// Orphaned: <HandlerName>" lines
	lines := strings.Split(orphanedSection, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "// Orphaned: ") {
			name := strings.TrimPrefix(line, "// Orphaned: ")
			names = append(names, name)
		}
	}

	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (g *Generator) generateNewHandlersCode(handlerNames []string) (string, error) {
	if len(handlerNames) == 0 {
		return "", nil
	}

	// Parse the resolver template to extract individual handler templates
	tmpl, err := template.ParseFS(templates, "templates/resolver.gotpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse resolver template: %w", err)
	}

	// Build template data with only the new handlers
	data := g.buildResolverTemplateData()

	// Filter to only include new handlers
	handlerSet := make(map[string]bool)
	for _, name := range handlerNames {
		handlerSet[name] = true
	}

	// Filter tools (HandlerName in data + "Tool" suffix should match required names)
	if tools, ok := data["Tools"].([]map[string]interface{}); ok {
		filteredTools := []map[string]interface{}{}
		for _, tool := range tools {
			if handlerName, ok := tool["HandlerName"].(string); ok {
				if handlerSet[handlerName+"Tool"] {
					filteredTools = append(filteredTools, tool)
				}
			}
		}
		data["Tools"] = filteredTools
	}

	// Filter resources (HandlerName in data + "Resource" suffix should match required names)
	if resources, ok := data["Resources"].([]map[string]interface{}); ok {
		filteredResources := []map[string]interface{}{}
		for _, resource := range resources {
			if handlerName, ok := resource["HandlerName"].(string); ok {
				if handlerSet[handlerName+"Resource"] {
					filteredResources = append(filteredResources, resource)
				}
			}
		}
		data["Resources"] = filteredResources
		data["HasResources"] = len(filteredResources) > 0
	}

	// Filter prompts (HandlerName in data + "Prompt" suffix should match required names)
	if prompts, ok := data["Prompts"].([]map[string]interface{}); ok {
		filteredPrompts := []map[string]interface{}{}
		for _, prompt := range prompts {
			if handlerName, ok := prompt["HandlerName"].(string); ok {
				if handlerSet[handlerName+"Prompt"] {
					filteredPrompts = append(filteredPrompts, prompt)
				}
			}
		}
		data["Prompts"] = filteredPrompts
		data["HasPrompts"] = len(filteredPrompts) > 0
	}

	// Generate only handler methods (not the full file structure)
	// NOTE: Must match the naming in resolver.gotpl template
	handlersOnlyTmpl, err := template.New("handlers").Parse(`
{{- range .Tools }}

{{- if .HasInputType }}
func (r *{{ $.ResolverType }}) {{ .HandlerName }}Tool(ctx context.Context, req *mcp.CallToolRequest, input *{{ .InputType }}) (*mcp.CallToolResult, {{ if .HasOutputType }}{{ .OutputType }}{{ else }}map[string]any{{ end }}, error) {
	return nil, {{ if .HasOutputType }}{{ .OutputType }}{}{{ else }}nil{{ end }}, fmt.Errorf("{{ .Name }} not implemented")
}
{{- else }}
func (r *{{ $.ResolverType }}) {{ .HandlerName }}Tool(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, {{ if .HasOutputType }}{{ .OutputType }}{{ else }}map[string]any{{ end }}, error) {
	return nil, {{ if .HasOutputType }}{{ .OutputType }}{}{{ else }}nil{{ end }}, fmt.Errorf("{{ .Name }} not implemented")
}
{{- end }}
{{- end }}

{{- range .Resources }}

func (r *{{ $.ResolverType }}) {{ .HandlerName }}Resource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return nil, fmt.Errorf("{{ .Name }} not implemented")
}
{{- end }}

{{- range .Prompts }}

{{- if .HasArgsType }}
func (r *{{ $.ResolverType }}) {{ .HandlerName }}Prompt(ctx context.Context, req *mcp.GetPromptRequest, args {{ .ArgsType }}) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("{{ .Name }} not implemented")
}
{{- else }}
func (r *{{ $.ResolverType }}) {{ .HandlerName }}Prompt(ctx context.Context, req *mcp.GetPromptRequest, args map[string]string) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("{{ .Name }} not implemented")
}
{{- end }}
{{- end }}
`)
	if err != nil {
		return "", fmt.Errorf("failed to create handlers template: %w", err)
	}

	var buf bytes.Buffer
	if err := handlersOnlyTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute handlers template: %w", err)
	}

	_ = tmpl // Keep using template for future enhancements
	return buf.String(), nil
}

func (g *Generator) getRequiredHandlerNames() []string {
	var names []string

	for _, tool := range g.spec.Tools {
		names = append(names, toHandlerName(tool.Name)+"Tool")
	}

	for _, resource := range g.spec.Resources {
		names = append(names, toHandlerName(resource.Name)+"Resource")
	}

	for _, prompt := range g.spec.Prompts {
		names = append(names, toHandlerName(prompt.Name)+"Prompt")
	}

	return names
}

func (g *Generator) buildServerTemplateData() map[string]interface{} {
	// Compute type prefix if model package is different from exec package
	modelPackage := g.config.Model.Package
	execPackage := g.config.Exec.Package
	typePrefix := ""
	modelImportPath := ""

	// Server needs to import models if they're in a different package
	if modelPackage != execPackage {
		parts := strings.Split(modelPackage, "/")
		typePrefix = parts[len(parts)-1] + "."

		// Compute the full import path for the model package
		modelImportPath = g.computeModelImportPath()
	}

	tools := make([]map[string]interface{}, 0, len(g.spec.Tools))
	hasTypedTools := false
	for _, tool := range g.spec.Tools {
		toolData := map[string]interface{}{
			"Name":        tool.Name,
			"Title":       tool.Title,
			"Description": tool.Description,
			"HandlerName": toHandlerName(tool.Name),
		}

		// Add hints if present
		if tool.Hints != nil {
			toolData["HasHints"] = true
			toolData["Readonly"] = tool.Hints.Readonly
			toolData["Destructive"] = tool.Hints.Destructive
			toolData["Idempotent"] = tool.Hints.Idempotent
			toolData["OpenWorld"] = tool.Hints.OpenWorld
		}

		// Add input type information and schema code
		if tool.InputSchema != nil {
			inputTypeName := typePrefix + toPascalCase(tool.Name) + "Input"
			toolData["InputType"] = inputTypeName
			toolData["HasInputType"] = true

			// Add schema variable name with proper prefix
			schemaVarName := typePrefix + toHandlerName(tool.Name) + "ToolInputSchema"
			toolData["InputSchemaVar"] = schemaVarName

			resolvedSchema := tool.InputSchema
			if config.IsSchemaRef(tool.InputSchema) && len(tool.InputSchema.Ref) > 0 && tool.InputSchema.Ref[0] == '#' {
				resolved, err := g.spec.ResolveSchemaRef(tool.InputSchema.Ref)
				if err == nil {
					resolvedSchema = resolved
				}
			}

			schemaCode := g.generateSchemaCode(resolvedSchema)
			toolData["InputSchemaCode"] = schemaCode

			hasTypedTools = true
		}

		// Add output type information and schema code
		if tool.OutputSchema != nil {
			outputTypeName := typePrefix + toPascalCase(tool.Name) + "Output"
			toolData["OutputType"] = outputTypeName
			toolData["HasOutputType"] = true

			// Add schema variable name with proper prefix
			schemaVarName := typePrefix + toHandlerName(tool.Name) + "ToolOutputSchema"
			toolData["OutputSchemaVar"] = schemaVarName

			resolvedSchema := tool.OutputSchema
			if config.IsSchemaRef(tool.OutputSchema) && len(tool.OutputSchema.Ref) > 0 && tool.OutputSchema.Ref[0] == '#' {
				resolved, err := g.spec.ResolveSchemaRef(tool.OutputSchema.Ref)
				if err == nil {
					resolvedSchema = resolved
				}
			}

			schemaCode := g.generateSchemaCode(resolvedSchema)
			toolData["OutputSchemaCode"] = schemaCode
		}

		tools = append(tools, toolData)
	}

	resources := make([]map[string]interface{}, 0, len(g.spec.Resources))
	for _, resource := range g.spec.Resources {
		resData := map[string]interface{}{
			"Name":        resource.Name,
			"Description": resource.Description,
			"HandlerName": toHandlerName(resource.Name),
			"MimeType":    resource.MimeType,
			"Readonly":    resource.Readonly,
		}

		if resource.URI != "" {
			resData["URI"] = resource.URI
		} else if resource.URITemplate != "" {
			resData["URITemplate"] = resource.URITemplate
			params := extractURIParams(resource.URITemplate)
			resData["URIParams"] = params
		}

		resources = append(resources, resData)
	}

	prompts := make([]map[string]interface{}, 0, len(g.spec.Prompts))
	for _, prompt := range g.spec.Prompts {
		args := make([]map[string]interface{}, 0, len(prompt.Arguments))
		for _, arg := range prompt.Arguments {
			args = append(args, map[string]interface{}{
				"Name":        arg.Name,
				"Description": arg.Description,
				"Required":    arg.Required,
			})
		}

		promptData := map[string]interface{}{
			"Name":        prompt.Name,
			"Description": prompt.Description,
			"HandlerName": toHandlerName(prompt.Name),
			"Arguments":   args,
		}

		// Add args type if there are arguments
		if len(prompt.Arguments) > 0 {
			argsTypeName := typePrefix + toPascalCase(prompt.Name) + "Args"
			promptData["ArgsType"] = argsTypeName
			promptData["HasArgsType"] = true
		}

		prompts = append(prompts, promptData)
	}

	data := map[string]interface{}{
		"Package":       g.config.Exec.Package,
		"ServerName":    g.spec.Info.Title,
		"ServerVersion": g.spec.Info.Version,
		"ResolverType":  g.config.Resolver.Type,
		"Tools":         tools,
		"Resources":     resources,
		"Prompts":       prompts,
		"HasResources":  len(resources) > 0,
		"HasPrompts":    len(prompts) > 0,
		"HasTypedTools": hasTypedTools,
	}

	// Add imports if packages are different from exec package
	var imports []map[string]string
	if modelPackage != execPackage && modelImportPath != "" {
		imports = append(imports, map[string]string{
			"Path":  modelImportPath,
			"Alias": "",
		})
	}
	// Never import resolver package - use interfaces to avoid circular imports
	if len(imports) > 0 {
		data["Imports"] = imports
	}

	return data
}

func (g *Generator) buildResolverTemplateData() map[string]interface{} {
	// Resolver template data is similar to server template data, but uses resolver package
	modelPackage := g.config.Model.Package
	resolverPackage := g.config.Resolver.Package
	typePrefix := ""
	modelImportPath := ""

	// Resolver needs to import models if they're in a different package
	if modelPackage != resolverPackage {
		parts := strings.Split(modelPackage, "/")
		typePrefix = parts[len(parts)-1] + "."

		// Compute the full import path for the model package
		modelImportPath = g.computeModelImportPath()
	}

	tools := make([]map[string]interface{}, 0, len(g.spec.Tools))
	hasTypedTools := false
	for _, tool := range g.spec.Tools {
		toolData := map[string]interface{}{
			"Name":        tool.Name,
			"Title":       tool.Title,
			"Description": tool.Description,
			"HandlerName": toHandlerName(tool.Name),
		}

		// Add hints if present
		if tool.Hints != nil {
			toolData["HasHints"] = true
			toolData["Readonly"] = tool.Hints.Readonly
			toolData["Destructive"] = tool.Hints.Destructive
			toolData["Idempotent"] = tool.Hints.Idempotent
			toolData["OpenWorld"] = tool.Hints.OpenWorld
		}

		// Add input type information
		if tool.InputSchema != nil {
			inputTypeName := typePrefix + toPascalCase(tool.Name) + "Input"
			toolData["InputType"] = inputTypeName
			toolData["HasInputType"] = true
			hasTypedTools = true
		}

		// Add output type information
		if tool.OutputSchema != nil {
			outputTypeName := typePrefix + toPascalCase(tool.Name) + "Output"
			toolData["OutputType"] = outputTypeName
			toolData["HasOutputType"] = true
		}

		tools = append(tools, toolData)
	}

	resources := make([]map[string]interface{}, 0, len(g.spec.Resources))
	for _, resource := range g.spec.Resources {
		resData := map[string]interface{}{
			"Name":        resource.Name,
			"Description": resource.Description,
			"HandlerName": toHandlerName(resource.Name),
			"MimeType":    resource.MimeType,
			"Readonly":    resource.Readonly,
		}

		if resource.URI != "" {
			resData["URI"] = resource.URI
		} else if resource.URITemplate != "" {
			resData["URITemplate"] = resource.URITemplate
			params := extractURIParams(resource.URITemplate)
			resData["URIParams"] = params
		}

		resources = append(resources, resData)
	}

	prompts := make([]map[string]interface{}, 0, len(g.spec.Prompts))
	for _, prompt := range g.spec.Prompts {
		args := make([]map[string]interface{}, 0, len(prompt.Arguments))
		for _, arg := range prompt.Arguments {
			args = append(args, map[string]interface{}{
				"Name":        arg.Name,
				"Description": arg.Description,
				"Required":    arg.Required,
			})
		}

		promptData := map[string]interface{}{
			"Name":        prompt.Name,
			"Description": prompt.Description,
			"HandlerName": toHandlerName(prompt.Name),
			"Arguments":   args,
		}

		// Add args type if there are arguments
		if len(prompt.Arguments) > 0 {
			argsTypeName := typePrefix + toPascalCase(prompt.Name) + "Args"
			promptData["ArgsType"] = argsTypeName
			promptData["HasArgsType"] = true
		}

		prompts = append(prompts, promptData)
	}

	data := map[string]interface{}{
		"Package":       g.config.Resolver.Package,
		"ServerName":    g.spec.Info.Title,
		"ServerVersion": g.spec.Info.Version,
		"ResolverType":  g.config.Resolver.Type,
		"Tools":         tools,
		"Resources":     resources,
		"Prompts":       prompts,
		"HasResources":  len(resources) > 0,
		"HasPrompts":    len(prompts) > 0,
		"HasTypedTools": hasTypedTools,
	}

	// Add model package import if different from resolver package
	if modelPackage != resolverPackage && modelImportPath != "" {
		var imports []map[string]string
		imports = append(imports, map[string]string{
			"Path":  modelImportPath,
			"Alias": "",
		})
		data["Imports"] = imports
	}

	return data
}

// computeModelImportPath computes the full import path for the model package
func (g *Generator) computeModelImportPath() string {
	return g.computeImportPath(g.config.Model.Package, g.config.Model.Filename)
}

// computeResolverImportPath computes the full import path for the resolver package
func (g *Generator) computeResolverImportPath() string {
	return g.computeImportPath(g.config.Resolver.Package, g.config.Resolver.Filename)
}

// computeImportPath computes the full import path for a package
func (g *Generator) computeImportPath(pkgName, filename string) string {
	// If the package is already a full path (contains slashes), use it as-is
	if strings.Contains(pkgName, "/") {
		return pkgName
	}

	// Make output path absolute
	absOutput, err := filepath.Abs(g.config.Output)
	if err != nil {
		return pkgName
	}

	// Find the closest go.mod to the output directory
	modulePath, moduleRoot, err := findClosestGoMod(absOutput)
	if err != nil {
		// If we can't read go.mod, fall back to using the package name directly
		return pkgName
	}

	// Compute the relative path from module root to output directory
	relPath, err := filepath.Rel(moduleRoot, absOutput)
	if err != nil {
		// If we can't compute relative path, fall back to package name
		return pkgName
	}

	// Compute the import path based on module + relative path + filename dir
	// Example: demo + generated + types = demo/generated/types
	fileDir := filepath.Dir(filename)
	if fileDir == "." {
		// If filename has no directory component, the files are in the output root
		// Import path should be module + output relative path
		return filepath.ToSlash(filepath.Join(modulePath, relPath))
	}

	// Otherwise, use the directory from the filename
	return filepath.ToSlash(filepath.Join(modulePath, relPath, fileDir))
}

// findClosestGoMod finds the closest go.mod file by walking up from the given directory
// Returns the module path and the directory containing go.mod
func findClosestGoMod(startDir string) (modulePath string, moduleRoot string, err error) {
	// Make startDir absolute
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", "", err
	}

	// Walk up the directory tree looking for go.mod
	currentDir := absDir
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, parse it using the official modfile package
			data, err := os.ReadFile(goModPath)
			if err != nil {
				return "", "", err
			}

			parsed, err := modfile.Parse(goModPath, data, nil)
			if err != nil {
				return "", "", fmt.Errorf("failed to parse %s: %w", goModPath, err)
			}

			if parsed.Module == nil || parsed.Module.Mod.Path == "" {
				return "", "", fmt.Errorf("no module directive found in %s", goModPath)
			}

			return parsed.Module.Mod.Path, currentDir, nil
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached the root directory
			return "", "", fmt.Errorf("no go.mod found in any parent directory of %s", absDir)
		}
		currentDir = parent
	}
}

func toHandlerName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}

func (g *Generator) generateSchemaCode(s *config.Schema) string {
	schemaJSON, err := json.Marshal(s)
	if err != nil {
		return "nil"
	}

	return fmt.Sprintf("mustUnmarshalSchema(`%s`)", string(schemaJSON))
}

func extractURIParams(uriTemplate string) []map[string]interface{} {
	var params []map[string]interface{}
	start := -1
	for i, ch := range uriTemplate {
		if ch == '{' {
			start = i + 1
		} else if ch == '}' && start >= 0 {
			paramName := uriTemplate[start:i]
			params = append(params, map[string]interface{}{
				"Name":        paramName,
				"Description": fmt.Sprintf("Parameter from URI template: %s", paramName),
			})
			start = -1
		}
	}

	return params
}

func extractGoTypeAnnotation(s *config.Schema) string {
	if s == nil || s.Extra == nil {
		return ""
	}

	if goType, ok := s.Extra["go.probo.inc/mcpgen/type"]; ok {
		if goTypeStr, ok := goType.(string); ok {
			return goTypeStr
		}
	}

	return ""
}

func parseTypeMapping(modelStr string) *CustomTypeMapping {
	mapping := &CustomTypeMapping{
		GoType: modelStr,
	}

	if strings.Contains(modelStr, "/") {
		parts := strings.Split(modelStr, ".")
		if len(parts) >= 2 {
			typeName := parts[len(parts)-1]
			importPath := strings.TrimSuffix(modelStr, "."+typeName)
			mapping.GoType = typeName
			mapping.ImportPath = importPath

			if strings.Contains(importPath, "/") {
				pkgParts := strings.Split(importPath, "/")
				pkgAlias := pkgParts[len(pkgParts)-1]
				mapping.GoType = pkgAlias + "." + typeName
			}
		}
	} else if strings.Contains(modelStr, ".") {
		parts := strings.Split(modelStr, ".")
		if len(parts) == 2 {
			mapping.ImportPath = parts[0]
			mapping.GoType = modelStr
		}
	}

	return mapping
}
