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

//go:build windows

package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatWindowsInstalled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		code   int
		stdout string
		stderr string
		want   string
	}{
		{
			name: "installed",
			code: 0,
			want: "installed: true",
		},
		{
			name:   "service does not exist",
			code:   windowsServiceNotFound,
			stderr: "The specified service does not exist as an installed service.\n",
			want:   "installed: false\nThe specified service does not exist as an installed service.",
		},
		{
			name:   "missing sc.exe",
			code:   -1,
			stderr: "exec: \"sc.exe\": executable file not found in %PATH%",
			want:   "installed: unknown\nexec: \"sc.exe\": executable file not found in %PATH%",
		},
		{
			name:   "access denied",
			code:   5,
			stderr: "Access is denied.",
			want:   "installed: unknown\nAccess is denied.",
		},
		{
			name:   "stdout when stderr is empty",
			code:   1,
			stdout: "OpenService FAILED 5",
			want:   "installed: unknown\nOpenService FAILED 5",
		},
		{
			name: "empty output uses the exit code",
			code: -1,
			want: "installed: unknown\nexit -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, formatWindowsInstalled(tt.code, tt.stdout, tt.stderr))
		})
	}
}

func TestWindowsEventLogArgs(t *testing.T) {
	t.Parallel()

	args := windowsEventLogArgs()
	require.Equal(t, []string{
		"qe",
		"Application",
		`/q:*[System[Provider[@Name='ProboAgent']]]`,
		fmt.Sprintf("/c:%d", supportLogTailLines),
		"/rd:true",
		"/f:text",
	}, args)
	assert.NotContains(t, args, "cl")
}
