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
	"os/exec"
	"strings"
)

const windowsDiagnosticLogPath = "Application log (" + WindowsEventSource + ")"

func readDiagnosticLog(limit int) (string, string, error) {
	text, err := readWindowsEventTail(limit)
	if err != nil {
		return "", windowsDiagnosticLogPath, err
	}

	return text, windowsDiagnosticLogPath, nil
}

func readWindowsEventTail(limit int) (string, error) {
	command := exec.Command("wevtutil.exe", wevtutilQueryArgs(limit)...)

	out, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			detail := strings.TrimSpace(string(exit.Stderr))
			if detail == "" {
				return "", fmt.Errorf("cannot query Windows event log: %w", err)
			}

			return "", fmt.Errorf("cannot query Windows event log: %s: %w", detail, err)
		}

		return "", fmt.Errorf("cannot run wevtutil: %w", err)
	}

	messages, err := eventMessagesFromXML(decodeCommandOutput(out))
	if err != nil {
		return "", err
	}

	return joinNewestEventLines(messages, limit), nil
}
