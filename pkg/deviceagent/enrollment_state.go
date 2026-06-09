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

package deviceagent

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// World-readable so the user-session tray helper can detect enrollment
	// without reading the API key.
	EnrollmentMarkerName = "enrolled"
)

func EnrollmentMarkerPath(dir string) string {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	return filepath.Join(dir, EnrollmentMarkerName)
}

func IsEnrolled(dir string) bool {
	_, err := os.Stat(EnrollmentMarkerPath(dir))
	return err == nil
}

func MarkEnrolled(dir string) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	path := EnrollmentMarkerPath(dir)
	if err := os.WriteFile(path, []byte("ok\n"), 0o644); err != nil {
		return fmt.Errorf("cannot write enrollment marker: %w", err)
	}

	return nil
}

func ClearEnrollmentMarker(dir string) error {
	if err := os.Remove(EnrollmentMarkerPath(dir)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove enrollment marker: %w", err)
	}

	return nil
}
