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
	"strconv"
	"strings"
)

func init() {
	Register(KeyDiskEncryption, windowsDiskEncryption)
	Register(KeyScreenLock, windowsScreenLock)
	Register(KeyFirewallEnabled, windowsFirewall)
	Register(KeyTimeSync, windowsTimeSync)
	Register(KeyOSVersion, windowsOSVersion)
	Register(KeyAutoUpdate, windowsAutoUpdate)
	Register(KeyPasswordPolicy, windowsPasswordPolicy)
	Register(KeyRemoteLogin, windowsRemoteLogin)
	Register(KeyMalwareProtection, windowsMalwareProtection)
}

const psNoProfile = "-NoProfile"

func powershell(ctx context.Context, script string) CmdResult {
	return RunCommand(ctx, "powershell.exe", psNoProfile, "-Command", script)
}

func windowsDiskEncryption(ctx context.Context) Result {
	out := powershell(
		ctx,
		`(Get-BitLockerVolume | Where-Object { $_.VolumeType -eq 'OperatingSystem' } | `+
			`Sort-Object MountPoint | `+
			`ForEach-Object { "$($_.MountPoint)=$($_.ProtectionStatus)" }) -join ";"`,
	)

	ev := map[string]any{"backend": "Get-BitLockerVolume"}
	if out.Err != nil {
		lower := strings.ToLower(out.Stderr + " " + errString(out.Err))
		if strings.Contains(lower, "not recognized") ||
			strings.Contains(lower, "not found") {
			ev["note"] = "Get-BitLockerVolume not available"

			return fail(ev)
		}

		ev["error"] = out.Err.Error()
		ev["stderr"] = out.Stderr

		return unknown(ev)
	}

	volumes, allProtected := parseWindowsBitLockerVolumes(out.Stdout)

	ev["volumes"] = volumes
	if len(volumes) == 0 {
		ev["note"] = "Get-BitLockerVolume not available"

		return fail(ev)
	}

	if allProtected {
		return pass(ev)
	}

	return fail(ev)
}

func windowsScreenLock(ctx context.Context) Result {
	// HKCU resolves to the SYSTEM hive when the agent runs as LocalSystem,
	// so we first look for a machine-wide policy and then enumerate every
	// loaded interactive user hive under HKU.
	machine := powershell(
		ctx,
		`(Get-ItemProperty 'HKLM:\SOFTWARE\Policies\Microsoft\Windows\Control Panel\Desktop' `+
			`-ErrorAction SilentlyContinue).ScreenSaverIsSecure`,
	)
	if machine.Err == nil {
		v := strings.TrimSpace(machine.Stdout)
		if v != "" {
			ev := map[string]any{
				"backend":                "machine_policy",
				"screen_saver_is_secure": v,
			}
			if v == "1" {
				return pass(ev)
			}

			return fail(ev)
		}
	}

	users := powershell(
		ctx,
		`Get-ChildItem 'Registry::HKEY_USERS' | `+
			`Where-Object { $_.PSChildName -match '^S-1-5-21-' } | `+
			`ForEach-Object { `+
			`  $path = "Registry::HKEY_USERS\$($_.PSChildName)\Control Panel\Desktop"; `+
			`  $key = Get-ItemProperty $path -ErrorAction SilentlyContinue; `+
			`  "$($_.PSChildName)=$($key.ScreenSaverIsSecure)" `+
			`}`,
	)
	if users.Err != nil {
		return unknown(
			map[string]any{
				"backend":              "hkey_users",
				"error":                users.Err.Error(),
				"stderr":               users.Stderr,
				"machine_policy_error": errString(machine.Err),
			},
		)
	}

	ev := map[string]any{
		"backend": "hkey_users",
		"raw":     truncate(users.Stdout, 400),
	}
	users_, anyDisabled, anyEnabled := parseWindowsUserScreenLock(users.Stdout)

	ev["users"] = users_
	if len(users_) == 0 {
		ev["note"] = "no interactive user hives loaded"
		return unknown(ev)
	}

	if anyEnabled && !anyDisabled {
		return pass(ev)
	}

	return fail(ev)
}

