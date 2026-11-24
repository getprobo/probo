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

package deviceagent_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/deviceagent"
	"go.probo.inc/probo/pkg/deviceagent/checks"
)

func TestFormatStatus_OmitsAPIKey(t *testing.T) {
	t.Parallel()

	text := deviceagent.FormatStatus(
		"/var/lib/probo-agent",
		&deviceagent.Config{HeartbeatInterval: time.Minute},
		[]deviceagent.Enrollment{{
			ServerURL: "https://eu.probo.com",
			DeviceID:  "device-123",
			APIKey:    "super-secret-key",
		}},
	)

	assert.NotContains(t, text, "super-secret-key")
	assert.Contains(t, text, "https://eu.probo.com")
	assert.Contains(t, text, "device-123")
	assert.Contains(t, text, "API key on disk:      true")
}

func TestScrubLog_DropsCredentialLines(t *testing.T) {
	t.Parallel()

	text := strings.Join(
		[]string{
			"agent started",
			"Authorization: Bearer secret-token",
			"api_key=abc",
			"posture collected",
			"API key on disk: true",
		},
		"\n",
	)

	scrubbed := deviceagent.ScrubLog(text)

	assert.Contains(t, scrubbed, "agent started")
	assert.Contains(t, scrubbed, "posture collected")
	assert.Contains(t, scrubbed, "API key on disk: true")
	assert.NotContains(t, scrubbed, "Bearer")
	assert.NotContains(t, scrubbed, "secret-token")
	assert.NotContains(t, scrubbed, "api_key")
}

func TestFormatDiagnosticsReport_OmitsSecrets(t *testing.T) {
	t.Parallel()

	status := deviceagent.FormatStatus(
		"/var/lib/probo-agent",
		&deviceagent.Config{},
		[]deviceagent.Enrollment{{
			ServerURL: "https://us.probo.com",
			DeviceID:  "device-456",
			APIKey:    "another-secret-key",
		}},
	)
	report := deviceagent.FormatDiagnosticsReport(
		deviceagent.DiagnosticsReport{
			Version:  "1.2.3",
			GOOS:     "darwin",
			GOARCH:   "arm64",
			Hostname: "host",
			Time:     time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC),
			Status:   status,
			Collect: deviceagent.FormatCollect(
				[]checks.Result{{
					CheckKey: "disk_encryption",
					Status:   checks.StatusPass,
					Evidence: map[string]any{"on": true},
				}},
				nil,
			),
			LogPath: "/var/log/probo-agent.log",
			LogText: "heartbeat ok\nAuthorization: Bearer hidden\n",
		},
	)

	assert.NotContains(t, report, "another-secret-key")
	assert.NotContains(t, report, "Bearer")
	assert.Contains(t, report, "https://us.probo.com")
	assert.Contains(t, report, "device-456")
	assert.Contains(t, report, "API key on disk:      true")
	assert.Contains(t, report, "disk_encryption")
	assert.Contains(t, report, "heartbeat ok")
	assert.Contains(t, report, "2026-11-24T12:00:00Z")
}

func TestReadLogTail_ReturnsTheEnd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	current := filepath.Join(dir, "probo-agent.log")
	require.NoError(t, os.WriteFile(current, []byte("from-current\n"), 0o600))

	text, err := deviceagent.ReadLogTail(current, 64)
	require.NoError(t, err)
	assert.Equal(t, "from-current\n", text)

	_, err = deviceagent.ReadLogTail(filepath.Join(dir, "missing.log"), 64)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "log file is missing")
}

func TestReadLogTail_DropsAPartialFirstLine(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "probo-agent.log")
	require.NoError(t, os.WriteFile(path, []byte("aaaa\nbbbb\ncccc\n"), 0o600))

	t.Run("keeps a line that starts at the window", func(t *testing.T) {
		t.Parallel()

		text, err := deviceagent.ReadLogTail(path, 10)
		require.NoError(t, err)
		assert.Equal(t, "bbbb\ncccc\n", text)
	})

	t.Run("drops a line that the window splits", func(t *testing.T) {
		t.Parallel()

		text, err := deviceagent.ReadLogTail(path, 12)
		require.NoError(t, err)
		assert.Equal(t, "bbbb\ncccc\n", text)
	})

	t.Run("keeps a blank line that starts at the window", func(t *testing.T) {
		t.Parallel()

		blank := filepath.Join(dir, "blank.log")
		require.NoError(t, os.WriteFile(blank, []byte("aaaa\n\nbbbb\n"), 0o600))

		text, err := deviceagent.ReadLogTail(blank, 6)
		require.NoError(t, err)
		assert.Equal(t, "\nbbbb\n", text)
	})
}
