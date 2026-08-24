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

package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/uri"
)

func TestServiceProtectedResource(t *testing.T) {
	t.Parallel()

	service := &Service{baseURL: "https://auth.example.com"}

	tests := []struct {
		name    string
		values  []string
		want    []uri.URI
		wantErr bool
	}{
		{
			name: "resource omitted defaults to issuer",
			want: []uri.URI{"https://auth.example.com"},
		},
		{
			name:   "issuer resource",
			values: []string{"https://auth.example.com"},
			want:   []uri.URI{"https://auth.example.com"},
		},
		{
			name: "duplicate resources deduplicated",
			values: []string{
				"https://auth.example.com",
				"https://auth.example.com",
			},
			want: []uri.URI{"https://auth.example.com"},
		},
		{
			name:    "unknown resource",
			values:  []string{"https://other.example.com/api/mcp/v1"},
			wantErr: true,
		},
		{
			name:    "empty resource",
			values:  []string{""},
			wantErr: true,
		},
		{
			name: "one unsupported resource",
			values: []string{
				"https://auth.example.com/api/mcp/v1",
				"https://auth.example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := service.protectedResources(tt.values)
				if tt.wantErr {
					require.ErrorIs(t, err, ErrInvalidTarget)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestResourcesSubset(t *testing.T) {
	t.Parallel()

	root := uri.URI("https://auth.example.com")
	files := uri.URI("https://files.example.com")

	assert.False(t, resourcesSubset(nil, nil))
	assert.True(t, resourcesSubset([]uri.URI{root}, []uri.URI{root}))
	assert.True(
		t,
		resourcesSubset(
			[]uri.URI{root, files},
			[]uri.URI{files, root},
		),
	)
	assert.True(t, resourcesSubset([]uri.URI{root}, []uri.URI{root, files}))
	assert.False(t, resourcesSubset([]uri.URI{root}, nil))
	assert.False(t, resourcesSubset([]uri.URI{root}, []uri.URI{files}))
	assert.False(t, resourcesSubset([]uri.URI{root, files}, []uri.URI{root}))
}
