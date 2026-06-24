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

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSPDirectives(t *testing.T) {
	t.Parallel()

	t.Run(
		"parses multiple directives",
		func(t *testing.T) {
			t.Parallel()

			raw := "default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'unsafe-inline'"
			directives := parseCSPDirectives(raw)

			require.Len(t, directives, 3)
			assert.Equal(t, "default-src", directives[0].Name)
			assert.Equal(t, []string{"'self'"}, directives[0].Values)
			assert.Equal(t, "script-src", directives[1].Name)
			assert.Equal(t, []string{"'self'", "https://cdn.example.com"}, directives[1].Values)
			assert.Equal(t, "style-src", directives[2].Name)
			assert.Equal(t, []string{"'unsafe-inline'"}, directives[2].Values)
		},
	)

	t.Run(
		"handles empty string",
		func(t *testing.T) {
			t.Parallel()

			directives := parseCSPDirectives("")
			assert.Empty(t, directives)
		},
	)

	t.Run(
		"handles directive without values",
		func(t *testing.T) {
			t.Parallel()

			raw := "upgrade-insecure-requests"
			directives := parseCSPDirectives(raw)

			require.Len(t, directives, 1)
			assert.Equal(t, "upgrade-insecure-requests", directives[0].Name)
			assert.Empty(t, directives[0].Values)
		},
	)

	t.Run(
		"ignores trailing semicolons",
		func(t *testing.T) {
			t.Parallel()

			raw := "default-src 'self';"
			directives := parseCSPDirectives(raw)

			require.Len(t, directives, 1)
			assert.Equal(t, "default-src", directives[0].Name)
		},
	)
}
