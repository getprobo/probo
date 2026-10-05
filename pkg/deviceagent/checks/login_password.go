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
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Auto-login sources. The server reports their names as-is, so a rename is a
// wire change.
const (
	autoLoginSourceWinlogon    = "winlogon"
	autoLoginSourceLoginwindow = "loginwindow"
	autoLoginSourceGDM         = "gdm"
	autoLoginSourceLightDM     = "lightdm"
	autoLoginSourceSDDM        = "sddm"
	autoLoginSourceGetty       = "getty"
)

// loginPasswordEvidence shapes the evidence every platform sends. A count
// below zero means the platform cannot observe it and is left out.
func loginPasswordEvidence(autoLoginSources []string, accountsWithoutPassword int) map[string]any {
	if autoLoginSources == nil {
		autoLoginSources = []string{}
	}

	ev := map[string]any{
		"auto_login":         len(autoLoginSources) > 0,
		"auto_login_sources": autoLoginSources,
	}
	if accountsWithoutPassword >= 0 {
		ev["accounts_without_password"] = accountsWithoutPassword
	}

	return ev
}

// loginPasswordResult passes when nothing opens a session without a password.
func loginPasswordResult(ev map[string]any, autoLoginSources []string, accountsWithoutPassword int) Result {
	if len(autoLoginSources) > 0 || accountsWithoutPassword > 0 {
		return fail(ev)
	}

	return pass(ev)
}

// parseINISections reads key=value pairs per [section] from the INI dialect
// GDM, LightDM and SDDM share. Keys keep their case; later files passed in the
// same body override earlier ones.
func parseINISections(body string) map[string]map[string]string {
	sections := map[string]map[string]string{}
	section := ""

	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if sections[section] == nil {
			sections[section] = map[string]string{}
		}

		sections[section][strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return sections
}

func iniTrue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes":
		return true
	}

	return false
}

// gdmAutoLogin reports whether GDM logs a user in without a password, either
// immediately or after the timed-login delay.
func gdmAutoLogin(body string) bool {
	daemon := parseINISections(body)["daemon"]

	return iniTrue(daemon["AutomaticLoginEnable"]) || iniTrue(daemon["TimedLoginEnable"])
}

// lightdmAutoLogin reports whether any seat section names an autologin user.
// Sections are [Seat:*], [Seat:seat0] or the legacy [SeatDefaults].
func lightdmAutoLogin(body string) bool {
	for name, values := range parseINISections(body) {
		if !strings.HasPrefix(name, "Seat") {
			continue
		}

		if values["autologin-user"] != "" {
			return true
		}
	}

	return false
}

// sddmAutoLogin reports whether the [Autologin] section names a user.
func sddmAutoLogin(body string) bool {
	return parseINISections(body)["Autologin"]["User"] != ""
}

// gettyAutoLogin reports whether a systemd getty drop-in starts agetty with
// --autologin (or its -a short form).
func gettyAutoLogin(body string) bool {
	for line := range strings.SplitSeq(body, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.TrimSpace(key) != "ExecStart" {
			continue
		}

		for _, arg := range strings.Fields(value) {
			if arg == "-a" || arg == "--autologin" || strings.HasPrefix(arg, "--autologin=") {
				return true
			}
		}
	}

	return false
}

// loginShell reports whether a shell lets the account open a session.
func loginShell(shell string) bool {
	shell = strings.TrimSpace(shell)
	if shell == "" {
		// login(1) falls back to /bin/sh.
		return true
	}

	switch path := shell[strings.LastIndex(shell, "/")+1:]; path {
	case "nologin", "false":
		return false
	}

	return true
}

// countLinuxAccountsWithoutPassword counts accounts with a login shell whose
// password field is empty. The field lives in /etc/shadow when /etc/passwd
// holds "x". Locked ("!", "*") fields are not empty. Only the count leaves
// this function; hashes are never retained.
func countLinuxAccountsWithoutPassword(passwd, shadow string) int {
	empty := map[string]bool{}

	for line := range strings.SplitSeq(shadow, "\n") {
		fields := strings.Split(strings.TrimSpace(line), ":")
		if len(fields) >= 2 && fields[0] != "" {
			empty[fields[0]] = fields[1] == ""
		}
	}

	count := 0

	for line := range strings.SplitSeq(passwd, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "+") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 7 || !loginShell(fields[6]) {
			continue
		}

		switch fields[1] {
		case "":
			count++
		case "x":
			if empty[fields[0]] {
				count++
			}
		}
	}

	return count
}

// countFreeBSDAccountsWithoutPassword counts master.passwd accounts with a
// login shell and an empty password field. Fields are name, password, uid,
// gid, class, change, expire, gecos, home, shell.
func countFreeBSDAccountsWithoutPassword(masterPasswd string) int {
	count := 0

	for line := range strings.SplitSeq(masterPasswd, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 10 || fields[1] != "" || !loginShell(fields[9]) {
			continue
		}

		count++
	}

	return count
}

