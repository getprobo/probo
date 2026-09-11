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

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

const attachParentProcess = ^uint32(0)

var (
	modKernel32       = windows.NewLazySystemDLL("kernel32.dll")
	procAttachConsole = modKernel32.NewProc("AttachConsole")
)

// GUI-subsystem binaries start with NUL stdio; attach the parent console
// so CLI commands still print from cmd.exe.
func bindParentConsole() {
	stdoutRedirected := redirectedStdHandle(windows.STD_OUTPUT_HANDLE)
	stderrRedirected := redirectedStdHandle(windows.STD_ERROR_HANDLE)
	stdinRedirected := redirectedStdHandle(windows.STD_INPUT_HANDLE)

	if stdoutRedirected && stderrRedirected && stdinRedirected {
		return
	}

	attachParentProcessConsole()

	if !stdoutRedirected {
		rebindStdHandle(windows.STD_OUTPUT_HANDLE, "CONOUT$", os.O_WRONLY, &os.Stdout)
	}

	if !stderrRedirected {
		rebindStdHandle(windows.STD_ERROR_HANDLE, "CONOUT$", os.O_WRONLY, &os.Stderr)
	}

	if !stdinRedirected {
		rebindStdHandle(windows.STD_INPUT_HANDLE, "CONIN$", os.O_RDONLY, &os.Stdin)
	}
}

func attachParentProcessConsole() {
	_, _, _ = procAttachConsole.Call(uintptr(attachParentProcess))
}

func redirectedStdHandle(stdhandle uint32) bool {
	h, err := windows.GetStdHandle(stdhandle)
	if err != nil || h == 0 || h == windows.InvalidHandle {
		return false
	}

	ft, err := windows.GetFileType(h)

	return isRedirectedFileType(ft, err)
}

func isRedirectedFileType(ft uint32, err error) bool {
	if err != nil {
		return false
	}

	return ft == windows.FILE_TYPE_DISK || ft == windows.FILE_TYPE_PIPE
}

func rebindStdHandle(stdhandle uint32, name string, flag int, dst **os.File) {
	f, err := os.OpenFile(name, flag, 0)
	if err != nil {
		return
	}

	if err := windows.SetStdHandle(stdhandle, windows.Handle(f.Fd())); err != nil {
		_ = f.Close()
		return
	}

	*dst = f
}
