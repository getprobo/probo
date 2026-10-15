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
	"strconv"
	"strings"
)

// A delay is a reading, not a failed check. Only an off lock is a failure.
func darwinScreenLockModeStatus(mode string) Status {
	if mode == "off" {
		return StatusFail
	}

	return StatusPass
}

// A non-negative length is a reading. Only a missing PASS_MIN_LEN is a failure.
func linuxPasswordPolicyFromDefs(minLen, maxDays string) Result {
	ev := map[string]any{
		"pass_min_len":  minLen,
		"pass_max_days": maxDays,
	}
	if minLen == "" {
		ev["parse_error"] = "PASS_MIN_LEN not set"
		return fail(ev)
	}

	minLenValue, err := strconv.Atoi(minLen)
	if err != nil || minLenValue < 0 {
		ev["parse_error"] = "invalid PASS_MIN_LEN value"
		return unknown(ev)
	}

	ev["pass_min_len_value"] = minLenValue

	return pass(ev)
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
