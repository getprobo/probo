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

package console_v1

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
)

func TestBuildEnrollmentURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		baseURL         string
		token           string
		wantServerURL   string
		wantErrContains string
	}{
		{
			name:          "strips path from base URL",
			baseURL:       "https://us.probo.com/console",
			token:         "secret-token",
			wantServerURL: "https://us.probo.com",
		},
		{
			name:          "keeps non-default port",
			baseURL:       "http://127.0.0.1:8080/api",
			token:         "tok",
			wantServerURL: "http://127.0.0.1:8080",
		},
		{
			name:            "nil base URL",
			token:           "tok",
			wantErrContains: "base URL is required",
		},
		{
			name:            "empty token",
			baseURL:         "https://us.probo.com",
			wantErrContains: "enrollment token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var base *baseurl.BaseURL

			if tt.baseURL != "" {
				parsed, err := baseurl.Parse(tt.baseURL)
				require.NoError(t, err)

				base = parsed
			}

			got, err := buildEnrollmentURLs(base, tt.token)
			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrContains)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantServerURL, got.ServerURL)

			parsed, err := url.Parse(got.EnrollmentURL)
			require.NoError(t, err)
			assert.Equal(t, "probo", parsed.Scheme)
			assert.Equal(t, "enroll", parsed.Host)
			assert.Equal(t, tt.wantServerURL, parsed.Query().Get("server"))
			assert.Equal(t, tt.token, parsed.Query().Get("token"))
		})
	}
}
