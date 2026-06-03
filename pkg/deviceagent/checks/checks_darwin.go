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

package checks

import (
	"context"
	"os"
	"strconv"
	"strings"
)

func init() {
	Register(KeyDiskEncryption, darwinDiskEncryption)
	Register(KeyScreenLock, darwinScreenLock)
	Register(KeyFirewallEnabled, darwinFirewall)
	Register(KeyTimeSync, darwinTimeSync)
	Register(KeyOSVersion, darwinOSVersion)
	Register(KeyAutoUpdate, darwinAutoUpdate)
	Register(KeyPasswordPolicy, darwinPasswordPolicy)
	Register(KeyRemoteLogin, darwinRemoteLogin)
	Register(KeyMalwareProtection, darwinMalwareProtection)
}

func darwinDiskEncryption(ctx context.Context) Result {
	out := RunCommand(ctx, "fdesetup", "status")
	if out.Err != nil {
		return unknown(
			map[string]any{
				"error":  out.Err.Error(),
				"stderr": out.Stderr,
			},
		)
	}

	on := strings.Contains(strings.ToLower(out.Stdout), "filevault is on")

	ev := map[string]any{"raw": out.Stdout}
	if on {
		return pass(ev)
	}

	return fail(ev)
}

func darwinScreenLock(ctx context.Context) Result {
	if CommandExists("sysadminctl") {
		status := RunCommand(ctx, "sysadminctl", "-screenLock", "status", "-password", "-")
		rawCombined := strings.TrimSpace(status.Stdout + "\n" + status.Stderr)
		ev := map[string]any{
			"backend":    "sysadminctl",
			"raw":        rawCombined,
			"raw_stdout": status.Stdout,
			"raw_stderr": status.Stderr,
		}

		mode, seconds, ok := darwinScreenLockMode(rawCombined)
		if ok {
			ev["mode"] = mode
			if mode == "seconds" && seconds >= 0 {
				ev["seconds"] = seconds
			}

			if mode == "immediate" {
				return pass(ev)
			}

			return fail(ev)
		}

		if status.Err != nil {
			ev["error"] = status.Err.Error()
		}
	}

	ask, askSource := darwinReadScreenSaverDefault(ctx, "askForPassword")

	ev := map[string]any{}
	if askSource != "" {
		ev["source"] = askSource
	}

	if ask.Err != nil {
		if darwinDefaultsMissing(ask) {
			ev["ask_for_password"] = "0"
			ev["note"] = "askForPassword is unset or unavailable"

			return fail(ev)
		}

		ev["error"] = ask.Err.Error()
		ev["stderr"] = ask.Stderr

		return unknown(ev)
	}

	enabled := strings.TrimSpace(ask.Stdout) == "1"

	delayCmd, delaySource := darwinReadScreenSaverDefault(ctx, "askForPasswordDelay")

	ev["ask_for_password"] = ask.Stdout
	if delayCmd.Err == nil {
		ev["ask_for_password_delay"] = delayCmd.Stdout
		if delaySource != "" && delaySource != askSource {
			ev["delay_source"] = delaySource
		}
	}

	if enabled {
		return pass(ev)
	}

	return fail(ev)
}

func darwinScreenLockMode(raw string) (string, int, bool) {
	lower := strings.ToLower(raw)
	if strings.Contains(lower, "immediate") {
		return "immediate", 0, true
	}

	if strings.Contains(lower, "off") {
		return "off", -1, true
	}

	if before, _, ok := strings.Cut(lower, "seconds"); ok {
		prefix := strings.Fields(before)
		if len(prefix) == 0 {
			return "seconds", -1, true
		}

		n, err := strconv.Atoi(prefix[len(prefix)-1])
		if err != nil {
			return "seconds", -1, true
		}

		return "seconds", n, true
	}

	return "", 0, false
}

func darwinFirewall(ctx context.Context) Result {
	out := RunCommand(
		ctx,
		"defaults",
		"read",
		"/Library/Preferences/com.apple.alf",
		"globalstate",
	)
	if out.Err == nil {
		state := strings.TrimSpace(out.Stdout)

		ev := map[string]any{"backend": "defaults", "global_state": state}
		if state == "1" || state == "2" {
			return pass(ev)
		}

		return fail(ev)
	}

	fallback := RunCommand(ctx, "/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")

	ev := map[string]any{
		"backend":         "socketfilterfw",
		"raw":             fallback.Stdout,
		"defaults_error":  errString(out.Err),
		"defaults_stderr": out.Stderr,
	}
	if fallback.Err != nil {
		ev["error"] = fallback.Err.Error()
		ev["stderr"] = fallback.Stderr

		return unknown(ev)
	}

	if darwinStateIndicatesEnabled(fallback.Stdout) {
		return pass(ev)
	}

	if darwinStateIndicatesDisabled(fallback.Stdout) {
		return fail(ev)
	}

	return unknown(ev)
}

