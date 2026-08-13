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
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const logonWithProfile = 0x00000001

var (
	modAdvapi32                 = windows.NewLazySystemDLL("advapi32.dll")
	procCreateProcessWithTokenW = modAdvapi32.NewProc("CreateProcessWithTokenW")
)

func startTrayBestEffort(exePath string, runDir string) {
	if err := startTrayNow(exePath, runDir); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"warning: could not start tray helper; it will start at next logon: %v\n",
			err,
		)
	}
}

func startTrayNow(exePath string, runDir string) error {
	token, err := unelevatedUserToken()
	if err != nil {
		return err
	}

	defer func() { _ = token.Close() }()

	return startProcessAsUser(token, exePath, trayRunCommand(exePath, runDir))
}

func unelevatedUserToken() (windows.Token, error) {
	if isCurrentProcessLocalSystem() {
		return interactiveUserToken()
	}

	var processToken windows.Token
	if err := windows.OpenProcessToken(
		windows.CurrentProcess(),
		windows.TOKEN_QUERY|windows.TOKEN_DUPLICATE,
		&processToken,
	); err != nil {
		return 0, fmt.Errorf("cannot open process token: %w", err)
	}

	defer func() { _ = processToken.Close() }()

	source := processToken
	if processToken.IsElevated() {
		linked, err := processToken.GetLinkedToken()
		if err != nil {
			return 0, fmt.Errorf("cannot get unelevated linked token: %w", err)
		}

		defer func() { _ = linked.Close() }()

		source = linked
	}

	return duplicatePrimaryToken(source)
}

func interactiveUserToken() (windows.Token, error) {
	var lastErr error

	for _, sessionID := range interactiveSessionCandidates() {
		var token windows.Token
		if err := windows.WTSQueryUserToken(sessionID, &token); err != nil {
			lastErr = err
			continue
		}

		return token, nil
	}

	if lastErr != nil {
		return 0, fmt.Errorf("cannot query interactive user token: %w", lastErr)
	}

	return 0, fmt.Errorf("no interactive user session available")
}

func duplicatePrimaryToken(token windows.Token) (windows.Token, error) {
	var primary windows.Token
	if err := windows.DuplicateTokenEx(
		token,
		windows.TOKEN_QUERY|
			windows.TOKEN_DUPLICATE|
			windows.TOKEN_IMPERSONATE|
			windows.TOKEN_ASSIGN_PRIMARY|
			windows.TOKEN_ADJUST_DEFAULT|
			windows.TOKEN_ADJUST_SESSIONID,
		nil,
		windows.SecurityImpersonation,
		windows.TokenPrimary,
		&primary,
	); err != nil {
		return 0, fmt.Errorf("cannot duplicate user token: %w", err)
	}

	return primary, nil
}

func createProcessWithTokenWFlags(hasEnv bool) uint32 {
	if !hasEnv {
		return 0
	}

	// CreateProcessWithTokenW rejects CREATE_NO_WINDOW and
	// CREATE_BREAKAWAY_FROM_JOB (not in its documented dwCreationFlags).
	return windows.CREATE_UNICODE_ENVIRONMENT
}

func startProcessAsUser(token windows.Token, exePath string, commandLine string) error {
	var env *uint16
	if err := windows.CreateEnvironmentBlock(&env, token, false); err != nil {
		env = nil
	} else {
		defer func() { _ = windows.DestroyEnvironmentBlock(env) }()
	}

	desktop, err := windows.UTF16PtrFromString(`winsta0\default`)
	if err != nil {
		return fmt.Errorf("cannot encode desktop name: %w", err)
	}

	appName, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return fmt.Errorf("cannot encode executable path: %w", err)
	}

	cmdLine, err := windows.UTF16PtrFromString(commandLine)
	if err != nil {
		return fmt.Errorf("cannot encode tray command: %w", err)
	}

	si := windows.StartupInfo{
		Desktop:    desktop,
		Flags:      windows.STARTF_USESHOWWINDOW,
		ShowWindow: windows.SW_HIDE,
	}
	si.Cb = uint32(unsafe.Sizeof(si))

	var pi windows.ProcessInformation
	tokenFlags := createProcessWithTokenWFlags(env != nil)

	var tokenErr error
	for _, logonFlags := range []uint32{0, logonWithProfile} {
		tokenErr = createProcessWithTokenW(
			token,
			logonFlags,
			appName,
			cmdLine,
			tokenFlags,
			env,
			&si,
			&pi,
		)
		if tokenErr == nil {
			_ = windows.CloseHandle(pi.Thread)
			_ = windows.CloseHandle(pi.Process)

			return nil
		}
	}

	asUserFlags := tokenFlags | windows.CREATE_NO_WINDOW | windows.CREATE_BREAKAWAY_FROM_JOB
	if err := windows.CreateProcessAsUser(
		token,
		appName,
		cmdLine,
		nil,
		nil,
		false,
		asUserFlags,
		env,
		nil,
		&si,
		&pi,
	); err != nil {
		return fmt.Errorf("cannot start tray helper: %v (also as user: %w)", tokenErr, err)
	}

	_ = windows.CloseHandle(pi.Thread)
	_ = windows.CloseHandle(pi.Process)

	return nil
}

func createProcessWithTokenW(
	token windows.Token,
	logonFlags uint32,
	appName *uint16,
	commandLine *uint16,
	creationFlags uint32,
	env *uint16,
	startupInfo *windows.StartupInfo,
	outProcInfo *windows.ProcessInformation,
) error {
	r1, _, err := procCreateProcessWithTokenW.Call(
		uintptr(token),
		uintptr(logonFlags),
		uintptr(unsafe.Pointer(appName)),
		uintptr(unsafe.Pointer(commandLine)),
		uintptr(creationFlags),
		uintptr(unsafe.Pointer(env)),
		0,
		uintptr(unsafe.Pointer(startupInfo)),
		uintptr(unsafe.Pointer(outProcInfo)),
	)
	if r1 == 0 {
		return err
	}

	return nil
}
