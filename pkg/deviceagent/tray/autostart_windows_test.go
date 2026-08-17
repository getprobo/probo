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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
)

func TestTrayRunCommand_QuotesPaths(t *testing.T) {
	t.Parallel()

	got := trayRunCommand(
		`C:\Program Files\Probo\probo-agent.exe`,
		`C:\ProgramData\Probo\run`,
	)

	assert.Equal(
		t,
		`"C:\Program Files\Probo\probo-agent.exe" tray --run-dir "C:\ProgramData\Probo\run"`,
		got,
	)
}

func TestCreateProcessWithTokenWFlags_OmitsUnsupportedBits(t *testing.T) {
	t.Parallel()

	flags := createProcessWithTokenWFlags(true)

	assert.Equal(t, uint32(0), flags&windows.CREATE_NO_WINDOW)
	assert.Equal(t, uint32(0), flags&windows.CREATE_BREAKAWAY_FROM_JOB)
	assert.Equal(t, uint32(windows.CREATE_UNICODE_ENVIRONMENT), flags&windows.CREATE_UNICODE_ENVIRONMENT)

	flags = createProcessWithTokenWFlags(false)
	assert.Equal(t, uint32(0), flags)
}
