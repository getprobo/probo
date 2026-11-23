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
	"strconv"
	"strings"
)

// parseWindowsJoinedPairs parses "K=V;K2=V2" produced by PowerShell
// `-join ";"`. Malformed segments and empty keys are skipped. SplitN
// with n=2 keeps BitLocker keys like "C:" intact in "C:=On".
func parseWindowsJoinedPairs(s string) map[string]string {
	out := map[string]string{}

	for pair := range strings.SplitSeq(s, ";") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if name == "" {
			continue
		}

		out[name] = value
	}

	return out
}

func windowsAllValuesEqualFold(m map[string]string, want string) bool {
	if len(m) == 0 {
		return false
	}

	for _, v := range m {
		if !strings.EqualFold(v, want) {
			return false
		}
	}

	return true
}

// parseWindowsBitLockerVolumes parses "C:=On;D:=Off" from Get-BitLockerVolume
// output, returning per-volume ProtectionStatus and whether every volume is
// protected.
func parseWindowsBitLockerVolumes(s string) (map[string]string, bool) {
	volumes := parseWindowsJoinedPairs(s)

	return volumes, windowsAllValuesEqualFold(volumes, "on")
}

// windowsFirewallOn reports whether every firewall profile is enabled, and
// whether the values could be read at all. A profile set that is empty or holds
// anything other than a boolean must not be reported as a disabled firewall.
func windowsFirewallOn(profiles map[string]string) (bool, bool) {
	if len(profiles) == 0 {
		return false, false
	}

	on := true

	for _, value := range profiles {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "1":
		case "false", "0":
			on = false
		default:
			return false, false
		}
	}

	return on, true
}

func windowsTimeSyncOn(serviceStart, typ string) bool {
	switch strings.TrimSpace(serviceStart) {
	case "2", "3":
	default:
		return false
	}

	switch strings.ToUpper(strings.TrimSpace(typ)) {
	case "NTP", "NT5DS", "ALLSYNC":
		return true
	}

	return false
}

// windowsScreenLockOn resolves screen lock from machine-wide policy, which
// applies to every user and is readable without a loaded user hive. It names
// the source that decided so a change of source is not mistaken for a change of
// state, and reports whether any source answered at all.
// Nothing here reads "Do not display the lock screen": it replaces the glance
// screen with the credential prompt and does not stop the host locking, so it
// cannot answer this check either way.
func windowsScreenLockOn(values map[string]string) (string, bool, bool) {
	if seconds, ok := windowsInt(values["InactivityTimeoutSecs"]); ok && seconds > 0 {
		return "machine_inactivity_limit", true, true
	}

	if minutes, ok := windowsInt(values["MaxInactivityTimeDeviceLock"]); ok && minutes > 0 {
		return "mdm_device_lock", true, true
	}

	on, known := windowsScreenSaverLockOn(
		values["ScreenSaverIsSecure"],
		values["ScreenSaveActive"],
		values["ScreenSaveTimeOut"],
	)
	if known {
		return "machine_policy", on, true
	}

	return "", false, false
}

// windowsInteractiveUserSID reports whether a HKEY_USERS child is a real
// interactive account. Local and AD users are S-1-5-21-*; Entra users are
// S-1-12-1-*. The _Classes suffix is a per-user COM hive, not a login.
func windowsInteractiveUserSID(sid string) bool {
	sid = strings.TrimSpace(sid)
	if sid == "" {
		return false
	}

	upper := strings.ToUpper(sid)
	if strings.HasSuffix(upper, "_CLASSES") {
		return false
	}

	return strings.HasPrefix(upper, "S-1-5-21-") || strings.HasPrefix(upper, "S-1-12-1-")
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

		if !windowsInteractiveUserSID(sid) {
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

// windowsScreenSaverLockOn evaluates the screensaver trio shared by the machine
// policy and the per-user hives. A secure screensaver only locks anything when
// it is also active with a non-zero timeout.
func windowsScreenSaverLockOn(secure, active, timeout string) (bool, bool) {
	switch strings.TrimSpace(secure) {
	case "0":
		return false, true
	case "1":
		if strings.TrimSpace(active) != "1" {
			return false, true
		}

		seconds, ok := windowsInt(timeout)

		return ok && seconds > 0, true
	}

	return false, false
}

func windowsInt(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, false
	}

	return n, true
}

// parseWindowsSeceditMinPasswordLength reads the first MinimumPasswordLength
// assignment from a secedit INF export. The key is locale-stable; surrounding
// whitespace and a missing key must not look like a zero-length policy.
func parseWindowsSeceditMinPasswordLength(inf string) (int, bool) {
	for line := range strings.SplitSeq(inf, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}

		if !strings.EqualFold(strings.TrimSpace(key), "MinimumPasswordLength") {
			continue
		}

		return windowsInt(value)
	}

	return 0, false
}

// windowsPasswordPolicyOn picks the stricter of the local SAM policy and the
// MDM DeviceLock PIN length. Both sources are optional: domain-joined hosts
// often have only secedit, Entra-joined hosts often have only DeviceLock.
func windowsPasswordPolicyOn(inf, mdm string) (int, string, bool) {
	samLen, samOK := parseWindowsSeceditMinPasswordLength(inf)
	mdmLen, mdmOK := windowsInt(mdm)

	switch {
	case samOK && mdmOK:
		length := max(mdmLen, samLen)

		return length, "max", true
	case samOK:
		return samLen, "secedit", true
	case mdmOK:
		return mdmLen, "mdm_device_lock", true
	default:
		return 0, "", false
	}
}

