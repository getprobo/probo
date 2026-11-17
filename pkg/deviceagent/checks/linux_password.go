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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	linuxPwqualityConfPath  = "/etc/security/pwquality.conf"
	linuxPwqualityDropinDir = "/etc/security/pwquality.conf.d"

	// linuxPasswordMinLen is the shortest minimum that still counts as a policy.
	linuxPasswordMinLen = 8

	// Defaults from pam_pwquality(8) and pam_cracklib(8) when minlen is unset.
	linuxPwqualityDefaultMinLen = 8
	linuxCracklibDefaultMinLen  = 9
)

const (
	linuxPasswordSourcePwqualityConf = "pwquality.conf"
	linuxPasswordSourcePwqualityDir  = "pwquality.conf.d"
	linuxPasswordSourcePAM           = "pam"
	linuxPasswordSourceLoginDefs     = "login.defs"
	linuxPasswordSourcePwquality     = "pam_pwquality"
	linuxPasswordSourceCracklib      = "pam_cracklib"
	linuxPasswordSourceNone          = "none"
)

// Debian uses common-password. Fedora and RHEL split local changes
// (system-auth) from remote ones (password-auth).
var linuxPAMPasswordPaths = []string{
	"/etc/pam.d/common-password",
	"/etc/pam.d/system-auth",
	"/etc/pam.d/password-auth",
}

type (
	linuxPwqualityDropin struct {
		Name string
		Body string
	}

	linuxPasswordReading struct {
		value  int
		source string
		known  bool
		broken bool
	}

	pamQualityModule struct {
		name      string
		minLen    int
		hasMinLen bool
		invalid   bool
	}
)

func linuxPasswordPolicy(ctx context.Context) Result {
	conf, _, err := readOptionalFile(linuxPwqualityConfPath)
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}

	dropins, err := readPwqualityDropins(linuxPwqualityDropinDir)
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}

	pamBodies := make([]string, 0, len(linuxPAMPasswordPaths))
	for _, path := range linuxPAMPasswordPaths {
		body, present, readErr := readOptionalFile(path)
		if readErr != nil {
			return unknown(
				map[string]any{
					"error": readErr.Error(),
					"path":  path,
				},
			)
		}

		if !present {
			continue
		}

		pamBodies = append(pamBodies, body)
	}

	loginDefs, _, err := readOptionalFile("/etc/login.defs")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}

	return linuxPasswordPolicyFromSources(conf, dropins, pamBodies, loginDefs)
}

func readOptionalFile(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("cannot read %s: %w", path, err)
	}

	return string(data), true, nil
}

func readPwqualityDropins(dir string) ([]linuxPwqualityDropin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot list pwquality drop-ins: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}

		names = append(names, entry.Name())
	}

	slices.Sort(names)

	dropins := make([]linuxPwqualityDropin, 0, len(names))
	for _, name := range names {
		body, _, readErr := readOptionalFile(filepath.Join(dir, name))
		if readErr != nil {
			return nil, fmt.Errorf("cannot read pwquality drop-in: %w", readErr)
		}

		dropins = append(dropins, linuxPwqualityDropin{Name: name, Body: body})
	}

	return dropins, nil
}

// linuxPasswordPolicyFromSources uses the weakest password-change path.
// login.defs applies only when the stack loads neither quality module.
func linuxPasswordPolicyFromSources(
	conf string,
	dropins []linuxPwqualityDropin,
	pamBodies []string,
	loginDefs string,
) Result {
	confValue, confSource, confKnown, confInvalid := mergedPwqualityMinLen(conf, dropins)
	maxDays := loginDefsLookup(loginDefs, "PASS_MAX_DAYS")

	if len(pamBodies) == 0 {
		return linuxPasswordPolicyResult(loginDefsPasswordReading(loginDefs, true), maxDays)
	}

	readings := make([]linuxPasswordReading, 0, len(pamBodies))
	for _, body := range pamBodies {
		modules := pamPasswordQualityModules(body)
		if len(modules) == 0 {
			readings = append(readings, loginDefsPasswordReading(loginDefs, false))
			continue
		}

		reading, ok := strictestPasswordModule(
			modules,
			confValue,
			confSource,
			confKnown,
			confInvalid,
		)
		if !ok {
			return linuxPasswordPolicyResult(linuxPasswordReading{broken: true}, maxDays)
		}

		readings = append(readings, reading)
	}

	return linuxPasswordPolicyResult(weakestPasswordReading(readings), maxDays)
}

