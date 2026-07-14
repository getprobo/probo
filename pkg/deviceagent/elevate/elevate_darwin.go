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

//go:build darwin

package elevate

import (
	"fmt"
	"os/exec"
	"strings"

	"go.probo.inc/probo/pkg/deviceagent/checks"
)

func runElevatedInstall(opts InstallOptions, enrollmentToken string) error {
	parts := []string{
		shellQuote(opts.ExePath),
		"install",
		"--server",
		shellQuote(opts.ServerURL),
		"--enrollment-token",
		shellQuote(enrollmentToken),
	}
	if opts.ConfigDir != "" {
		parts = append(parts, "--dir", shellQuote(opts.ConfigDir))
	}

	shellCmd := strings.Join(parts, " ")

	script := fmt.Sprintf(
		`do shell script %s with administrator privileges`,
		applescriptQuote(shellCmd),
	)

	candidates := checks.CommandCandidates("osascript")
	if len(candidates) == 0 {
		return fmt.Errorf("command %q not available at expected absolute path", "osascript")
	}

	out, err := exec.Command(candidates[0], "-e", script).CombinedOutput()

	return commandError(out, err)
}

func shellQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `'"'"'`) + "'"
}

func applescriptQuote(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)

	return `"` + v + `"`
}