// parseWindowsUserScreenLock parses one "SID=<value>" line per user from
// the registry enumeration and reports whether each user has screen
// saver locking enabled.
func parseWindowsUserScreenLock(s string) (map[string]string, bool, bool) {
	users := map[string]string{}

	var anyEnabled, anyDisabled bool

	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		idx := strings.LastIndex(line, "=")
		if idx < 0 {
			continue
		}

		sid := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		if sid == "" {
			continue
		}

		users[sid] = value
		switch value {
		case "1":
			anyEnabled = true
		default:
			anyDisabled = true
		}
	}

	return users, anyDisabled, anyEnabled
}

func windowsFirewall(ctx context.Context) Result {
	primary := powershell(
		ctx,
		`(Get-NetFirewallProfile -PolicyStore ActiveStore | `+
			`Sort-Object Name | `+
			`ForEach-Object { "$($_.Name)=$($_.Enabled)" }) -join ";"`,
	)
	if primary.Err == nil && strings.TrimSpace(primary.Stdout) != "" {
		ev := map[string]any{
			"backend": "Get-NetFirewallProfile",
			"raw":     primary.Stdout,
		}
		profiles, allEnabled := parseWindowsFirewallProfiles(primary.Stdout)

		ev["profiles"] = profiles
		if allEnabled {
			return pass(ev)
		}

		return fail(ev)
	}

	fallback := RunCommand(ctx, "netsh", "advfirewall", "show", "allprofiles", "state")
	if fallback.Err != nil {
		return unknown(
			map[string]any{
				"error":            errString(fallback.Err),
				"stderr":           fallback.Stderr,
				"powershell_error": errString(primary.Err),
			},
		)
	}

	ev := map[string]any{
		"backend": "netsh",
		"raw":     truncate(fallback.Stdout, 600),
	}
	stateLines, anyOff := parseNetshFirewallStates(fallback.Stdout)

	ev["state_lines"] = stateLines
	if len(stateLines) > 0 && !anyOff {
		return pass(ev)
	}

	return fail(ev)
}

// parseNetshFirewallStates extracts per-profile "State <ON|OFF>" lines
// from `netsh advfirewall show allprofiles state`. It is whitespace- and
// case-insensitive.
func parseNetshFirewallStates(s string) ([]string, bool) {
	var states []string

	anyOff := false

	for line := range strings.SplitSeq(s, "\n") {
		trimmed := strings.TrimSpace(line)

		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, "state") {
			continue
		}

		fields := strings.Fields(lower)
		if len(fields) < 2 {
			continue
		}

		value := fields[len(fields)-1]

		states = append(states, value)
		if value != "on" {
			anyOff = true
		}
	}

	return states, anyOff
}

func windowsTimeSync(ctx context.Context) Result {
	out := powershell(
		ctx,
		`$s = Get-Service w32time; `+
			`$p = Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\W32Time\Parameters'; `+
			`"$($s.Status);$($p.Type);$($p.NtpServer)"`,
	)
	if out.Err != nil {
		return unknown(
			map[string]any{
				"error":  out.Err.Error(),
				"stderr": out.Stderr,
			},
		)
	}

	parts := strings.SplitN(strings.TrimSpace(out.Stdout), ";", 3)

	var status, typ, ntpServer string
	if len(parts) >= 1 {
		status = strings.TrimSpace(parts[0])
	}

	if len(parts) >= 2 {
		typ = strings.TrimSpace(parts[1])
	}

	if len(parts) >= 3 {
		ntpServer = strings.TrimSpace(parts[2])
	}

	ev := map[string]any{
		"backend":        "w32time",
		"w32time_status": status,
		"w32time_type":   typ,
	}
	if ntpServer != "" {
		ev["ntp_server"] = ntpServer
	}

	if windowsTimeSyncOn(status, typ) {
		return pass(ev)
	}

	if status == "" {
		return unknown(ev)
	}

	return fail(ev)
}

func windowsOSVersion(ctx context.Context) Result {
	out := powershell(ctx, `(Get-CimInstance Win32_OperatingSystem).Version`)
	if out.Err != nil {
		return unknown(map[string]any{"error": out.Err.Error()})
	}

	caption := powershell(ctx, `(Get-CimInstance Win32_OperatingSystem).Caption`)

	return pass(
		map[string]any{
			"version": out.Stdout,
			"caption": caption.Stdout,
		},
	)
}

