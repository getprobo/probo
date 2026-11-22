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

package searchquery_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/searchquery"
)

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run(
		"two tags and a quoted name",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse(`tag:Production tag:"eu west" type:document`)
			require.NoError(t, err)
			require.Len(t, query.Clauses, 3)

			tag0, ok := query.Clauses[0].(searchquery.TagClause)
			require.True(t, ok)
			assert.Equal(t, "Production", tag0.Name)

			tag1, ok := query.Clauses[1].(searchquery.TagClause)
			require.True(t, ok)
			assert.Equal(t, "eu west", tag1.Name)

			type0, ok := query.Clauses[2].(searchquery.TypeClause)
			require.True(t, ok)
			assert.Equal(t, "document", type0.Name)
		},
	)

	t.Run(
		"empty query",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse("   ")
			require.NoError(t, err)
			assert.Empty(t, query.Clauses)
		},
	)

	t.Run(
		"unknown qualifier",
		func(t *testing.T) {
			t.Parallel()

			_, err := searchquery.Parse("tag:Production status:open")
			require.Error(t, err)

			var parseErr *searchquery.Error
			require.ErrorAs(t, err, &parseErr)
			assert.Equal(t, `unknown qualifier "status"`, parseErr.Message)
		},
	)

	t.Run(
		"bare word",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse("access-control")
			require.NoError(t, err)
			require.Len(t, query.Clauses, 1)

			text, ok := query.Clauses[0].(searchquery.TextClause)
			require.True(t, ok)
			assert.Equal(t, "access-control", text.Value)
		},
	)

	t.Run(
		"text with tag and type",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse(`access tag:Production type:document`)
			require.NoError(t, err)
			require.Len(t, query.Clauses, 3)

			text, ok := query.Clauses[0].(searchquery.TextClause)
			require.True(t, ok)
			assert.Equal(t, "access", text.Value)
		},
	)

	t.Run(
		"unterminated quote",
		func(t *testing.T) {
			t.Parallel()

			_, err := searchquery.Parse(`tag:"open`)
			require.Error(t, err)

			var parseErr *searchquery.Error
			require.ErrorAs(t, err, &parseErr)
			assert.Equal(t, "unterminated quote", parseErr.Message)
		},
	)

	t.Run(
		"quoted name with colon",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse(`tag:"foo:bar"`)
			require.NoError(t, err)
			require.Len(t, query.Clauses, 1)

			tag, ok := query.Clauses[0].(searchquery.TagClause)
			require.True(t, ok)
			assert.Equal(t, "foo:bar", tag.Name)
		},
	)

	t.Run(
		"quoted name and type",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse(`tag:"eu west" type:risk`)
			require.NoError(t, err)
			require.Len(t, query.Clauses, 2)

			tag, ok := query.Clauses[0].(searchquery.TagClause)
			require.True(t, ok)
			assert.Equal(t, "eu west", tag.Name)

			typeClause, ok := query.Clauses[1].(searchquery.TypeClause)
			require.True(t, ok)
			assert.Equal(t, "risk", typeClause.Name)
		},
	)

	t.Run(
		"empty string",
		func(t *testing.T) {
			t.Parallel()

			query, err := searchquery.Parse("")
			require.NoError(t, err)
			assert.Empty(t, query.Clauses)
		},
	)
}

func TestParse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		message string
		pos     int
	}{
		{name: "unterminated quote at start", raw: `"open`, message: "unterminated quote", pos: 0},
		{name: "string then colon", raw: `"tag":Production`, message: "unexpected token", pos: 5},
		{name: "leading colon", raw: ":tag", message: "unexpected token", pos: 0},
		{name: "leading equals", raw: "=tag", message: "unexpected token", pos: 0},
		{name: "missing tag name", raw: "tag:", message: "tag name is required", pos: 4},
		{name: "equals instead of name", raw: "tag:=prod", message: "tag name is required", pos: 4},
		{name: "empty tag name", raw: `tag:""`, message: "tag name is required", pos: 6},
		{name: "key equals value syntax", raw: "tag:env=Production", message: "unexpected token", pos: 7},
		{name: "unterminated tag name", raw: `tag:"open`, message: "unterminated quote", pos: 4},
		{name: "trailing backslash in name", raw: "tag:\"a\\", message: "unterminated quote", pos: 4},
		{name: "unknown escape in name", raw: `tag:"a\q"`, message: "invalid escape sequence", pos: 6},
		{name: "unterminated quote after a clause", raw: `tag:Production "open`, message: "unterminated quote", pos: 15},
		{name: "missing type name", raw: "type:", message: "type name is required", pos: 5},
		{name: "empty type name", raw: `type:""`, message: "type name is required", pos: 7},
		{name: "unterminated type name", raw: `type:"doc`, message: "unterminated quote", pos: 5},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				_, err := searchquery.Parse(tt.raw)
				require.Error(t, err)

				var parseErr *searchquery.Error
				require.ErrorAs(t, err, &parseErr)
				assert.Equal(t, tt.message, parseErr.Message)
				assert.Equal(t, tt.pos, parseErr.Pos)
				assert.Equal(t, tt.message, err.Error())
			},
		)
	}
}
