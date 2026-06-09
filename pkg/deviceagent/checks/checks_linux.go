// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"path/filepath"
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
	user := linuxConsoleUser(ctx)
	desktop := linuxDesktopSession()

	probes := linuxScreenLockProbes(desktop)

	var (
		anyTool    bool
		lastResult Result
	)

	for _, probe := range probes {
		result, tried := probe(ctx, user)
		if !tried {
			continue
		}

		anyTool = true

		switch result.Status {
		case StatusPass, StatusFail:
			return result
		case StatusUnknown:
			lastResult = result
		}
	}

	if !anyTool {
		return notApplicable(
			map[string]any{
				"note": "no known desktop screen lock tool found (likely headless host)",
			},
		)
	}

	if lastResult.Status == StatusUnknown {
		return lastResult
	}

	return unknown(
		map[string]any{
			"note": "desktop screen lock tools present but policy could not be read",
		},
	)
}

var linuxGsettingsLockSchemas = []struct {
	schema  string
	backend string
}{
	{"org.gnome.desktop.screensaver", "gnome"},
	{"org.cinnamon.desktop.screensaver", "cinnamon"},
	{"org.mate.screensaver", "mate"},
	{"org.ukui.screensaver", "ukui"},
}

func linuxScreenLockProbes(desktop string) []func(context.Context, string) (Result, bool) {
	gsettings := linuxScreenLockGsettings
	kde := linuxScreenLockKDE
	xfce := linuxScreenLockXFCE
	i3 := linuxScreenLockI3

	switch {
	case linuxDesktopPrefersKDE(desktop):
		return []func(context.Context, string) (Result, bool){kde, i3, gsettings, xfce}
	case linuxDesktopPrefersXFCE(desktop):
		return []func(context.Context, string) (Result, bool){xfce, i3, gsettings, kde}
	case linuxDesktopPrefersI3(desktop):
		return []func(context.Context, string) (Result, bool){i3, gsettings, kde, xfce}
	default:
		return []func(context.Context, string) (Result, bool){gsettings, i3, kde, xfce}
	}
}

func linuxScreenLockGsettings(ctx context.Context, user string) (Result, bool) {
	if !CommandExists("gsettings") {
		return Result{}, false
	}

	for _, schema := range linuxOrderGsettingsSchemas(linuxDesktopSession()) {
		out := linuxRunAsUser(
			ctx,
			user,
			"gsettings",
			"get",
			schema.schema,
			"lock-enabled",
		)
		if out.Err != nil {
			continue
		}

		combined := strings.ToLower(out.Stderr + "\n" + out.Stdout)
		if strings.Contains(combined, "no such key") ||
			strings.Contains(combined, "no such schema") {
			continue
		}

		val := strings.TrimSpace(out.Stdout)

		ev := map[string]any{
			"backend":      schema.backend,
			"schema":       schema.schema,
			"lock_enabled": val,
		}
		if user != "" {
			ev["console_user"] = user
		}

		if val == "true" {
			return pass(ev), true
		}

		return fail(ev), true
	}

	return Result{}, false
}

func linuxScreenLockKDE(ctx context.Context, user string) (Result, bool) {
	cmd := linuxKReadConfigCommand()
	if cmd == "" {
		return Result{}, false
	}

	out := linuxRunAsUser(
		ctx,
		user,
		cmd,
		"--file",
		"kscreenlockerrc",
		"--group",
		"Daemon",
		"--key",
		"Autolock",
	)
	if out.Err != nil {
		return unknown(
			map[string]any{
				"backend": "kde",
				"error":   out.Err.Error(),
			},
		), true
	}

	val := strings.TrimSpace(out.Stdout)
	if val == "" {
		return Result{}, false
	}

	ev := map[string]any{
		"backend":  "kde",
		"autolock": val,
	}
	if user != "" {
		ev["console_user"] = user
	}

	if strings.EqualFold(val, "true") {
		return pass(ev), true
	}

	return fail(ev), true
}

func linuxScreenLockXFCE(ctx context.Context, user string) (Result, bool) {
	if !CommandExists("xfconf-query") {
		return Result{}, false
	}

	paths := []string{"/lock/enabled", "/saver/enabled"}
	for _, path := range paths {
		out := linuxRunAsUser(
			ctx,
			user,
			"xfconf-query",
			"-c",
			"xfce4-screensaver",
			"-p",
			path,
		)
		if out.Err != nil {
			continue
		}

		val := strings.TrimSpace(out.Stdout)
		if val == "" {
			continue
		}

		ev := map[string]any{
			"backend": "xfce",
			"path":    path,
			"enabled": val,
		}
		if user != "" {
			ev["console_user"] = user
		}

		switch strings.ToLower(val) {
		case "true", "1", "yes":
			return pass(ev), true
		case "false", "0", "no":
			return fail(ev), true
		}
	}

	return Result{}, false
}

