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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/deviceagent/checks"
)

const (
	// DiagnosticLogTailBytes is the maximum log text included in a report.
	DiagnosticLogTailBytes = 64 << 10
)

var secretLogLine = regexp.MustCompile(`(?i)api[_-]?key|authorization|\bbearer\b`)

// DiagnosticsReport is the text of one support report before formatting.
type DiagnosticsReport struct {
	Version  string
	GOOS     string
	GOARCH   string
	Hostname string
	Time     time.Time

	Status    string
	StatusErr error

	Collect    string
	CollectErr error

	LogPath string
	LogText string
	LogErr  error
}

// FormatStatus prints the local agent state. API keys are reported only
// as present or absent.
func FormatStatus(dir string, cfg *Config, items []Enrollment) string {
	var b strings.Builder

	if len(items) == 0 {
		b.WriteString("Enrollments:          0\n")
	}

	for i, item := range items {
		fmt.Fprintf(&b, "Enrollment %d:\n", i+1)
		fmt.Fprintf(&b, "  Server URL:           %s\n", item.ServerURL)
		fmt.Fprintf(&b, "  Device ID:            %s\n", item.DeviceID)
		fmt.Fprintf(&b, "  API key on disk:      %v\n", item.APIKey != "")
	}

	if cfg != nil {
		fmt.Fprintf(&b, "Heartbeat interval:   %s\n", cfg.HeartbeatInterval)
		fmt.Fprintf(&b, "Posture interval:     %s\n", cfg.PostureInterval)
		fmt.Fprintf(&b, "Update interval:      %s\n", cfg.UpdateInterval)
		fmt.Fprintf(&b, "Auto-update enabled:  %v\n", !cfg.UpdatesDisabled)
		fmt.Fprintf(&b, "Allow prereleases:    %v\n", cfg.AllowPrereleases)
	}

	fmt.Fprintf(&b, "Config directory:     %s\n", dir)

	return b.String()
}

// FormatCollect prints posture check results. A non-nil err is included
// after the rows that finished.
func FormatCollect(results []checks.Result, err error) string {
	var b strings.Builder

	for _, result := range results {
		fmt.Fprintf(&b, "%-20s %-15s %v\n", result.CheckKey, result.Status, result.Evidence)
	}

	if err != nil {
		fmt.Fprintf(&b, "collect error: %s\n", err)
	}

	if len(results) == 0 && err == nil {
		b.WriteString("(no checks)\n")
	}

	return b.String()
}

// ScrubLog drops lines that contain credential material.
func ScrubLog(text string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")

	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if secretLogLine.MatchString(line) {
			continue
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}

// ReadLogTail returns the end of path, up to limit bytes.
func ReadLogTail(path string, limit int) (string, error) {
	if path == "" {
		return "", errors.New("log file is missing")
	}

	if limit <= 0 {
		return "", errors.New("log tail size must be positive")
	}

	tail, err := readFileTail(path, limit)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("log file is missing: %s", path)
		}

		return "", fmt.Errorf("cannot read log: %w", err)
	}

	return string(tail), nil
}

func readFileTail(path string, limit int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat log: %w", err)
	}

	size := info.Size()

	offset := int64(0)
	if size > int64(limit) {
		offset = size - int64(limit)
	}

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("cannot seek log: %w", err)
	}

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read log: %w", err)
	}

	// A window that starts after a newline already begins on a full line.
	// Drop the first line only when the seek split it.
	if offset > 0 {
		var prev [1]byte

		if _, err := file.ReadAt(prev[:], offset-1); err != nil {
			return nil, fmt.Errorf("cannot read the log byte before the tail: %w", err)
		}

		if prev[0] != '\n' {
			if i := bytes.IndexByte(buf, '\n'); i >= 0 {
				buf = buf[i+1:]
			}
		}
	}

	return buf, nil
}

// FormatDiagnosticsReport renders a support report. Enrollment API keys
// are omitted. Log lines that look like credentials are omitted.
func FormatDiagnosticsReport(report DiagnosticsReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "version: %s\n", report.Version)
	fmt.Fprintf(&b, "os: %s\n", report.GOOS)
	fmt.Fprintf(&b, "arch: %s\n", report.GOARCH)
	fmt.Fprintf(&b, "hostname: %s\n", report.Hostname)
	fmt.Fprintf(&b, "time: %s\n", report.Time.Format(time.RFC3339))

	b.WriteString("\nstatus:\n")

	if report.StatusErr != nil {
		fmt.Fprintf(&b, "%s\n", report.StatusErr)
	} else {
		writeText(&b, report.Status)
	}

	b.WriteString("\ncollect:\n")
	writeText(&b, report.Collect)

	if report.CollectErr != nil && !strings.Contains(report.Collect, report.CollectErr.Error()) {
		fmt.Fprintf(&b, "collect error: %s\n", report.CollectErr)
	}

	b.WriteString("\nlog:\n")

	if report.LogPath != "" {
		fmt.Fprintf(&b, "path: %s\n", report.LogPath)
	}

	if report.LogErr != nil {
		fmt.Fprintf(&b, "%s\n", report.LogErr)
	}

	writeText(&b, ScrubLog(report.LogText))

	return b.String()
}

func writeText(b *strings.Builder, text string) {
	if text == "" {
		return
	}

	b.WriteString(text)

	if !strings.HasSuffix(text, "\n") {
		b.WriteByte('\n')
	}
}

// BuildDiagnostics collects status, posture checks, and the log tail.
// Section failures are written into the report.
func BuildDiagnostics(ctx context.Context, dir, version string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	report := DiagnosticsReport{
		Version:  version,
		GOOS:     runtime.GOOS,
		GOARCH:   runtime.GOARCH,
		Hostname: hostname,
		Time:     time.Now(),
	}

	cfg, cfgErr := LoadConfig(dir)

	items, enrollErr := LoadEnrollments(dir)
	switch {
	case cfgErr != nil:
		report.StatusErr = cfgErr
	case enrollErr != nil:
		report.StatusErr = enrollErr
	default:
		report.Status = FormatStatus(dir, cfg, items)
	}

	logger := log.NewLogger(
		log.WithName("probo-agent"),
		log.WithOutput(io.Discard),
	)
	agent := New(dir, version, logger)
	results, collectErr := agent.CollectOnce(ctx)
	report.Collect = FormatCollect(results, collectErr)
	report.CollectErr = collectErr

	text, logPath, logErr := readDiagnosticLog(DiagnosticLogTailBytes)
	report.LogPath = logPath
	report.LogText = text
	report.LogErr = logErr

	return FormatDiagnosticsReport(report)
}
