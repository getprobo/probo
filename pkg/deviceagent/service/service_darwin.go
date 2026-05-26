// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package service

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const plistPath = "/Library/LaunchDaemons/com.getprobo.agent.plist"

const launchdPlistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{xml .Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{xml .ExePath}}</string>
        <string>run</string>
        <string>--dir</string>
        <string>{{xml .Dir}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/probo-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/probo-agent.log</string>
    <key>UserName</key>
    <string>root</string>
    <key>GroupName</key>
    <string>wheel</string>
</dict>
</plist>
`

func xmlEscape(v string) (string, error) {
	var sb strings.Builder
	if err := xml.EscapeText(&sb, []byte(v)); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// Install writes and boots the launchd plist.
func Install(cfg Config) error {
	if cfg.ExePath == "" {
		return errors.New("executable path is required")
	}

	if cfg.Dir == "" {
		return errors.New("state directory is required")
	}

	if cfg.Label == "" {
		cfg.Label = DefaultLabel
	}

	tmpl, err := template.New("plist").Funcs(template.FuncMap{"xml": xmlEscape}).Parse(launchdPlistTmpl)
	if err != nil {
		return fmt.Errorf("cannot parse plist template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return fmt.Errorf("cannot ensure launch daemons directory: %w", err)
	}

	f, err := os.OpenFile(plistPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("cannot write plist (need root?): %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("cannot render plist: %w", err)
	}

	// `bootout` first keeps install idempotent.
	_ = exec.Command("launchctl", "bootout", "system", plistPath).Run()
	if out, err := exec.Command("launchctl", "bootstrap", "system", plistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run launchctl bootstrap: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

// Uninstall bootouts and removes the launchd plist.
func Uninstall(cfg Config) error {
	_ = exec.Command("launchctl", "bootout", "system", plistPath).Run()
	if err := os.Remove(plistPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove plist: %w", err)
	}

	return nil
}
