// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package update

import (
	"fmt"
	"strings"
)

// AssetLayout describes the names used by the release pipeline for a
// (goos, goarch) combination. The fields mirror what the
// release-probo-agent.yaml workflow produces.
type AssetLayout struct {
	// ArchiveName is the file name of the published archive
	// (e.g. probo-agent_Linux_x86_64.tar.gz).
	ArchiveName string
	// ArchiveDir is the top-level directory inside the archive
	// (e.g. probo-agent_Linux_x86_64).
	ArchiveDir string
	// BinaryName is the agent binary file name inside the archive
	// (e.g. probo-agent or probo-agent.exe).
	BinaryName string
	// IsZip is true for Windows builds, which ship as zip archives.
	// Other platforms ship as gzipped tar.
	IsZip bool
}

// LayoutFor returns the asset layout for a given (goos, goarch).
//
// The mapping is the inverse of the case statements in the release
// workflow: linux/Linux, darwin/Darwin, windows/Windows, freebsd/Freebsd
// and amd64 -> x86_64 (others kept as-is).
func LayoutFor(goos, goarch string) (AssetLayout, error) {
	osLabel, err := osLabel(goos)
	if err != nil {
		return AssetLayout{}, err
	}

	archLabel := archLabel(goarch)
	dir := fmt.Sprintf("probo-agent_%s_%s", osLabel, archLabel)

	binary := "probo-agent"
	isZip := false
	ext := "tar.gz"

	if goos == "windows" {
		binary += ".exe"
		isZip = true
		ext = "zip"
	}

	return AssetLayout{
		ArchiveName: fmt.Sprintf("%s.%s", dir, ext),
		ArchiveDir:  dir,
		BinaryName:  binary,
		IsZip:       isZip,
	}, nil
}

func osLabel(goos string) (string, error) {
	switch strings.ToLower(goos) {
	case "linux":
		return "Linux", nil
	case "darwin":
		return "Darwin", nil
	case "windows":
		return "Windows", nil
	case "freebsd":
		return "Freebsd", nil
	}

	return "", fmt.Errorf("unsupported GOOS %q for auto-update", goos)
}

func archLabel(goarch string) string {
	if goarch == "amd64" {
		return "x86_64"
	}

	return goarch
}
