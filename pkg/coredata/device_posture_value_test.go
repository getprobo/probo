// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package coredata_test

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// devicePostureValueCase is one evidence fixture and the value it must yield.
// Fixtures are copied from real command output on each platform, since the
// parser dispatches on the tool that produced them.
type devicePostureValueCase struct {
	name       string
	checkKey   string
	evidence   map[string]any
	wantKind   coredata.DevicePostureValueKind
	wantText   string
	wantNumber *int
}

func TestParseDevicePostureValue_OSVersion(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "darwin sw_vers",
				checkKey: "OS_VERSION",
				evidence: map[string]any{
					"product_version": "15.4",
					"build_version":   "24E248",
				},
				wantKind: coredata.DevicePostureValueKindText,
				wantText: "15.4",
			},
			{
				name:     "linux os-release prefers pretty name",
				checkKey: "OS_VERSION",
				evidence: map[string]any{
					"pretty_name": "Ubuntu 24.04.2 LTS",
					"version_id":  "24.04",
					"id":          "ubuntu",
				},
				wantKind: coredata.DevicePostureValueKindText,
				wantText: "Ubuntu 24.04.2 LTS",
			},
			{
				name:     "read failure is unknown",
				checkKey: "OS_VERSION",
				evidence: map[string]any{"error": "sw_vers failed"},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
		},
	)
}

func TestParseDevicePostureValue_DiskEncryption(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "linux lsblk crypt mapping",
				checkKey: "DISK_ENCRYPTION",
				evidence: map[string]any{
					"crypttab_present": false,
					"lsblk":            "nvme0n1 disk\nnvme0n1p3 part crypto_LUKS\nnvme0n1p3_crypt crypt LVM2_member",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "linux no encrypted device",
				checkKey: "DISK_ENCRYPTION",
				evidence: map[string]any{
					"crypttab_present": false,
					"lsblk":            "nvme0n1 disk\nnvme0n1p2 part ext4 /",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "linux lsblk unavailable is unknown",
				checkKey: "DISK_ENCRYPTION",
				evidence: map[string]any{
					"crypttab_present": false,
					"lsblk_error":      "exec: \"lsblk\": executable file not found in $PATH",
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "windows manage-bde missing does not surface the note",
				checkKey: "DISK_ENCRYPTION",
				evidence: map[string]any{"note": "manage-bde not found"},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "freebsd geli active provider",
				checkKey: "DISK_ENCRYPTION",
				evidence: map[string]any{
					"raw": "        Name  Status  Components\nada0p4.eli  ACTIVE  ada0p4",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
		},
	)
}

func TestParseDevicePostureValue_ScreenLock(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "darwin sysadminctl delay",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"backend": "sysadminctl",
					"mode":    "seconds",
					"seconds": float64(900),
					"raw":     "screenLock delay is 900 seconds",
				},
				wantKind:   coredata.DevicePostureValueKindSeconds,
				wantNumber: new(900),
			},
			{
				name:     "darwin sysadminctl off",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"backend": "sysadminctl",
					"mode":    "off",
					"raw":     "screenLock is off",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "darwin defaults zero delay is immediate",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"ask_for_password":       "1",
					"ask_for_password_delay": "0",
				},
				wantKind: coredata.DevicePostureValueKindImmediate,
			},
			{
				name:     "linux gnome lock enabled",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"backend":      "gnome",
					"schema":       "org.gnome.desktop.screensaver",
					"lock_enabled": "true",
					"console_user": "alice",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "linux i3 idle lock",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"backend":      "i3",
					"mechanism":    "xautolock",
					"locker":       "i3lock",
					"idle_minutes": float64(10),
				},
				wantKind:   coredata.DevicePostureValueKindSeconds,
				wantNumber: new(600),
			},
			{
				name:     "windows one user hive insecure",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"backend": "hkey_users",
					"users": map[string]any{
						"S-1-5-21-1004336348-1177238915-682003330-1001": "1",
						"S-1-5-21-1004336348-1177238915-682003330-1002": "0",
					},
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "windows no user hive loaded",
				checkKey: "SCREEN_LOCK",
				evidence: map[string]any{
					"backend": "hkey_users",
					"users":   map[string]any{},
					"note":    "no interactive user hives loaded",
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
		},
	)
}

