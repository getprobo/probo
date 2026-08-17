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

package tray

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func stopInteractiveAgentProcessesBestEffort() {
	if err := stopInteractiveAgentProcesses(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"warning: could not stop tray helper: %v\n",
			err,
		)
	}
}

func stopInteractiveAgentProcesses() error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("cannot snapshot processes: %w", err)
	}

	defer func() { _ = windows.CloseHandle(snapshot) }()

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return fmt.Errorf("cannot read process snapshot: %w", err)
	}

	selfPID := uint32(os.Getpid())

	for {
		baseName := windows.UTF16ToString(entry.ExeFile[:])

		var sessionID uint32
		if err := windows.ProcessIdToSessionId(entry.ProcessID, &sessionID); err == nil &&
			shouldStopInteractiveAgentProcess(
				baseName,
				entry.ProcessID,
				selfPID,
				sessionID,
			) {
			if err := terminatePID(entry.ProcessID); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"warning: cannot stop tray helper (pid %d): %v\n",
					entry.ProcessID,
					err,
				)
			}
		}

		err := windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	return nil
}

func terminatePID(pid uint32) error {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return fmt.Errorf("cannot open process %d: %w", pid, err)
	}

	defer func() { _ = windows.CloseHandle(handle) }()

	if err := windows.TerminateProcess(handle, 1); err != nil {
		return fmt.Errorf("cannot terminate process %d: %w", pid, err)
	}

	return nil
}
