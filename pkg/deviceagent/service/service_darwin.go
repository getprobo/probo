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

const (
	plistPath = "/Library/LaunchDaemons/com.probo.agent.plist"

	helperLabel      = "com.probo.agent.helper"
	helperPlistPath  = "/Library/LaunchDaemons/" + helperLabel + ".plist"
	helperBinaryPath = "/Library/PrivilegedHelperTools/" + helperLabel
)

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

func removeLaunchDaemonPlist(path string) error {
	_ = exec.Command("launchctl", "bootout", "system", path).Run()
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove plist %s: %w", path, err)
	}

	return nil
}

func removeManagedPath(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove %s: %w", path, err)
	}

	return nil
}

// removePrivilegedHelper boots out and deletes the PKG-installed XPC helper.
// Missing artifacts are treated as success so uninstall stays idempotent.
func removePrivilegedHelper() error {
	_ = exec.Command("launchctl", "bootout", "system/"+helperLabel).Run()
	_ = exec.Command("launchctl", "bootout", "system", helperPlistPath).Run()

	if err := removeManagedPath(helperPlistPath); err != nil {
		return fmt.Errorf("cannot remove privileged helper plist: %w", err)
	}

	if err := removeManagedPath(helperBinaryPath); err != nil {
		return fmt.Errorf("cannot remove privileged helper binary: %w", err)
	}

	return nil
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

// Uninstall bootouts and removes the agent LaunchDaemon and the privileged
// XPC helper installed by the macOS PKG.
func Uninstall(cfg Config) error {
	_ = cfg

	if err := removeLaunchDaemonPlist(plistPath); err != nil {
		return err
	}

	if err := removePrivilegedHelper(); err != nil {
		return fmt.Errorf("cannot remove privileged helper: %w", err)
	}

	return nil
}
