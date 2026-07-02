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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const (
	systemdUnitPath = "/etc/systemd/system/probo-agent.service"
)

const systemdUnitTmpl = `[Unit]
Description=Probo device posture agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{.ExePath}} run --dir {{.Dir}}
Restart=always
RestartSec=10
# 75 is the exit code emitted after a successful self-update.
# Treat it as a normal exit so the unit restarts without entering
# the "failed" state.
SuccessExitStatus=75
User=root
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
`

func Install(cfg Config) error {
	if cfg.ExePath == "" {
		return errors.New("executable path is required")
	}

	if cfg.Dir == "" {
		return errors.New("state directory is required")
	}

	tmpl, err := template.New("unit").Parse(systemdUnitTmpl)
	if err != nil {
		return fmt.Errorf("cannot parse unit template: %w", err)
	}

	f, err := os.OpenFile(systemdUnitPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("cannot write systemd unit (need root?): %w", err)
	}

	defer func() { _ = f.Close() }()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("cannot render systemd unit: %w", err)
	}

	if out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run systemctl daemon-reload: %w: %s", err, strings.TrimSpace(string(out)))
	}

	if out, err := exec.Command("systemctl", "enable", "--now", "probo-agent.service").CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run systemctl enable --now: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func Uninstall(cfg Config) error {
	_ = exec.Command("systemctl", "disable", "--now", "probo-agent.service").Run()

	if err := os.Remove(systemdUnitPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove systemd unit: %w", err)
	}

	_ = exec.Command("systemctl", "daemon-reload").Run()

	return nil
}
