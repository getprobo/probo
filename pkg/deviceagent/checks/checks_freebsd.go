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
	Register(KeyLoginPassword, freebsdLoginPassword)
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

// freebsdPasswordPolicy reads minpasswordlen from the default login class,
// which pam_unix enforces, and pam_passwdqc when /etc/pam.d/passwd enables it
// (it ships commented out).
func freebsdPasswordPolicy(ctx context.Context) Result {
	data, err := os.ReadFile("/etc/login.conf")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}

	lengths := passwordLengths{}
	if n, ok := parseLoginConfMinPasswordLen(string(data)); ok {
		lengths[passwordSourceLoginConf] = n
	}

	if pam, err := os.ReadFile("/etc/pam.d/passwd"); err == nil {
		for source, n := range pamPasswordLengths(parsePAMPasswordRules(string(pam)), "") {
			// pam_unix reads its minimum from login.conf, already counted.
			if source != passwordSourcePAMUnix {
				lengths[source] = n
			}
		}
	}

	ev, length := passwordPolicyEvidence(lengths)

	return passwordPolicyResult(ev, length)
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

// freebsdLoginPassword reports console auto-login through gettytab al=,
// display managers installed from ports, and master.passwd accounts with an
// empty password field. Hashes never leave the host.
func freebsdLoginPassword(ctx context.Context) Result {
	master, err := os.ReadFile("/etc/master.passwd")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}

	gettytab, _ := os.ReadFile("/etc/gettytab")
	ttys, _ := os.ReadFile("/etc/ttys")

	sources := displayManagerAutoLogin("/usr/local/")
	if freebsdGettyAutoLogin(string(gettytab), string(ttys)) {
		sources = append(sources, autoLoginSourceGetty)
	}

	empty := countFreeBSDAccountsWithoutPassword(string(master))
	ev := loginPasswordEvidence(sources, empty)

	return loginPasswordResult(ev, sources, empty)
}
