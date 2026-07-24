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

//go:build darwin || windows

package tray

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBrowserURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:  "accepts https",
			input: "https://us.probo.com/enroll",
		},
		{
			name:  "accepts http",
			input: "http://localhost:3000/enroll",
		},
		{
			name:  "accepts uppercase https",
			input: "HTTPS://eu.probo.com/enroll",
		},
		{
			name:    "rejects file scheme",
			input:   "file:///etc/passwd",
			wantErr: `unsupported URL scheme "file"`,
		},
		{
			name:    "rejects javascript scheme",
			input:   "javascript:alert(1)",
			wantErr: `unsupported URL scheme "javascript"`,
		},
		{
			name:    "rejects custom scheme",
			input:   "slack://open",
			wantErr: `unsupported URL scheme "slack"`,
		},
		{
			name:    "rejects empty input",
			input:   "",
			wantErr: `unsupported URL scheme ""`,
		},
		{
			name:    "rejects scheme-less input",
			input:   "us.probo.com/enroll",
			wantErr: `unsupported URL scheme ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateBrowserURL(tt.input)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestOpenBrowser_RejectsDisallowedScheme(t *testing.T) {
	t.Parallel()

	err := openBrowser("file:///tmp/evil")
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot open browser URL")
	assert.ErrorContains(t, err, `unsupported URL scheme "file"`)
}
