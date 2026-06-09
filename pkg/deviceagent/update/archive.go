// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

const (
	maxExtractedFileSize = 200 * 1024 * 1024 // 200 MiB hard cap per file
)

// extractBinary extracts the agent binary at
// `<ArchiveDir>/<BinaryName>` from archivePath into workDir and
// returns the absolute path of the written binary.
func extractBinary(archivePath string, layout AssetLayout, workDir string) (string, error) {
	wantPath := path.Join(layout.ArchiveDir, layout.BinaryName)
	dest := filepath.Join(workDir, "probo-agent.new")

	if layout.IsZip {
		if err := extractZipFile(archivePath, wantPath, dest); err != nil {
			return "", err
		}
	} else {
		if err := extractTarGzFile(archivePath, wantPath, dest); err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(dest); err != nil {
		return "", fmt.Errorf("update: extracted binary missing: %w", err)
	}

	return dest, nil
}

func extractTarGzFile(archivePath, wantPath, dest string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}

	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("cannot read gzip: %w", err)
	}

	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return fmt.Errorf("cannot read tar entry: %w", err)
		}

		if path.Clean(hdr.Name) != wantPath {
			continue
		}

		if hdr.Typeflag != tar.TypeReg {
			return fmt.Errorf("update: %s is not a regular file", wantPath)
		}

		return writeStream(dest, tr, 0o755)
	}

	return fmt.Errorf("update: %s missing from archive", wantPath)
}

func extractZipFile(archivePath, wantPath, dest string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open zip: %w", err)
	}

	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if path.Clean(f.Name) != wantPath {
			continue
		}

		if f.FileInfo().IsDir() {
			return fmt.Errorf("update: %s is a directory", wantPath)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("cannot open %s in zip: %w", wantPath, err)
		}

		err = writeStream(dest, rc, 0o755)
		_ = rc.Close()

		return err
	}

	return fmt.Errorf("update: %s missing from archive", wantPath)
}

func writeStream(dest string, src io.Reader, mode os.FileMode) error {
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", dest, err)
	}

	if _, err := io.Copy(out, io.LimitReader(src, maxExtractedFileSize+1)); err != nil {
		_ = out.Close()
		return fmt.Errorf("cannot write %s: %w", dest, err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("cannot close %s: %w", dest, err)
	}

	stat, err := os.Stat(dest)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", dest, err)
	}

	if stat.Size() > maxExtractedFileSize {
		_ = os.Remove(dest)
		return fmt.Errorf("update: extracted file exceeds %d bytes", maxExtractedFileSize)
	}

	return nil
}
