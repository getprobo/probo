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

package coredata

import (
	"strings"
)

// One parser per posture check. Each dispatches on the tool the agent used, and
// the tools are enumerated in pkg/deviceagent/checks — read those alongside
// these branches, since they define what each backend's output means.

func parseOSVersionValue(ev map[string]any) DevicePostureValue {
	return textValue(
		firstNonEmptyString(
			stringEvidence(ev, "product_version"),
			stringEvidence(ev, "pretty_name"),
			stringEvidence(ev, "version_id"),
			stringEvidence(ev, "version"),
			stringEvidence(ev, "caption"),
			stringEvidence(ev, "release"),
		),
	)
}

func parseDiskEncryptionValue(ev map[string]any) DevicePostureValue {
	// Linux is the only platform reporting crypttab, so the key doubles as the
	// platform discriminator.
	if present, ok := boolEvidence(ev, "crypttab_present"); ok {
		return parseLinuxDiskEncryptionValue(ev, present)
	}

	if backendOf(ev) == "get-bitlockervolume" {
		return parseWindowsBitLockerValue(ev)
	}

	raw := lowerStringEvidence(ev, "raw")
	switch {
	case raw == "":
		return unknownValue()
	case strings.Contains(raw, "filevault is on"):
		return onOffValue(true)
	case strings.Contains(raw, "filevault is off"):
		return onOffValue(false)
	case strings.Contains(raw, "components"):
		// FreeBSD geli prints a "Name Status Components" table with one row per
		// encrypted provider; ACTIVE is the only status meaning attached.
		return onOffValue(strings.Contains(raw, "active"))
	}

	return unknownValue()
}

func parseWindowsBitLockerValue(ev map[string]any) DevicePostureValue {
	if stringEvidence(ev, "error") != "" {
		return unknownValue()
	}

	allOn, any := allValuesMatch(stringMapEvidence(ev, "volumes"), "on")
	if !any {
		// No OS volumes, or the BitLocker cmdlet is missing (Home).
		return onOffValue(false)
	}

	return onOffValue(allOn)
}

func parseLinuxDiskEncryptionValue(
	ev map[string]any,
	crypttabPresent bool,
) DevicePostureValue {
	if crypttabPresent && len(stringSliceEvidence(ev, "crypttab_lines")) > 0 {
		return onOffValue(true)
	}

	if lsblk := stringEvidence(ev, "lsblk"); lsblk != "" {
		return onOffValue(lsblkHasCryptDevice(lsblk))
	}

	return unknownValue()
}

// lsblkHasCryptDevice reports whether `lsblk -o NAME,TYPE,... -r` listed a
// device of type crypt, which is how a LUKS mapping appears.
func lsblkHasCryptDevice(raw string) bool {
	for line := range strings.SplitSeq(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if fields[1] == "crypt" {
			return true
		}
	}

	return false
}

func parseScreenLockValue(ev map[string]any) DevicePostureValue {
	switch backendOf(ev) {
	case "sysadminctl":
		return parseDarwinScreenLockModeValue(ev)
	case "gnome", "cinnamon", "mate", "ukui":
		return boolKeyValue(ev, "lock_enabled")
	case "kde":
		return boolKeyValue(ev, "autolock")
	case "xfce":
		return boolKeyValue(ev, "enabled")
	case "i3":
		return parseI3ScreenLockValue(ev)
	case "machine_policy":
		return boolKeyValue(ev, "screen_saver_is_secure")
	case "hkey_users":
		return parseWindowsUserScreenLockValue(ev)
	}

	// macOS falls back to the com.apple.screensaver defaults, which set no
	// backend key.
	if ask, ok := boolEvidence(ev, "ask_for_password"); ok {
		if !ask {
			return onOffValue(false)
		}

		if delay, ok := numberEvidence(ev, "ask_for_password_delay"); ok {
			return screenLockDelayValue(delay)
		}

		return onOffValue(true)
	}

	return unknownValue()
}

func parseDarwinScreenLockModeValue(ev map[string]any) DevicePostureValue {
	switch stringEvidence(ev, "mode") {
	case "immediate":
		return DevicePostureValue{Kind: DevicePostureValueKindImmediate}
	case "off":
		return onOffValue(false)
	case "seconds":
		if seconds, ok := numberEvidence(ev, "seconds"); ok {
			return screenLockDelayValue(seconds)
		}

		return onOffValue(true)
	}

	return unknownValue()
}

