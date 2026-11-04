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

package api

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

var (
	operationLead = regexp.MustCompile(`^(?:query|mutation|subscription)\b`)
	// Source templates may contain %s or ${...} and still need a name in
	// the document text. Runtime code uses the GraphQL parser instead.
	namedOperationLead = regexp.MustCompile(
		`(?m)^[ \t]*(?:query|mutation|subscription)\s+[_A-Za-z][_0-9A-Za-z]*\b`,
	)
)

func TestShippedGraphQLOperationsAreNamed(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)

	var failures []string

	for _, dir := range []string{
		filepath.Join(root, "pkg", "cmd"),
		filepath.Join(root, "pkg", "cli"),
		filepath.Join(root, "packages", "n8n-node"),
	} {
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			name := entry.Name()
			if entry.IsDir() {
				if name == "node_modules" || name == "dist" || name == "vendor" {
					return fs.SkipDir
				}

				return nil
			}

			if strings.HasSuffix(name, "_test.go") {
				return nil
			}

			ext := filepath.Ext(name)
			if ext != ".go" && ext != ".ts" && ext != ".tsx" {
				return nil
			}

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			for _, doc := range graphqlDocuments(string(src), ext != ".go") {
				if namedOperationLead.MatchString(doc) {
					continue
				}

				failures = append(failures, fmt.Sprintf("%s: %s", rel, firstLine(doc)))
			}

			return nil
		})
		require.NoError(t, err)
	}

	if len(failures) > 0 {
		t.Fatalf("GraphQL operations without a name:\n%s", strings.Join(failures, "\n"))
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir)

		dir = parent
	}
}

func graphqlDocuments(src string, allowSingleQuote bool) []string {
	var docs []string

	for i := 0; i < len(src); {
		if strings.HasPrefix(src[i:], "//") {
			nl := strings.IndexByte(src[i:], '\n')
			if nl < 0 {
				break
			}

			i += nl + 1

			continue
		}

		if strings.HasPrefix(src[i:], "/*") {
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				break
			}

			i += end + 4

			continue
		}

		switch src[i] {
		case '`':
			end := strings.IndexByte(src[i+1:], '`')
			if end < 0 {
				return docs
			}

			if doc := src[i+1 : i+1+end]; isGraphQLDocument(doc) {
				docs = append(docs, doc)
			}

			i += end + 2
		case '"':
			lit, next := readQuoted(src, i, '"')
			if next < 0 {
				i++

				continue
			}

			if isGraphQLDocument(lit) {
				docs = append(docs, lit)
			}

			i = next
		case '\'':
			if !allowSingleQuote {
				i = skipGoRune(src, i)

				continue
			}

			lit, next := readQuoted(src, i, '\'')
			if next < 0 {
				i++

				continue
			}

			if isGraphQLDocument(lit) {
				docs = append(docs, lit)
			}

			i = next
		default:
			i++
		}
	}

	return docs
}

func isGraphQLDocument(src string) bool {
	trimmed := strings.TrimSpace(src)
	if operationLead.MatchString(trimmed) && strings.Contains(trimmed, "{") {
		return true
	}

	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") || strings.Contains(trimmed, `"`) {
		return false
	}

	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			return true
		}
	}

	return false
}

func skipGoRune(src string, i int) int {
	j := i + 1
	if j >= len(src) {
		return len(src)
	}

	if src[j] == '\\' {
		j++
		if j >= len(src) {
			return len(src)
		}

		switch src[j] {
		case 'x':
			j += 2
		case 'u':
			j += 4
		case 'U':
			j += 8
		}

		j++
	} else {
		j++
	}

	if j < len(src) && src[j] == '\'' {
		return j + 1
	}

	return i + 1
}

func readQuoted(src string, i int, quote byte) (string, int) {
	var b strings.Builder

	for j := i + 1; j < len(src); j++ {
		if src[j] == '\\' && j+1 < len(src) {
			b.WriteByte(src[j+1])
			j++

			continue
		}

		if src[j] == quote {
			return b.String(), j + 1
		}

		if src[j] == '\n' {
			return "", -1
		}

		b.WriteByte(src[j])
	}

	return "", -1
}

func firstLine(doc string) string {
	line := strings.TrimSpace(doc)
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}

	line = strings.TrimSpace(line)
	if len(line) > 120 {
		line = line[:120]
	}

	return line
}
