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
	"os/user"
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
}

func RegisterAutoStart(exePath string) error {
	if exePath == "" {
		return fmt.Errorf("executable path is required")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot resolve home directory: %w", err)
	}

	agentsDir := filepath.Join(home, "Library", "LaunchAgents")
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
		},
	); err != nil {
		return fmt.Errorf("cannot render LaunchAgent plist: %w", err)
	}

	uid, err := currentGUIUID()
	if err != nil {
		return err
	}

	target := fmt.Sprintf("gui/%s/%s", uid, trayLabel)
	_ = exec.Command("launchctl", "bootout", target).Run()
	if out, err := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%s", uid), plistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run launchctl bootstrap: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func currentGUIUID() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("cannot resolve current user: %w", err)
	}

	if u.Uid != "" {
		return u.Uid, nil
	}

	out, err := exec.Command("id", "-u").Output()
	if err != nil {
		return "", fmt.Errorf("cannot resolve uid: %w", err)
	}

	uid := strings.TrimSpace(string(out))
	if _, err := strconv.Atoi(uid); err != nil {
		return "", fmt.Errorf("invalid uid %q", uid)
	}

	return uid, nil
}
