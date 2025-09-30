// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// Package version provides build version information and user agent generation
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// BuildInfo contains version and build information
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
}

// GetBuildInfo returns the current build information
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version:   "dev",
		Commit:    "unknown",
		BuildDate: "unknown",
		GoVersion: runtime.Version(),
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	// Get module version
	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		info.Version = bi.Main.Version
	}

	// Extract VCS information from build settings
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			info.Commit = s.Value
			// Truncate commit hash to 7 characters for brevity
			if len(info.Commit) > 7 {
				info.Commit = info.Commit[:7]
			}
		case "vcs.time":
			info.BuildDate = s.Value
		case "vcs.modified":
			// If working tree is modified, append -dirty to commit
			if s.Value == "true" && info.Commit != "unknown" {
				info.Commit += "-dirty"
			}
		}
	}

	return info
}

// UserAgent returns a formatted user agent string for the given component
func UserAgent(component string) string {
	info := GetBuildInfo()

	if info.Version == "dev" && info.Commit == "unknown" {
		// Simple format for development builds without VCS info
		return fmt.Sprintf("Probo/dev (%s) Go/%s", component, info.GoVersion)
	}

	// Full format with all available information
	parts := []string{
		fmt.Sprintf("Probo/%s", info.Version),
	}

	// Add component and metadata
	metadata := []string{component}
	if info.Commit != "unknown" {
		metadata = append(metadata, fmt.Sprintf("commit=%s", info.Commit))
	}
	if info.BuildDate != "unknown" {
		metadata = append(metadata, fmt.Sprintf("built=%s", info.BuildDate))
	}

	parts = append(parts, fmt.Sprintf("(%s)", strings.Join(metadata, "; ")))
	parts = append(parts, fmt.Sprintf("Go/%s", info.GoVersion))

	return strings.Join(parts, " ")
}

// Version returns the version string
func Version() string {
	info := GetBuildInfo()
	return info.Version
}

// FullVersion returns a detailed version string with all build information
func FullVersion() string {
	info := GetBuildInfo()

	parts := []string{
		fmt.Sprintf("Version: %s", info.Version),
	}

	if info.Commit != "unknown" {
		parts = append(parts, fmt.Sprintf("Commit: %s", info.Commit))
	}

	if info.BuildDate != "unknown" {
		parts = append(parts, fmt.Sprintf("Built: %s", info.BuildDate))
	}

	parts = append(parts, fmt.Sprintf("Go: %s", info.GoVersion))

	return strings.Join(parts, "\n")
}