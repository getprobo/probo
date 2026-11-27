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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// supportBundleMode keeps the bundle owner-readable. It holds the
	// host serial number and hardware UUID.
	supportBundleMode = 0o600
)

type SupportBundleRequest struct {
	Dir          string
	RunDir       string
	Version      string
	Now          time.Time
	DescribeUnit func(context.Context) string
	ReadLog      func(context.Context) (string, error)
}

// SupportBundleFileName is probo-support-<UTC timestamp>.txt.
func SupportBundleFileName(now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}

	return "probo-support-" + now.UTC().Format("20060102T150405Z") + ".txt"
}

// WriteSupportBundle writes the user-session support file and returns its path.
// It tries the signed-in user's Desktop, then their home directory.
func WriteSupportBundle(ctx context.Context, req SupportBundleRequest) (string, error) {
	if req.Dir == "" {
		req.Dir = DefaultConfigDir()
	}

	if req.RunDir == "" {
		req.RunDir = EnrollmentRunDir(req.Dir)
	}

	if req.Now.IsZero() {
		req.Now = time.Now()
	}

	body := renderSupportBundle(ctx, req)

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home directory: %w", err)
	}

	return writeSupportBundleFile(home, SupportBundleFileName(req.Now), []byte(body))
}

func renderSupportBundle(ctx context.Context, req SupportBundleRequest) string {
	var b strings.Builder

	writeBundleSection(&b, "Host", formatHostSection(req.Version, CollectHostInfo()))
	writeBundleSection(&b, "Enrollment marker", formatEnrollmentMarker(req.RunDir))
	writeBundleSection(&b, "Service unit", describeBundleUnit(ctx, req.DescribeUnit))
	writeBundleSection(&b, "Log tail", readLogTail(ctx, req.ReadLog))

	return b.String()
}

func writeBundleSection(b *strings.Builder, title, body string) {
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", len(title)))
	b.WriteString("\n")
	b.WriteString(strings.TrimRight(body, "\n"))
	b.WriteString("\n\n")
}

func formatHostSection(version string, host HostInfo) string {
	serial := "absent"
	if host.SerialNumber != nil && *host.SerialNumber != "" {
		serial = *host.SerialNumber
	}

	return fmt.Sprintf(
		"Agent version: %s\nOS: %s\nHostname: %s\nPlatform: %s\nOS version: %s\nHardware UUID: %s\nSerial number: %s\nThe tray Connected row is the enrollment marker only.",
		version,
		runtime.GOOS,
		host.Hostname,
		host.Platform,
		host.OSVersion,
		host.HardwareUUID,
		serial,
	)
}

func formatEnrollmentMarker(runDir string) string {
	enrolled, err := IsEnrolled(runDir)
	if err != nil {
		return err.Error()
	}

	if enrolled {
		return "present"
	}

	return "absent"
}

func describeBundleUnit(ctx context.Context, describe func(context.Context) string) string {
	if describe == nil {
		return "unavailable"
	}

	return describe(ctx)
}

func readLogTail(ctx context.Context, read func(context.Context) (string, error)) string {
	if read == nil {
		return "unavailable"
	}

	text, err := read(ctx)
	if err != nil {
		return err.Error()
	}

	return text
}

// FormatStatus prints the same local fields as `probo-agent status`.
func FormatStatus(dir string) (string, error) {
	cfg, err := LoadConfig(dir)
	if err != nil {
		return "", fmt.Errorf("cannot load config: %w", err)
	}

	items, err := LoadEnrollments(dir)
	if err != nil {
		return "", fmt.Errorf("cannot load enrollments: %w", err)
	}

	var b strings.Builder

	if len(items) == 0 {
		fmt.Fprintf(&b, "Enrollments:          0\n")
	}

	for i, item := range items {
		fmt.Fprintf(&b, "Enrollment %d:\n", i+1)
		fmt.Fprintf(&b, "  Server URL:           %s\n", item.ServerURL)
		fmt.Fprintf(&b, "  Device ID:            %s\n", item.DeviceID)
		fmt.Fprintf(&b, "  API key on disk:      %v\n", item.APIKey != "")
	}

	fmt.Fprintf(&b, "Heartbeat interval:   %s\n", cfg.HeartbeatInterval)
	fmt.Fprintf(&b, "Posture interval:     %s\n", cfg.PostureInterval)
	fmt.Fprintf(&b, "Update interval:      %s\n", cfg.UpdateInterval)
	fmt.Fprintf(&b, "Auto-update enabled:  %v\n", !cfg.UpdatesDisabled)
	fmt.Fprintf(&b, "Allow prereleases:    %v\n", cfg.AllowPrereleases)
	fmt.Fprintf(&b, "Config directory:     %s\n", dir)

	return b.String(), nil
}

func writeSupportBundleFile(home, name string, body []byte) (string, error) {
	var errs []error

	for _, dir := range supportBundleDirs(home) {
		path, err := tryWriteSupportFile(dir, name, body)
		if err == nil {
			return path, nil
		}

		errs = append(errs, err)
	}

	return "", fmt.Errorf("cannot write support bundle: %w", errors.Join(errs...))
}

func supportBundleDirs(home string) []string {
	return []string{
		filepath.Join(home, "Desktop"),
		home,
	}
}

func tryWriteSupportFile(dir, name string, body []byte) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("cannot use %s: %w", dir, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("cannot use %s: not a directory", dir)
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, supportBundleMode); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", path, err)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}

	return abs, nil
}
