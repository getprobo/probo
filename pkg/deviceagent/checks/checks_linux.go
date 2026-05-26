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
	"strconv"
	"strings"
)

func init() {
	Register(KeyDiskEncryption, linuxDiskEncryption)
	Register(KeyScreenLock, linuxScreenLock)
	Register(KeyFirewallEnabled, linuxFirewall)
	Register(KeyTimeSync, linuxTimeSync)
	Register(KeyOSVersion, linuxOSVersion)
	Register(KeyAutoUpdate, linuxAutoUpdate)
	Register(KeyPasswordPolicy, linuxPasswordPolicy)
	Register(KeyRemoteLogin, linuxRemoteLogin)
	Register(KeyMalwareProtection, linuxMalwareProtection)
}

func linuxDiskEncryption(ctx context.Context) Result {
	ev := map[string]any{}

	if data, err := os.ReadFile("/etc/crypttab"); err == nil {
		body := strings.TrimSpace(string(data))
		ev["crypttab_present"] = true
		ev["crypttab_lines"] = nonCommentLines(body)
		if len(nonCommentLines(body)) > 0 {
			return pass(ev)
		}
	} else {
		ev["crypttab_present"] = false
	}

	lsblk := RunCommand(ctx, "lsblk", "-o", "NAME,TYPE,FSTYPE,MOUNTPOINT", "-r")
	if lsblk.Err == nil {
		ev["lsblk"] = truncate(lsblk.Stdout, 800)
		lines := strings.SplitSeq(lsblk.Stdout, "\n")
		for line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			if fields[1] == "crypt" {
				return pass(ev)
			}
		}
	} else {
		ev["lsblk_error"] = lsblk.Err.Error()
	}

	if lsblk.Err != nil {
		return unknown(ev)
	}
	return fail(ev)
}

func linuxScreenLock(ctx context.Context) Result {
	if !CommandExists("gsettings") {
		return notApplicable(
			map[string]any{
				"note": "gsettings not installed (likely headless host)",
			},
		)
	}
	idle := RunCommand(ctx, "gsettings", "get", "org.gnome.desktop.screensaver", "lock-enabled")
	if idle.Err != nil {
		return unknown(
			map[string]any{
				"error": idle.Err.Error(),
			},
		)
	}
	on := strings.TrimSpace(idle.Stdout) == "true"
	ev := map[string]any{"lock_enabled": idle.Stdout}
	if on {
		return pass(ev)
	}
	return fail(ev)
}

func linuxFirewall(ctx context.Context) Result {
	if CommandExists("ufw") {
		out := RunCommand(ctx, "ufw", "status")
		if out.Err == nil {
			active := strings.Contains(strings.ToLower(out.Stdout), "status: active")
			ev := map[string]any{"backend": "ufw", "raw": out.Stdout}
			if active {
				return pass(ev)
			}
			return fail(ev)
		}
	}
	if CommandExists("firewall-cmd") {
		out := RunCommand(ctx, "firewall-cmd", "--state")
		ev := map[string]any{"backend": "firewalld", "raw": out.Stdout}
		if out.Err == nil && strings.Contains(strings.ToLower(out.Stdout), "running") {
			return pass(ev)
		}
		return fail(ev)
	}

	if CommandExists("nft") {
		out := RunCommand(ctx, "nft", "list", "ruleset")
		ev := map[string]any{
			"backend":       "nftables",
			"rules_excerpt": truncate(out.Stdout, 400),
		}
		if out.Err != nil {
			ev["error"] = out.Err.Error()
			return unknown(ev)
		}
		if strings.Contains(out.Stdout, "chain ") {
			return pass(ev)
		}
		return fail(ev)
	}

	if CommandExists("iptables") {
		out := RunCommand(ctx, "iptables", "-S", "INPUT")
		ev := map[string]any{"backend": "iptables"}
		if out.Err != nil {
			ev["error"] = out.Err.Error()
			return unknown(ev)
		}
		policy, rules := parseIptablesInput(out.Stdout)
		ev["input_policy"] = policy
		ev["input_rules"] = rules
		if policy == "DROP" || policy == "REJECT" {
			return pass(ev)
		}

		if rules == 0 {
			return fail(ev)
		}
		// ACCEPT policy with some rules means the operator is filtering,
		// but we cannot tell from -S whether the rules are restrictive
		// or permissive without modelling the chain.
		return unknown(ev)
	}
	return unknown(
		map[string]any{
			"note": "no known firewall tool found",
		},
	)
}

// parseIptablesInput extracts the INPUT chain policy and rule count from
// `iptables -S INPUT` output.
func parseIptablesInput(s string) (string, int) {
	var (
		policy string
		rules  int
	)

	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "-P INPUT"):
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				policy = strings.ToUpper(fields[2])
			}
		case strings.HasPrefix(line, "-A INPUT"):
			rules++
		}
	}

	return policy, rules
}

func linuxTimeSync(ctx context.Context) Result {
	if !CommandExists("timedatectl") {
		return unknown(
			map[string]any{
				"note": "timedatectl not installed",
			},
		)
	}
	out := RunCommand(ctx, "timedatectl", "show")
	if out.Err != nil {
		return unknown(map[string]any{"error": out.Err.Error()})
	}
	ev := map[string]any{"raw": truncate(out.Stdout, 400)}
	if strings.Contains(out.Stdout, "NTPSynchronized=yes") {
		return pass(ev)
	}
	return fail(ev)
}