// darwinReadScreenSaverDefault prefers console-user settings when running as root.
func darwinReadScreenSaverDefault(ctx context.Context, key string) (CmdResult, string) {
	consoleUser := darwinConsoleUser(ctx)
	if os.Geteuid() == 0 && consoleUser != "" {
		var consoleMissing CmdResult

		consoleMissingSource := ""

		if CommandExists("sudo") {
			consoleUserCurrentHost := RunCommand(
				ctx,
				"sudo",
				"-u",
				consoleUser,
				"defaults",
				"-currentHost",
				"read",
				"com.apple.screensaver",
				key,
			)
			if consoleUserCurrentHost.Err == nil {
				return consoleUserCurrentHost, "console_user_current_host:" + consoleUser
			}

			if !darwinDefaultsMissing(consoleUserCurrentHost) {
				return consoleUserCurrentHost, "console_user_current_host:" + consoleUser
			}

			if consoleMissingSource == "" {
				consoleMissing = consoleUserCurrentHost
				consoleMissingSource = "console_user_current_host:" + consoleUser
			}

			consoleUserDomain := RunCommand(
				ctx,
				"sudo",
				"-u",
				consoleUser,
				"defaults",
				"read",
				"com.apple.screensaver",
				key,
			)
			if consoleUserDomain.Err == nil {
				return consoleUserDomain, "console_user:" + consoleUser
			}

			if !darwinDefaultsMissing(consoleUserDomain) {
				return consoleUserDomain, "console_user:" + consoleUser
			}

			if consoleMissingSource == "" {
				consoleMissing = consoleUserDomain
				consoleMissingSource = "console_user:" + consoleUser
			}
		}

		plistPath := "/Users/" + consoleUser + "/Library/Preferences/com.apple.screensaver.plist"

		consoleUserOut := RunCommand(ctx, "defaults", "read", plistPath, key)
		if consoleUserOut.Err == nil {
			return consoleUserOut, "console_user_plist:" + consoleUser
		}

		if !darwinDefaultsMissing(consoleUserOut) {
			return consoleUserOut, "console_user_plist:" + consoleUser
		}

		if consoleMissingSource == "" {
			consoleMissing = consoleUserOut
			consoleMissingSource = "console_user_plist:" + consoleUser
		}

		if consoleMissingSource != "" {
			return consoleMissing, consoleMissingSource
		}
	}

	currentHost := RunCommand(ctx, "defaults", "-currentHost", "read", "com.apple.screensaver", key)
	if currentHost.Err == nil {
		return currentHost, "current_user_current_host"
	}

	currentUser := RunCommand(ctx, "defaults", "read", "com.apple.screensaver", key)
	if currentUser.Err == nil {
		return currentUser, "current_user"
	}

	if !darwinDefaultsMissing(currentUser) {
		return currentUser, "current_user"
	}

	if !darwinDefaultsMissing(currentHost) {
		return currentHost, "current_user_current_host"
	}

	return currentUser, "current_user"
}

func darwinDefaultsMissing(out CmdResult) bool {
	lower := strings.ToLower(out.Stderr + "\n" + out.Stdout)

	return strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "could not find") ||
		strings.Contains(lower, "does not exist in domain")
}

func darwinConsoleUser(ctx context.Context) string {
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" && sudoUser != "root" {
		return sudoUser
	}

	out := RunCommand(ctx, "stat", "-f", "%Su", "/dev/console")
	if out.Err != nil {
		return ""
	}

	user := strings.TrimSpace(out.Stdout)
	if user == "" || user == "root" || user == "loginwindow" {
		return ""
	}

	return user
}

func darwinStateIndicatesEnabled(raw string) bool {
	lower := strings.ToLower(raw)

	return strings.Contains(lower, "enabled") ||
		strings.Contains(lower, "state = 1") ||
		strings.Contains(lower, "state = 2")
}

func darwinStateIndicatesDisabled(raw string) bool {
	lower := strings.ToLower(raw)
	return strings.Contains(lower, "disabled") || strings.Contains(lower, "state = 0")
}

