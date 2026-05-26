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
