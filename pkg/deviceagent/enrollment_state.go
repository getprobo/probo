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
	"fmt"
	"os"
	"path/filepath"
)

const (
	// EnrollmentRunDirMode is searchable by the user-session tray helper so
	// IsEnrolled can stat the marker without access to the secrets directory.
	EnrollmentRunDirMode = 0o755

	// EnrollmentMarkerMode is owner-only; tray detection uses os.Stat via
	// the searchable run directory, not read access to the marker contents.
	EnrollmentMarkerMode = 0o600

	// EnrollmentMarkerName is the runtime enrollment flag filename.
	EnrollmentMarkerName = "enrolled"
)

// EnrollmentRunDir returns the runtime directory for the public enrollment
// marker and the short-lived enrolling.lock. Production installs use
// DefaultEnrollmentRunDir(); custom --dir values get an isolated sibling
// run tree for dev and tests.
func EnrollmentRunDir(configDir string) string {
	if configDir == "" {
		configDir = DefaultConfigDir()
	}

	if configDir == DefaultConfigDir() {
		return DefaultEnrollmentRunDir()
	}

	return filepath.Join(filepath.Dir(configDir), "run", filepath.Base(configDir))
}

func enrollmentMarkerPath(runDir string) string {
	if runDir == "" {
		runDir = DefaultEnrollmentRunDir()
	}

	return filepath.Join(runDir, EnrollmentMarkerName)
}

// IsEnrolled reports whether the enrollment marker exists in runDir.
func IsEnrolled(runDir string) (bool, error) {
	_, err := os.Stat(enrollmentMarkerPath(runDir))
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("cannot stat enrollment marker: %w", err)
}

// MarkEnrolled writes the enrollment marker under runDir.
func MarkEnrolled(runDir string) error {
	if runDir == "" {
		runDir = DefaultEnrollmentRunDir()
	}

	if err := os.MkdirAll(runDir, EnrollmentRunDirMode); err != nil {
		return fmt.Errorf("cannot create enrollment run dir: %w", err)
	}

	if err := os.Chmod(runDir, EnrollmentRunDirMode); err != nil {
		return fmt.Errorf("cannot set enrollment run dir permissions: %w", err)
	}

	path := enrollmentMarkerPath(runDir)
	if err := os.WriteFile(path, []byte("ok\n"), EnrollmentMarkerMode); err != nil {
		return fmt.Errorf("cannot write enrollment marker: %w", err)
	}

	if err := os.Chmod(path, EnrollmentMarkerMode); err != nil {
		return fmt.Errorf("cannot set enrollment marker permissions: %w", err)
	}

	return nil
}

// ClearEnrollmentMarker removes the public enrollment marker from runDir.
func ClearEnrollmentMarker(runDir string) error {
	if err := os.Remove(enrollmentMarkerPath(runDir)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove enrollment marker: %w", err)
	}

	return nil
}