func TestParseDevicePostureValue_Firewall(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "linux iptables accept policy with rules is unknown",
				checkKey: "FIREWALL_ENABLED",
				evidence: map[string]any{
					"backend":      "iptables",
					"input_policy": "ACCEPT",
					"input_rules":  float64(7),
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "windows netsh without any state line is unknown",
				checkKey: "FIREWALL_ENABLED",
				evidence: map[string]any{
					"backend":     "netsh",
					"state_lines": []any{},
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "no firewall tool found",
				checkKey: "FIREWALL_ENABLED",
				evidence: map[string]any{"note": "no known firewall tool found"},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "empty evidence is unknown",
				checkKey: "FIREWALL_ENABLED",
				evidence: map[string]any{},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
		},
	)
}

func TestParseDevicePostureValue_TimeSync(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "linux timedatectl synchronized",
				checkKey: "TIME_SYNC",
				evidence: map[string]any{
					"raw": "Timezone=Europe/Paris\nLocalRTC=no\nCanNTP=yes\nNTP=yes\nNTPSynchronized=yes",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "linux timedatectl not synchronized",
				checkKey: "TIME_SYNC",
				evidence: map[string]any{
					"raw": "Timezone=Europe/Paris\nLocalRTC=no\nCanNTP=yes\nNTP=yes\nNTPSynchronized=no",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "windows w32tm with a real source",
				checkKey: "TIME_SYNC",
				evidence: map[string]any{
					"raw": "Leap Indicator: 0(no warning)\nStratum: 4 (secondary reference)\nSource: time.windows.com,0x8\nPoll Interval: 10",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "windows w32tm falling back to the local clock",
				checkKey: "TIME_SYNC",
				evidence: map[string]any{
					"raw": "Leap Indicator: 3(not synchronized)\nStratum: 0 (unspecified)\nSource: Local CMOS Clock\nPoll Interval: 10",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				// "is not running" contains "is running".
				name:     "freebsd ntpd stopped",
				checkKey: "TIME_SYNC",
				evidence: map[string]any{"raw": "ntpd is not running."},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "timedatectl missing is unknown",
				checkKey: "TIME_SYNC",
				evidence: map[string]any{"note": "timedatectl not installed"},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
		},
	)
}

func TestParseDevicePostureValue_AutoUpdate(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "darwin every preference enabled",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"backend": "defaults",
					"AutomaticCheckEnabled": map[string]any{
						"source":  "system",
						"value":   "1",
						"enabled": true,
					},
					"AutomaticDownload": map[string]any{
						"source":  "default",
						"enabled": true,
					},
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "darwin one preference disabled",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"backend":       "defaults",
					"disabled_keys": []any{"AutomaticDownload"},
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "darwin unreadable preference",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"backend":            "defaults",
					"indeterminate_keys": []any{"ConfigDataInstall"},
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "linux unattended upgrades enabled",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"backend": "unattended-upgrades",
					"raw":     "APT::Periodic::Update-Package-Lists \"1\";\nAPT::Periodic::Unattended-Upgrade \"1\";\n",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "linux unattended upgrades disabled",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"backend": "unattended-upgrades",
					"raw":     "APT::Periodic::Update-Package-Lists \"0\";\nAPT::Periodic::Unattended-Upgrade \"0\";\n",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				// "disabled" contains "enabled".
				name:     "linux dnf automatic timer disabled",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"backend": "dnf-automatic",
					"state":   "disabled",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "windows policy installs automatically",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"no_auto_update": "0",
					"au_options":     "4",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "windows policy only notifies",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"no_auto_update": "",
					"au_options":     "2",
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "windows without policy falls back to the service",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"no_auto_update": "",
					"au_options":     "",
					"wuauserv":       "running",
				},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				name:     "freebsd has no auto update mechanism",
				checkKey: "AUTO_UPDATE",
				evidence: map[string]any{
					"note": "FreeBSD relies on operator-driven freebsd-update",
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
		},
	)
}