func parseI3ScreenLockValue(ev map[string]any) DevicePostureValue {
	if stringEvidence(ev, "error") != "" {
		return unknownValue()
	}

	// No idle lock command in the config at all.
	if stringEvidence(ev, "mechanism") == "" {
		return onOffValue(false)
	}

	if minutes, ok := numberEvidence(ev, "idle_minutes"); ok && minutes > 0 {
		return secondsValue(minutes * 60)
	}

	return onOffValue(true)
}

func parseWindowsUserScreenLockValue(ev map[string]any) DevicePostureValue {
	allSecure, any := allValuesMatch(stringMapEvidence(ev, "users"), "1")
	if !any {
		return unknownValue()
	}

	return onOffValue(allSecure)
}

func screenLockDelayValue(seconds int) DevicePostureValue {
	if seconds <= 0 {
		return DevicePostureValue{Kind: DevicePostureValueKindImmediate}
	}

	return secondsValue(seconds)
}

func parseFirewallValue(ev map[string]any) DevicePostureValue {
	switch backendOf(ev) {
	case "defaults":
		return parseDarwinFirewallStateValue(ev)
	case "socketfilterfw":
		return parseSocketFilterFWValue(ev)
	case "ufw":
		return parseUFWValue(ev)
	case "firewalld":
		return parseFirewalldValue(ev)
	case "nftables":
		return parseNftablesValue(ev)
	case "iptables":
		return parseIptablesValue(ev)
	case "get-netfirewallprofile":
		return parseWindowsFirewallProfilesValue(ev)
	case "netsh":
		return parseNetshFirewallValue(ev)
	}

	// FreeBSD pfctl sets no backend key.
	raw := lowerStringEvidence(ev, "raw")
	switch {
	case strings.Contains(raw, "status: enabled"):
		return onOffValue(true)
	case strings.Contains(raw, "status: disabled"):
		return onOffValue(false)
	}

	return unknownValue()
}

// parseDarwinFirewallStateValue reads com.apple.alf globalstate, where 1 blocks
// incoming connections and 2 blocks all but essential services.
func parseDarwinFirewallStateValue(ev map[string]any) DevicePostureValue {
	switch stringEvidence(ev, "global_state") {
	case "1", "2":
		return onOffValue(true)
	case "0":
		return onOffValue(false)
	}

	return unknownValue()
}

func parseSocketFilterFWValue(ev map[string]any) DevicePostureValue {
	raw := lowerStringEvidence(ev, "raw")
	switch {
	case strings.Contains(raw, "disabled"), strings.Contains(raw, "state = 0"):
		return onOffValue(false)
	case strings.Contains(raw, "enabled"),
		strings.Contains(raw, "state = 1"),
		strings.Contains(raw, "state = 2"):
		return onOffValue(true)
	}

	return unknownValue()
}

// parseUFWValue reads `ufw status`. The whole "status: <state>" phrase has to
// match: "inactive" contains "active", so a bare substring test reads a
// disabled firewall as enabled.
func parseUFWValue(ev map[string]any) DevicePostureValue {
	raw := lowerStringEvidence(ev, "raw")
	switch {
	case strings.Contains(raw, "status: inactive"):
		return onOffValue(false)
	case strings.Contains(raw, "status: active"):
		return onOffValue(true)
	}

	return unknownValue()
}

// parseFirewalldValue reads `firewall-cmd --state`, which prints "running" or
// "not running" — so the negative has to be tested first.
func parseFirewalldValue(ev map[string]any) DevicePostureValue {
	raw := lowerStringEvidence(ev, "raw")
	switch {
	case strings.Contains(raw, "not running"):
		return onOffValue(false)
	case strings.Contains(raw, "running"):
		return onOffValue(true)
	}

	return unknownValue()
}

func parseNftablesValue(ev map[string]any) DevicePostureValue {
	if stringEvidence(ev, "error") != "" {
		return unknownValue()
	}

	excerpt := stringEvidence(ev, "rules_excerpt")
	if excerpt == "" {
		return unknownValue()
	}

	return onOffValue(strings.Contains(excerpt, "chain "))
}

