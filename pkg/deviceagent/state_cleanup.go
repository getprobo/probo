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
	"os"
	"path/filepath"
)

const (
	// SigstoreCacheDirName is the TUF metadata cache under the agent
	// state directory.
	SigstoreCacheDirName = "sigstore-cache"

	// configFileTmpName is the atomic-write leftover from SaveConfig.
	configFileTmpName = ConfigFileName + ".tmp"
)

// RemoveLocalState deletes known agent local state for an install: allowlisted
// files under the state dir and EnrollmentRunDir, plus the sigstore-cache
// subdirectory. It never recursively deletes either directory itself.
// Missing paths are ignored. Filesystem roots and empty paths are rejected.
//
// Every path is attempted even if an earlier removal fails.
//
// enrolling.lock is intentionally retained: flock binds to the inode,
// so unlinking the path while LoadOrExchangeAPIKey holds the lock would
// let a concurrent caller create a new inode and bypass serialization.
func RemoveLocalState(dir string) error {
	if dir == "" {
		return errors.New("state directory is empty")
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("cannot resolve state directory: %w", err)
	}

	cleaned := filepath.Clean(abs)
	if isFilesystemRoot(cleaned) {
		return fmt.Errorf("refusing to remove filesystem root %q", cleaned)
	}

	// Use the original dir (not Abs/Clean) so custom relative --dir values
	// resolve the same sibling run tree as MarkEnrolled / IsEnrolled.
	runDir := EnrollmentRunDir(dir)
	runDirAbs, err := filepath.Abs(runDir)
	if err != nil {
		return fmt.Errorf("cannot resolve enrollment run dir: %w", err)
	}

	runDir = filepath.Clean(runDirAbs)
	if isFilesystemRoot(runDir) {
		return fmt.Errorf("refusing to remove enrollment run dir %q", runDir)
	}

	knownPaths := []string{
		filepath.Join(cleaned, ConfigFileName),
		filepath.Join(cleaned, configFileTmpName),
		filepath.Join(cleaned, KeyFileName),
		filepath.Join(cleaned, KeyFileName+".tmp"),
		filepath.Join(cleaned, KeyFileName+".old"),
		filepath.Join(cleaned, pendingPosturesFileName),
		filepath.Join(cleaned, pendingPosturesFileName+".tmp"),
		filepath.Join(runDir, EnrollmentMarkerName),
	}

	var errs error
	for _, path := range knownPaths {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = errors.Join(errs, fmt.Errorf("cannot remove %s: %w", path, err))
		}
	}

	cacheDir := filepath.Join(cleaned, SigstoreCacheDirName)
	if err := os.RemoveAll(cacheDir); err != nil {
		errs = errors.Join(errs, fmt.Errorf("cannot remove %s: %w", SigstoreCacheDirName, err))
	}

	// Best-effort: remove the dirs only when empty. Leave them alone when
	// foreign files remain or removal is otherwise refused.
	_ = os.Remove(cleaned)
	_ = os.Remove(runDir)

	return errs
}

func isFilesystemRoot(path string) bool {
	cleaned := filepath.Clean(path)
	return cleaned != "" && cleaned == filepath.Dir(cleaned)
}
