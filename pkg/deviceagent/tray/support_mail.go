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

//go:build darwin || windows

package tray

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const (
	supportAddress = "support@probo.com"

	confirmDiagnosticsMessage = "Open your email app with a diagnostic report for support@probo.com? " +
		"macOS or Windows will ask for an administrator password. " +
		"Collection can take a few minutes. " +
		"You can review the message before you send it. " +
		"The report does not include API keys."

	diagnosticsNotCreated = "Diagnostic report was not created."

	mailtoLimitDarwin  = 12000
	mailtoLimitWindows = 1800
)

var errDiagnosticsNotAuthorized = errors.New("administrator authorization was cancelled")

func sendDiagnostics(opts Options) {
	ok, err := confirmSendDiagnostics()
	if err != nil {
		showEnrollmentError(err.Error())
		return
	}

	if !ok {
		return
	}

	report, err := runElevatedDiagnostics(opts.ExePath)
	if err != nil {
		if errors.Is(err, errDiagnosticsNotAuthorized) {
			showEnrollmentError(diagnosticsNotCreated)
			return
		}

		showEnrollmentError(err.Error())

		return
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}

	if err := openSupportDraft(report, opts.Version, hostname); err != nil {
		showEnrollmentError(err.Error())
	}
}

func openSupportDraft(report, version, hostname string) error {
	subject := fmt.Sprintf("Probo agent diagnostics (%s, %s)", version, hostname)

	raw, overflow := draftMailto(subject, report, mailtoLimit())
	if !overflow {
		return openMailto(raw)
	}

	path, err := saveDiagnosticsReport(report)
	if err != nil {
		return err
	}

	_ = revealPath(path)

	body := fmt.Sprintf(
		"The full diagnostic report is saved at:\n%s\n\nAttach that file to this message.",
		path,
	)

	short, overflow := draftMailto(subject, body, mailtoLimit())
	if overflow {
		return fmt.Errorf("cannot open email. The report is saved at %s", path)
	}

	if err := openMailto(short); err != nil {
		return fmt.Errorf("cannot open email. The report is saved at %s: %w", path, err)
	}

	return nil
}

func draftMailto(subject, body string, limit int) (string, bool) {
	raw := supportMailto(subject, body)
	if len(raw) <= limit {
		return raw, false
	}

	return "", true
}

func supportMailto(subject, body string) string {
	query := url.Values{}
	query.Set("subject", subject)
	query.Set("body", body)

	// Form encoding writes a space as "+". In a mailto URL, "+" is a
	// literal plus, so mail clients would show it in the draft.
	encoded := strings.ReplaceAll(query.Encode(), "+", "%20")

	return "mailto:" + supportAddress + "?" + encoded
}

func validateSupportMailto(raw string) error {
	prefix := "mailto:" + supportAddress + "?"
	if !strings.HasPrefix(raw, prefix) {
		return errors.New("refusing to open a mail draft for a different recipient")
	}

	if strings.ContainsAny(raw, "\r\n") {
		return errors.New("refusing to open a mail draft with a raw newline")
	}

	return nil
}

func mailtoLimit() int {
	if runtime.GOOS == "windows" {
		return mailtoLimitWindows
	}

	return mailtoLimitDarwin
}

func saveDiagnosticsReport(report string) (string, error) {
	file, err := os.CreateTemp("", "probo-agent-diagnostics-*.txt")
	if err != nil {
		return "", fmt.Errorf("cannot create diagnostics report: %w", err)
	}

	path := file.Name()
	if _, err := file.WriteString(report); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("cannot write diagnostics report: %w", err)
	}

	if err := file.Close(); err != nil {
		return "", fmt.Errorf("cannot close diagnostics report: %w", err)
	}

	return path, nil
}

func elevationWasCancelled(output []byte) bool {
	text := strings.ToLower(string(output))

	return strings.Contains(text, "user canceled") ||
		strings.Contains(text, "user cancelled") ||
		strings.Contains(text, "canceled by the user") ||
		strings.Contains(text, "cancelled by the user") ||
		strings.Contains(text, "(-128)")
}
