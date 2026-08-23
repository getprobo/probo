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

package provider_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This package describes how to reach a provider and how to authenticate to it.
// What anyone does once connected belongs to the domain that wants it, and that
// domain registers its own capability factories (pkg/accessreview/drivers is
// the first). Importing a domain from here would put every future consumer's
// types on Registration and make this package grow with each one, so the
// direction is pinned by a test rather than by a comment.
//
// Adding an import listed below means the design drifted back: give the
// capability to the consuming domain's registry instead.
func TestPackageImportsNoConsumerDomain(t *testing.T) {
	t.Parallel()

	forbidden := []string{
		"go.probo.inc/probo/pkg/accessreview",
		"go.probo.inc/probo/pkg/iam",
		"go.probo.inc/probo/pkg/probo",
	}

	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	fset := token.NewFileSet()

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}

		// An external test package (provider_test) may import a consumer
		// domain: it sits above this package, not inside it.
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, parser.ImportsOnly)
		require.NoErrorf(t, err, "cannot parse %s", name)

		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			require.NoError(t, err)

			for _, prefix := range forbidden {
				assert.Falsef(
					t,
					path == prefix || strings.HasPrefix(path, prefix+"/"),
					"%s imports the consumer domain %q", name, path,
				)
			}
		}
	}
}
