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
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"go.probo.inc/probo/pkg/deviceagent/checks"
	"go.probo.inc/probo/pkg/deviceagent/win32"
	"golang.org/x/sys/windows"
)

func confirmSendDiagnostics() (bool, error) {
	ret, err := win32.MessageBox(
		"Probo Device Posture Agent",
		confirmDiagnosticsMessage,
		windows.MB_YESNO|windows.MB_ICONINFORMATION,
	)
	if err != nil {
		return false, fmt.Errorf("cannot confirm diagnostic report: %w", err)
	}

	return ret == win32.IDYes, nil
}

func runElevatedDiagnostics(exePath string) (string, error) {
	if exePath == "" {
		resolved, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("cannot resolve executable path: %w", err)
		}

		exePath = resolved
	}

	candidates := checks.CommandCandidates("powershell")
	if len(candidates) == 0 {
		return "", fmt.Errorf("command %q not available at expected absolute path", "powershell")
	}

	pipe, err := listenDiagnosticsPipe()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(
		candidates[0],
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		elevatedDiagnosticsScript(exePath, pipe.path),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		pipe.close()
		<-pipe.done

		if elevationWasCancelled(out) {
			return "", errDiagnosticsNotAuthorized
		}

		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return "", fmt.Errorf("cannot run diagnostics: %w", err)
		}

		return "", fmt.Errorf("cannot run diagnostics: %s: %w", msg, err)
	}

	result := <-pipe.done
	pipe.close()
	if result.err != nil {
		return "", result.err
	}

	return result.report, nil
}

type pipeRead struct {
	report string
	err    error
}

type diagnosticsPipe struct {
	path   string
	handle windows.Handle
	done   chan pipeRead
	once   sync.Once
}

func listenDiagnosticsPipe() (*diagnosticsPipe, error) {
	path, err := diagnosticsPipePath()
	if err != nil {
		return nil, err
	}

	handle, err := createDiagnosticsPipe(path)
	if err != nil {
		return nil, err
	}

	pipe := &diagnosticsPipe{
		path:   path,
		handle: handle,
		done:   make(chan pipeRead, 1),
	}

	go pipe.serve()

	return pipe, nil
}

func diagnosticsPipePath() (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("cannot create diagnostics pipe name: %w", err)
	}

	return `\\.\pipe\probo-agent-diagnostics-` + hex.EncodeToString(buf[:]), nil
}

func createDiagnosticsPipe(path string) (windows.Handle, error) {
	sid, err := currentUserSID()
	if err != nil {
		return 0, err
	}

	var pinner runtime.Pinner

	defer pinner.Unpin()

	pinner.Pin(sid)

	acl, err := windows.ACLFromEntries(
		[]windows.EXPLICIT_ACCESS{
			{
				AccessPermissions: windows.GENERIC_ALL,
				AccessMode:        windows.GRANT_ACCESS,
				Inheritance:       windows.NO_INHERITANCE,
				Trustee: windows.TRUSTEE{
					TrusteeForm:  windows.TRUSTEE_IS_SID,
					TrusteeType:  windows.TRUSTEE_IS_USER,
					TrusteeValue: windows.TrusteeValueFromSID(sid),
				},
			},
		},
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot build diagnostics pipe DACL: %w", err)
	}

	descriptor, err := windows.NewSecurityDescriptor()
	if err != nil {
		return 0, fmt.Errorf("cannot create diagnostics pipe security descriptor: %w", err)
	}

	if err := descriptor.SetDACL(acl, true, false); err != nil {
		return 0, fmt.Errorf("cannot set diagnostics pipe DACL: %w", err)
	}

	if err := descriptor.SetControl(
		windows.SE_DACL_PROTECTED,
		windows.SE_DACL_PROTECTED,
	); err != nil {
		return 0, fmt.Errorf("cannot protect diagnostics pipe DACL: %w", err)
	}

	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("cannot encode diagnostics pipe name: %w", err)
	}

	sa := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: descriptor,
	}
	handle, err := windows.CreateNamedPipe(
		name,
		windows.PIPE_ACCESS_INBOUND,
		windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT|windows.PIPE_REJECT_REMOTE_CLIENTS,
		1,
		64<<10,
		64<<10,
		0,
		sa,
	)
	runtime.KeepAlive(sid)
	runtime.KeepAlive(acl)
	runtime.KeepAlive(descriptor)
	runtime.KeepAlive(sa)
	if err != nil {
		return 0, fmt.Errorf("cannot create diagnostics pipe: %w", err)
	}

	return handle, nil
}

func currentUserSID() (*windows.SID, error) {
	token := windows.GetCurrentProcessToken()

	user, err := token.GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("cannot read current user: %w", err)
	}

	sid, err := user.User.Sid.Copy()
	runtime.KeepAlive(user)
	if err != nil {
		return nil, fmt.Errorf("cannot copy current user SID: %w", err)
	}

	return sid, nil
}

func (p *diagnosticsPipe) serve() {
	err := windows.ConnectNamedPipe(p.handle, nil)
	if err != nil && !errors.Is(err, windows.ERROR_PIPE_CONNECTED) {
		p.done <- pipeRead{err: fmt.Errorf("cannot connect diagnostics pipe: %w", err)}

		return
	}

	report, err := readDiagnosticsPipe(p.handle)
	if err != nil {
		p.done <- pipeRead{err: err}

		return
	}

	p.done <- pipeRead{report: report}
}

func readDiagnosticsPipe(handle windows.Handle) (string, error) {
	var b bytes.Buffer

	buf := make([]byte, 32<<10)

	for {
		var done uint32

		err := windows.ReadFile(handle, buf, &done, nil)
		if done > 0 {
			if _, writeErr := b.Write(buf[:done]); writeErr != nil {
				return "", fmt.Errorf("cannot buffer diagnostics report: %w", writeErr)
			}
		}

		if err != nil {
			if errors.Is(err, windows.ERROR_BROKEN_PIPE) {
				return b.String(), nil
			}

			return "", fmt.Errorf("cannot read diagnostics report: %w", err)
		}

		if done == 0 {
			return b.String(), nil
		}
	}
}

func (p *diagnosticsPipe) close() {
	p.once.Do(func() {
		_ = windows.CloseHandle(p.handle)
	})
}

func elevatedDiagnosticsScript(exePath, pipePath string) string {
	quotedExe := powershellQuote(exePath)
	quotedPipe := powershellQuote(pipePath)

	return fmt.Sprintf(
		`$ErrorActionPreference = 'Stop'; $p = Start-Process -FilePath %s -ArgumentList @('diagnostics','--output',%s) -Verb RunAs -WindowStyle Hidden -Wait -PassThru; if ($null -eq $p) { exit 1 }; if ($p.ExitCode -ne 0) { exit $p.ExitCode }`,
		quotedExe,
		quotedPipe,
	)
}

func powershellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func openMailto(raw string) error {
	if err := validateSupportMailto(raw); err != nil {
		return err
	}

	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", raw).Start(); err != nil {
		return fmt.Errorf("cannot open email: %w", err)
	}

	return nil
}

func revealPath(path string) error {
	if err := exec.Command("explorer", "/select,"+path).Start(); err != nil {
		return fmt.Errorf("cannot reveal diagnostics report: %w", err)
	}

	return nil
}