func linuxOSVersion(ctx context.Context) Result {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}
	body := string(data)
	ev := map[string]any{
		"pretty_name": kvLookup(body, "PRETTY_NAME"),
		"version_id":  kvLookup(body, "VERSION_ID"),
		"id":          kvLookup(body, "ID"),
	}
	return pass(ev)
}

func linuxAutoUpdate(ctx context.Context) Result {
	if _, err := os.Stat("/etc/apt/apt.conf.d/20auto-upgrades"); err == nil {
		data, _ := os.ReadFile("/etc/apt/apt.conf.d/20auto-upgrades")
		body := string(data)
		ev := map[string]any{
			"backend": "unattended-upgrades",
			"raw":     body,
		}
		if strings.Contains(body, `"1"`) {
			return pass(ev)
		}
		return fail(ev)
	}
	if CommandExists("systemctl") {
		out := RunCommand(ctx, "systemctl", "is-enabled", "dnf-automatic.timer")
		if out.Err == nil {
			ev := map[string]any{"backend": "dnf-automatic", "state": out.Stdout}
			if strings.TrimSpace(out.Stdout) == "enabled" {
				return pass(ev)
			}
			return fail(ev)
		}
	}
	return notApplicable(
		map[string]any{
			"note": "no known auto-update mechanism",
		},
	)
}

func linuxPasswordPolicy(ctx context.Context) Result {
	data, err := os.ReadFile("/etc/login.defs")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}
	body := string(data)
	minLen := loginDefsLookup(body, "PASS_MIN_LEN")
	maxDays := loginDefsLookup(body, "PASS_MAX_DAYS")
	ev := map[string]any{
		"pass_min_len":  minLen,
		"pass_max_days": maxDays,
	}
	if minLen == "" {
		ev["parse_error"] = "PASS_MIN_LEN not set"
		return fail(ev)
	}

	minLenValue, err := strconv.Atoi(minLen)
	if err != nil {
		ev["parse_error"] = "invalid PASS_MIN_LEN value"
		return unknown(ev)
	}

	if minLenValue >= 8 {
		ev["pass_min_len_value"] = minLenValue
		return pass(ev)
	}

	ev["pass_min_len_value"] = minLenValue

	return fail(ev)
}

func linuxRemoteLogin(ctx context.Context) Result {
	if !CommandExists("systemctl") {
		return unknown(map[string]any{"note": "systemctl unavailable"})
	}
	state := RunCommand(ctx, "systemctl", "is-active", "ssh.service")
	stateAlt := RunCommand(ctx, "systemctl", "is-active", "sshd.service")
	merged := strings.TrimSpace(state.Stdout)
	if merged == "" {
		merged = strings.TrimSpace(stateAlt.Stdout)
	}
	ev := map[string]any{"is_active": merged}
	switch merged {
	case "active":
		return fail(ev)
	case "inactive", "failed":
		return pass(ev)
	case "":
		return notApplicable(ev)
	}
	return unknown(ev)
}

// linuxMalwareProtection tracks AV/EDR agent services, not MAC frameworks.
func linuxMalwareProtection(ctx context.Context) Result {
	candidates := []struct {
		unit string
		name string
	}{
		{"clamav-daemon.service", "ClamAV"},
		{"clamd.service", "ClamAV"},
		{"clamd@scan.service", "ClamAV"},
		{"falcon-sensor.service", "CrowdStrike Falcon"},
		{"sentinelone.service", "SentinelOne"},
		{"sentineld.service", "SentinelOne"},
		{"sav-protect.service", "Sophos"},
		{"sophos-spl.service", "Sophos"},
		{"esets.service", "ESET"},
		{"mdatp.service", "Microsoft Defender for Endpoint"},
		{"wazuh-agent.service", "Wazuh"},
		{"ossec.service", "OSSEC"},
		{"elastic-agent.service", "Elastic Agent"},
		{"osqueryd.service", "osquery"},
	}

	if !CommandExists("systemctl") {
		return unknown(
			map[string]any{
				"note": "systemctl not available; cannot enumerate endpoint agents",
			},
		)
	}

	var active, installed []string
	for _, c := range candidates {
		state := strings.TrimSpace(
			RunCommand(ctx, "systemctl", "is-active", c.unit).Stdout)
		switch state {
		case "active":
			active = append(active, c.name)
		case "inactive", "failed", "activating", "deactivating":
			installed = append(installed, c.name)
		}
	}

	ev := map[string]any{
		"active":    active,
		"installed": installed,
	}
	if len(active) > 0 {
		return pass(ev)
	}
	if len(installed) > 0 {
		return fail(ev)
	}
	return unknown(ev)
}

func nonCommentLines(s string) []string {
	out := []string{}
	for line := range strings.SplitSeq(s, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		out = append(out, t)
	}
	return out
}

func kvLookup(body, key string) string {
	for line := range strings.SplitSeq(body, "\n") {
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		if strings.TrimSpace(line[:eq]) == key {
			v := strings.TrimSpace(line[eq+1:])
			v = strings.Trim(v, `"`)
			return v
		}
	}
	return ""
}

func loginDefsLookup(body, key string) string {
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == key {
			return fields[1]
		}
	}
	return ""
}
