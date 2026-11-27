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

//go:build linux

package service

import (
	"context"
	"fmt"
	"strings"
)

// DescribeInstalledUnit records whether the systemd unit file is installed.
func DescribeInstalledUnit(_ context.Context) string {
	return formatPathExists(systemdUnitPath)
}

// ReadServiceLogTail returns the last 200 journal records for the service.
func ReadServiceLogTail(ctx context.Context) (string, error) {
	ctx, cancel := probeContext(ctx)
	defer cancel()

	stdout, stderr, code := runCaptured(
		ctx,
		"journalctl",
		"-u",
		DefaultUnixName,
		"-n",
		"200",
		"--no-pager",
	)
	if code != 0 {
		detail := strings.TrimSpace(stderr)
		if detail == "" {
			detail = strings.TrimSpace(stdout)
		}

		if detail == "" {
			detail = fmt.Sprintf("exit %d", code)
		}

		return "", fmt.Errorf("cannot read journal: %s", detail)
	}

	if strings.TrimSpace(stderr) != "" {
		return stdout + "\nstderr:\n" + stderr, nil
	}

	return stdout, nil
}
