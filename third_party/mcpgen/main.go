package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.probo.inc/mcpgen/internal/codegen"
	"go.probo.inc/mcpgen/internal/config"
)

var version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mcpgen",
	Short: "A code generator for Model Context Protocol (MCP) servers",
	Long: `mcpgen is a gqlgen-like code generator for building MCP servers in Go.
It generates type-safe Go code from JSON Schema definitions for tools, resources, and prompts.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of mcpgen",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcpgen %s\n", version)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go code from mcpgen configuration",
	Long: `Reads mcpgen.yaml (or mcpgen.yml) configuration file and generates:
  - Type-safe Go structs from JSON Schemas
  - MCP server boilerplate code
  - Handler function stubs for tools, resources, and prompts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		return runGenerate(configFile)
	},
}

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new MCP server project",
	Long:  `Creates a new MCP server project with example configuration and file structure.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "my-mcp-server"
		if len(args) > 0 {
			name = args[0]
		}
		return runInit(name)
	},
}

func init() {
	generateCmd.Flags().StringP("config", "c", "mcpgen.yaml", "Path to config file")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
}

func runGenerate(configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if configFile == "mcpgen.yaml" {
			if _, err := os.Stat("mcpgen.yml"); err == nil {
				configFile = "mcpgen.yml"
			}
		}
	}

	fmt.Printf("Loading configuration from %s...\n", configFile)

	cfg, spec, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Generating code for %s v%s...\n", spec.Info.Title, spec.Info.Version)

	gen := codegen.New(cfg, spec)

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	fmt.Println("✓ Code generation completed successfully!")
	return nil
}

func runInit(name string) error {
	fmt.Printf("Initializing new MCP server project: %s\n", name)

	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	configContent := `# mcpgen configuration
# Path to MCP API specification
spec: schema.yaml

# Output directory for generated code
output: generated

# Resolver configuration
resolver:
  package: generated
  filename: resolver.go
  type: Resolver
  preserve: true

# Model configuration
model:
  package: generated
  filename: models.go
`

	configPath := filepath.Join(name, "mcpgen.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	schemaContent := fmt.Sprintf(`# MCP API Specification
# This file contains the pure MCP API definition

info:
  title: %s
  version: 1.0.0
  description: An example MCP server

# Reusable schema components
components:
  schemas:
    ExampleInput:
      type: object
      properties:
        message:
          type: string
          description: The message to process
      required: [message]

# MCP Tools
tools:
  - name: example_tool
    title: Example Tool
    description: An example tool that processes messages
    hints:
      readonly: false
      destructive: false
      idempotent: true
    inputSchema:
      $ref: "#/components/schemas/ExampleInput"

# MCP Resources
resources: []

# MCP Prompts
prompts: []
`, name)

	schemaPath := filepath.Join(name, "schema.yaml")
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	fmt.Printf("\n✓ Project initialized successfully!\n\n")
	fmt.Printf("Files created:\n")
	fmt.Printf("  - mcpgen.yaml (code generation configuration)\n")
	fmt.Printf("  - schema.yaml (MCP API specification)\n\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  # Edit schema.yaml to define your tools, resources, and prompts\n")
	fmt.Printf("  mcpgen generate\n")

	return nil
}
