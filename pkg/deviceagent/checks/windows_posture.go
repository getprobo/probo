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

import "strings"

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

// parseWindowsFirewallProfiles parses "Domain=True;Private=True;Public=True"
// from Get-NetFirewallProfile output, returning per-profile state and
// whether every profile is enabled.
func parseWindowsFirewallProfiles(s string) (map[string]string, bool) {
	profiles := parseWindowsJoinedPairs(s)

	return profiles, windowsAllValuesEqualFold(profiles, "true")
}

func windowsTimeSyncOn(status, typ string) bool {
	if !strings.EqualFold(strings.TrimSpace(status), "Running") {
		return false
	}

	switch strings.ToUpper(strings.TrimSpace(typ)) {
	case "NTP", "NT5DS", "ALLSYNC":
		return true
	}

	return false
}
