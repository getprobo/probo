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

package checks

import (
	"context"
	"os"
	"strings"
)

func init() {
	Register(KeyDiskEncryption, freebsdDiskEncryption)
	Register(KeyScreenLock, freebsdScreenLock)
	Register(KeyFirewallEnabled, freebsdFirewall)
	Register(KeyTimeSync, freebsdTimeSync)
	Register(KeyOSVersion, freebsdOSVersion)
	Register(KeyAutoUpdate, freebsdAutoUpdate)
	Register(KeyPasswordPolicy, freebsdPasswordPolicy)
	Register(KeyRemoteLogin, freebsdRemoteLogin)
	Register(KeyMalwareProtection, freebsdMalwareProtection)
}

func freebsdDiskEncryption(ctx context.Context) Result {
	if !CommandExists("geli") {
		return unknown(map[string]any{"note": "geli command not found"})
	}
	out := RunCommand(ctx, "geli", "status")
	ev := map[string]any{"raw": out.Stdout, "stderr": out.Stderr}
	if out.Err != nil {
		return unknown(ev)
	}
	if strings.Contains(out.Stdout, "ACTIVE") {
		return pass(ev)
	}
	return fail(ev)
}

func freebsdScreenLock(ctx context.Context) Result {
	if CommandExists("xscreensaver-command") {
		out := RunCommand(ctx, "xscreensaver-command", "-version")
		if out.Err == nil {
			return pass(map[string]any{"raw": out.Stdout})
		}
	}
	return notApplicable(
		map[string]any{
			"note": "FreeBSD does not have a unified screen lock policy",
		},
	)
}

func freebsdFirewall(ctx context.Context) Result {
	if !CommandExists("pfctl") {
		return unknown(map[string]any{"note": "pfctl not found"})
	}
	out := RunCommand(ctx, "pfctl", "-si")
	ev := map[string]any{"raw": truncate(out.Stdout, 400)}
	if out.Err != nil {
		return unknown(ev)
	}
	if strings.Contains(out.Stdout, "Status: Enabled") {
		return pass(ev)
	}
	return fail(ev)
}

func freebsdTimeSync(ctx context.Context) Result {
	out := RunCommand(ctx, "service", "ntpd", "status")
	ev := map[string]any{"raw": out.Stdout, "stderr": out.Stderr}
	if out.Err != nil {
		return fail(ev)
	}
	if strings.Contains(strings.ToLower(out.Stdout), "is running") {
		return pass(ev)
	}
	return fail(ev)
}

func freebsdOSVersion(ctx context.Context) Result {
	out := RunCommand(ctx, "uname", "-r")
	if out.Err != nil {
		return unknown(map[string]any{"error": out.Err.Error()})
	}
	return pass(map[string]any{"release": out.Stdout})
}

func freebsdAutoUpdate(ctx context.Context) Result {
	return notApplicable(
		map[string]any{
			"note": "FreeBSD relies on operator-driven freebsd-update",
		},
	)
}

func freebsdPasswordPolicy(ctx context.Context) Result {
	data, err := os.ReadFile("/etc/login.conf")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}
	body := string(data)
	hasPolicy := strings.Contains(body, "minpasswordlen=") ||
		strings.Contains(body, "passwordtime=")
	ev := map[string]any{
		"login_conf_snippet": truncate(body, 400),
	}
	if hasPolicy {
		return pass(ev)
	}
	return fail(ev)
}

func freebsdMalwareProtection(ctx context.Context) Result {
	if !CommandExists("clamd") && !CommandExists("clamdscan") {
		return notApplicable(
			map[string]any{
				"note": "clamav not installed",
			},
		)
	}
	out := RunCommand(ctx, "service", "clamav_clamd", "status")
	ev := map[string]any{"raw": out.Stdout, "stderr": out.Stderr}
	if out.Err != nil {
		return unknown(ev)
	}
	if strings.Contains(strings.ToLower(out.Stdout), "is running") {
		return pass(ev)
	}
	return fail(ev)
}

func freebsdRemoteLogin(ctx context.Context) Result {
	out := RunCommand(ctx, "service", "sshd", "status")
	ev := map[string]any{"raw": out.Stdout, "stderr": out.Stderr}
	if out.Err != nil {
		return unknown(ev)
	}
	if strings.Contains(strings.ToLower(out.Stdout), "is running") {
		return fail(ev)
	}
	return pass(ev)
}