// freebsdGettyAutoLogin reports whether a terminal enabled in /etc/ttys runs
// getty with a gettytab entry that sets al= (autologin user).
func freebsdGettyAutoLogin(gettytab, ttys string) bool {
	autoLogin := map[string]bool{}

	for _, entry := range termcapEntries(gettytab) {
		names, caps, ok := strings.Cut(entry, ":")
		if !ok {
			continue
		}

		hasAutoLogin := false

		for capability := range strings.SplitSeq(caps, ":") {
			key, value, ok := strings.Cut(strings.TrimSpace(capability), "=")
			if ok && key == "al" && value != "" {
				hasAutoLogin = true
			}
		}

		if !hasAutoLogin {
			continue
		}

		for name := range strings.SplitSeq(names, "|") {
			autoLogin[strings.TrimSpace(name)] = true
		}
	}

	if len(autoLogin) == 0 {
		return false
	}

	for line := range strings.SplitSeq(ttys, "\n") {
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		// name "command type" type status ...
		open := strings.Index(line, `"`)
		closing := strings.LastIndex(line, `"`)
		if open < 0 || closing <= open {
			continue
		}

		command := strings.Fields(line[open+1 : closing])
		rest := strings.Fields(line[closing+1:])

		if len(command) < 2 || !strings.HasSuffix(command[0], "getty") || len(rest) < 2 {
			continue
		}

		// onifexists is how the stock ttys enables the virtual consoles.
		if (rest[1] == "on" || rest[1] == "onifexists") && autoLogin[command[1]] {
			return true
		}
	}

	return false
}

// parsePlistTopLevel returns the top-level string and boolean values of an
// XML property list, such as `defaults export` prints. Nested dicts are
// skipped.
func parsePlistTopLevel(out string) (map[string]string, map[string]bool, error) {
	strs := map[string]string{}
	bools := map[string]bool{}

	start := strings.Index(out, "<plist")
	if start < 0 {
		return strs, bools, nil
	}

	decoder := xml.NewDecoder(strings.NewReader(out[start:]))
	decoder.Strict = false

	var (
		depth   int
		key     string
		inKey   bool
		inValue bool
	)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return strs, bools, nil
		}

		if err != nil {
			return strs, bools, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "dict", "array":
				depth++
			case "key":
				inKey = depth == 1
			case "string":
				inValue = depth == 1
			case "true", "false":
				if depth == 1 && key != "" {
					bools[key] = t.Name.Local == "true"
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "dict", "array":
				depth--
			case "key":
				inKey = false
			case "string":
				inValue = false
			}
		case xml.CharData:
			switch {
			case inKey:
				key = strings.TrimSpace(string(t))
			case inValue && key != "":
				strs[key] = strings.TrimSpace(string(t))
			}
		}
	}
}

// displayManagerAutoLoginFiles maps each display manager to its configuration,
// relative to the configuration root ("/" on Linux, "/usr/local/" on FreeBSD).
// Vendor drop-ins come first so local files override them.
var displayManagerAutoLoginFiles = []struct {
	source string
	globs  []string
	parse  func(string) bool
}{
	{
		source: autoLoginSourceGDM,
		globs:  []string{"etc/gdm3/custom.conf", "etc/gdm3/daemon.conf", "etc/gdm/custom.conf"},
		parse:  gdmAutoLogin,
	},
	{
		source: autoLoginSourceLightDM,
		globs: []string{
			"share/lightdm/lightdm.conf.d/*.conf",
			"usr/share/lightdm/lightdm.conf.d/*.conf",
			"etc/lightdm/lightdm.conf",
			"etc/lightdm/lightdm.conf.d/*.conf",
		},
		parse: lightdmAutoLogin,
	},
	{
		source: autoLoginSourceSDDM,
		globs: []string{
			"lib/sddm/sddm.conf.d/*.conf",
			"usr/lib/sddm/sddm.conf.d/*.conf",
			"etc/sddm.conf",
			"etc/sddm.conf.d/*.conf",
		},
		parse: sddmAutoLogin,
	},
}

// displayManagerAutoLogin returns the display managers under root configured
// to log a user in without a password. Each manager's files are joined in
// load order so a later file can switch auto-login back off.
func displayManagerAutoLogin(root string) []string {
	var sources []string

	for _, dm := range displayManagerAutoLoginFiles {
		var body strings.Builder

		for _, glob := range dm.globs {
			matches, _ := filepath.Glob(filepath.Join(root, glob))
			sort.Strings(matches)

			for _, name := range matches {
				if data, err := os.ReadFile(name); err == nil {
					body.Write(data)
					body.WriteString("\n")
				}
			}
		}

		if dm.parse(body.String()) {
			sources = append(sources, dm.source)
		}
	}

	return sources
}
