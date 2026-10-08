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

package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDarwinDefaultsMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		stderr string
		want   bool
	}{
		{
			name:   "legacy missing domain",
			stderr: "The domain/default pair of (/Library/Managed Preferences/com.apple.SoftwareUpdate, AutomaticCheckEnabled) does not exist",
			want:   true,
		},
		{
			name:   "missing key",
			stderr: "Error: Could not find key 'AutomaticCheckEnabled' in domain 'com.apple.SoftwareUpdate'.",
			want:   true,
		},
		{
			name:   "macos 27 missing domain",
			stderr: "Error: Domain 'com.apple.SoftwareUpdate' not found.",
			want:   true,
		},
		{
			name:   "command unavailable",
			stderr: "defaults: command not found",
			want:   false,
		},
		{
			name:   "unrelated failure",
			stderr: "Error: exit status 1",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := darwinDefaultsMissing(CmdResult{Stderr: tt.stderr})
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestDarwinSoftwareUpdateResult_MissingManagedDomainUsesDefault(t *testing.T) {
	t.Parallel()

	prefs := make(map[string]darwinSoftwareUpdatePref, len(darwinSoftwareUpdateKeys))
	for _, key := range darwinSoftwareUpdateKeys {
		prefs[key] = darwinSoftwareUpdatePref{source: darwinPrefSourceDefault}
	}

	result := darwinSoftwareUpdateResult(prefs)

	require.Equal(t, StatusPass, result.Status)
	_, indeterminate := result.Evidence["indeterminate_keys"]
	assert.False(t, indeterminate)
}
