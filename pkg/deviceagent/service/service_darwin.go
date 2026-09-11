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
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"go.probo.inc/probo/pkg/deviceagent"
)

const (
	plistPath = "/Library/LaunchDaemons/com.probo.agent.plist"

	helperLabel      = "com.probo.agent.helper"
	helperPlistPath  = "/Library/LaunchDaemons/" + helperLabel + ".plist"
	helperBinaryPath = "/Library/PrivilegedHelperTools/" + helperLabel
)

var (
	//go:embed launchd.plist.tmpl
	launchdPlistTmpl string

	launchdPlist = template.Must(
		template.New("plist").Funcs(template.FuncMap{"xml": xmlEscape}).Parse(launchdPlistTmpl),
	)
)

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
	if err := WritePlist(cfg); err != nil {
		return err
	}

	return reloadNow()
}

// WritePlist renders the LaunchDaemon plist without touching launchd.
// Use this when the running daemon cannot boot itself out.
func WritePlist(cfg Config) error {
	if cfg.ExePath == "" {
		return errors.New("executable path is required")
	}

	if cfg.Dir == "" {
		return errors.New("state directory is required")
	}

	if cfg.Label == "" {
		cfg.Label = DefaultLabel
	}

	if !deviceagent.IsCanonicalExecutable(cfg.ExePath) {
		return fmt.Errorf("executable path must be %s", deviceagent.DefaultExecutablePath())
	}

	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return fmt.Errorf("cannot ensure launch daemons directory: %w", err)
	}

	f, err := os.OpenFile(plistPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("cannot write plist (need root?): %w", err)
	}

	defer func() { _ = f.Close() }()

	if err := launchdPlist.Execute(f, cfg); err != nil {
		return fmt.Errorf("cannot render plist: %w", err)
	}

	return nil
}

func reloadNow() error {
	// `bootout` first keeps install idempotent.
	_ = exec.Command("launchctl", "bootout", "system", plistPath).Run()
	if out, err := exec.Command("launchctl", "bootstrap", "system", plistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run launchctl bootstrap: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

// ScheduleReload bootouts and bootstraps the agent LaunchDaemon from a
// detached process so a running daemon can change its own Program path.
func ScheduleReload() error {
	script := "sleep 1; /bin/launchctl bootout system " + plistPath +
		"; /bin/launchctl bootstrap system " + plistPath
	cmd := exec.Command("/bin/sh", "-c", script)

	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot schedule launchd reload: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("cannot detach launchd reload: %w", err)
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

	if err := deviceagent.RemovePrivilegedExecutable(); err != nil {
		return fmt.Errorf("cannot remove privileged executable: %w", err)
	}

	return nil
}

// InstalledExePath returns ProgramArguments[0] from the agent LaunchDaemon
// plist, or "" when the plist is missing or unreadable.
func InstalledExePath() string {
	out, err := exec.Command(
		"/usr/libexec/PlistBuddy",
		"-c",
		"Print :ProgramArguments:0",
		plistPath,
	).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
