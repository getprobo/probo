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

// windowsScreenLockMachineScript reads every machine-wide lock source in one
// invocation. Three separate calls would cost up to 30s against a 25s
// per-check budget. Each source applies to all users and needs no loaded user
// hive.
const windowsScreenLockMachineScript = `` +
	`$s = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System' -ErrorAction SilentlyContinue; ` +
	`$d = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\PolicyManager\current\device\DeviceLock' -ErrorAction SilentlyContinue; ` +
	`$c = Get-ItemProperty 'HKLM:\SOFTWARE\Policies\Microsoft\Windows\Control Panel\Desktop' -ErrorAction SilentlyContinue; ` +
	`"InactivityTimeoutSecs=$($s.InactivityTimeoutSecs);` +
	`MaxInactivityTimeDeviceLock=$($d.MaxInactivityTimeDeviceLock);ScreenSaverIsSecure=$($c.ScreenSaverIsSecure);` +
	`ScreenSaveActive=$($c.ScreenSaveActive);ScreenSaveTimeOut=$($c.ScreenSaveTimeOut)"`

const windowsScreenLockUsersScript = `Get-ChildItem 'Registry::HKEY_USERS' | ` +
	`Where-Object { $_.PSChildName -match '^S-1-5-21-' } | ` +
	`ForEach-Object { ` +
	`  $path = "Registry::HKEY_USERS\$($_.PSChildName)\Control Panel\Desktop"; ` +
	`  $key = Get-ItemProperty $path -ErrorAction SilentlyContinue; ` +
	`  "$($_.PSChildName)=$($key.ScreenSaverIsSecure):$($key.ScreenSaveActive):$($key.ScreenSaveTimeOut)" ` +
	`}`

func windowsScreenLock(ctx context.Context) Result {
	// HKCU resolves to the SYSTEM hive when the agent runs as LocalSystem, so we
	// exhaust machine-wide policy before enumerating loaded interactive hives.
	machine := powershell(ctx, windowsScreenLockMachineScript)
	if machine.Err == nil {
		policy := parseWindowsJoinedPairs(machine.Stdout)
		if source, on, known := windowsScreenLockOn(policy); known {
			ev := map[string]any{
				"backend":              source,
				"policy":               policy,
				"screen_lock_enforced": on,
			}
			if on {
				return pass(ev)
			}

			return fail(ev)
		}
	}

	users := powershell(ctx, windowsScreenLockUsersScript)
	if users.Err != nil {
		return unknown(
			map[string]any{
				"backend":              "hkey_users",
				"error":                errString(users.Err),
				"stderr":               users.Stderr,
				"machine_policy_error": errString(machine.Err),
			},
		)
	}

	ev := map[string]any{
		"backend": "hkey_users",
		"raw":     truncate(users.Stdout, 400),
	}
	perUser, anyDisabled, anyEnabled := parseWindowsUserScreenLock(users.Stdout)

	ev["users"] = perUser
	if len(perUser) == 0 {
		ev["note"] = "no interactive user hives loaded"
		return unknown(ev)
	}

	enforced := anyEnabled && !anyDisabled
	ev["screen_lock_enforced"] = enforced

	if enforced {
		return pass(ev)
	}

	return fail(ev)
}

// parseWindowsUserScreenLock parses one "SID=secure:active:timeout" line per
// user from the registry enumeration and reports whether each user has screen
// saver locking enforced. A user whose values cannot be read counts as
// disabled, so one unprotected account fails the host.
func parseWindowsUserScreenLock(s string) (map[string]string, bool, bool) {
	users := map[string]string{}

	var anyEnabled, anyDisabled bool

	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}

		sid := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		if sid == "" {
			continue
		}

		users[sid] = value

		secure, active, timeout := splitWindowsUserScreenLock(value)
		if on, known := windowsScreenSaverLockOn(secure, active, timeout); known && on {
			anyEnabled = true
			continue
		}

		anyDisabled = true
	}

	return users, anyDisabled, anyEnabled
}

func splitWindowsUserScreenLock(value string) (string, string, string) {
	parts := strings.SplitN(value, ":", 3)
	for len(parts) < 3 {
		parts = append(parts, "")
	}

	return parts[0], parts[1], parts[2]
}

