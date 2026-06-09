// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

//go:build windows

package tray

import (
	"fmt"
	"os/exec"
	"strings"
)

func runElevatedInstall(opts Options, token string) error {
	args := []string{
		"install",
		"--server",
		opts.ServerURL,
		"--enrollment-token",
		token,
	}

	argList := make([]string, len(args))
	for i, arg := range args {
		argList[i] = "'" + escapePowerShellSingleQuoted(arg) + "'"
	}

	script := fmt.Sprintf(
		`$p = Start-Process -FilePath %q -ArgumentList @(%s) -Verb RunAs -Wait -PassThru; if ($p.ExitCode -ne 0) { exit $p.ExitCode }`,
		opts.ExePath,
		strings.Join(argList, ","),
	)

	out, err := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		script,
	).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}

		return fmt.Errorf("%s", msg)
	}

	return nil
}
