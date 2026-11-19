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
	linuxPAMDir             = "/etc/pam.d"

	// linuxPAMIncludeDepth matches the point at which a PAM include cycle
	// is no longer a readable stack.
	linuxPAMIncludeDepth = 16

	// linuxPasswordMinLen is the shortest minimum that still counts as a policy.
	linuxPasswordMinLen = 8

	// Defaults from pam_pwquality(8) and pam_cracklib(8) when minlen is unset.
	linuxPwqualityDefaultMinLen = 8
	linuxCracklibDefaultMinLen  = 9

	// libpwquality raises any lower minlen to PWQ_BASE_MIN_LENGTH.
	linuxPwqualityBaseMinLen = 6
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
	pamFiles := map[string]string{}

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
		pamFiles[filepath.Base(path)] = body
	}

	if readErr := readPAMIncludes(pamFiles); readErr != nil {
		return unknown(map[string]any{"error": readErr.Error()})
	}

	loginDefs, _, err := readOptionalFile("/etc/login.defs")
	if err != nil {
		return unknown(map[string]any{"error": err.Error()})
	}

	return linuxPasswordPolicyFromSources(conf, dropins, pamBodies, pamFiles, loginDefs)
}

// readPAMIncludes loads files named by include and substack lines. A name
// that is not a single pam.d entry is left unresolved.
func readPAMIncludes(files map[string]string) error {
	seen := map[string]bool{}

	for range linuxPAMIncludeDepth {
		var names []string

		for _, body := range files {
			names = append(names, pamIncludeNames(body)...)
		}

		loaded := false

		for _, name := range names {
			if !pamConfigName(name) || seen[name] {
				continue
			}

			seen[name] = true
			if _, ok := files[name]; ok {
				continue
			}

			body, present, err := readOptionalFile(filepath.Join(linuxPAMDir, name))
			if err != nil {
				return fmt.Errorf("cannot read pam include: %w", err)
			}

			if !present {
				continue
			}

			files[name] = body
			loaded = true
		}

		if !loaded {
			return nil
		}
	}

	return nil
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
// Include and substack lines use pamIncludes, keyed by the pam.d basename.
func linuxPasswordPolicyFromSources(
	conf string,
	dropins []linuxPwqualityDropin,
	pamBodies []string,
	pamIncludes map[string]string,
	loginDefs string,
) Result {
	confValue, confSource, confKnown, confInvalid := mergedPwqualityMinLen(conf, dropins)
	maxDays := loginDefsLookup(loginDefs, "PASS_MAX_DAYS")

	if len(pamBodies) == 0 {
		return linuxPasswordPolicyResult(loginDefsPasswordReading(loginDefs, true), maxDays)
	}

	readings := make([]linuxPasswordReading, 0, len(pamBodies))
	for _, body := range pamBodies {
		modules, resolved := pamPasswordQualityModules(body, pamIncludes)
		if !resolved {
			ev := map[string]any{
				"note": "password stack could not be resolved",
			}
			if maxDays != "" {
				ev["pass_max_days"] = maxDays
			}

			return unknown(ev)
		}

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
	cracklib := m.name == "pam_cracklib.so"
	if m.invalid || (cracklib && m.hasMinLen && m.minLen < 0) {
		return linuxPasswordReading{}, false
	}

	if m.hasMinLen {
		value := m.minLen
		if !cracklib {
			value = max(value, linuxPwqualityBaseMinLen)
		}

		return linuxPasswordReading{
			value:  value,
			source: linuxPasswordSourcePAM,
			known:  true,
		}, true
	}

	if cracklib {
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
			value:  max(confValue, linuxPwqualityBaseMinLen),
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

// libpwquality reads the drop-ins first and pwquality.conf last, so the
// main file wins.
func mergedPwqualityMinLen(
	conf string,
	dropins []linuxPwqualityDropin,
) (int, string, bool, bool) {
	ordered := slices.Clone(dropins)
	slices.SortFunc(ordered, func(a, b linuxPwqualityDropin) int {
		return strings.Compare(a.Name, b.Name)
	})

	parts := make([]linuxPwqualityDropin, 0, len(ordered)+1)
	for _, dropin := range ordered {
		parts = append(parts, linuxPwqualityDropin{
			Name: linuxPasswordSourcePwqualityDir,
			Body: dropin.Body,
		})
	}

	parts = append(parts, linuxPwqualityDropin{
		Name: linuxPasswordSourcePwqualityConf,
		Body: conf,
	})

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
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "minlen") {
			continue
		}

		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
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

func pamPasswordQualityModules(
	body string,
	includes map[string]string,
) ([]pamQualityModule, bool) {
	return pamPasswordModules(body, includes, map[string]bool{}, 0)
}

func pamPasswordModules(
	body string,
	includes map[string]string,
	stack map[string]bool,
	depth int,
) ([]pamQualityModule, bool) {
	if depth > linuxPAMIncludeDepth {
		return nil, false
	}

	var modules []pamQualityModule

	body = strings.ReplaceAll(body, "\\\n", " ")
	for line := range strings.SplitSeq(body, "\n") {
		name, included := pamIncludeName(line)
		if included {
			if !pamConfigName(name) || stack[name] {
				return nil, false
			}

			nested, ok := includes[name]
			if !ok {
				return nil, false
			}

			stack[name] = true
			found, ok := pamPasswordModules(nested, includes, stack, depth+1)
			delete(stack, name)

			if !ok {
				return nil, false
			}

			modules = append(modules, found...)

			continue
		}

		module, ok := pamPasswordQualityModule(line)
		if !ok {
			continue
		}

		modules = append(modules, module)
	}

	return modules, true
}

func pamIncludeNames(body string) []string {
	body = strings.ReplaceAll(body, "\\\n", " ")

	var names []string

	for line := range strings.SplitSeq(body, "\n") {
		name, ok := pamIncludeName(line)
		if !ok {
			continue
		}

		names = append(names, name)
	}

	return names
}

// pamIncludeName matches @include, and password include or substack.
func pamIncludeName(line string) (string, bool) {
	if hash := strings.IndexByte(line, '#'); hash >= 0 {
		line = line[:hash]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return "", false
	}

	if strings.HasPrefix(strings.ToLower(line), "@include") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.EqualFold(fields[0], "@include") {
			return filepath.Base(fields[1]), true
		}

		return "", false
	}

	tokens := pamTokens(line)
	if len(tokens) < 3 {
		return "", false
	}

	if !strings.EqualFold(strings.TrimPrefix(tokens[0], "-"), "password") {
		return "", false
	}

	if !strings.EqualFold(tokens[1], "include") && !strings.EqualFold(tokens[1], "substack") {
		return "", false
	}

	return filepath.Base(tokens[2]), true
}

func pamConfigName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.Contains(name, "/")
}

// pamPasswordQualityModule returns a quality module only when its control
// rejects the password on failure; optional and sufficient lines enforce nothing.
func pamPasswordQualityModule(line string) (pamQualityModule, bool) {
	if hash := strings.IndexByte(line, '#'); hash >= 0 {
		line = line[:hash]
	}

	tokens := pamTokens(strings.TrimSpace(line))
	if len(tokens) < 3 {
		return pamQualityModule{}, false
	}

	if !strings.EqualFold(strings.TrimPrefix(tokens[0], "-"), "password") {
		return pamQualityModule{}, false
	}

	if !pamControlRejectsFailure(tokens[1]) {
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

// pamControlRejectsFailure mirrors libpam: bracket values left unset
// default to bad, and quality modules fail with authtok_err.
func pamControlRejectsFailure(control string) bool {
	switch strings.ToLower(control) {
	case "required", "requisite":
		return true
	}

	if !strings.HasPrefix(control, "[") {
		return false
	}

	actions := map[string]string{}

	for pair := range strings.FieldsSeq(strings.Trim(control, "[]")) {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}

		actions[strings.ToLower(key)] = strings.ToLower(value)
	}

	action, ok := actions["authtok_err"]
	if !ok {
		action, ok = actions["default"]
	}

	if !ok {
		return true
	}

	return action == "bad" || action == "die"
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
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "minlen") {
			continue
		}

		found = true

		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
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
