// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

//go:build windows

package update

import (
	"errors"
	"fmt"
	"os"
)

const oldSuffix = ".old"

// replaceBinary swaps dst with src on Windows.
//
// Windows blocks deletion / replacement of the running .exe but does
// allow renaming a locked .exe out of the way. We:
//
//  1. Stage src as `<dst>.new` (same directory, so the final rename is
//     just a metadata update and won't cross volumes).
//  2. Move the running binary to `<dst>.old` (NTFS lets us rename a
//     locked exe).
//  3. Move `<dst>.new` into place at `<dst>`.
//
// On the next start the agent's main() calls CleanupAfterRestart to
// best-effort delete `<dst>.old`.
func replaceBinary(dst, src string) error {
	staging := dst + ".new"
	if err := copyFile(src, staging); err != nil {
		return err
	}

	oldPath := dst + oldSuffix
	_ = os.Remove(oldPath)

	if err := os.Rename(dst, oldPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(staging)
		return fmt.Errorf("cannot move running binary aside: %w", err)
	}

	if err := os.Rename(staging, dst); err != nil {
		// Try to roll back the running binary swap.
		_ = os.Rename(oldPath, dst)
		_ = os.Remove(staging)

		return fmt.Errorf("cannot install new binary at %s: %w", dst, err)
	}

	return nil
}

// CleanupAfterRestart removes the previous-version binary left behind
// by replaceBinary. Best-effort: callers ignore errors, so a still-locked
// `<exePath>.old` is fine and will be retried on the next boot.
func CleanupAfterRestart(exePath string) {
	if exePath == "" {
		return
	}

	_ = os.Remove(exePath + oldSuffix)
}
