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
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"
)

// devicePostureValueTextMax bounds a TEXT value. Evidence literals we surface
// are short by nature (a version, an engine name, a few agent names); the cap
// only guards against a pathological host.
const devicePostureValueTextMax = 80

// DevicePostureValue is the observation a posture check made, in a shape a
// client can localize. Status is deliberately absent: whether the observation
// is acceptable is a ruleset decision, not a property of the measurement.
type DevicePostureValue struct {
	Kind   DevicePostureValueKind
	Text   string
	Number *int
}

// ParseDevicePostureValue derives the observed value of a posture check from
// the evidence the agent recorded.
//
// Evidence shapes differ per platform and per tool, so each check dispatches on
// the "backend" key the agent sets (or, where it sets none, on a distinctive
// key). Unrecognised evidence yields UNKNOWN — never a guess, and never raw
// command output, which can carry usernames and file paths.
func ParseDevicePostureValue(
	checkKey string,
	evidence json.RawMessage,
) DevicePostureValue {
	ev := decodeEvidenceMap(evidence)
	if len(ev) == 0 {
		return unknownValue()
	}

	switch DevicePostureCheckKey(checkKey) {
	case DevicePostureCheckKeyOSVersion:
		return parseOSVersionValue(ev)
	case DevicePostureCheckKeyDiskEncryption:
		return parseDiskEncryptionValue(ev)
	case DevicePostureCheckKeyScreenLock:
		return parseScreenLockValue(ev)
	case DevicePostureCheckKeyFirewallEnabled:
		return parseFirewallValue(ev)
	case DevicePostureCheckKeyTimeSync:
		return parseTimeSyncValue(ev)
	case DevicePostureCheckKeyAutoUpdate:
		return parseAutoUpdateValue(ev)
	case DevicePostureCheckKeyPasswordPolicy:
		return parsePasswordPolicyValue(ev)
	case DevicePostureCheckKeyRemoteLogin:
		return parseRemoteLoginValue(ev)
	case DevicePostureCheckKeyMalwareProtection:
		return parseMalwareProtectionValue(ev)
	}

	return unknownValue()
}

func decodeEvidenceMap(evidence json.RawMessage) map[string]any {
	if len(evidence) == 0 {
		return nil
	}

	var ev map[string]any
	if err := json.Unmarshal(evidence, &ev); err != nil {
		return nil
	}

	return ev
}

func onOffValue(on bool) DevicePostureValue {
	if on {
		return DevicePostureValue{Kind: DevicePostureValueKindOn}
	}

	return DevicePostureValue{Kind: DevicePostureValueKindOff}
}

func unknownValue() DevicePostureValue {
	return DevicePostureValue{Kind: DevicePostureValueKindUnknown}
}

func noneValue() DevicePostureValue {
	return DevicePostureValue{Kind: DevicePostureValueKindNone}
}

func configuredValue() DevicePostureValue {
	return DevicePostureValue{Kind: DevicePostureValueKindConfigured}
}

func textValue(text string) DevicePostureValue {
	text = truncateValue(text, devicePostureValueTextMax)
	if text == "" {
		return unknownValue()
	}

	return DevicePostureValue{
		Kind: DevicePostureValueKindText,
		Text: text,
	}
}

func secondsValue(seconds int) DevicePostureValue {
	return DevicePostureValue{
		Kind:   DevicePostureValueKindSeconds,
		Number: new(seconds),
	}
}

func minPasswordLengthValue(length int) DevicePostureValue {
	return DevicePostureValue{
		Kind:   DevicePostureValueKindMinPasswordLength,
		Number: new(length),
	}
}

// backendOf returns the tool the agent used to gather the evidence. Checks that
// probe a single tool on every platform do not set it.
func backendOf(ev map[string]any) string {
	return strings.ToLower(stringEvidence(ev, "backend"))
}

// hasAnyKey discriminates platforms for checks where the agent sets no backend
// key but the key set itself is distinctive.
func hasAnyKey(ev map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := ev[key]; ok {
			return true
		}
	}

	return false
}

func lowerStringEvidence(ev map[string]any, key string) string {
	return strings.ToLower(stringEvidence(ev, key))
}

func stringEvidence(ev map[string]any, key string) string {
	v, ok := ev[key]
	if !ok || v == nil {
		return ""
	}

	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "true"
		}

		return "false"
	default:
		return ""
	}
}

func boolEvidence(ev map[string]any, key string) (bool, bool) {
	v, ok := ev[key]
	if !ok || v == nil {
		return false, false
	}

	switch typed := v.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "yes", "on":
			return true, true
		case "0", "false", "no", "off":
			return false, true
		}
	case float64:
		return typed != 0, true
	}

	return false, false
}

func numberEvidence(ev map[string]any, key string) (int, bool) {
	v, ok := ev[key]
	if !ok || v == nil {
		return 0, false
	}

	switch typed := v.(type) {
	case float64:
		return int(typed), true
	case int:
		return typed, true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}

		return n, true
	}

	return 0, false
}

func stringSliceEvidence(ev map[string]any, key string) []string {
	v, ok := ev[key]
	if !ok || v == nil {
		return nil
	}

	switch typed := v.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}

		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			s, ok := item.(string)
			if !ok {
				continue
			}

			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}

		return out
	default:
		return nil
	}
}

// stringMapEvidence reads a per-subject map such as the Windows firewall
// "profiles" or the per-user screen lock "users".
func stringMapEvidence(ev map[string]any, key string) map[string]string {
	v, ok := ev[key]
	if !ok || v == nil {
		return nil
	}

	typed, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	out := make(map[string]string, len(typed))

	for name, raw := range typed {
		s, ok := raw.(string)
		if !ok {
			continue
		}

		out[name] = strings.TrimSpace(s)
	}

	return out
}

// allValuesMatch reports whether the map is non-empty and every value equals
// want, case-insensitively.
func allValuesMatch(values map[string]string, want string) (allMatch bool, any bool) {
	if len(values) == 0 {
		return false, false
	}

	for _, v := range values {
		if !strings.EqualFold(v, want) {
			return false, true
		}
	}

	return true, true
}

func allEntriesMatch(values []string, want string) (allMatch bool, any bool) {
	if len(values) == 0 {
		return false, false
	}

	for _, v := range values {
		if !strings.EqualFold(v, want) {
			return false, true
		}
	}

	return true, true
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}

	return ""
}

// parseAssignedInt finds assignments like "minpasswordlen=8".
func parseAssignedInt(raw, key string) (int, bool) {
	if raw == "" || key == "" {
		return 0, false
	}

	lower := strings.ToLower(raw)
	key = strings.ToLower(key) + "="

	idx := strings.Index(lower, key)
	if idx < 0 {
		return 0, false
	}

	return parseLeadingInt(raw[idx+len(key):])
}

func parseLeadingInt(s string) (int, bool) {
	s = strings.TrimLeft(s, " \t")
	if s == "" {
		return 0, false
	}

	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}

	if end == 0 {
		return 0, false
	}

	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0, false
	}

	return n, true
}

// truncateValue collapses whitespace and caps the length on a rune boundary so
// the result stays valid UTF-8 for JSON encoding.
func truncateValue(v string, max int) string {
	v = strings.Join(strings.Fields(v), " ")
	if len(v) <= max {
		return v
	}

	cut := max - len("…")
	for cut > 0 && !utf8.RuneStart(v[cut]) {
		cut--
	}

	return v[:cut] + "…"
}
