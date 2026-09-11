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

//go:build darwin

package deviceagent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// EnsurePrivilegedExecutable copies src to DefaultExecutablePath when
// needed, then locks down the destination tree as root:wheel 0755.
func EnsurePrivilegedExecutable(src string) (string, error) {
	return ensurePrivilegedExecutable(src, DefaultExecutablePath())
}

// RemovePrivilegedExecutable deletes the canonical Darwin binary, the
// leftover /usr/local/bin copy, and an empty /Library/Probo directory.
func RemovePrivilegedExecutable() error {
	if err := removePath(DefaultExecutablePath()); err != nil {
		return err
	}

	if err := RemoveLegacyExecutable(); err != nil {
		return err
	}

	if err := os.Remove(privilegedExecutableDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		if !errors.Is(err, syscall.ENOTEMPTY) {
			return fmt.Errorf("cannot remove %s: %w", privilegedExecutableDir, err)
		}
	}

	return nil
}

// RemoveLegacyExecutable deletes /usr/local/bin/probo-agent when present.
func RemoveLegacyExecutable() error {
	return removePath(LegacyExecutablePath())
}

// IsCanonicalExecutable reports whether path is the privileged Darwin
// install location.
func IsCanonicalExecutable(path string) bool {
	return sameExecutable(path, DefaultExecutablePath())
}

// IsLegacyExecutable reports whether path is the former /usr/local/bin
// install location.
func IsLegacyExecutable(path string) bool {
	return sameExecutable(path, LegacyExecutablePath())
}

// ShouldMigrateExecutable reports whether a running Darwin agent should
// copy itself to the privileged path and rewrite launchd.
func ShouldMigrateExecutable(exe, plistProgram string) bool {
	if IsCanonicalExecutable(exe) {
		return false
	}

	if IsCanonicalExecutable(plistProgram) {
		return false
	}

	if IsLegacyExecutable(exe) {
		return true
	}

	if plistProgram == "" {
		return false
	}

	return sameExecutable(exe, plistProgram)
}

func ensurePrivilegedExecutable(src, dst string) (string, error) {
	if src == "" {
		return "", errors.New("executable path is required")
	}

	if dst == "" {
		return "", errors.New("destination path is required")
	}

	if err := ensurePrivilegedDir(filepath.Dir(dst)); err != nil {
		return "", err
	}

	if !sameExecutable(src, dst) {
		if err := installExecutable(src, dst); err != nil {
			return "", err
		}
	}

	if err := lockDownPath(dst, 0o755); err != nil {
		return "", err
	}

	if err := validatePrivilegedFile(dst); err != nil {
		return "", fmt.Errorf("cannot trust privileged executable: %w", err)
	}

	return dst, nil
}

func ensurePrivilegedDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", dir, err)
	}

	if err := lockDownPath(dir, 0o755); err != nil {
		return err
	}

	if err := validatePrivilegedDir(dir); err != nil {
		return fmt.Errorf("cannot trust privileged directory: %w", err)
	}

	return nil
}

func installExecutable(src, dst string) error {
	staging := dst + ".new"
	_ = os.Remove(staging)

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", src, err)
	}

	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(staging, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", staging, err)
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(staging)

		return fmt.Errorf("cannot copy to %s: %w", staging, err)
	}

	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(staging)

		return fmt.Errorf("cannot fsync %s: %w", staging, err)
	}

	if err := out.Close(); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("cannot close %s: %w", staging, err)
	}

	if err := os.Rename(staging, dst); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("cannot install %s: %w", dst, err)
	}

	return nil
}

func lockDownPath(path string, mode os.FileMode) error {
	if os.Geteuid() == 0 {
		if err := os.Chown(path, 0, 0); err != nil {
			return fmt.Errorf("cannot chown %s: %w", path, err)
		}
	}

	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("cannot chmod %s: %w", path, err)
	}

	return nil
}

func validatePrivilegedFile(path string) error {
	return validatePrivilegedNode(path, false)
}

func validatePrivilegedDir(path string) error {
	return validatePrivilegedNode(path, true)
}

func validatePrivilegedNode(path string, wantDir bool) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", path, err)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s is a symbolic link", path)
	}

	if wantDir && !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	if !wantDir && (fi.IsDir() || !fi.Mode().IsRegular()) {
		return fmt.Errorf("%s is not a regular file", path)
	}

	return nil
}

func sameExecutable(a, b string) bool {
	return normalizeExecPath(a) == normalizeExecPath(b)
}

func normalizeExecPath(path string) string {
	cleaned := filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(cleaned); err == nil {
		return resolved
	}

	return cleaned
}

func removePath(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove %s: %w", path, err)
	}

	return nil
}
