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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	SigstoreCacheDirName = "sigstore-cache"
	configFileTmpName    = ConfigFileName + ".tmp"
)

var stateCleanupNames = []string{
	ConfigFileName,
	configFileTmpName,
	KeyFileName,
	KeyFileName + ".tmp",
	KeyFileName + ".old",
	pendingPosturesFileName,
	pendingPosturesFileName + ".tmp",
}

// RemoveLocalState removes allowlisted state/run files and sigstore-cache.
// Missing paths are ignored; filesystem roots are refused. Does not remove
// enrolling.lock (flock is inode-based).
func RemoveLocalState(dir string) error {
	if dir == "" {
		return errors.New("state directory is empty")
	}

	stateDir, err := resolveCleanupDir(dir)
	if err != nil {
		return err
	}

	runDir, err := resolveCleanupDir(EnrollmentRunDir(dir))
	if err != nil {
		return err
	}

	return errors.Join(
		removeAllUnder(stateDir, SigstoreCacheDirName),
		cleanDir(stateDir, stateCleanupNames),
		cleanDir(runDir, []string{EnrollmentMarkerName}),
	)
}

// resolveCleanupDir resolves dir and rejects filesystem roots.
// Missing paths return ("", nil).
func resolveCleanupDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	cleaned := filepath.Clean(abs)
	if isFilesystemRoot(cleaned) {
		return "", fmt.Errorf("refusing to remove filesystem root %q", cleaned)
	}

	real, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}

		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	if isFilesystemRoot(real) {
		return "", fmt.Errorf("refusing to remove filesystem root %q", real)
	}

	return real, nil
}

func cleanDir(dir string, names []string) error {
	if dir == "" {
		return nil
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot open %s: %w", dir, err)
	}

	defer func() { _ = root.Close() }()

	var errs error

	for _, name := range names {
		if err := root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = errors.Join(errs, fmt.Errorf("cannot remove %s: %w", name, err))
		}
	}

	_ = os.Remove(dir)

	return errs
}

func removeAllUnder(dir, name string) error {
	if dir == "" {
		return nil
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot open %s: %w", dir, err)
	}

	defer func() { _ = root.Close() }()

	if err := root.RemoveAll(name); err != nil {
		return fmt.Errorf("cannot remove %s: %w", name, err)
	}

	return nil
}

func isFilesystemRoot(path string) bool {
	cleaned := filepath.Clean(path)
	return cleaned != "" && cleaned == filepath.Dir(cleaned)
}