// windowsFirewallCOMScript reads the effective profile state through
// INetFwPolicy2, whose profile constants are 1 domain, 2 private and 4 public.
// It needs no NetSecurity module load and returns booleans rather than the
// localized prose that `netsh advfirewall` prints.
const windowsFirewallCOMScript = `$ErrorActionPreference = 'Stop'; ` +
	`$fw = New-Object -ComObject HNetCfg.FwPolicy2; ` +
	`"Domain=$($fw.FirewallEnabled(1));Private=$($fw.FirewallEnabled(2));Public=$($fw.FirewallEnabled(4))"`

func windowsFirewall(ctx context.Context) Result {
	primary := powershell(
		ctx,
		`(Get-NetFirewallProfile -PolicyStore ActiveStore | `+
			`Sort-Object Name | `+
			`ForEach-Object { "$($_.Name)=$($_.Enabled)" }) -join ";"`,
	)
	if primary.Err == nil && strings.TrimSpace(primary.Stdout) != "" {
		return windowsFirewallResult(
			map[string]any{
				"backend": "Get-NetFirewallProfile",
				"raw":     primary.Stdout,
			},
			primary.Stdout,
		)
	}

	ev := map[string]any{
		"backend":         "HNetCfg.FwPolicy2",
		"degraded":        true,
		"primary_backend": "Get-NetFirewallProfile",
		"primary_error":   errString(primary.Err),
	}
	if primary.TimedOut {
		ev["primary_timed_out"] = true
	}

	fallback := powershell(ctx, windowsFirewallCOMScript)
	if fallback.Err != nil {
		ev["error"] = errString(fallback.Err)
		ev["stderr"] = fallback.Stderr

		if fallback.TimedOut {
			ev["timed_out"] = true
		}

		return unknown(ev)
	}

	ev["raw"] = fallback.Stdout

	return windowsFirewallResult(ev, fallback.Stdout)
}

func windowsFirewallResult(ev map[string]any, raw string) Result {
	profiles := parseWindowsJoinedPairs(raw)
	ev["profiles"] = profiles

	on, known := windowsFirewallOn(profiles)
	if !known {
		return unknown(ev)
	}

	if on {
		return pass(ev)
	}

	return fail(ev)
}

func windowsTimeSync(ctx context.Context) Result {
	out := powershell(
		ctx,
		`$s = Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\W32Time'; `+
			`$p = Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\W32Time\Parameters'; `+
			`"$($s.Start);$($p.Type);$($p.NtpServer)"`,
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

	var serviceStart, typ, ntpServer string
	if len(parts) >= 1 {
		serviceStart = strings.TrimSpace(parts[0])
	}

	if len(parts) >= 2 {
		typ = strings.TrimSpace(parts[1])
	}

	if len(parts) >= 3 {
		ntpServer = strings.TrimSpace(parts[2])
	}

	ev := map[string]any{
		"backend":               "w32time",
		"w32time_service_start": serviceStart,
		"w32time_type":          typ,
	}
	if ntpServer != "" {
		ev["ntp_server"] = ntpServer
	}

	if windowsTimeSyncOn(serviceStart, typ) {
		return pass(ev)
	}

	if serviceStart == "" {
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
			`$svc = Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\wuauserv'; `+
			`"$($au.NoAutoUpdate);$($au.AUOptions);$($svc.Start)"`,
	)

	ev := map[string]any{}
	if out.Err != nil {
		ev["error"] = out.Err.Error()
		ev["stderr"] = out.Stderr

		return unknown(ev)
	}

	parts := strings.SplitN(strings.TrimSpace(out.Stdout), ";", 3)

	var noAutoUpdate, auOptions, serviceStart string
	if len(parts) >= 1 {
		noAutoUpdate = strings.TrimSpace(parts[0])
	}

	if len(parts) >= 2 {
		auOptions = strings.TrimSpace(parts[1])
	}

	if len(parts) >= 3 {
		serviceStart = strings.TrimSpace(parts[2])
	}

	ev["no_auto_update"] = noAutoUpdate
	ev["au_options"] = auOptions
	ev["wuauserv_service_start"] = serviceStart

	on, known := windowsAutoUpdateOn(noAutoUpdate, auOptions, serviceStart)
	if !known {
		return unknown(ev)
	}

	if on {
		return pass(ev)
	}

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