func linuxPasswordPolicyResult(reading linuxPasswordReading, maxDays string) Result {
	ev := map[string]any{}
	if maxDays != "" {
		ev["pass_max_days"] = maxDays
	}

	if reading.broken {
		ev["parse_error"] = "invalid password minimum length"
		return unknown(ev)
	}

	if !reading.known {
		ev["note"] = "password minimum length could not be determined"
		return unknown(ev)
	}

	ev["source"] = reading.source
	ev["pass_min_len"] = strconv.Itoa(reading.value)
	ev["pass_min_len_value"] = reading.value

	switch linuxPasswordPolicyVerdict(reading.value, true) {
	case StatusPass:
		return pass(ev)
	case StatusFail:
		return fail(ev)
	default:
		return unknown(ev)
	}
}

func loginDefsPasswordReading(body string, absentIsUnknown bool) linuxPasswordReading {
	raw := loginDefsLookup(body, "PASS_MIN_LEN")
	if raw == "" {
		if absentIsUnknown {
			return linuxPasswordReading{}
		}

		return linuxPasswordReading{
			source: linuxPasswordSourceNone,
			known:  true,
		}
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return linuxPasswordReading{broken: true}
	}

	return linuxPasswordReading{
		value:  value,
		source: linuxPasswordSourceLoginDefs,
		known:  true,
	}
}

func weakestPasswordReading(readings []linuxPasswordReading) linuxPasswordReading {
	var weakest linuxPasswordReading

	for i, reading := range readings {
		if reading.broken || !reading.known {
			return reading
		}

		if i == 0 || reading.value < weakest.value {
			weakest = reading
		}
	}

	return weakest
}

func strictestPasswordModule(
	modules []pamQualityModule,
	confValue int,
	confSource string,
	confKnown bool,
	confInvalid bool,
) (linuxPasswordReading, bool) {
	var strictest linuxPasswordReading

	for i, module := range modules {
		reading, ok := module.reading(confValue, confSource, confKnown, confInvalid)
		if !ok {
			return linuxPasswordReading{}, false
		}

		if i == 0 || reading.value > strictest.value {
			strictest = reading
		}
	}

	return strictest, len(modules) > 0
}

func (m pamQualityModule) reading(
	confValue int,
	confSource string,
	confKnown bool,
	confInvalid bool,
) (linuxPasswordReading, bool) {
	if m.invalid || (m.hasMinLen && m.minLen < 0) {
		return linuxPasswordReading{}, false
	}

	if m.hasMinLen {
		return linuxPasswordReading{
			value:  m.minLen,
			source: linuxPasswordSourcePAM,
			known:  true,
		}, true
	}

	if m.name == "pam_cracklib.so" {
		return linuxPasswordReading{
			value:  linuxCracklibDefaultMinLen,
			source: linuxPasswordSourceCracklib,
			known:  true,
		}, true
	}

	if confInvalid {
		return linuxPasswordReading{}, false
	}

	if confKnown {
		return linuxPasswordReading{
			value:  confValue,
			source: confSource,
			known:  true,
		}, true
	}

	return linuxPasswordReading{
		value:  linuxPwqualityDefaultMinLen,
		source: linuxPasswordSourcePwquality,
		known:  true,
	}, true
}

