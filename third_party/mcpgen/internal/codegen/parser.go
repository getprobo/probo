package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

type HandlerInfo struct {
	Name       string
	RecvType   string
	SourceCode string
	IsOrphaned bool
}

type ResolverParser struct {
	filePath string
	fset     *token.FileSet
	file     *ast.File
}

func NewResolverParser(filePath string) (*ResolverParser, error) {
	fset := token.NewFileSet()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("resolver file not found: %s", filePath)
	}

	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resolver file: %w", err)
	}

	return &ResolverParser{
		filePath: filePath,
		fset:     fset,
		file:     file,
	}, nil
}

func (p *ResolverParser) ExtractHandlers(resolverType string) (map[string]*HandlerInfo, error) {
	handlers := make(map[string]*HandlerInfo)

	// Extract from both old wrapper types and new direct Resolver type
	allowedTypes := map[string]bool{
		// Old wrapper types (for backward compatibility during migration)
		"toolResolver":      true,
		"*toolResolver":     true,
		"promptResolver":    true,
		"*promptResolver":   true,
		"resourceResolver":  true,
		"*resourceResolver": true,
		// New direct Resolver type
		resolverType:  true,
		"*" + resolverType: true,
	}

	for _, decl := range p.file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Recv == nil {
			continue
		}

		recvType := p.getReceiverType(funcDecl.Recv)
		if !allowedTypes[recvType] {
			continue
		}

		methodName := funcDecl.Name.Name
		sourceCode, err := p.extractFunctionSource(funcDecl)
		if err != nil {
			return nil, fmt.Errorf("failed to extract source for %s: %w", methodName, err)
		}

		// Transform receiver type from old wrapper types to main Resolver type
		sourceCode = TransformReceiverType(sourceCode, resolverType)

		handlers[methodName] = &HandlerInfo{
			Name:       methodName,
			RecvType:   "*" + resolverType, // Always use main Resolver type
			SourceCode: sourceCode,
			IsOrphaned: false,
		}
	}

	return handlers, nil
}

func (p *ResolverParser) getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	switch typ := field.Type.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.StarExpr:
		if ident, ok := typ.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	}

	return ""
}

func (p *ResolverParser) extractFunctionSource(funcDecl *ast.FuncDecl) (string, error) {
	var buf strings.Builder

	cfg := printer.Config{
		Mode:     printer.UseSpaces | printer.TabIndent,
		Tabwidth: 8,
	}

	if err := cfg.Fprint(&buf, p.fset, funcDecl); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// TransformReceiverType rewrites the receiver type in handler source code from old wrapper types
// (toolResolver, promptResolver, resourceResolver) to the main Resolver type
func TransformReceiverType(sourceCode, resolverType string) string {
	// Replace old wrapper types with main Resolver type
	sourceCode = strings.ReplaceAll(sourceCode, "*toolResolver)", "*"+resolverType+")")
	sourceCode = strings.ReplaceAll(sourceCode, "*promptResolver)", "*"+resolverType+")")
	sourceCode = strings.ReplaceAll(sourceCode, "*resourceResolver)", "*"+resolverType+")")
	return sourceCode
}

func IdentifyOrphanedHandlers(existingHandlers map[string]*HandlerInfo, requiredHandlers []string) {
	requiredSet := make(map[string]bool)
	for _, name := range requiredHandlers {
		requiredSet[name] = true
	}

	for name, handler := range existingHandlers {
		if !requiredSet[name] {
			handler.IsOrphaned = true
		}
	}
}

func FormatOrphanedHandlers(handlers map[string]*HandlerInfo) string {
	var orphaned []*HandlerInfo

	for _, handler := range handlers {
		if handler.IsOrphaned {
			orphaned = append(orphaned, handler)
		}
	}

	if len(orphaned) == 0 {
		return ""
	}

	var buf strings.Builder
	buf.WriteString("\n\n// ==============================================================================\n")
	buf.WriteString("// Orphaned Handlers\n")
	buf.WriteString("// ==============================================================================\n")
	buf.WriteString("// The following handlers were found in the resolver file but are no longer\n")
	buf.WriteString("// defined in the MCP specification. They have been preserved here as comments\n")
	buf.WriteString("// in case you need to reference or restore them.\n")
	buf.WriteString("// ==============================================================================\n\n")

	for _, handler := range orphaned {
		buf.WriteString(fmt.Sprintf("// Orphaned: %s\n", handler.Name))
		buf.WriteString("// Uncomment and update signature if you want to restore this handler.\n")

		lines := strings.Split(handler.SourceCode, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				buf.WriteString("// ")
			}
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	return buf.String()
}
