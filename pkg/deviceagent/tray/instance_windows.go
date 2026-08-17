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
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

const trayInstanceMutexName = `Local\ProboAgentTray`

func acquireTrayInstance() (func(), error) {
	name, err := windows.UTF16PtrFromString(trayInstanceMutexName)
	if err != nil {
		return nil, fmt.Errorf("cannot encode tray instance mutex name: %w", err)
	}

	handle, err := windows.CreateMutex(nil, false, name)
	if handle == 0 {
		return nil, fmt.Errorf("cannot create tray instance mutex: %w", err)
	}

	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		_ = windows.CloseHandle(handle)

		return nil, errTrayAlreadyRunning
	}

	return func() {
		_ = windows.CloseHandle(handle)
	}, nil
}
