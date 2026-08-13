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

package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

// RunGlobalQueuePackage serializes test packages that exercise process-global
// database queues. Those workers intentionally claim across every tenant, so
// separate `go test` package processes sharing probod_test would otherwise
// consume each other's fixtures.
func RunGlobalQueuePackage(m *testing.M) int {
	lockPath := filepath.Join(os.TempDir(), "probo-global-queue-tests.lock")

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot open global queue test lock: %v\n", err)
		return 1
	}

	defer func() {
		_ = lockFile.Close()
	}()

	if err := unix.Flock(int(lockFile.Fd()), unix.LOCK_EX); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot acquire global queue test lock: %v\n", err)
		return 1
	}

	defer func() {
		_ = unix.Flock(int(lockFile.Fd()), unix.LOCK_UN)
	}()

	return m.Run()
}
