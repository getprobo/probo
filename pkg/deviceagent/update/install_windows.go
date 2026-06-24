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
