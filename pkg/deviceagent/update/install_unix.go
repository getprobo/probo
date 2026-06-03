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
