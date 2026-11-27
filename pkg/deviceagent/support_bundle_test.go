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

package deviceagent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportBundleFileName(t *testing.T) {
	t.Parallel()

	name := SupportBundleFileName(time.Date(2026, 11, 26, 15, 35, 4, 0, time.FixedZone("CET", 3600)))
	assert.Equal(t, "probo-support-20261126T143504Z.txt", name)
}

func TestWriteSupportBundleFile_PrefersDesktop(t *testing.T) {
	t.Parallel()

	home := t.TempDir()

	desktop := filepath.Join(home, "Desktop")
	require.NoError(t, os.Mkdir(desktop, 0o755))

	downloads := filepath.Join(home, "Downloads")
	require.NoError(t, os.Mkdir(downloads, 0o755))

	name := "probo-support-20261126T143500Z.txt"
	path, err := writeSupportBundleFile(home, name, []byte("bundle\n"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(desktop, name), path)

	_, err = os.Stat(filepath.Join(downloads, name))
	assert.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Stat(filepath.Join(home, name))
	assert.ErrorIs(t, err, os.ErrNotExist)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(supportBundleMode), info.Mode().Perm())
	}
}

func TestWriteSupportBundleFile_FallsBackToHome(t *testing.T) {
	t.Parallel()

	t.Run("desktop missing", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		name := "probo-support-20261126T143500Z.txt"
		path, err := writeSupportBundleFile(home, name, []byte("bundle\n"))
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, name), path)
	})

	t.Run("desktop is not a directory", func(t *testing.T) {
		t.Parallel()

		home := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(home, "Desktop"), []byte("x"), 0o644))

		name := "probo-support-20261126T143500Z.txt"
		path, err := writeSupportBundleFile(home, name, []byte("bundle\n"))
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, name), path)
	})
}

func TestFormatStatus_OmitsAPIKey(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "agent")
	require.NoError(t, SaveConfig(dir, &Config{}))
	require.NoError(t, SaveEnrollments(dir, []Enrollment{{
		ServerURL: "https://example.test",
		DeviceID:  "device-1",
		APIKey:    "super-secret-key",
	}}))

	text, err := FormatStatus(dir)
	require.NoError(t, err)
	assert.Contains(t, text, "https://example.test")
	assert.Contains(t, text, "device-1")
	assert.Contains(t, text, "API key on disk:      true")
	assert.Contains(t, text, "Config directory:     "+dir)
	assert.NotContains(t, text, "super-secret-key")
}

func TestFormatEnrollmentMarker(t *testing.T) {
	t.Parallel()

	runDir := filepath.Join(t.TempDir(), "run")
	require.NoError(t, os.MkdirAll(runDir, 0o755))
	assert.Equal(t, "absent", formatEnrollmentMarker(runDir))

	require.NoError(t, MarkEnrolled(runDir))
	assert.Equal(t, "present", formatEnrollmentMarker(runDir))
}

func TestRenderSupportBundle_SectionOrder(t *testing.T) {
	t.Parallel()

	runDir := filepath.Join(t.TempDir(), "run")
	require.NoError(t, os.MkdirAll(runDir, 0o755))

	body := renderSupportBundle(context.Background(), SupportBundleRequest{
		RunDir:  runDir,
		Version: "1.2.3",
		DescribeUnit: func(context.Context) string {
			return "exists: false"
		},
		ReadLog: func(context.Context) (string, error) {
			return "", errors.New("cannot read journal: permission denied")
		},
	})

	sections := []string{
		"Host",
		"Enrollment marker",
		"Service unit",
		"Log tail",
	}

	last := -1

	for _, section := range sections {
		at := strings.Index(body, section+"\n")
		require.Greater(t, at, last, section)
		last = at
	}

	assert.Contains(t, body, "Agent version: 1.2.3")
	assert.Contains(t, body, "The tray Connected row is the enrollment marker only.")
	assert.Contains(t, body, "absent")
	assert.Contains(t, body, "exists: false")
	assert.Contains(t, body, "cannot read journal: permission denied")
	assert.NotContains(t, body, "Downloads")
	assert.NotContains(t, body, "support-status.json")
}
