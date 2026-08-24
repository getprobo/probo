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

package types

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

func TestParseResource(t *testing.T) {
	t.Parallel()

	t.Run(
		"single resource",
		func(t *testing.T) {
			t.Parallel()

			resource, err := parseResource(
				url.Values{"resource": {"https://auth.example.com/api/mcp/v1"}},
			)
			require.NoError(t, err)
			assert.Equal(t, "https://auth.example.com/api/mcp/v1", resource)
		},
	)

	t.Run(
		"multiple resources rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := parseResource(
				url.Values{
					"resource": {
						"https://auth.example.com/api/mcp/v1",
						"https://other.example.com/api/mcp/v1",
					},
				},
			)
			require.ErrorIs(t, err, oauth2.ErrInvalidTarget)
		},
	)

	t.Run(
		"empty resource rejected",
		func(t *testing.T) {
			t.Parallel()

			_, err := parseResource(url.Values{"resource": {""}})
			require.ErrorIs(t, err, oauth2.ErrInvalidTarget)
		},
	)
}
