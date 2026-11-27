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

package tray

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"go.probo.inc/probo/pkg/deviceagent"
)

func diagnosticsArgs(runDir string) ([]string, error) {
	if runDir == "" {
		runDir = deviceagent.DefaultEnrollmentRunDir()
	}

	dir := deviceagent.ConfigDirForRunDir(runDir)
	if dir == "" {
		return nil, errors.New("cannot map run directory to a state directory")
	}

	return []string{
		"diagnostics",
		"--dir",
		dir,
		"--run-dir",
		runDir,
	}, nil
}

// saveSupportBundle runs diagnostics as the current user. The tray does
// not write the file and does not elevate.
func saveSupportBundle(opts Options) (string, error) {
	if opts.ExePath == "" {
		return "", errors.New("agent executable is unknown")
	}

	args, err := diagnosticsArgs(opts.RunDir)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(opts.ExePath, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}

		return "", errors.New(message)
	}

	path := strings.TrimSpace(stdout.String())
	if path == "" {
		return "", errors.New("diagnostics did not print a path")
	}

	return path, nil
}
