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

//go:build !windows

package update

import (
	"fmt"
	"os"
)

// replaceBinary replaces the file at dst with src.
//
// On Unix the rename is atomic: the kernel keeps the running
// executable mapped via its inode, while the destination path now
// points at the new binary on disk. The next exec (after the
// supervisor restarts the process) loads the new code.
//
// We try a same-directory rename first, then fall back to a
// copy + atomic rename when src and dst live on different
// filesystems (e.g. when /tmp is a tmpfs separate from /usr/local/bin).
func replaceBinary(dst, src string) error {
	if err := os.Chmod(src, 0o755); err != nil {
		return fmt.Errorf("cannot chmod new binary: %w", err)
	}

	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Cross-filesystem fallback: copy into <dst>.new, fsync,
	// then rename within the destination directory.
	staging := dst + ".new"
	if err := copyFile(src, staging); err != nil {
		return err
	}

	if err := os.Chmod(staging, 0o755); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("cannot chmod staged binary: %w", err)
	}

	if err := os.Rename(staging, dst); err != nil {
		_ = os.Remove(staging)
		return fmt.Errorf("cannot atomically replace %s: %w", dst, err)
	}

	return nil
}

// CleanupAfterRestart removes any leftover .old binary from a
// previous Windows-style swap. On Unix this is a no-op.
func CleanupAfterRestart(_ string) {}