func windowsAutoUpdate(ctx context.Context) Result {
	out := powershell(
		ctx,
		`$au = Get-ItemProperty 'HKLM:\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU' `+
			`-ErrorAction SilentlyContinue; `+
			`"$($au.NoAutoUpdate);$($au.AUOptions)"`,
	)

	ev := map[string]any{}
	if out.Err != nil {
		ev["error"] = out.Err.Error()
		ev["stderr"] = out.Stderr

		return unknown(ev)
	}

	parts := strings.SplitN(strings.TrimSpace(out.Stdout), ";", 2)

	var noAutoUpdate, auOptions string
	if len(parts) >= 1 {
		noAutoUpdate = strings.TrimSpace(parts[0])
	}

	if len(parts) >= 2 {
		auOptions = strings.TrimSpace(parts[1])
	}

	ev["no_auto_update"] = noAutoUpdate
	ev["au_options"] = auOptions

	// NoAutoUpdate=1 explicitly disables automatic updates via policy.
	if noAutoUpdate == "1" {
		return fail(ev)
	}

	// AUOptions semantics:
	//   2 — notify before download (no auto-install)
	//   3 — auto download, prompt to install
	//   4 — auto download + auto install (target SOC posture)
	//   5 — managed by local administrators
	switch auOptions {
	case "3", "4", "5":
		return pass(ev)
	case "2":
		return fail(ev)
	}

	// No managed policy. The Windows Update service must at least be
	// running for the OS default of auto-install to take effect.
	svc := RunCommand(ctx, "sc.exe", "query", "wuauserv")
	if svc.Err != nil {
		ev["wuauserv_error"] = svc.Err.Error()
		return unknown(ev)
	}

	if strings.Contains(svc.Stdout, "RUNNING") {
		ev["wuauserv"] = "running"
		return pass(ev)
	}

	ev["wuauserv"] = "stopped"

	return fail(ev)
}

func windowsPasswordPolicy(ctx context.Context) Result {
	out := powershell(
		ctx,
		`([ADSI]"WinNT://$env:COMPUTERNAME").InvokeGet('MinPasswordLength')`,
	)
	if out.Err != nil {
		return unknown(
			map[string]any{
				"error":  out.Err.Error(),
				"stderr": out.Stderr,
			},
		)
	}

	minLen, err := strconv.Atoi(strings.TrimSpace(out.Stdout))
	if err != nil {
		return unknown(
			map[string]any{
				"error": "invalid MinPasswordLength",
				"raw":   truncate(out.Stdout, 80),
			},
		)
	}

	ev := map[string]any{"min_password_length": minLen}
	if minLen > 0 {
		return pass(ev)
	}

	return fail(ev)
}

func windowsMalwareProtection(ctx context.Context) Result {
	out := powershell(
		ctx,
		`$s = Get-MpComputerStatus; `+
			`"$($s.AntivirusEnabled);$($s.RealTimeProtectionEnabled);`+
			`$($s.AMServiceEnabled);$($s.AntivirusSignatureLastUpdated)"`,
	)
	if out.Err != nil {
		return unknown(
			map[string]any{
				"error":  out.Err.Error(),
				"stderr": out.Stderr,
			},
		)
	}

	parts := strings.Split(out.Stdout, ";")

	ev := map[string]any{"raw": out.Stdout}
	if len(parts) < 3 {
		return unknown(ev)
	}

	antivirusOn := strings.EqualFold(strings.TrimSpace(parts[0]), "True")
	realtimeOn := strings.EqualFold(strings.TrimSpace(parts[1]), "True")
	serviceOn := strings.EqualFold(strings.TrimSpace(parts[2]), "True")

	ev["antivirus_enabled"] = antivirusOn
	ev["real_time_protection"] = realtimeOn

	ev["am_service_enabled"] = serviceOn
	if len(parts) >= 4 {
		ev["signatures_last_updated"] = strings.TrimSpace(parts[3])
	}

	if antivirusOn && (realtimeOn || serviceOn) {
		return pass(ev)
	}

	return fail(ev)
}

func windowsRemoteLogin(ctx context.Context) Result {
	out := powershell(
		ctx,
		`(Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Control\Terminal Server').fDenyTSConnections`,
	)
	if out.Err != nil {
		return unknown(map[string]any{"error": out.Err.Error()})
	}

	ev := map[string]any{"fdeny_ts_connections": out.Stdout}
	if strings.TrimSpace(out.Stdout) == "1" {
		return pass(ev)
	}

	return fail(ev)
}