func mergedPwqualityMinLen(
	conf string,
	dropins []linuxPwqualityDropin,
) (int, string, bool, bool) {
	parts := []linuxPwqualityDropin{{
		Name: linuxPasswordSourcePwqualityConf,
		Body: conf,
	}}

	ordered := slices.Clone(dropins)
	slices.SortFunc(ordered, func(a, b linuxPwqualityDropin) int {
		return strings.Compare(a.Name, b.Name)
	})

	for _, dropin := range ordered {
		parts = append(parts, linuxPwqualityDropin{
			Name: linuxPasswordSourcePwqualityDir,
			Body: dropin.Body,
		})
	}

	var (
		value   int
		source  string
		known   bool
		invalid bool
	)

	for _, part := range parts {
		parsed, ok, bad := pwqualityMinLen(part.Body)
		if bad {
			value = 0
			source = part.Name
			known = false
			invalid = true

			continue
		}

		if !ok {
			continue
		}

		value = parsed
		source = part.Name
		known = true
		invalid = false
	}

	return value, source, known, invalid
}

func pwqualityMinLen(body string) (int, bool, bool) {
	found := false
	invalid := false
	value := 0

	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if hash := strings.IndexByte(line, '#'); hash >= 0 {
			line = strings.TrimSpace(line[:hash])
		}

		key, raw, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "minlen" {
			continue
		}

		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || parsed < 0 {
			found = true
			invalid = true

			continue
		}

		value = parsed
		found = true
		invalid = false
	}

	if invalid {
		return 0, false, true
	}

	return value, found, false
}

func pamPasswordQualityModules(body string) []pamQualityModule {
	var modules []pamQualityModule

	for line := range strings.SplitSeq(body, "\n") {
		module, ok := pamPasswordQualityModule(line)
		if !ok {
			continue
		}

		modules = append(modules, module)
	}

	return modules
}

func pamPasswordQualityModule(line string) (pamQualityModule, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return pamQualityModule{}, false
	}

	tokens := pamTokens(line)
	if len(tokens) < 3 {
		return pamQualityModule{}, false
	}

	if strings.TrimPrefix(tokens[0], "-") != "password" {
		return pamQualityModule{}, false
	}

	if tokens[1] == "include" || tokens[1] == "substack" {
		return pamQualityModule{}, false
	}

	name := pamModuleBase(tokens[2])
	if name != "pam_pwquality.so" && name != "pam_cracklib.so" {
		return pamQualityModule{}, false
	}

	minLen, hasMinLen, invalid := pamMinLenArg(tokens[3:])

	return pamQualityModule{
		name:      name,
		minLen:    minLen,
		hasMinLen: hasMinLen,
		invalid:   invalid,
	}, true
}

func pamModuleBase(token string) string {
	token = strings.TrimPrefix(token, "-")
	if slash := strings.LastIndex(token, "/"); slash >= 0 {
		return token[slash+1:]
	}

	return token
}

func pamMinLenArg(args []string) (int, bool, bool) {
	found := false
	invalid := false
	value := 0

	for _, arg := range args {
		key, raw, ok := strings.Cut(arg, "=")
		if !ok || strings.TrimSpace(key) != "minlen" {
			continue
		}

		found = true

		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || parsed < 0 {
			invalid = true
			continue
		}

		value = parsed
		invalid = false
	}

	return value, found, invalid
}

func pamTokens(line string) []string {
	var tokens []string

	for i := 0; i < len(line); {
		if line[i] == ' ' || line[i] == '\t' {
			i++
			continue
		}

		if line[i] == '[' {
			end := strings.IndexByte(line[i:], ']')
			if end < 0 {
				tokens = append(tokens, line[i:])
				break
			}

			tokens = append(tokens, line[i:i+end+1])
			i += end + 1

			continue
		}

		start := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' {
			i++
		}

		tokens = append(tokens, line[start:i])
	}

	return tokens
}
