// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

//go:build windows

package tray

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	mbOK              = 0x00000000
	mbIconWarning     = 0x00000030
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