func darwinTimeSync(ctx context.Context) Result {
	out := RunCommand(ctx, "systemsetup", "-getusingnetworktime")
	if out.Err != nil || needsAdmin(out.Stdout) {
		return unknown(
			map[string]any{
				"raw":   out.Stdout,
				"error": errString(out.Err),
			},
		)
	}

	on := strings.Contains(strings.ToLower(out.Stdout), "on")

	ev := map[string]any{"raw": out.Stdout}
	if on {
		return pass(ev)
	}

	return fail(ev)
}

func darwinOSVersion(ctx context.Context) Result {
	out := RunCommand(ctx, "sw_vers", "-productVersion")
	if out.Err != nil || out.Stdout == "" {
		return unknown(map[string]any{"error": "sw_vers failed"})
	}

	build := RunCommand(ctx, "sw_vers", "-buildVersion")
	ev := map[string]any{
		"product_version": out.Stdout,
		"build_version":   build.Stdout,
	}

	return pass(ev)
}

func darwinAutoUpdate(ctx context.Context) Result {
	primary := RunCommand(
		ctx,
		"defaults",
		"read",
		"/Library/Preferences/com.apple.SoftwareUpdate",
		"AutomaticCheckEnabled",
	)
	if primary.Err == nil {
		ev := map[string]any{
			"backend":                 "defaults",
			"automatic_check_enabled": primary.Stdout,
		}
		if strings.TrimSpace(primary.Stdout) == "1" {
			return pass(ev)
		}

		return fail(ev)
	}

	fallback := RunCommand(ctx, "softwareupdate", "--schedule")

	ev := map[string]any{
		"backend":         "softwareupdate",
		"raw":             fallback.Stdout,
		"defaults_error":  errString(primary.Err),
		"defaults_stderr": primary.Stderr,
	}
	if fallback.Err != nil ||
		needsAdmin(fallback.Stdout) ||
		needsAdmin(fallback.Stderr) {
		ev["error"] = errString(fallback.Err)
		ev["stderr"] = fallback.Stderr

		return unknown(ev)
	}

	lower := strings.ToLower(fallback.Stdout)
	switch {
	case strings.Contains(lower, "is turned on"),
		strings.Contains(lower, "automatic check is on"):
		return pass(ev)
	case strings.Contains(lower, "is turned off"),
		strings.Contains(lower, "automatic check is off"):
		return fail(ev)
	}

	return unknown(ev)
}

func darwinPasswordPolicy(ctx context.Context) Result {
	out := RunCommand(ctx, "pwpolicy", "-getaccountpolicies")
	if out.Err != nil {
		return unknown(
			map[string]any{
				"error":  out.Err.Error(),
				"stderr": out.Stderr,
			},
		)
	}

	lower := strings.ToLower(out.Stdout)

	ev := map[string]any{"raw_truncated": truncate(out.Stdout, 400)}
	if strings.Contains(lower, "no account policies") || lower == "" {
		return fail(ev)
	}

	return pass(ev)
}

func darwinRemoteLogin(ctx context.Context) Result {
	out := RunCommand(ctx, "systemsetup", "-getremotelogin")
	if out.Err != nil || needsAdmin(out.Stdout) {
		return unknown(
			map[string]any{
				"raw":   out.Stdout,
				"error": errString(out.Err),
			},
		)
	}

	off := strings.Contains(strings.ToLower(out.Stdout), "off")

	ev := map[string]any{"raw": out.Stdout}
	if off {
		return pass(ev)
	}

	return fail(ev)
}

func darwinMalwareProtection(ctx context.Context) Result {
	candidates := []string{
		"/Library/Apple/System/Library/CoreServices/XProtect.bundle/Contents/Resources/XProtect.meta.plist",
		"/System/Library/CoreServices/XProtect.bundle/Contents/Resources/XProtect.meta.plist",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}

		ev := map[string]any{"engine": "XProtect", "plist": path}

		version := RunCommand(
			ctx,
			"defaults",
			"read",
			strings.TrimSuffix(path, ".plist"),
			"Version",
		)
		if version.Err == nil {
			ev["version"] = version.Stdout
		}

		return pass(ev)
	}

	return fail(
		map[string]any{
			"engine": "XProtect",
			"note":   "XProtect.meta.plist not found in expected locations",
		},
	)
}

// needsAdmin checks systemsetup's stdout for privilege errors.
func needsAdmin(stdout string) bool {
	return strings.Contains(strings.ToLower(stdout), "administrator access")
}
