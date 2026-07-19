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

//go:build darwin

package tray

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const (
	trayLabel     = "com.probo.agent.tray"
	trayPlistPath = "/Library/LaunchAgents/com.probo.agent.tray.plist"
)

var (
	//go:embed launchagent.plist.tmpl
	launchAgentPlistTmpl string

	launchAgentPlist = template.Must(
		template.New("launchagent").Funcs(template.FuncMap{"xml": xmlEscape}).Parse(launchAgentPlistTmpl),
	)
)

func xmlEscape(v string) (string, error) {
	var sb strings.Builder
	if err := xml.EscapeText(&sb, []byte(v)); err != nil {
		return "", err
	}

	return sb.String(), nil
}

type launchAgentData struct {
	Label   string
	ExePath string
	RunDir  string
}

func RegisterAutoStart(exePath string, runDir string) error {
	if exePath == "" {
		return fmt.Errorf("executable path is required")
	}

	if runDir == "" {
		return fmt.Errorf("enrollment run directory is required")
	}

	current, err := launchAgentIsCurrent(trayPlistPath, exePath, runDir)
	if err != nil {
		return err
	}

	if current {
		return nil
	}

	if err := writeTrayLaunchAgentPlist(exePath, runDir); err != nil {
		return err
	}

	uids := activeGUIUserUIDs()
	if len(uids) == 0 {
		return nil
	}

	bootstrapTrayForUIDs(uids)

	return nil
}

func renderLaunchAgentPlist(exePath string, runDir string) ([]byte, error) {
	var buf bytes.Buffer
	if err := launchAgentPlist.Execute(
		&buf,
		launchAgentData{
			Label:   trayLabel,
			ExePath: exePath,
			RunDir:  runDir,
		},
	); err != nil {
		return nil, fmt.Errorf("cannot render LaunchAgent plist: %w", err)
	}

	return buf.Bytes(), nil
}

// launchAgentIsCurrent reports whether plistPath already contains the
// LaunchAgent definition for exePath and runDir.
func launchAgentIsCurrent(plistPath string, exePath string, runDir string) (bool, error) {
	existing, err := os.ReadFile(plistPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("cannot read LaunchAgent plist: %w", err)
	}

	desired, err := renderLaunchAgentPlist(exePath, runDir)
	if err != nil {
		return false, err
	}

	return bytes.Equal(existing, desired), nil
}

func writeTrayLaunchAgentPlist(exePath string, runDir string) error {
	agentsDir := filepath.Dir(trayPlistPath)
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		return fmt.Errorf("cannot ensure launch agents directory: %w", err)
	}

	desired, err := renderLaunchAgentPlist(exePath, runDir)
	if err != nil {
		return err
	}

	if err := os.WriteFile(trayPlistPath, desired, 0o644); err != nil {
		return fmt.Errorf("cannot write plist (need root?): %w", err)
	}

	return nil
}

// UnregisterAutoStart stops the tray LaunchAgent and removes its plist.
func UnregisterAutoStart() error {
	bootoutTrayForUIDs(activeGUIUserUIDs())

	if err := os.Remove(trayPlistPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove plist: %w", err)
	}

	return nil
}

// bootstrapTrayForUIDs best-effort loads the tray LaunchAgent into each
// GUI session. Failures are warnings only: the plist is enough for the
// next login (same policy as the macOS PKG postinstall script).
func bootstrapTrayForUIDs(uids []int) {
	for _, uid := range uids {
		target := fmt.Sprintf("gui/%d/%s", uid, trayLabel)

		_ = exec.Command("launchctl", "bootout", target).Run()

		if out, err := exec.Command(
			"launchctl",
			"bootstrap",
			fmt.Sprintf("gui/%d", uid),
			trayPlistPath,
		).CombinedOutput(); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"warning: could not start tray helper for uid %d; it will start at next GUI login: %v: %s\n",
				uid,
				err,
				strings.TrimSpace(string(out)),
			)
		}
	}
}

func bootoutTrayForUIDs(uids []int) {
	for _, uid := range uids {
		target := fmt.Sprintf("gui/%d/%s", uid, trayLabel)
		_ = exec.Command("launchctl", "bootout", target).Run()
	}
}

func activeGUIUserUIDs() []int {
	names := loggedInUsernames()
	if len(names) == 0 {
		if uid, err := currentGUIUserUID(); err == nil {
			return []int{uid}
		}

		return nil
	}

	seen := make(map[int]struct{}, len(names))
	uids := make([]int, 0, len(names))

	for _, name := range names {
		uid, err := uidForUsername(name)
		if err != nil {
			continue
		}

		if _, ok := seen[uid]; ok {
			continue
		}

		seen[uid] = struct{}{}
		uids = append(uids, uid)
	}

	if len(uids) == 0 {
		if uid, err := currentGUIUserUID(); err == nil {
			return []int{uid}
		}
	}

	return uids
}

func loggedInUsernames() []string {
	out, err := exec.Command("users").Output()
	if err != nil {
		return nil
	}

	return parseLoggedInUsernames(string(out))
}

func parseLoggedInUsernames(usersOutput string) []string {
	tokens := strings.Fields(usersOutput)

	seen := make(map[string]struct{}, len(tokens))
	names := make([]string, 0, len(tokens))

	for _, name := range tokens {
		name = strings.TrimSpace(name)
		if name == "" || name == "root" || name == "loginwindow" {
			continue
		}

		if _, ok := seen[name]; ok {
			continue
		}

		seen[name] = struct{}{}
		names = append(names, name)
	}

	return names
}

func currentGUIUserUID() (int, error) {
	name := strings.TrimSpace(consoleUserName())
	if name == "" || name == "root" || name == "loginwindow" {
		name = strings.TrimSpace(os.Getenv("USER"))
	}

	if name == "" || name == "root" || name == "loginwindow" {
		return 0, fmt.Errorf("cannot resolve active GUI user")
	}

	return uidForUsername(name)
}

func uidForUsername(name string) (int, error) {
	uidOut, err := exec.Command("id", "-u", name).Output()
	if err != nil {
		return 0, fmt.Errorf("cannot resolve uid for %s: %w", name, err)
	}

	uidRaw := strings.TrimSpace(string(uidOut))

	uid, err := strconv.Atoi(uidRaw)
	if err != nil {
		return 0, fmt.Errorf("invalid uid %q", uidRaw)
	}

	return uid, nil
}

func consoleUserName() string {
	out, err := exec.Command("stat", "-f", "%Su", "/dev/console").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
