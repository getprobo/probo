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
)

type secureDirKind int

const (
	secureDirRoot secureDirKind = iota
	secureDirRun
	secureDirAgent
)

func ensureSecureRunDir(runDir string) error {
	if runDir == "" {
		runDir = DefaultEnrollmentRunDir()
	}

	if err := ensureProtectedWindowsTree(runDir, secureDirRun); err != nil {
		return err
	}

	if err := os.MkdirAll(runDir, EnrollmentRunDirMode); err != nil {
		return fmt.Errorf("cannot create enrollment run dir: %w", err)
	}

	if err := os.Chmod(runDir, EnrollmentRunDirMode); err != nil {
		return fmt.Errorf("cannot set enrollment run dir permissions: %w", err)
	}

	return nil
}

func ensureSecureAgentDir(dir string) error {
	if dir == "" {
		dir = DefaultConfigDir()
	}

	if err := ensureProtectedWindowsTree(dir, secureDirAgent); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create agent data dir: %w", err)
	}

	return nil
}