// parseIptablesValue reads the INPUT chain default policy. An ACCEPT policy
// carrying rules cannot be classified without modelling the whole chain, which
// is what the agent declines to do as well.
func parseIptablesValue(ev map[string]any) DevicePostureValue {
	if stringEvidence(ev, "error") != "" {
		return unknownValue()
	}

	switch strings.ToUpper(stringEvidence(ev, "input_policy")) {
	case "DROP", "REJECT":
		return onOffValue(true)
	case "ACCEPT":
		if rules, ok := numberEvidence(ev, "input_rules"); ok && rules == 0 {
			return onOffValue(false)
		}

		return unknownValue()
	}

	return unknownValue()
}

func parseWindowsFirewallProfilesValue(ev map[string]any) DevicePostureValue {
	allEnabled, any := allValuesMatch(stringMapEvidence(ev, "profiles"), "true")
	if !any {
		return unknownValue()
	}

	return onOffValue(allEnabled)
}

func parseNetshFirewallValue(ev map[string]any) DevicePostureValue {
	allOn, any := allEntriesMatch(stringSliceEvidence(ev, "state_lines"), "on")
	if !any {
		return unknownValue()
	}

	return onOffValue(allOn)
}

func parseTimeSyncValue(ev map[string]any) DevicePostureValue {
	if status := stringEvidence(ev, "w32time_status"); status != "" {
		return parseWindowsTimeSyncValue(status, stringEvidence(ev, "w32time_type"))
	}

	raw := lowerStringEvidence(ev, "raw")
	switch {
	case raw == "":
		return unknownValue()
	case strings.Contains(raw, "ntpsynchronized=yes"):
		return onOffValue(true)
	case strings.Contains(raw, "ntpsynchronized=no"):
		return onOffValue(false)
	case strings.Contains(raw, "network time: on"):
		return onOffValue(true)
	case strings.Contains(raw, "network time: off"):
		return onOffValue(false)
	case strings.Contains(raw, "is not running"):
		return onOffValue(false)
	case strings.Contains(raw, "is running"):
		return onOffValue(true)
	}

	return unknownValue()
}

func parseWindowsTimeSyncValue(status, typ string) DevicePostureValue {
	if !strings.EqualFold(status, "Running") {
		return onOffValue(false)
	}

	switch strings.ToUpper(strings.TrimSpace(typ)) {
	case "NTP", "NT5DS", "ALLSYNC":
		return onOffValue(true)
	}

	return onOffValue(false)
}

func parseAutoUpdateValue(ev map[string]any) DevicePostureValue {
	switch backendOf(ev) {
	case "defaults":
		return parseDarwinSoftwareUpdateValue(ev)
	case "unattended-upgrades":
		// Each APT periodic task is enabled with a quoted "1".
		return onOffValue(strings.Contains(stringEvidence(ev, "raw"), `"1"`))
	case "dnf-automatic":
		return unitStateValue(stringEvidence(ev, "state"))
	}

	// The Windows Update policy read sets no backend key.
	if hasAnyKey(ev, "no_auto_update", "au_options", "wuauserv") {
		return parseWindowsAutoUpdateValue(ev)
	}

	return unknownValue()
}

// parseDarwinSoftwareUpdateValue collapses the five Software Update preferences
// into one value: any preference off makes automatic updates off.
func parseDarwinSoftwareUpdateValue(ev map[string]any) DevicePostureValue {
	if len(stringSliceEvidence(ev, "disabled_keys")) > 0 {
		return onOffValue(false)
	}

	if len(stringSliceEvidence(ev, "indeterminate_keys")) > 0 {
		return unknownValue()
	}

	return onOffValue(true)
}

func parseWindowsAutoUpdateValue(ev map[string]any) DevicePostureValue {
	if stringEvidence(ev, "no_auto_update") == "1" {
		return onOffValue(false)
	}

	// AUOptions: 2 notifies only, 3 downloads, 4 downloads and installs, 5
	// delegates to local administrators.
	switch stringEvidence(ev, "au_options") {
	case "3", "4", "5":
		return onOffValue(true)
	case "2":
		return onOffValue(false)
	}

	// With no managed policy the value is whether the Windows Update service is
	// running to apply the OS default.
	switch stringEvidence(ev, "wuauserv") {
	case "running":
		return onOffValue(true)
	case "stopped":
		return onOffValue(false)
	}

	return unknownValue()
}

