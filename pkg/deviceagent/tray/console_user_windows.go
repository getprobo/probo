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

const invalidSessionID = 0xFFFFFFFF

func interactiveSessionCandidates() []uint32 {
	candidates := make([]uint32, 0, 4)

	var sessionID uint32

	if err := windows.ProcessIdToSessionId(windows.GetCurrentProcessId(), &sessionID); err == nil && sessionID != 0 {
		candidates = appendUniqueSessionID(candidates, sessionID)
	}

	consoleSessionID := windows.WTSGetActiveConsoleSessionId()
	if consoleSessionID != invalidSessionID {
		candidates = appendUniqueSessionID(candidates, consoleSessionID)
	}

	for _, id := range activeWTSSessionIDs() {
		candidates = appendUniqueSessionID(candidates, id)
	}

	return candidates
}

func activeWTSSessionIDs() []uint32 {
	var (
		sessions *windows.WTS_SESSION_INFO
		count    uint32
	)

	if err := windows.WTSEnumerateSessions(0, 0, 1, &sessions, &count); err != nil {
		return nil
	}

	defer windows.WTSFreeMemory(uintptr(unsafe.Pointer(sessions)))

	sessionSlice := unsafe.Slice(sessions, count)

	ids := make([]uint32, 0, len(sessionSlice))
	for _, session := range sessionSlice {
		if session.SessionID == 0 {
			continue
		}

		if session.State != windows.WTSActive && session.State != windows.WTSConnected {
			continue
		}

		ids = append(ids, session.SessionID)
	}

	return ids
}

func appendUniqueSessionID(ids []uint32, sessionID uint32) []uint32 {
	for _, id := range ids {
		if id == sessionID {
			return ids
		}
	}

	return append(ids, sessionID)
}

func isCurrentProcessLocalSystem() bool {
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return false
	}

	defer func() { _ = token.Close() }()

	tu, err := token.GetTokenUser()
	if err != nil {
		return false
	}

	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return false
	}

	return tu.User.Sid.Equals(systemSID)
}
