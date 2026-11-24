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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/deviceagent"
)

func TestDiagnosticsExecutable(t *testing.T) {
	t.Parallel()

	t.Run(
		"rejects a non-canonical path",
		func(t *testing.T) {
			t.Parallel()

			path, err := diagnosticsExecutable("/tmp/probo-agent")

			require.Error(t, err)
			assert.Empty(t, path)
			assert.Contains(t, err.Error(), deviceagent.DefaultExecutablePath())
		},
	)

	t.Run(
		"selects the canonical path",
		func(t *testing.T) {
			t.Parallel()

			path, err := diagnosticsExecutable("/Library/Probo/../Probo/probo-agent")

			require.NoError(t, err)
			assert.Equal(t, deviceagent.DefaultExecutablePath(), path)
		},
	)
}

func TestDiagnosticsShellCommand_QuotesOutput(t *testing.T) {
	t.Parallel()

	command := diagnosticsShellCommand(
		"/Library/Probo/probo-agent",
		"/tmp/a b/report.sock",
		4242,
	)

	assert.Contains(t, command, "'/Library/Probo/probo-agent'")
	assert.Contains(t, command, "'/tmp/a b/report.sock'")
	assert.Contains(t, command, "diagnostics --output")
	assert.Contains(t, command, "--output-pid 4242")
}

func TestAppleScriptString_EscapesQuoteAndBackslash(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `"a\"b\\c"`, appleScriptString(`a"b\c`))
}
