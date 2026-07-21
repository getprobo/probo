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

//go:build windows

package elevate

import (
	"fmt"
	"os/exec"
	"strings"

	"go.probo.inc/probo/pkg/deviceagent/checks"
	"golang.org/x/sys/windows"
)

func runElevatedInstall(opts InstallOptions, enrollmentToken string) error {
	args := []string{
		"install",
		"--server",
		opts.ServerURL,
		"--enrollment-token",
		enrollmentToken,
	}
	if opts.ConfigDir != "" {
		args = append(args, "--dir", opts.ConfigDir)
	}

	return runPowerShellCommand(elevatedStartProcess(opts.ExePath, args))
}

func runElevatedUninstall(opts UninstallOptions) error {
	args := []string{"uninstall"}
	if opts.ConfigDir != "" {
		args = append(args, "--dir", opts.ConfigDir)
	}

	return runPowerShellCommand(elevatedStartProcess(opts.ExePath, args))
}

// elevatedStartProcess builds a PowerShell script that launches exePath elevated
// via UAC. Launch failures (including UAC cancel) are terminating and exit
// nonzero before $p.ExitCode is inspected.
func elevatedStartProcess(exePath string, args []string) string {
	argList := make([]string, len(args))
	for i, arg := range args {
		argList[i] = "'" + escapePowerShellSingleQuoted(windows.EscapeArg(arg)) + "'"
	}

	return fmt.Sprintf(
		`$ErrorActionPreference = 'Stop'; $p = Start-Process -FilePath %s -ArgumentList @(%s) -Verb RunAs -Wait -PassThru; if ($null -eq $p) { exit 1 }; if ($p.ExitCode -ne 0) { exit $p.ExitCode }`,
		"'"+escapePowerShellSingleQuoted(exePath)+"'",
		strings.Join(argList, ","),
	)
}

func runPowerShellCommand(script string) error {
	candidates := checks.CommandCandidates("powershell.exe")
	if len(candidates) == 0 {
		return fmt.Errorf("command %q not available at expected absolute path", "powershell.exe")
	}

	out, err := exec.Command(
		candidates[0],
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		script,
	).CombinedOutput()

	return commandError(out, err)
}
