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
	"encoding/xml"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Password length sources. The server keys on these names, so a rename is a
// wire change.
const (
	passwordSourceNone          = "none"
	passwordSourcePwquality     = "pam_pwquality"
	passwordSourceCracklib      = "pam_cracklib"
	passwordSourcePasswdqc      = "pam_passwdqc"
	passwordSourcePAMUnix       = "pam_unix"
	passwordSourceLoginDefs     = "login_defs"
	passwordSourcePwpolicy      = "pwpolicy"
	passwordSourceLoginConf     = "login_conf"
	passwordSourceSecedit       = "secedit"
	passwordSourceMDMDeviceLock = "mdm_device_lock"
)

const (
	// pwqualityDefaultMinLen is the lowest default libpwquality has shipped
	// (8; older releases used 9), so an active module never overstates.
	pwqualityDefaultMinLen = 8
	// pamUnixDefaultMinLen is pam_unix's built-in minimum.
	pamUnixDefaultMinLen = 6
)

// passwordLengths records the minimum length each source enforces.
type passwordLengths map[string]int

// strictest returns the source with the longest minimum. Every source listed
// is enforced on its own, so the longest is the effective rule. Ties resolve
// by name so the reported source is stable.
func (l passwordLengths) strictest() (string, int, bool) {
	var (
		source string
		length int
		found  bool
	)

	for name, n := range l {
		if !found || n > length || (n == length && name < source) {
			source, length, found = name, n, true
		}
	}

	return source, length, found
}

// passwordPolicyEvidence shapes the evidence every platform sends: the
// effective minimum, the source that set it, and every source found.
func passwordPolicyEvidence(lengths passwordLengths) (map[string]any, int) {
	ev := map[string]any{"sources": map[string]int(lengths)}

	source, length, found := lengths.strictest()
	if !found {
		ev["backend"] = passwordSourceNone
		ev["min_password_length"] = 0

		return ev, 0
	}

	ev["backend"] = source
	ev["min_password_length"] = length

	return ev, length
}

// passwordPolicyResult reports whether a length rule exists at all. Whether
// the length is long enough is decided by the server.
func passwordPolicyResult(ev map[string]any, length int) Result {
	if length > 0 {
		return pass(ev)
	}

	return fail(ev)
}

// pamPasswordRule is one "password" line of a PAM service file.
type pamPasswordRule struct {
	module string
	args   map[string]string
}

// parsePAMPasswordRules returns the password-type rules of a PAM service file.
// The control field may be a bracketed list with spaces, and a leading "-"
// on the type only silences a missing module.
func parsePAMPasswordRules(body string) []pamPasswordRule {
	var rules []pamPasswordRule

	for line := range strings.SplitSeq(body, "\n") {
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		fields := strings.Fields(line)
		if len(fields) < 3 || strings.TrimPrefix(fields[0], "-") != "password" {
			continue
		}

		rest := fields[1:]
		if strings.HasPrefix(rest[0], "[") {
			for len(rest) > 0 && !strings.HasSuffix(rest[0], "]") {
				rest = rest[1:]
			}

			if len(rest) == 0 {
				continue
			}
		}

		rest = rest[1:]
		if len(rest) == 0 {
			continue
		}

		rule := pamPasswordRule{
			module: strings.TrimSuffix(path.Base(rest[0]), ".so"),
			args:   map[string]string{},
		}
		for _, arg := range rest[1:] {
			key, value, _ := strings.Cut(arg, "=")
			rule.args[key] = value
		}

		rules = append(rules, rule)
	}

	return rules
}

// parsePwqualityConfMinLen reads minlen from pwquality.conf syntax. Callers
// join the main file and the conf.d drop-ins in load order, so the last
// assignment wins as it does in libpwquality.
func parsePwqualityConfMinLen(body string) (int, bool) {
	var (
		minLen int
		found  bool
	)

	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "minlen" {
			continue
		}

		if n, ok := trimmedInt(value); ok {
			minLen, found = n, true
		}
	}

	return minLen, found
}

// pamPasswordLengths resolves the minimum each length-enforcing module in the
// PAM password stacks applies. Module arguments override pwquality.conf.
func pamPasswordLengths(rules []pamPasswordRule, pwqualityConf string) passwordLengths {
	lengths := passwordLengths{}
	confMinLen, confFound := parsePwqualityConfMinLen(pwqualityConf)

	setMax := func(source string, n int) {
		if current, ok := lengths[source]; !ok || n > current {
			lengths[source] = n
		}
	}

	for _, rule := range rules {
		minLen, hasMinLen := trimmedInt(rule.args["minlen"])

		switch rule.module {
		case "pam_pwquality":
			switch {
			case hasMinLen:
				setMax(passwordSourcePwquality, minLen)
			case confFound:
				setMax(passwordSourcePwquality, confMinLen)
			default:
				setMax(passwordSourcePwquality, pwqualityDefaultMinLen)
			}
		case "pam_cracklib":
			// Without minlen the default depends on credit settings that can
			// lower it, so only an explicit value is reported.
			if hasMinLen {
				setMax(passwordSourceCracklib, minLen)
			}
		case "pam_passwdqc":
			if n, ok := parsePasswdqcMin(rule.args["min"]); ok {
				setMax(passwordSourcePasswdqc, n)
			}
		case "pam_unix":
			if hasMinLen {
				setMax(passwordSourcePAMUnix, minLen)
			} else {
				setMax(passwordSourcePAMUnix, pamUnixDefaultMinLen)
			}
		}
	}

	return lengths
}

