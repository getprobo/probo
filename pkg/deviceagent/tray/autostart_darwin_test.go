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

//go:build darwin

package tray

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLoggedInUsernames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name:   "empty output",
			output: "",
			want:   []string{},
		},
		{
			name:   "single user",
			output: "alice\n",
			want:   []string{"alice"},
		},
		{
			name:   "multiple users",
			output: "alice bob\n",
			want:   []string{"alice", "bob"},
		},
		{
			name:   "duplicate users",
			output: "alice alice bob\n",
			want:   []string{"alice", "bob"},
		},
		{
			name:   "skips root and loginwindow",
			output: "root loginwindow alice\n",
			want:   []string{"alice"},
		},
		{
			name:   "whitespace only",
			output: "   \n",
			want:   []string{},
		},
		{
			name:   "extra whitespace between names",
			output: "  alice   bob  \n",
			want:   []string{"alice", "bob"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				got := parseLoggedInUsernames(tt.output)
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestLaunchAgentIsCurrent(t *testing.T) {
	t.Parallel()

	const (
		exePath = "/usr/local/bin/probo-agent"
		runDir  = "/var/run/probo-agent"
	)

	t.Run(
		"missing file",
		func(t *testing.T) {
			t.Parallel()

			current, err := launchAgentIsCurrent(
				filepath.Join(t.TempDir(), "missing.plist"),
				exePath,
				runDir,
			)
			require.NoError(t, err)
			assert.False(t, current)
		},
	)

	t.Run(
		"matching bytes",
		func(t *testing.T) {
			t.Parallel()

			desired, err := renderLaunchAgentPlist(exePath, runDir)
			require.NoError(t, err)

			plistPath := filepath.Join(t.TempDir(), "tray.plist")
			require.NoError(t, os.WriteFile(plistPath, desired, 0o644))

			current, err := launchAgentIsCurrent(plistPath, exePath, runDir)
			require.NoError(t, err)
			assert.True(t, current)
		},
	)

	t.Run(
		"mismatched bytes",
		func(t *testing.T) {
			t.Parallel()

			plistPath := filepath.Join(t.TempDir(), "tray.plist")
			require.NoError(
				t,
				os.WriteFile(plistPath, []byte("not the launch agent"), 0o644),
			)

			current, err := launchAgentIsCurrent(plistPath, exePath, runDir)
			require.NoError(t, err)
			assert.False(t, current)
		},
	)
}