func linuxScreenLockI3(ctx context.Context, user string) (Result, bool) {
	configPath := linuxI3ConfigPath(ctx, user)
	if configPath == "" {
		return Result{}, false
	}

	if _, err := os.Stat(configPath); err != nil {
		return Result{}, false
	}

	body := linuxReadI3Config(ctx, user, configPath, 0)
	if body == "" {
		return unknown(
			map[string]any{
				"backend": "i3",
				"config":  configPath,
				"error":   "cannot read i3 config",
			},
		), true
	}

	idleMinutes, locker, mechanism, ok := parseI3IdleLock(body)

	ev := map[string]any{
		"backend": "i3",
		"config":  configPath,
	}
	if user != "" {
		ev["console_user"] = user
	}

	if !ok {
		ev["note"] = "i3 config present but no idle screen lock command found"
		return fail(ev), true
	}

	ev["mechanism"] = mechanism
	ev["locker"] = locker

	if idleMinutes >= 0 {
		ev["idle_minutes"] = idleMinutes
	}

	return pass(ev), true
}

func linuxI3ConfigPath(ctx context.Context, user string) string {
	home := linuxUserHome(ctx, user)
	if home == "" {
		return ""
	}

	return filepath.Join(home, ".config", "i3", "config")
}

func linuxUserHome(ctx context.Context, user string) string {
	if user == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}

		return home
	}

	if CommandExists("getent") {
		out := RunCommand(ctx, "getent", "passwd", user)
		if out.Err == nil {
			fields := strings.Split(out.Stdout, ":")
			if len(fields) >= 6 && fields[5] != "" {
				return fields[5]
			}
		}
	}

	return filepath.Join("/home", user)
}

func linuxReadI3Config(ctx context.Context, user, path string, depth int) string {
	if depth > 4 || path == "" {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	body := string(data)

	var merged strings.Builder

	merged.WriteString(body)

	home := linuxUserHome(ctx, user)

	for line := range strings.SplitSeq(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, "include ") {
			continue
		}

		inc := strings.TrimSpace(trimmed[len("include "):])
		inc = strings.Trim(inc, `"`)

		inc = linuxExpandHome(inc, home)
		if nested := linuxReadI3Config(ctx, user, inc, depth+1); nested != "" {
			merged.WriteString("\n")
			merged.WriteString(nested)
		}
	}

	return merged.String()
}

func linuxExpandHome(path, home string) string {
	if home == "" {
		return path
	}

	switch {
	case strings.HasPrefix(path, "~/"):
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	case path == "~":
		return home
	default:
		return path
	}
}

// parseI3IdleLock scans an i3 config for idle screen lock via xautolock or xss-lock.
func parseI3IdleLock(config string) (idleMinutes int, locker, mechanism string, ok bool) {
	idleMinutes = -1

	for line := range strings.SplitSeq(config, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		lower := strings.ToLower(trimmed)
		if !strings.Contains(lower, "exec") {
			continue
		}

		if minutes, lockCmd, found := parseXautolockIdleLock(trimmed); found {
			return minutes, lockCmd, "xautolock", true
		}

		if lockCmd, found := parseXssLockIdleLock(trimmed); found {
			return -1, lockCmd, "xss-lock", true
		}
	}

	return idleMinutes, "", "", false
}

func parseXautolockIdleLock(line string) (minutes int, locker string, ok bool) {
	if !strings.Contains(strings.ToLower(line), "xautolock") {
		return 0, "", false
	}

	timeValue, hasTime := linuxParseFlagValue(line, "-time")
	if !hasTime {
		return 0, "", false
	}

	minutes, validTime := parseXautolockTime(timeValue)
	if !validTime || minutes <= 0 {
		return 0, "", false
	}

	locker, hasLocker := linuxParseFlagValue(line, "-locker")
	if !hasLocker || !linuxLooksLikeLockCommand(locker) {
		return 0, "", false
	}

	return minutes, locker, true
}

func parseXssLockIdleLock(line string) (locker string, ok bool) {
	if !strings.Contains(strings.ToLower(line), "xss-lock") {
		return "", false
	}

	_, after, ok0 := strings.Cut(line, "--")
	if !ok0 {
		return "", false
	}

	locker = strings.TrimSpace(after)
	locker = strings.Trim(locker, `"`)

	if locker == "" || !linuxLooksLikeLockCommand(locker) {
		return "", false
	}

	return locker, true
}