// parsePasswdqcMin reads "min=N0,N1,N2,N3,N4". Each field is the minimum for
// passwords of a given character-class mix, so the shortest enabled one is
// the shortest password passwdqc accepts.
func parsePasswdqcMin(s string) (int, bool) {
	if s == "" {
		return 0, false
	}

	var (
		shortest int
		found    bool
	)

	for field := range strings.SplitSeq(s, ",") {
		n, ok := trimmedInt(field)
		if !ok {
			continue
		}

		if !found || n < shortest {
			shortest, found = n, true
		}
	}

	return shortest, found
}

// parseLoginDefsMinLen reads PASS_MIN_LEN from /etc/login.defs.
func parseLoginDefsMinLen(body string) (int, bool) {
	return trimmedInt(loginDefsLookup(body, "PASS_MIN_LEN"))
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

var pwpolicyLengthPattern = regexp.MustCompile(`\.\{(\d+)`)

// parsePwpolicyMinLengths reads `pwpolicy -getaccountpolicies` and returns
// the minimum length of every password content rule. Rules are regular
// expressions such as ".{8,}"; configuration profiles also carry minLength or
// minChars parameters. Leading text before the plist is ignored.
func parsePwpolicyMinLengths(out string) ([]int, error) {
	start := strings.Index(out, "<?xml")
	if start < 0 {
		start = strings.Index(out, "<plist")
	}

	if start < 0 {
		return nil, nil
	}

	decoder := xml.NewDecoder(strings.NewReader(out[start:]))
	decoder.Strict = false

	var (
		lengths []int
		lastKey string
		inKey   bool
		inValue string
	)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return lengths, nil
		}

		if err != nil {
			return lengths, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "key":
				inKey = true
			case "string", "integer":
				inValue = t.Name.Local
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "key":
				inKey = false
			case "string", "integer":
				inValue = ""
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))

			switch {
			case inKey:
				lastKey = text
			case inValue == "string" && lastKey == "policyContent":
				if !strings.Contains(text, "policyAttributePassword") {
					continue
				}

				for _, m := range pwpolicyLengthPattern.FindAllStringSubmatch(text, -1) {
					if n, err := strconv.Atoi(m[1]); err == nil {
						lengths = append(lengths, n)
					}
				}
			case inValue == "integer" && (lastKey == "minimumLength" || lastKey == "minChars"):
				if n, err := strconv.Atoi(text); err == nil {
					lengths = append(lengths, n)
				}
			}
		}
	}
}

// parseLoginConfMinPasswordLen reads minpasswordlen from the "default" class
// of /etc/login.conf. Entries span lines joined by a trailing backslash and
// list their names before the first colon, separated by "|".
func parseLoginConfMinPasswordLen(body string) (int, bool) {
	entries := termcapEntries(body)

	for _, entry := range entries {
		names, caps, ok := strings.Cut(entry, ":")
		if !ok || !loginConfHasName(names, "default") {
			continue
		}

		for capability := range strings.SplitSeq(caps, ":") {
			key, value, ok := strings.Cut(strings.TrimSpace(capability), "=")
			if ok && key == "minpasswordlen" {
				return trimmedInt(value)
			}
		}

		return 0, false
	}

	return 0, false
}

func loginConfHasName(names, want string) bool {
	for name := range strings.SplitSeq(names, "|") {
		if strings.TrimSpace(name) == want {
			return true
		}
	}

	return false
}

// termcapEntries splits a termcap-style database (gettytab, login.conf) into
// entries, joining backslash continuations and skipping comments.
func termcapEntries(body string) []string {
	var (
		entries []string
		current strings.Builder
	)

	for line := range strings.SplitSeq(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if current.Len() == 0 && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			continue
		}

		if strings.HasSuffix(trimmed, `\`) {
			current.WriteString(strings.TrimSuffix(trimmed, `\`))
			continue
		}

		current.WriteString(trimmed)
		entries = append(entries, current.String())
		current.Reset()
	}

	if current.Len() > 0 {
		entries = append(entries, current.String())
	}

	return entries
}