func windowsAutoUpdateOn(noAutoUpdate, auOptions, serviceStart string) (bool, bool) {
	switch strings.TrimSpace(serviceStart) {
	case "4":
		return false, true
	case "2", "3":
	default:
		return false, false
	}

	if strings.TrimSpace(noAutoUpdate) == "1" {
		return false, true
	}

	switch strings.TrimSpace(auOptions) {
	case "", "3", "4", "5":
		return true, true
	case "2":
		return false, true
	default:
		return false, false
	}
}

func windowsNinjaAgentRunning(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "Running")
}

// A stopped NinjaOne agent is not patching, so Windows Update stays the reading.
func windowsAutoUpdateVerdict(
	ninjaStatus string,
	noAutoUpdate string,
	auOptions string,
	serviceStart string,
) (bool, bool) {
	if windowsNinjaAgentRunning(ninjaStatus) {
		return true, true
	}

	return windowsAutoUpdateOn(noAutoUpdate, auOptions, serviceStart)
}

type (
	windowsEndpointService struct {
		name     string
		protects bool
	}

	windowsEndpointProduct struct {
		name     string
		services []windowsEndpointService
	}
)

// Order is the order a running product is reported in.
var windowsEndpointProducts = []windowsEndpointProduct{
	{
		name: "SentinelOne",
		services: []windowsEndpointService{
			{name: "SentinelAgent", protects: true},
		},
	},
	{
		name: "CrowdStrike Falcon",
		services: []windowsEndpointService{
			{name: "csagent", protects: true},
			// The user-mode service stays up when the sensor itself is stopped.
			{name: "CSFalconService", protects: false},
		},
	},
	{
		name: "Bitdefender",
		services: []windowsEndpointService{
			{name: "EPSecurityService", protects: true},
			{name: "VSSERV", protects: true},
			{name: "EPProtectedService", protects: false},
		},
	},
	{
		name: "Webroot",
		services: []windowsEndpointService{
			{name: "WRSVC", protects: true},
		},
	},
}

func windowsEndpointServiceNames() []string {
	names := make([]string, 0)
	seen := map[string]bool{}

	for _, product := range windowsEndpointProducts {
		for _, svc := range product.services {
			key := strings.ToLower(svc.name)
			if seen[key] {
				continue
			}

			seen[key] = true

			names = append(names, svc.name)
		}
	}

	return names
}

func windowsMalwareProtectionScript() string {
	names := windowsEndpointServiceNames()

	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = "'" + name + "'"
	}

	return `$ErrorActionPreference = 'Stop'; ` +
		`$names = @(` + strings.Join(quoted, ",") + `); ` +
		`$parts = @(); ` +
		`foreach ($n in $names) { ` +
		`  $svc = Get-Service -Name $n -ErrorAction SilentlyContinue; ` +
		`  if ($null -ne $svc) { $parts += ('{0}={1}' -f $n, $svc.Status) } ` +
		`}; ` +
		`$defender = 'ERR'; ` +
		`try { ` +
		`  $s = Get-MpComputerStatus -ErrorAction Stop; ` +
		`  $defender = '{0};{1};{2};{3}' -f ` +
		`$s.AntivirusEnabled, $s.RealTimeProtectionEnabled, ` +
		`$s.AMServiceEnabled, $s.AntivirusSignatureLastUpdated ` +
		`} catch { }; ` +
		`Write-Output ('SERVICES=' + ($parts -join ';')); ` +
		`Write-Output ('DEFENDER=' + $defender)`
}

func parseWindowsMalwareProbe(stdout string) (map[string]string, string, bool) {
	var (
		servicesLine string
		defenderLine string
		sawServices  bool
		sawDefender  bool
	)

	normalized := strings.ReplaceAll(stdout, "\r\n", "\n")
	for line := range strings.SplitSeq(normalized, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "SERVICES="):
			servicesLine = strings.TrimPrefix(line, "SERVICES=")
			sawServices = true
		case strings.HasPrefix(line, "DEFENDER="):
			defenderLine = strings.TrimPrefix(line, "DEFENDER=")
			sawDefender = true
		}
	}

	if !sawServices || !sawDefender {
		return nil, "", false
	}

	return parseWindowsJoinedPairs(servicesLine), defenderLine, true
}

func windowsServiceStatus(services map[string]string, name string) (string, bool) {
	for key, status := range services {
		if strings.EqualFold(strings.TrimSpace(key), name) {
			return status, true
		}
	}

	return "", false
}

func classifyWindowsEndpointAgents(services map[string]string) ([]string, []string) {
	var active, installed []string

	for _, product := range windowsEndpointProducts {
		running := false
		present := false

		for _, svc := range product.services {
			status, ok := windowsServiceStatus(services, svc.name)
			if !ok {
				continue
			}

			present = true

			if svc.protects && strings.EqualFold(strings.TrimSpace(status), "Running") {
				running = true
			}
		}

		switch {
		case running:
			active = append(active, product.name)
		case present:
			installed = append(installed, product.name)
		}
	}

	return active, installed
}

func windowsMalwareProtectionResult(services map[string]string, defender string) Result {
	active, installed := classifyWindowsEndpointAgents(services)
	if len(active) > 0 || len(installed) > 0 {
		if active == nil {
			active = []string{}
		}

		if installed == nil {
			installed = []string{}
		}

		ev := map[string]any{
			"active":    active,
			"installed": installed,
		}
		if len(active) > 0 {
			return pass(ev)
		}

		return fail(ev)
	}

	return windowsDefenderMalwareResult(defender)
}

func windowsDefenderMalwareResult(defender string) Result {
	defender = strings.TrimSpace(defender)
	if defender == "" || strings.EqualFold(defender, "ERR") {
		ev := map[string]any{}
		if strings.EqualFold(defender, "ERR") {
			ev["error"] = "Get-MpComputerStatus failed"
		}

		return unknown(ev)
	}

	parts := strings.Split(defender, ";")

	ev := map[string]any{"raw": defender}
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
