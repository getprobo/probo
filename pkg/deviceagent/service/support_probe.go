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

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	supportLogTailLines = 200
	supportProbeTimeout = 15 * time.Second
)

// TailLines returns the last n lines. A trailing newline does not count
// as an extra empty record.
func TailLines(text string, n int) string {
	if n <= 0 {
		return ""
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return strings.Join(lines, "\n")
}

func probeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithTimeout(ctx, supportProbeTimeout)
}

func runCaptured(ctx context.Context, name string, args ...string) (string, string, int) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}

	if stderr.Len() == 0 {
		stderr.WriteString(err.Error())
	}

	return stdout.String(), stderr.String(), -1
}

func formatPathExists(path string) string {
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Sprintf("path: %s\nexists: true", path)
	}

	if errors.Is(err, os.ErrNotExist) {
		return fmt.Sprintf("path: %s\nexists: false", path)
	}

	return fmt.Sprintf("path: %s\nexists: error: %s", path, err.Error())
}