func TestParseDevicePostureValue_PasswordPolicy(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "linux login defs minimum length",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"pass_min_len":       "12",
					"pass_max_days":      "90",
					"pass_min_len_value": float64(12),
				},
				wantKind:   coredata.DevicePostureValueKindMinPasswordLength,
				wantNumber: new(12),
			},
			{
				name:     "linux login defs without a minimum",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"pass_min_len":  "",
					"pass_max_days": "99999",
					"parse_error":   "PASS_MIN_LEN not set",
				},
				wantKind: coredata.DevicePostureValueKindNone,
			},
			{
				name:     "windows net accounts minimum length",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"raw": "Force user logoff how long after time expires?:       Never\nMinimum password age (days):                    0\nMaximum password age (days):                    42\nMinimum password length:                        8\n",
				},
				wantKind:   coredata.DevicePostureValueKindMinPasswordLength,
				wantNumber: new(8),
			},
			{
				name:     "darwin pwpolicy without account policies",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"raw_truncated": "There are no account policies for all users.",
				},
				wantKind: coredata.DevicePostureValueKindNone,
			},
			{
				name:     "darwin pwpolicy with a policy plist",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"raw_truncated": "<dict><key>policyCategoryPasswordContent</key><array>...</array></dict>",
				},
				wantKind: coredata.DevicePostureValueKindConfigured,
			},
			{
				name:     "freebsd login conf minimum length",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"login_conf_snippet": "default:\\\n\t:passwd_format=sha512:\\\n\t:minpasswordlen=10:\\\n\t:passwordtime=90d:",
				},
				wantKind:   coredata.DevicePostureValueKindMinPasswordLength,
				wantNumber: new(10),
			},
			{
				name:     "freebsd login conf without a policy",
				checkKey: "PASSWORD_POLICY",
				evidence: map[string]any{
					"login_conf_snippet": "default:\\\n\t:passwd_format=sha512:",
				},
				wantKind: coredata.DevicePostureValueKindNone,
			},
		},
	)
}

func TestParseDevicePostureValue_RemoteLogin(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "linux without an ssh unit",
				checkKey: "REMOTE_LOGIN",
				evidence: map[string]any{"is_active": ""},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "freebsd sshd running",
				checkKey: "REMOTE_LOGIN",
				evidence: map[string]any{"raw": "sshd is running as pid 987."},
				wantKind: coredata.DevicePostureValueKindOn,
			},
			{
				// "is not running" contains "is running".
				name:     "freebsd sshd stopped",
				checkKey: "REMOTE_LOGIN",
				evidence: map[string]any{"raw": "sshd is not running."},
				wantKind: coredata.DevicePostureValueKindOff,
			},
		},
	)
}

func TestParseDevicePostureValue_MalwareProtection(t *testing.T) {
	t.Parallel()

	runDevicePostureValueCases(
		t,
		[]devicePostureValueCase{
			{
				name:     "darwin xprotect version never surfaces the plist path",
				checkKey: "MALWARE_PROTECTION",
				evidence: map[string]any{
					"engine":  "XProtect",
					"plist":   "/Library/Apple/System/Library/CoreServices/XProtect.bundle/Contents/Resources/XProtect.meta.plist",
					"version": "5260",
				},
				wantKind: coredata.DevicePostureValueKindText,
				wantText: "XProtect 5260",
			},
			{
				name:     "darwin xprotect missing",
				checkKey: "MALWARE_PROTECTION",
				evidence: map[string]any{
					"engine": "XProtect",
					"note":   "XProtect.meta.plist not found in expected locations",
				},
				wantKind: coredata.DevicePostureValueKindNone,
			},
			{
				name:     "linux running agents",
				checkKey: "MALWARE_PROTECTION",
				evidence: map[string]any{
					"active":    []any{"ClamAV", "osquery"},
					"installed": []any{},
				},
				wantKind: coredata.DevicePostureValueKindText,
				wantText: "ClamAV, osquery",
			},
			{
				name:     "linux agent installed but stopped",
				checkKey: "MALWARE_PROTECTION",
				evidence: map[string]any{
					"active":    []any{},
					"installed": []any{"ClamAV"},
				},
				wantKind: coredata.DevicePostureValueKindOff,
			},
			{
				name:     "linux no known agent",
				checkKey: "MALWARE_PROTECTION",
				evidence: map[string]any{
					"active":    []any{},
					"installed": []any{},
				},
				wantKind: coredata.DevicePostureValueKindUnknown,
			},
			{
				name:     "freebsd clamav not installed",
				checkKey: "MALWARE_PROTECTION",
				evidence: map[string]any{"note": "clamav not installed"},
				wantKind: coredata.DevicePostureValueKindNone,
			},
		},
	)
}

