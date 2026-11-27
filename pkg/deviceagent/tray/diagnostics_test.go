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

package tray

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/deviceagent"
)

func TestDiagnosticsArgs_UsesRunDir(t *testing.T) {
	t.Parallel()

	t.Run("default run directory", func(t *testing.T) {
		t.Parallel()

		args, err := diagnosticsArgs("")
		require.NoError(t, err)
		assert.Equal(t, []string{
			"diagnostics",
			"--dir",
			deviceagent.DefaultConfigDir(),
			"--run-dir",
			deviceagent.DefaultEnrollmentRunDir(),
		}, args)
	})

	t.Run("custom run directory", func(t *testing.T) {
		t.Parallel()

		runDir := filepath.Join(t.TempDir(), "run", "custom")
		args, err := diagnosticsArgs(runDir)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"diagnostics",
			"--dir",
			deviceagent.ConfigDirForRunDir(runDir),
			"--run-dir",
			runDir,
		}, args)
		assert.NotContains(t, args, "sudo")
		assert.NotContains(t, args, "Downloads")
	})

	t.Run("unrecognized run directory", func(t *testing.T) {
		t.Parallel()

		_, err := diagnosticsArgs(filepath.Join(t.TempDir(), "elsewhere"))
		require.Error(t, err)
	})
}

func TestSaveSupportBundle_RunsCommandAsCurrentUser(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("shell script double is unix-only")
	}

	dir := t.TempDir()
	record := filepath.Join(dir, "invoked")
	script := filepath.Join(dir, "probo-agent")
	quoted := strconv.Quote(record)
	body := fmt.Sprintf("#!/bin/sh\nid -u > %s\nprintf '%%s\\n' \"$@\" >> %s\necho /tmp/probo-support-test.txt\n", quoted, quoted)
	require.NoError(t, os.WriteFile(script, []byte(body), 0o755))

	path, err := saveSupportBundle(Options{
		ExePath: script,
		RunDir:  deviceagent.DefaultEnrollmentRunDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, "/tmp/probo-support-test.txt", path)

	recorded, err := os.ReadFile(record)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimRight(string(recorded), "\n"), "\n")
	require.GreaterOrEqual(t, len(lines), 1)
	assert.Equal(t, strconv.Itoa(os.Getuid()), lines[0])
	assert.Equal(t, []string{
		"diagnostics",
		"--dir",
		deviceagent.DefaultConfigDir(),
		"--run-dir",
		deviceagent.DefaultEnrollmentRunDir(),
	}, lines[1:])
}
