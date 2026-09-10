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

package win32

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
)

// IDYes is the MessageBox result when the user selects Yes.
const IDYes = 6

// MessageBox shows a native Windows dialog and returns the selected button ID.
func MessageBox(title, message string, flags uint32) (int32, error) {
	titleUTF16, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return 0, fmt.Errorf("cannot encode message box title: %w", err)
	}

	body := strings.ReplaceAll(message, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\n", "\r\n")

	messageUTF16, err := windows.UTF16PtrFromString(body)
	if err != nil {
		return 0, fmt.Errorf("cannot encode message box message: %w", err)
	}

	ret, err := windows.MessageBox(0, messageUTF16, titleUTF16, flags)
	if err != nil {
		return 0, fmt.Errorf("cannot show message box: %w", err)
	}

	return ret, nil
}
