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
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wtsUserName   = 5
	wtsDomainName = 7
)

var (
	modWtsapi32                     = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSQuerySessionInformationW = modWtsapi32.NewProc("WTSQuerySessionInformationW")
)

func wtsQuerySessionString(sessionID uint32, infoClass uint32) (string, error) {
	var (
		buffer        *uint16
		bytesReturned uint32
	)

	r0, _, err := procWTSQuerySessionInformationW.Call(
		0,
		uintptr(sessionID),
		uintptr(infoClass),
		uintptr(unsafe.Pointer(&buffer)),
		uintptr(unsafe.Pointer(&bytesReturned)),
	)
	if r0 == 0 {
		if err != syscall.Errno(0) {
			return "", err
		}

		return "", syscall.EINVAL
	}

	defer windows.WTSFreeMemory(uintptr(unsafe.Pointer(buffer)))

	return windows.UTF16PtrToString(buffer), nil
}

func sessionUserAndDomain(sessionID uint32) (string, string, error) {
	user, err := wtsQuerySessionString(sessionID, wtsUserName)
	if err != nil {
		return "", "", fmt.Errorf("cannot query session %d username: %w", sessionID, err)
	}

	user = strings.TrimSpace(user)
	if user == "" {
		return "", "", fmt.Errorf("session %d username is empty", sessionID)
	}

	domain, err := wtsQuerySessionString(sessionID, wtsDomainName)
	if err != nil {
		return "", "", fmt.Errorf("cannot query session %d domain: %w", sessionID, err)
	}

	return user, strings.TrimSpace(domain), nil
}