func parseXautolockTime(value string) (minutes int, ok bool) {
	value = strings.TrimSpace(value)
	if strings.Contains(value, ":") {
		parts := strings.Split(value, ":")
		if len(parts) != 2 {
			return 0, false
		}

		hours, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		mins, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

		if err1 != nil || err2 != nil {
			return 0, false
		}

		return hours*60 + mins, true
	}

	minutes, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}

	return minutes, true
}

func linuxParseFlagValue(line, flag string) (string, bool) {
	_, after, ok := strings.Cut(line, flag)
	if !ok {
		return "", false
	}

	rest := strings.TrimSpace(after)
	if rest == "" {
		return "", false
	}

	if rest[0] == '"' {
		end := strings.Index(rest[1:], `"`)
		if end < 0 {
			return "", false
		}

		return rest[1 : end+1], true
	}

	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return "", false
	}

	return fields[0], true
}

func linuxLooksLikeLockCommand(cmd string) bool {
	lower := strings.ToLower(cmd)

	for _, bin := range []string{
		"i3lock",
		"i3lock-color",
		"swaylock",
		"xlock",
		"slock",
		"gnome-screensaver-command",
		"loginctl",
	} {
		if strings.Contains(lower, bin) {
			return true
		}
	}

	return false
}

func linuxOrderGsettingsSchemas(desktop string) []struct {
	schema  string
	backend string
} {
	ordered := make([]struct {
		schema  string
		backend string
	}, 0, len(linuxGsettingsLockSchemas))
	seen := make(map[string]struct{}, len(linuxGsettingsLockSchemas))

	preferred := ""

	switch {
	case strings.Contains(desktop, "cinnamon"):
		preferred = "cinnamon"
	case strings.Contains(desktop, "mate"):
		preferred = "mate"
	case strings.Contains(desktop, "ukui"):
		preferred = "ukui"
	default:
		preferred = "gnome"
	}

	for _, schema := range linuxGsettingsLockSchemas {
		if schema.backend == preferred {
			ordered = append(ordered, schema)
			seen[schema.schema] = struct{}{}
		}
	}

	for _, schema := range linuxGsettingsLockSchemas {
		if _, ok := seen[schema.schema]; ok {
			continue
		}

		ordered = append(ordered, schema)
	}

	return ordered
}

func linuxKReadConfigCommand() string {
	switch {
	case CommandExists("kreadconfig6"):
		return "kreadconfig6"
	case CommandExists("kreadconfig5"):
		return "kreadconfig5"
	default:
		return ""
	}
}

func linuxDesktopSession() string {
	for _, key := range []string{"XDG_CURRENT_DESKTOP", "DESKTOP_SESSION", "GDMSESSION"} {
		if v := strings.ToLower(strings.TrimSpace(os.Getenv(key))); v != "" {
			return v
		}
	}

	return ""
}

func linuxDesktopPrefersKDE(desktop string) bool {
	desktop = strings.ToLower(desktop)

	return strings.Contains(desktop, "kde") || strings.Contains(desktop, "plasma")
}

func linuxDesktopPrefersXFCE(desktop string) bool {
	return strings.Contains(strings.ToLower(desktop), "xfce")
}

func linuxDesktopPrefersI3(desktop string) bool {
	desktop = strings.ToLower(desktop)

	return strings.Contains(desktop, "i3")
}

func linuxConsoleUser(ctx context.Context) string {
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" && sudoUser != "root" {
		return sudoUser
	}

	if CommandExists("loginctl") {
		seat := RunCommand(ctx, "loginctl", "show-seat", "seat0", "-p", "ActiveSession", "--value")

		sessionID := strings.TrimSpace(seat.Stdout)
		if sessionID != "" {
			name := RunCommand(ctx, "loginctl", "show-session", sessionID, "-p", "Name", "--value")

			user := strings.TrimSpace(name.Stdout)
			if user != "" && user != "root" {
				return user
			}
		}
	}

	out := RunCommand(ctx, "stat", "-c", "%U", "/dev/console")
	if out.Err != nil {
		return ""
	}

	user := strings.TrimSpace(out.Stdout)
	if user == "" || user == "root" {
		return ""
	}

	return user
}

func linuxRunAsUser(ctx context.Context, user, name string, args ...string) CmdResult {
	if user != "" && os.Geteuid() == 0 {
		if CommandExists("runuser") {
			runArgs := append([]string{"-u", user, "--", name}, args...)

			return RunCommand(ctx, "runuser", runArgs...)
		}

		if CommandExists("sudo") {
			runArgs := append([]string{"-u", user, "-H", name}, args...)

			return RunCommand(ctx, "sudo", runArgs...)
		}
	}

	return RunCommand(ctx, name, args...)
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
