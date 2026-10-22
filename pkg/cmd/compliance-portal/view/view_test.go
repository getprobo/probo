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

package view

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

func TestViewQuery_MatchesConsoleSchema(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob("../../../server/api/console/v1/graphql/*.graphql")
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	sources := make([]*ast.Source, 0, len(matches))
	for _, match := range matches {
		input, err := os.ReadFile(match)
		require.NoError(t, err)

		sources = append(
			sources,
			&ast.Source{
				Name:  filepath.Base(match),
				Input: string(input),
			},
		)
	}

	schema, err := gqlparser.LoadSchema(sources...)
	require.NoError(t, err)

	_, errs := gqlparser.LoadQueryWithRules(schema, viewQuery, rules.NewDefaultRules())
	require.Empty(t, errs)
	assert.NotNil(t, schema.Types["DNSRecordInstruction"])
	assert.NotNil(t, schema.Types["SlackbotInstallation"])
}
