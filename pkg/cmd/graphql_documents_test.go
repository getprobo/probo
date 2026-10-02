// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	gqlast "github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator/rules"
	"gopkg.in/yaml.v3"
)

const (
	consoleEndpoint = "/api/console/v1/graphql"
	connectEndpoint = "/api/connect/v1/graphql"

	consoleGqlgenConfig = "../server/api/console/v1/gqlgen.yaml"
	connectGqlgenConfig = "../server/api/connect/v1/gqlgen.yaml"
)

type graphQLDocument struct {
	file string
	name string
	body string
}

// TestCLIGraphQLDocuments_MatchServerSchema validates every GraphQL document
// embedded in a CLI command against the schema of the endpoint that command
// talks to, so a schema change that breaks a command fails the build instead of
// the user's terminal.
func TestCLIGraphQLDocuments_MatchServerSchema(t *testing.T) {
	t.Parallel()

	schemas := map[string]*gqlast.Schema{
		consoleEndpoint: loadSchema(t, consoleGqlgenConfig),
		connectEndpoint: loadSchema(t, connectGqlgenConfig),
	}

	files := collectCommandFiles(t)
	require.NotEmpty(t, files)

	documentCount := 0

	for _, file := range files {
		source, err := os.ReadFile(file)
		require.NoError(t, err)

		documents := extractGraphQLDocuments(t, file, source)
		if len(documents) == 0 {
			continue
		}

		endpoint, err := commandEndpoint(string(source))
		if !assert.NoErrorf(t, err, "%s embeds GraphQL documents", file) {
			continue
		}

		for _, document := range documents {
			documentCount++

			_, gqlErrs := gqlparser.LoadQueryWithRules(
				schemas[endpoint],
				document.body,
				rules.NewDefaultRules(),
			)
			for _, gqlErr := range gqlErrs {
				assert.Failf(
					t,
					"invalid GraphQL document",
					"%s: %s is not valid against %s: %s",
					document.file,
					document.name,
					endpoint,
					gqlErr.Error(),
				)
			}
		}
	}

	assert.NotZero(t, documentCount)

	t.Logf("validated %d GraphQL documents from %d files", documentCount, len(files))
}

// loadSchema builds the schema an API serves, from the globs its gqlgen
// configuration lists.
func loadSchema(t *testing.T, gqlgenConfig string) *gqlast.Schema {
	t.Helper()

	rawConfig, err := os.ReadFile(gqlgenConfig)
	require.NoError(t, err)

	var parsedConfig struct {
		Schema []string `yaml:"schema"`
	}

	require.NoError(t, yaml.Unmarshal(rawConfig, &parsedConfig))
	require.NotEmpty(t, parsedConfig.Schema)

	var sources []*gqlast.Source

	for _, glob := range parsedConfig.Schema {
		paths, err := filepath.Glob(filepath.Join(filepath.Dir(gqlgenConfig), glob))
		require.NoError(t, err)
		require.NotEmpty(t, paths, "no schema file matching %s in %s", glob, gqlgenConfig)

		for _, path := range paths {
			content, err := os.ReadFile(path)
			require.NoError(t, err)

			sources = append(sources, &gqlast.Source{Name: path, Input: string(content)})
		}
	}

	schema, gqlErr := gqlparser.LoadSchema(sources...)
	require.NoError(t, gqlErr)

	return schema
}

func collectCommandFiles(t *testing.T) []string {
	t.Helper()

	var files []string

	err := filepath.WalkDir(
		".",
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			files = append(files, path)

			return nil
		},
	)
	require.NoError(t, err)

	return files
}

func extractGraphQLDocuments(t *testing.T, file string, source []byte) []graphQLDocument {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), file, source, 0)
	require.NoError(t, err)

	var documents []graphQLDocument

	for _, decl := range parsed.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || (genDecl.Tok != token.CONST && genDecl.Tok != token.VAR) {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, value := range valueSpec.Values {
				lit, ok := value.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING || !strings.HasPrefix(lit.Value, "`") {
					continue
				}

				body, err := strconv.Unquote(lit.Value)
				require.NoError(t, err)

				if !isGraphQLDocument(body) {
					continue
				}

				documents = append(
					documents,
					graphQLDocument{
						file: file,
						name: valueSpec.Names[i].Name,
						body: body,
					},
				)
			}
		}
	}

	return documents
}

func isGraphQLDocument(body string) bool {
	trimmed := strings.TrimSpace(body)

	for _, keyword := range []string{"query", "mutation", "fragment"} {
		if strings.HasPrefix(trimmed, keyword+" ") || strings.HasPrefix(trimmed, keyword+"(") || strings.HasPrefix(trimmed, keyword+"{") {
			return true
		}
	}

	return false
}

func commandEndpoint(source string) (string, error) {
	usesConsole := strings.Contains(source, strconv.Quote(consoleEndpoint))
	usesConnect := strings.Contains(source, strconv.Quote(connectEndpoint))

	switch {
	case usesConsole && usesConnect:
		return "", fmt.Errorf("cannot pick a schema: file references both %s and %s", consoleEndpoint, connectEndpoint)
	case usesConsole:
		return consoleEndpoint, nil
	case usesConnect:
		return connectEndpoint, nil
	default:
		return "", fmt.Errorf("cannot pick a schema: file references neither %s nor %s", consoleEndpoint, connectEndpoint)
	}
}
