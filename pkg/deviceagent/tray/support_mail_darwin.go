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

//go:build darwin

package tray

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/deviceagent"
)

func confirmSendDiagnostics() (bool, error) {
	path, ok := osascriptPath()
	if !ok {
		return false, fmt.Errorf("osascript is not available")
	}

	script := fmt.Sprintf(
		`display alert "Probo Device Posture Agent" message %q buttons {"Cancel", "Open Email"} default button "Open Email" cancel button "Cancel"`,
		confirmDiagnosticsMessage,
	)

	out, err := exec.Command(path, "-e", script).CombinedOutput()
	if err != nil {
		if elevationWasCancelled(out) {
			return false, nil
		}

		return false, fmt.Errorf("cannot confirm diagnostic report: %w", err)
	}

	return bytes.Contains(out, []byte("Open Email")), nil
}

func diagnosticsExecutable(exePath string) (string, error) {
	if exePath == "" {
		resolved, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("cannot resolve executable path: %w", err)
		}

		exePath = resolved
	}

	if !deviceagent.IsCanonicalExecutable(exePath) {
		return "", fmt.Errorf("executable path must be %s", deviceagent.DefaultExecutablePath())
	}

	return deviceagent.DefaultExecutablePath(), nil
}

func runElevatedDiagnostics(exePath string) (string, error) {
	exePath, err := diagnosticsExecutable(exePath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve diagnostics executable: %w", err)
	}

	path, ok := osascriptPath()
	if !ok {
		return "", fmt.Errorf("osascript is not available")
	}

	// /tmp keeps the socket path under the macOS unix-socket length limit.
	// The directory is mode 0700: root can dial it, other users cannot.
	dir, err := os.MkdirTemp("/tmp", "probo-diag-")
	if err != nil {
		return "", fmt.Errorf("cannot create diagnostics socket directory: %w", err)
	}

	defer func() { _ = os.RemoveAll(dir) }()

	socketPath := filepath.Join(dir, "s")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return "", fmt.Errorf("cannot listen for diagnostics report: %w", err)
	}

	defer func() { _ = listener.Close() }()

	reports := readDiagnosticsSocket(listener)

	// 660 seconds outlasts the diagnostics command timeout so the process
	// can write a partial report and exit on its own.
	command := diagnosticsShellCommand(exePath, socketPath, os.Getpid())
	script := fmt.Sprintf(
		"with timeout of 660 seconds\ndo shell script %s with administrator privileges\nend timeout",
		appleScriptString(command),
	)

	cmd := exec.Command(path, "-e", script)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = listener.Close()

		<-reports

		if elevationWasCancelled(stderr.Bytes()) || elevationWasCancelled(stdout.Bytes()) {
			return "", errDiagnosticsNotAuthorized
		}

		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return "", fmt.Errorf("cannot run diagnostics: %w", err)
		}

		return "", fmt.Errorf("cannot run diagnostics: %s: %w", msg, err)
	}

	result := <-reports
	if result.err != nil {
		return "", result.err
	}

	return result.report, nil
}

type socketRead struct {
	report string
	err    error
}

func readDiagnosticsSocket(listener net.Listener) <-chan socketRead {
	reports := make(chan socketRead, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			reports <- socketRead{err: fmt.Errorf("cannot accept diagnostics report: %w", err)}

			return
		}

		defer func() { _ = conn.Close() }()

		data, err := io.ReadAll(conn)
		if err != nil {
			reports <- socketRead{err: fmt.Errorf("cannot read diagnostics report: %w", err)}

			return
		}

		reports <- socketRead{report: string(data)}
	}()

	return reports
}

func diagnosticsShellCommand(exePath, outputPath string, peerPID int) string {
	return shellSingleQuote(exePath) +
		" diagnostics --output " + shellSingleQuote(outputPath) +
		" --output-pid " + strconv.Itoa(peerPID)
}

func appleScriptString(value string) string {
	var b strings.Builder

	b.WriteByte('"')

	for _, r := range value {
		if r == '\\' || r == '"' {
			b.WriteByte('\\')
		}

		b.WriteRune(r)
	}

	b.WriteByte('"')

	return b.String()
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func openMailto(raw string) error {
	if err := validateSupportMailto(raw); err != nil {
		return err
	}

	if err := exec.Command("open", raw).Start(); err != nil {
		return fmt.Errorf("cannot open email: %w", err)
	}

	return nil
}

func revealPath(path string) error {
	if err := exec.Command("open", "-R", path).Start(); err != nil {
		return fmt.Errorf("cannot reveal diagnostics report: %w", err)
	}

	return nil
}
