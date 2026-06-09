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

//go:build darwin

package tray

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const (
	trayLabel     = "com.getprobo.agent.tray"
	trayPlistName = trayLabel + ".plist"
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
	Dir     string
}

func RegisterAutoStart(exePath string, dir string) error {
	if exePath == "" {
		return fmt.Errorf("executable path is required")
	}
	if dir == "" {
		return fmt.Errorf("state directory is required")
	}

	guiUser, err := currentGUIUser()
	if err != nil {
		return err
	}

	agentsDir := filepath.Join(guiUser.Home, "Library", "LaunchAgents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		return fmt.Errorf("cannot create LaunchAgents directory: %w", err)
	}

	plistPath := filepath.Join(agentsDir, trayPlistName)

	f, err := os.OpenFile(plistPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("cannot write LaunchAgent plist: %w", err)
	}

	defer func() { _ = f.Close() }()

	if err := launchAgentPlist.Execute(
		f,
		launchAgentData{
			Label:   trayLabel,
			ExePath: exePath,
			Dir:     dir,
		},
	); err != nil {
		return fmt.Errorf("cannot render LaunchAgent plist: %w", err)
	}

	if guiUser.Name != "" {
		_ = os.Chown(plistPath, guiUser.UID, -1)
	}

	target := fmt.Sprintf("gui/%d/%s", guiUser.UID, trayLabel)
	_ = exec.Command("launchctl", "bootout", target).Run()
	if out, err := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%d", guiUser.UID), plistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run launchctl bootstrap: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

type guiUser struct {
	Name string
	UID  int
	Home string
}

func currentGUIUser() (guiUser, error) {
	name := strings.TrimSpace(consoleUserName())
	if name == "" || name == "root" || name == "loginwindow" {
		name = strings.TrimSpace(os.Getenv("USER"))
	}

	if name == "" || name == "root" || name == "loginwindow" {
		return guiUser{}, fmt.Errorf("cannot resolve active GUI user")
	}

	uidOut, err := exec.Command("id", "-u", name).Output()
	if err != nil {
		return guiUser{}, fmt.Errorf("cannot resolve uid for %s: %w", name, err)
	}

	uidRaw := strings.TrimSpace(string(uidOut))
	uid, err := strconv.Atoi(uidRaw)
	if err != nil {
		return guiUser{}, fmt.Errorf("invalid uid %q", uidRaw)
	}

	homeOut, err := exec.Command("dscl", ".", "-read", "/Users/"+name, "NFSHomeDirectory").Output()
	if err != nil {
		return guiUser{}, fmt.Errorf("cannot resolve home directory for %s: %w", name, err)
	}

	home := strings.TrimSpace(strings.TrimPrefix(string(homeOut), "NFSHomeDirectory:"))
	if home == "" {
		return guiUser{}, fmt.Errorf("empty home directory for %s", name)
	}

	return guiUser{Name: name, UID: uid, Home: home}, nil
}

func consoleUserName() string {
	out, err := exec.Command("stat", "-f", "%Su", "/dev/console").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
