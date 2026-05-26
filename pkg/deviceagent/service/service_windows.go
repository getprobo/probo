// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package service

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Install registers and starts the Windows service via sc.exe.
func Install(cfg Config) error {
	if cfg.ExePath == "" {
		return errors.New("executable path is required")
	}

	if cfg.Dir == "" {
		return errors.New("state directory is required")
	}

	name := DefaultWindowsName

	bin := fmt.Sprintf(`"%s" run --dir "%s"`, cfg.ExePath, cfg.Dir)
	if out, err := exec.Command(
		"sc.exe",
		"create",
		name,
		"binPath=",
		bin,
		"start=",
		"auto",
		"DisplayName=",
		"Probo Device Posture Agent",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run sc.exe create: %w: %s", err, strings.TrimSpace(string(out)))
	}

	// Restart on failure.
	if out, err := exec.Command(
		"sc.exe",
		"failure",
		name,
		"reset=",
		"86400",
		"actions=",
		"restart/1000/restart/1000/restart/1000",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run sc.exe failure: %w: %s", err, strings.TrimSpace(string(out)))
	}

	if out, err := exec.Command("sc.exe", "start", name).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot run sc.exe start: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func Uninstall(cfg Config) error {
	name := DefaultWindowsName

	_ = exec.Command("sc.exe", "stop", name).Run()
	if out, err := exec.Command("sc.exe", "delete", name).CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if isWindowsServiceMissing(msg) {
			return nil
		}

		return fmt.Errorf("cannot run sc.exe delete: %w: %s", err, msg)
	}

	return nil
}
