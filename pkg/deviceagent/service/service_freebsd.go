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
	rcScriptPath = "/usr/local/etc/rc.d/probo_agent"
)

// FreeBSD rc.d script template.
const rcScriptTmpl = `#!/bin/sh
#
# PROVIDE: probo_agent
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name=probo_agent
rcvar=probo_agent_enable
desc="Probo device posture agent"
pidfile="/var/run/${name}.pid"
procname="{{.ExePath}}"
command=/usr/sbin/daemon
command_args="-r -P ${pidfile} -- \"{{.ExePath}}\" run --dir \"{{.Dir}}\""

load_rc_config $name
: ${probo_agent_enable:=YES}

run_rc_command "$1"
`

func Install(cfg Config) error {
	if cfg.ExePath == "" {
		return errors.New("executable path is required")
	}

	if cfg.Dir == "" {
		return errors.New("state directory is required")
	}

	if err := validateServicePaths(cfg); err != nil {
		return err
	}

	rcTmpl, err := template.New("rc").Parse(rcScriptTmpl)
	if err != nil {
		return fmt.Errorf("cannot parse rc.d template: %w", err)
	}

	sf, err := os.OpenFile(rcScriptPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("cannot write rc.d script (need root?): %w", err)
	}

	defer func() { _ = sf.Close() }()

	if err := rcTmpl.Execute(sf, cfg); err != nil {
		return fmt.Errorf("cannot render rc.d script: %w", err)
	}

	if out, err := exec.Command("service", "probo_agent", "enable").CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run service probo_agent enable: %w: %s", err, strings.TrimSpace(string(out)))
	}

	if out, err := exec.Command("service", "probo_agent", "start").CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run service probo_agent start: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func Uninstall(cfg Config) error {
	_ = exec.Command("service", "probo_agent", "stop").Run()
	_ = exec.Command("service", "probo_agent", "disable").Run()

	if err := os.Remove(rcScriptPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot remove rc.d script: %w", err)
	}

	return nil
}
