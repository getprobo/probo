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
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	mbOK              = 0x00000000
	mbIconInformation = 0x00000040
)

var (
	modUser32       = windows.NewLazySystemDLL("user32.dll")
	procMessageBoxW = modUser32.NewProc("MessageBoxW")
)

func nativeMessageBox(title, message string, flags uint32) {
	titleUTF16, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return
	}

	messageUTF16, err := windows.UTF16PtrFromString(message)
	if err != nil {
		return
	}

	_, _, _ = procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messageUTF16)),
		uintptr(unsafe.Pointer(titleUTF16)),
		uintptr(flags),
	)
}
