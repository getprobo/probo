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

package deviceagent

import (
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

// WriteDiagnosticsOutput writes the report to a unix socket. A regular
// file path is refused so an elevated run cannot replace an arbitrary file.
// The report is written only when the connected process id is peerPID.
func WriteDiagnosticsOutput(path, report string, peerPID int) error {
	if path == "" {
		return fmt.Errorf("diagnostics output path is empty")
	}

	if peerPID <= 0 {
		return fmt.Errorf("diagnostics listener pid is required")
	}

	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("cannot open diagnostics output: %w", err)
	}

	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("diagnostics output must be a socket: %s", path)
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		return fmt.Errorf("cannot connect diagnostics output: %w", err)
	}

	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		_ = conn.Close()

		return fmt.Errorf("cannot use diagnostics socket connection")
	}

	if err := verifyDiagnosticsPeer(unixConn, peerPID); err != nil {
		_ = conn.Close()

		return err
	}

	_, err = io.WriteString(conn, report)
	closeErr := conn.Close()

	if err != nil {
		return fmt.Errorf("cannot write diagnostics report: %w", err)
	}

	if closeErr != nil {
		return fmt.Errorf("cannot close diagnostics output: %w", closeErr)
	}

	return nil
}

func verifyDiagnosticsPeer(conn *net.UnixConn, peerPID int) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("cannot access diagnostics socket: %w", err)
	}

	var (
		pid    int
		pidErr error
	)

	if err := raw.Control(func(fd uintptr) {
		pid, pidErr = unix.GetsockoptInt(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERPID)
	}); err != nil {
		return fmt.Errorf("cannot control diagnostics socket: %w", err)
	}

	if pidErr != nil {
		return fmt.Errorf("cannot read diagnostics listener pid: %w", pidErr)
	}

	if pid != peerPID {
		return fmt.Errorf("diagnostics listener pid does not match")
	}

	return nil
}
