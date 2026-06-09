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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"go.probo.inc/probo/pkg/deviceagent"
)

const createNoWindow = 0x08000000

func promptEnrollmentNative(exePath string) (serverURL, token string, ok bool) {
	uiPath := enrollUIPath(exePath)
	if err := ensureEnrollUI(uiPath); err != nil {
		showError("Enrollment unavailable", err)
		return "", "", false
	}

	cmd := exec.Command(uiPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNoWindow,
	}

	out, err := cmd.Output()
	if err != nil {
		return "", "", false
	}

	var result enrollmentUIResult
	if err := json.Unmarshal(out, &result); err != nil {
		showError("Enrollment failed", fmt.Errorf("cannot decode enrollment UI response: %w", err))
		return "", "", false
	}

	serverURL, err = deviceagent.NormalizeServerURL(result.ServerURL)
	if err != nil {
		showError("Invalid server URL", err)
		return "", "", false
	}

	token = strings.TrimSpace(result.EnrollmentToken)
	if token == "" {
		showError("Enrollment failed", fmt.Errorf("enrollment token is required"))
		return "", "", false
	}

	return serverURL, token, true
}

func enrollUIPath(exePath string) string {
	name := enrollUIBinaryName + ".exe"

	if exePath != "" {
		candidate := filepath.Join(filepath.Dir(exePath), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}

	return filepath.Join(programFiles, "Probo", name)
}

func showError(title string, err error) {
	nativeMessageBox(title, err.Error(), mbOK|mbIconWarning)
}

func showAbout(version string) {
	message := fmt.Sprintf(
		"Version %s\r\n\r\nReports device posture to your Probo workspace.",
		version,
	)
	nativeMessageBox("Probo Device Posture Agent", message, mbOK|mbIconInformation)
}
