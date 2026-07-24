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
	// EnrollmentLockFileName serializes concurrent LoadOrExchangeAPIKey
	// calls that exchange enrollment tokens and write agent.key. It lives
	// under EnrollmentRunDir so a reboot clears it with the run tree.
	EnrollmentLockFileName = "enrolling.lock"

	// enrollmentLockMode is owner-only; only the elevated install path
	// creates and holds this lock.
	enrollmentLockMode = 0o600
)

// EnrollmentLockPath returns the absolute path of the enrollment lock file
// under EnrollmentRunDir for the given state directory.
func EnrollmentLockPath(configDir string) string {
	return filepath.Join(EnrollmentRunDir(configDir), EnrollmentLockFileName)
}

// AcquireEnrollmentLock takes an exclusive lock on enrolling.lock under
// EnrollmentRunDir(configDir) for the duration of credential exchange. The
// returned release function unlocks and closes the lock file. The path is
// left in place so waiters that already opened the inode stay serialized
// with later acquirers; the kernel releases the lock on close (including
// after crashes). Callers must not unlink this path while a holder may
// exist.
func AcquireEnrollmentLock(configDir string) (release func(), err error) {
	runDir := EnrollmentRunDir(configDir)

	if err := os.MkdirAll(runDir, EnrollmentRunDirMode); err != nil {
		return nil, fmt.Errorf("cannot create enrollment run dir: %w", err)
	}

	if err := os.Chmod(runDir, EnrollmentRunDirMode); err != nil {
		return nil, fmt.Errorf("cannot set enrollment run dir permissions: %w", err)
	}

	path := EnrollmentLockPath(configDir)

	file, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		enrollmentLockMode,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open enrollment lock: %w", err)
	}

	if err := lockFileExclusive(file); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("cannot acquire enrollment lock: %w", err)
	}

	return func() {
		_ = unlockFile(file)
		_ = file.Close()
	}, nil
}
