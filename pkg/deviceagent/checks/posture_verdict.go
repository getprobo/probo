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

package checks

import (
	"strings"
)

// A delay is a reading, not a failed check. Only an off lock is a failure.
func darwinScreenLockModeStatus(mode string) Status {
	if mode == "off" {
		return StatusFail
	}

	return StatusPass
}

// A minimum of linuxPasswordMinLen or more is a policy. Shorter settings
// and an explicit zero are failures. An unread length stays unknown.
func linuxPasswordPolicyVerdict(length int, known bool) Status {
	if !known || length < 0 {
		return StatusUnknown
	}

	if length < linuxPasswordMinLen {
		return StatusFail
	}

	return StatusPass
}

func linuxRemoteLoginStatus(state string) Status {
	switch strings.TrimSpace(state) {
	case "active", "activating":
		return StatusFail
	case "inactive", "failed", "deactivating":
		return StatusPass
	case "":
		return StatusNotApplicable
	default:
		return StatusUnknown
	}
}

// A recorded length is a reading, including zero. An unread length stays unknown.
func windowsPasswordPolicyVerdict(minLen int, known bool) Status {
	if !known || minLen < 0 {
		return StatusUnknown
	}

	return StatusPass
}