// devicePostureAgreementCase pairs evidence with the status the agent itself
// derived from it, copied from the branch in pkg/deviceagent/checks that emits
// that evidence.
type devicePostureAgreementCase struct {
	name        string
	checkKey    string
	evidence    map[string]any
	agentStatus coredata.DevicePostureStatus
}

// passingKindByCheckKey is the state a passing check observed. Remote login is
// the inverted one: a reachable SSH server is the finding, so PASS means OFF.
var passingKindByCheckKey = map[string]coredata.DevicePostureValueKind{
	"FIREWALL_ENABLED":   coredata.DevicePostureValueKindOn,
	"DISK_ENCRYPTION":    coredata.DevicePostureValueKindOn,
	"TIME_SYNC":          coredata.DevicePostureValueKindOn,
	"MALWARE_PROTECTION": coredata.DevicePostureValueKindOn,
	"REMOTE_LOGIN":       coredata.DevicePostureValueKindOff,
}

// TestParseDevicePostureValue_AgreesWithAgentStatus is the regression guard for
// the whole parser. For a check whose value is a state rather than a
// measurement, the agent's PASS/FAIL and the parsed ON/OFF read the same bit
// out of the same evidence, so the two cannot disagree without one of them
// being wrong. Reading `ufw status` with a bare "active" substring test used to
// report a disabled firewall as ON against the agent's own FAIL.
func TestParseDevicePostureValue_AgreesWithAgentStatus(t *testing.T) {
	t.Parallel()

	cases := []devicePostureAgreementCase{
		{
			name:     "darwin alf blocks incoming",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":      "defaults",
				"global_state": "1",
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "darwin alf off",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":      "defaults",
				"global_state": "0",
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "darwin socketfilterfw enabled",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend": "socketfilterfw",
				"raw":     "Firewall is enabled. (State = 1)",
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "darwin socketfilterfw disabled",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend": "socketfilterfw",
				"raw":     "Firewall is disabled. (State = 0)",
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "linux ufw active",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend": "ufw",
				"raw":     "Status: active",
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "linux ufw inactive",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend": "ufw",
				"raw":     "Status: inactive",
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "linux firewalld running",
			checkKey:    "FIREWALL_ENABLED",
			evidence:    map[string]any{"backend": "firewalld", "raw": "running"},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "linux firewalld not running",
			checkKey:    "FIREWALL_ENABLED",
			evidence:    map[string]any{"backend": "firewalld", "raw": "not running"},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "linux nftables with a chain",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":       "nftables",
				"rules_excerpt": "table inet filter {\n\tchain input {\n\t\tpolicy drop;\n\t}\n}",
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "linux nftables empty ruleset",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":       "nftables",
				"rules_excerpt": "table inet filter {\n}",
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "linux iptables drop policy",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":      "iptables",
				"input_policy": "DROP",
				"input_rules":  float64(4),
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "linux iptables accept policy without rules",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":      "iptables",
				"input_policy": "ACCEPT",
				"input_rules":  float64(0),
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "windows every profile enabled",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend": "Get-NetFirewallProfile",
				"raw":     "Domain=True;Private=True;Public=True",
				"profiles": map[string]any{
					"Domain":  "True",
					"Private": "True",
					"Public":  "True",
				},
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "windows public profile disabled",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend": "Get-NetFirewallProfile",
				"raw":     "Domain=True;Private=True;Public=False",
				"profiles": map[string]any{
					"Domain":  "True",
					"Private": "True",
					"Public":  "False",
				},
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "windows netsh every profile on",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":     "netsh",
				"state_lines": []any{"on", "on"},
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "windows netsh one profile off",
			checkKey: "FIREWALL_ENABLED",
			evidence: map[string]any{
				"backend":     "netsh",
				"state_lines": []any{"on", "off"},
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "freebsd pfctl enabled",
			checkKey:    "FIREWALL_ENABLED",
			evidence:    map[string]any{"raw": "Status: Enabled for 3 days 04:21:16"},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "freebsd pfctl disabled",
			checkKey:    "FIREWALL_ENABLED",
			evidence:    map[string]any{"raw": "Status: Disabled"},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "darwin filevault on",
			checkKey:    "DISK_ENCRYPTION",
			evidence:    map[string]any{"raw": "FileVault is On."},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "darwin filevault off",
			checkKey:    "DISK_ENCRYPTION",
			evidence:    map[string]any{"raw": "FileVault is Off."},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "linux crypttab with an entry",
			checkKey: "DISK_ENCRYPTION",
			evidence: map[string]any{
				"crypttab_present": true,
				"crypttab_lines": []any{
					"nvme0n1p3_crypt UUID=6c2f none luks,discard",
				},
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "windows bitlocker fully encrypted",
			checkKey: "DISK_ENCRYPTION",
			evidence: map[string]any{
				"raw": "Conversion Status: Fully Encrypted\n    Percentage Encrypted: 100%",
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "windows bitlocker fully decrypted",
			checkKey: "DISK_ENCRYPTION",
			evidence: map[string]any{
				"raw": "Conversion Status: Fully Decrypted\n    Percentage Encrypted: 0%",
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "darwin network time on",
			checkKey:    "TIME_SYNC",
			evidence:    map[string]any{"raw": "Network Time: On"},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "darwin network time off",
			checkKey:    "TIME_SYNC",
			evidence:    map[string]any{"raw": "Network Time: Off"},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "darwin remote login off",
			checkKey:    "REMOTE_LOGIN",
			evidence:    map[string]any{"raw": "Remote Login: Off"},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "darwin remote login on",
			checkKey:    "REMOTE_LOGIN",
			evidence:    map[string]any{"raw": "Remote Login: On"},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "linux sshd inactive",
			checkKey:    "REMOTE_LOGIN",
			evidence:    map[string]any{"is_active": "inactive"},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "linux sshd active",
			checkKey:    "REMOTE_LOGIN",
			evidence:    map[string]any{"is_active": "active"},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:        "windows terminal services denied",
			checkKey:    "REMOTE_LOGIN",
			evidence:    map[string]any{"fdeny_ts_connections": "1"},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:        "windows terminal services allowed",
			checkKey:    "REMOTE_LOGIN",
			evidence:    map[string]any{"fdeny_ts_connections": "0"},
			agentStatus: coredata.DevicePostureStatusFail,
		},
		{
			name:     "windows defender with real time protection",
			checkKey: "MALWARE_PROTECTION",
			evidence: map[string]any{
				"antivirus_enabled":    true,
				"real_time_protection": true,
				"am_service_enabled":   true,
			},
			agentStatus: coredata.DevicePostureStatusPass,
		},
		{
			name:     "windows defender disabled",
			checkKey: "MALWARE_PROTECTION",
			evidence: map[string]any{
				"antivirus_enabled":    false,
				"real_time_protection": false,
				"am_service_enabled":   false,
			},
			agentStatus: coredata.DevicePostureStatusFail,
		},
	}

	for _, tt := range cases {
		t.Run(tt.checkKey+" "+tt.name, func(t *testing.T) {
			t.Parallel()

			passingKind, ok := passingKindByCheckKey[tt.checkKey]
			require.True(t, ok, "no passing kind declared for %s", tt.checkKey)

			wantKind := passingKind
			if tt.agentStatus == coredata.DevicePostureStatusFail {
				wantKind = oppositeDevicePostureValueKind(passingKind)
			}

			raw, err := json.Marshal(tt.evidence)
			require.NoError(t, err)

			value := coredata.ParseDevicePostureValue(tt.checkKey, raw)

			assert.Equal(
				t,
				wantKind,
				value.Kind,
				"the agent read this evidence as %s, so the value must be %s",
				tt.agentStatus,
				wantKind,
			)
		})
	}
}

func oppositeDevicePostureValueKind(
	kind coredata.DevicePostureValueKind,
) coredata.DevicePostureValueKind {
	if kind == coredata.DevicePostureValueKindOn {
		return coredata.DevicePostureValueKindOff
	}

	return coredata.DevicePostureValueKindOn
}

func TestParseDevicePostureValue_UnknownCheckKey(t *testing.T) {
	t.Parallel()

	value := coredata.ParseDevicePostureValue(
		"SOMETHING_NEW",
		json.RawMessage(`{"raw":"whatever the agent sent"}`),
	)

	assert.Equal(t, coredata.DevicePostureValueKindUnknown, value.Kind)
	assert.Empty(t, value.Text)
}

func TestParseDevicePostureValue_MalformedEvidence(t *testing.T) {
	t.Parallel()

	for name, evidence := range map[string]string{
		"empty":         "",
		"empty object":  "{}",
		"not an object": `["raw"]`,
		"invalid json":  `{"raw":`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value := coredata.ParseDevicePostureValue(
				"DISK_ENCRYPTION",
				json.RawMessage(evidence),
			)

			assert.Equal(t, coredata.DevicePostureValueKindUnknown, value.Kind)
		})
	}
}

// TestParseDevicePostureValue_NeverSurfacesRawEvidence guards the privacy
// property of the parser: evidence carries usernames, file paths and command
// output, and none of it may reach the value. Anything the parser cannot
// classify is UNKNOWN.
func TestParseDevicePostureValue_NeverSurfacesRawEvidence(t *testing.T) {
	t.Parallel()

	secrets := []string{"alice", "/home/alice", "manage-bde", "netsh", "gsettings"}

	evidences := map[string]map[string]any{
		"linux screen lock with console user": {
			"backend":      "gnome",
			"schema":       "org.gnome.desktop.screensaver",
			"lock_enabled": "true",
			"console_user": "alice",
		},
		"linux screen lock with an unreadable schema": {
			"backend":      "gsettings",
			"console_user": "alice",
		},
		"linux i3 config path": {
			"backend":      "i3",
			"config":       "/home/alice/.config/i3/config",
			"console_user": "alice",
			"mechanism":    "xautolock",
			"locker":       "i3lock",
			"idle_minutes": float64(5),
		},
		"windows bitlocker unavailable": {
			"note": "manage-bde not found",
		},
		"windows firewall via netsh": {
			"backend":     "netsh",
			"raw":         "State ON",
			"state_lines": []any{"on"},
		},
	}

	for name, evidence := range evidences {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(evidence)
			require.NoError(t, err)

			for _, checkKey := range coredata.DevicePostureCheckKeys() {
				value := coredata.ParseDevicePostureValue(checkKey.String(), raw)

				for _, secret := range secrets {
					assert.NotContains(
						t,
						value.Text,
						secret,
						"check %s leaked evidence into the value",
						checkKey,
					)
				}
			}
		})
	}
}

