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

package service

import (
	"context"
	"fmt"
	"strings"
)

// windowsServiceNotFound is the sc.exe exit for ERROR_SERVICE_DOES_NOT_EXIST.
const windowsServiceNotFound = 1060

// DescribeInstalledUnit records whether the Windows service is installed.
// sc.exe query exits 0 when the service exists, including when it is
// stopped, and 1060 when it does not exist. Any other failure is unknown.
func DescribeInstalledUnit(ctx context.Context) string {
	ctx, cancel := probeContext(ctx)
	defer cancel()

	stdout, stderr, code := runCaptured(ctx, "sc.exe", "query", DefaultWindowsName)

	return formatWindowsInstalled(code, stdout, stderr)
}

func formatWindowsInstalled(code int, stdout, stderr string) string {
	if code == 0 {
		return "installed: true"
	}

	detail := strings.TrimSpace(stderr)
	if detail == "" {
		detail = strings.TrimSpace(stdout)
	}

	if detail == "" {
		detail = fmt.Sprintf("exit %d", code)
	}

	if code == windowsServiceNotFound {
		return "installed: false\n" + detail
	}

	return "installed: unknown\n" + detail
}

// windowsEventLogArgs is wevtutil qe for the newest ProboAgent events.
func windowsEventLogArgs() []string {
	return []string{
		"qe",
		"Application",
		`/q:*[System[Provider[@Name='ProboAgent']]]`,
		fmt.Sprintf("/c:%d", supportLogTailLines),
		"/rd:true",
		"/f:text",
	}
}

// ReadServiceLogTail returns the newest 200 Application events for ProboAgent.
func ReadServiceLogTail(ctx context.Context) (string, error) {
	ctx, cancel := probeContext(ctx)
	defer cancel()

	args := windowsEventLogArgs()
	stdout, stderr, code := runCaptured(ctx, "wevtutil", args...)
	if code != 0 {
		detail := strings.TrimSpace(stderr)
		if detail == "" {
			detail = strings.TrimSpace(stdout)
		}

		if detail == "" {
			detail = fmt.Sprintf("exit %d", code)
		}

		return "", fmt.Errorf("cannot read event log: %s", detail)
	}

	return stdout, nil
}
