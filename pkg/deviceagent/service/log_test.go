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

package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewsyslogEntry_RotatesOneMegabyteSevenCopies(t *testing.T) {
	t.Parallel()

	fields := strings.Fields(darwinNewsyslogConf())
	require.Len(t, fields, 9)
	assert.Equal(t, serviceLogPath, fields[0])
	assert.Equal(t, "root:wheel", fields[1])
	assert.Equal(t, "644", fields[2])
	assert.Equal(t, "7", fields[3])
	assert.Equal(t, "1024", fields[4])
	assert.Equal(t, "*", fields[5])
	assert.Equal(t, "ZC", fields[6])
	assert.Equal(t, darwinPIDPath, fields[7])
	assert.Equal(t, "1", fields[8])

	freebsd := strings.Fields(freebsdNewsyslogConf())
	require.Len(t, freebsd, 9)
	assert.Equal(t, freebsdSyslogPIDPath, freebsd[7])
}

func TestFreeBSDSyslogConf_SelectsAgentProgram(t *testing.T) {
	t.Parallel()

	text := freebsdSyslogConf()
	assert.Contains(t, text, "!probo-agent\n")
	assert.Contains(t, text, serviceLogPath)
	assert.True(t, strings.HasSuffix(text, "!*\n"))
}

func TestLaunchdPlist_DoesNotCaptureStandardStreams(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("launchd.plist.tmpl")
	require.NoError(t, err)

	text := string(data)
	assert.NotContains(t, text, "StandardOutPath")
	assert.NotContains(t, text, "StandardErrorPath")
	assert.NotContains(t, text, serviceLogPath)
}

func TestRemovePrefixedLogs_RemovesRotatedCopiesOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	keep := []string{
		"probo-agent-install.log",
		"other.log",
	}
	remove := []string{
		serviceLogName,
		serviceLogName + ".0.gz",
		serviceLogName + ".6.bz2",
	}

	for _, name := range append(keep, remove...) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644))
	}

	require.NoError(t, removePrefixedLogs(dir, serviceLogName))

	for _, name := range remove {
		_, err := os.Stat(filepath.Join(dir, name))
		assert.ErrorIs(t, err, os.ErrNotExist)
	}

	for _, name := range keep {
		_, err := os.Stat(filepath.Join(dir, name))
		assert.NoError(t, err)
	}
}

func TestReopen_AfterCloseDoesNotRecreate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, serviceLogName)
	file, err := openLogFile(path)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	require.NoError(t, os.Remove(path))

	err = file.reopen()
	assert.ErrorIs(t, err, os.ErrClosed)

	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestWriteRootFile_UsesMode644(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "probo-agent.conf")
	require.NoError(t, writeRootFile(path, "hello\n"))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())

	require.NoError(t, removeFile(path))
	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
	require.NoError(t, removeFile(path))
}

func TestLogEventLevel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "error", logEventLevel(`{"level":"ERROR","msg":"heartbeat failed"}`))
	assert.Equal(t, "warning", logEventLevel(`{"level":"WARN","msg":"posture flush will retry"}`))
	assert.Equal(t, "info", logEventLevel(`{"level":"INFO","msg":"agent starting"}`))
}
