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

package deviceagent

import (
	"errors"
	"fmt"
	"os"
)

func replaceRegularFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("cannot write temp file: %w", err)
	}

	old := path + ".old"
	_ = os.Remove(old)

	if err := os.Rename(path, old); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(tmp)
		return fmt.Errorf("cannot move existing file aside: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Rename(old, path)
		_ = os.Remove(tmp)

		return fmt.Errorf("cannot replace file: %w", err)
	}

	_ = os.Remove(old)

	return nil
}
