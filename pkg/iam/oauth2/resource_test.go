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
		want    *uri.URI
		wantErr bool
	}{
		{
			name: "resource omitted",
		},
		{
			name:   "issuer resource",
			values: []string{"https://auth.example.com"},
			want:   new(uri.URI("https://auth.example.com")),
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
			name: "multiple resources",
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

				got, err := service.protectedResource(tt.values)
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

func TestResourceMatches(t *testing.T) {
	t.Parallel()

	resource := uri.URI("https://auth.example.com")

	assert.True(t, resourceMatches(nil, ""))
	assert.True(t, resourceMatches(&resource, resource.String()))
	assert.False(t, resourceMatches(nil, resource.String()))
	assert.False(t, resourceMatches(&resource, ""))
	assert.False(t, resourceMatches(&resource, "https://other.example.com"))
}
