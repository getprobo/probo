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