// TestParseDevicePostureValue_TextIsValidUTF8 covers the truncation boundary:
// cutting a long value mid-rune would produce invalid UTF-8 and break JSON
// encoding of the response.
func TestParseDevicePostureValue_TextIsValidUTF8(t *testing.T) {
	t.Parallel()

	evidence, err := json.Marshal(
		map[string]any{"pretty_name": strings.Repeat("é", 200)},
	)
	require.NoError(t, err)

	value := coredata.ParseDevicePostureValue("OS_VERSION", evidence)

	require.Equal(t, coredata.DevicePostureValueKindText, value.Kind)
	assert.True(t, utf8.ValidString(value.Text))
	assert.LessOrEqual(t, len(value.Text), 80)
}

func runDevicePostureValueCases(t *testing.T, cases []devicePostureValueCase) {
	t.Helper()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tt.evidence)
			require.NoError(t, err)

			value := coredata.ParseDevicePostureValue(tt.checkKey, raw)

			assert.Equal(t, tt.wantKind, value.Kind)
			assert.Equal(t, tt.wantText, value.Text)

			if tt.wantNumber == nil {
				assert.Nil(t, value.Number)

				return
			}

			require.NotNil(t, value.Number)
			assert.Equal(t, *tt.wantNumber, *value.Number)
		})
	}
}
