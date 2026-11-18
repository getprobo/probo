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

func TestLexer_Next(t *testing.T) {
	t.Parallel()

	t.Run(
		"tag and type with quoted value",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer(`tag:env=Production type:"eu west"`)
			got := collectTokens(lexer)

			require.Len(t, got, 9)
			assert.Equal(t, searchquery.Ident, got[0].Kind)
			assert.Equal(t, "tag", got[0].Literal)
			assert.Equal(t, searchquery.Colon, got[1].Kind)
			assert.Equal(t, searchquery.Ident, got[2].Kind)
			assert.Equal(t, "env", got[2].Literal)
			assert.Equal(t, searchquery.Equal, got[3].Kind)
			assert.Equal(t, searchquery.Ident, got[4].Kind)
			assert.Equal(t, "Production", got[4].Literal)
			assert.Equal(t, searchquery.Ident, got[5].Kind)
			assert.Equal(t, "type", got[5].Literal)
			assert.Equal(t, searchquery.Colon, got[6].Kind)
			assert.Equal(t, searchquery.String, got[7].Kind)
			assert.Equal(t, "eu west", got[7].Literal)
			assert.Equal(t, searchquery.EOF, got[8].Kind)
		},
	)

	t.Run(
		"escaped quote and backslash",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer(`tag:note="say \"hi\" \\ok"`)
			got := collectTokens(lexer)

			require.Len(t, got, 6)
			assert.Equal(t, searchquery.String, got[4].Kind)
			assert.Equal(t, `say "hi" \ok`, got[4].Literal)
			assert.Equal(t, searchquery.EOF, got[5].Kind)
		},
	)

	t.Run(
		"unterminated quote",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer(`tag:env="open`)
			got := collectTokens(lexer)

			require.Len(t, got, 5)
			assert.Equal(t, searchquery.Illegal, got[4].Kind)
			assert.Equal(t, "unterminated quote", got[4].Literal)
		},
	)

	t.Run(
		"unknown escape",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer(`tag:"a\q"`)
			got := collectTokens(lexer)

			require.Len(t, got, 3)
			assert.Equal(t, searchquery.Illegal, got[2].Kind)
			assert.Equal(t, "invalid escape sequence", got[2].Literal)
			assert.Equal(t, 6, got[2].Pos)
		},
	)

	t.Run(
		"whitespace including newlines",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer("tag:env=prod\n\ttype:risk")
			got := collectTokens(lexer)

			require.Len(t, got, 9)
			assert.Equal(t, "tag", got[0].Literal)
			assert.Equal(t, 0, got[0].Pos)
			assert.Equal(t, "type", got[5].Literal)
			assert.Equal(t, searchquery.EOF, got[8].Kind)
		},
	)

	t.Run(
		"carriage return is whitespace",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer("tag:env=prod\r\ntype:risk")
			got := collectTokens(lexer)

			require.Len(t, got, 9)
			assert.Equal(t, "prod", got[4].Literal)
			assert.Equal(t, "type", got[5].Literal)
		},
	)

	t.Run(
		"trailing backslash",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer("tag:env=\"a\\")
			got := collectTokens(lexer)

			require.Len(t, got, 5)
			assert.Equal(t, searchquery.Illegal, got[4].Kind)
			assert.Equal(t, "unterminated quote", got[4].Literal)
			assert.Equal(t, 8, got[4].Pos)
		},
	)

	t.Run(
		"empty input",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer("")
			got := collectTokens(lexer)

			require.Len(t, got, 1)
			assert.Equal(t, searchquery.EOF, got[0].Kind)
			assert.Equal(t, 0, got[0].Pos)
		},
	)

	t.Run(
		"quote ends an ident",
		func(t *testing.T) {
			t.Parallel()

			lexer := searchquery.NewLexer(`ab"cd"`)
			got := collectTokens(lexer)

			require.Len(t, got, 3)
			assert.Equal(t, searchquery.Ident, got[0].Kind)
			assert.Equal(t, "ab", got[0].Literal)
			assert.Equal(t, searchquery.String, got[1].Kind)
			assert.Equal(t, "cd", got[1].Literal)
			assert.Equal(t, searchquery.EOF, got[2].Kind)
		},
	)
}

func TestKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind searchquery.Kind
		want string
	}{
		{kind: searchquery.EOF, want: "EOF"},
		{kind: searchquery.Ident, want: "Ident"},
		{kind: searchquery.String, want: "String"},
		{kind: searchquery.Colon, want: "Colon"},
		{kind: searchquery.Equal, want: "Equal"},
		{kind: searchquery.Illegal, want: "Illegal"},
		{kind: searchquery.Kind(99), want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(
			tt.want,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, tt.kind.String())
			},
		)
	}
}

func collectTokens(lexer *searchquery.Lexer) []searchquery.Token {
	tokens := make([]searchquery.Token, 0)

	for {
		tok := lexer.Next()
		tokens = append(tokens, tok)

		if tok.Kind == searchquery.EOF || tok.Kind == searchquery.Illegal {
			return tokens
		}
	}
}