func parsePasswordPolicyValue(ev map[string]any) DevicePostureValue {
	// Linux reads PASS_MIN_LEN from /etc/login.defs.
	if minLen, ok := numberEvidence(ev, "pass_min_len_value"); ok {
		return minPasswordLengthValue(minLen)
	}

	if minLen, ok := numberEvidence(ev, "pass_min_len"); ok {
		return minPasswordLengthValue(minLen)
	}

	if minLen, ok := numberEvidence(ev, "min_password_length"); ok {
		return minPasswordLengthValue(minLen)
	}

	if parseError := lowerStringEvidence(ev, "parse_error"); parseError != "" {
		if strings.Contains(parseError, "not set") {
			return noneValue()
		}

		return unknownValue()
	}

	// FreeBSD /etc/login.conf.
	if snippet := stringEvidence(ev, "login_conf_snippet"); snippet != "" {
		if minLen, ok := parseAssignedInt(snippet, "minpasswordlen"); ok {
			return minPasswordLengthValue(minLen)
		}

		if strings.Contains(strings.ToLower(snippet), "passwordtime=") {
			return configuredValue()
		}

		return noneValue()
	}

	// macOS pwpolicy returns the policy plist, which has no single figure.
	if _, ok := ev["raw_truncated"]; ok {
		raw := lowerStringEvidence(ev, "raw_truncated")
		if raw == "" || strings.Contains(raw, "no account policies") {
			return noneValue()
		}

		return configuredValue()
	}

	return unknownValue()
}

// parseRemoteLoginValue reports whether remote login is reachable, so On is the
// insecure observation here.
func parseRemoteLoginValue(ev map[string]any) DevicePostureValue {
	// Windows: fDenyTSConnections=1 refuses Terminal Server connections.
	if deny, ok := boolEvidence(ev, "fdeny_ts_connections"); ok {
		return onOffValue(!deny)
	}

	// Linux reports the ssh unit state.
	if _, ok := ev["is_active"]; ok {
		return unitActiveValue(stringEvidence(ev, "is_active"))
	}

	raw := lowerStringEvidence(ev, "raw")
	switch {
	case strings.Contains(raw, "remote login: on"):
		return onOffValue(true)
	case strings.Contains(raw, "remote login: off"):
		return onOffValue(false)
	case strings.Contains(raw, "is not running"):
		return onOffValue(false)
	case strings.Contains(raw, "is running"):
		return onOffValue(true)
	}

	return unknownValue()
}

func parseMalwareProtectionValue(ev map[string]any) DevicePostureValue {
	// Windows Defender.
	if antivirus, ok := boolEvidence(ev, "antivirus_enabled"); ok {
		realtime, _ := boolEvidence(ev, "real_time_protection")
		service, _ := boolEvidence(ev, "am_service_enabled")

		return onOffValue(antivirus && (realtime || service))
	}

	// macOS XProtect. The plist path is never surfaced.
	if engine := stringEvidence(ev, "engine"); engine != "" {
		if strings.Contains(lowerStringEvidence(ev, "note"), "not found") {
			return noneValue()
		}

		if version := stringEvidence(ev, "version"); version != "" {
			return textValue(engine + " " + version)
		}

		return textValue(engine)
	}

	// Linux endpoint agents. Running agents are the value; agents installed but
	// not running mean the protection is off.
	if _, ok := ev["active"]; ok {
		if active := stringSliceEvidence(ev, "active"); len(active) > 0 {
			return textValue(strings.Join(active, ", "))
		}

		if len(stringSliceEvidence(ev, "installed")) > 0 {
			return onOffValue(false)
		}

		return unknownValue()
	}

	// FreeBSD clamav.
	raw := lowerStringEvidence(ev, "raw")
	switch {
	case strings.Contains(raw, "is not running"):
		return onOffValue(false)
	case strings.Contains(raw, "is running"):
		return onOffValue(true)
	}

	if strings.Contains(lowerStringEvidence(ev, "note"), "not installed") {
		return noneValue()
	}

	return unknownValue()
}

// unitStateValue maps `systemctl is-enabled` output.
func unitStateValue(state string) DevicePostureValue {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "":
		return unknownValue()
	case "enabled", "enabled-runtime":
		return onOffValue(true)
	default:
		return onOffValue(false)
	}
}

// unitActiveValue maps `systemctl is-active` output.
func unitActiveValue(state string) DevicePostureValue {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "active", "activating":
		return onOffValue(true)
	case "inactive", "failed", "deactivating":
		return onOffValue(false)
	}

	return unknownValue()
}

func boolKeyValue(ev map[string]any, key string) DevicePostureValue {
	v, ok := boolEvidence(ev, key)
	if !ok {
		return unknownValue()
	}

	return onOffValue(v)
}
