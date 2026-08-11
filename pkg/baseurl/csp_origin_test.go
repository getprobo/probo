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

package baseurl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
)

func TestCSPOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "https host",
			input: "https://app.example.com",
			want:  "https://app.example.com",
		},
		{
			name:  "http host with port",
			input: "http://localhost:8080",
			want:  "http://localhost:8080",
		},
		{
			name:  "ipv6 with port",
			input: "https://[::1]:8443",
			want:  "https://[::1]:8443",
		},
		{
			name:  "strips trailing slash path",
			input: "https://app.example.com/",
			want:  "https://app.example.com",
		},
		{
			name:  "strips path prefix",
			input: "https://app.example.com/api",
			want:  "https://app.example.com",
		},
		{
			name:  "strips nested path prefix",
			input: "https://app.example.com/probo/console",
			want:  "https://app.example.com",
		},
		{
			name:    "rejects semicolon host injection",
			input:   "https://evil.com;frame-ancestors",
			wantErr: true,
		},
		{
			name:    "rejects comma in host",
			input:   "https://evil.com,https://other.com",
			wantErr: true,
		},
		{
			name:    "rejects query",
			input:   "https://app.example.com?x=1",
			wantErr: true,
		},
		{
			name:    "rejects userinfo",
			input:   "https://user:pass@app.example.com", // trufflehog:ignore
			wantErr: true,
		},
		{
			name:    "rejects wildcard host",
			input:   "https://*.example.com",
			wantErr: true,
		},
		{
			name:    "rejects empty port",
			input:   "https://example.com:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := baseurl.CSPOrigin(tt.input)
				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestBaseURL_CSPOrigin_RejectsSemicolonHost(t *testing.T) {
	t.Parallel()

	// Parse still accepts this host; CSPOrigin must not.
	b, err := baseurl.Parse("https://evil.com;frame-ancestors")
	require.NoError(t, err)

	_, err = b.CSPOrigin()
	require.Error(t, err)
}

func TestBaseURL_CSPOrigin_DiscardsPath(t *testing.T) {
	t.Parallel()

	b, err := baseurl.Parse("https://app.example.com/probo")
	require.NoError(t, err)

	got, err := b.CSPOrigin()
	require.NoError(t, err)
	assert.Equal(t, "https://app.example.com", got)
}
