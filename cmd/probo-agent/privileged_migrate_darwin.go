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

// Temporary migration entrypoint for the privileged Darwin launchd daemon.
// Remove this file once every deployed agent has moved to the canonical path.

package main

import (
	"fmt"
	"os"

	"go.probo.inc/probo/pkg/deviceagent"
	"go.probo.inc/probo/pkg/deviceagent/service"
)

func migratePrivilegedDaemon(exe, dir string) error {
	if os.Geteuid() != 0 {
		return nil
	}

	if deviceagent.IsCanonicalExecutable(exe) {
		if err := deviceagent.RemoveLegacyExecutable(); err != nil {
			return fmt.Errorf("cannot remove legacy executable: %w", err)
		}

		return nil
	}

	if !deviceagent.ShouldMigrateExecutable(exe, service.InstalledExePath()) {
		return nil
	}

	dest, err := deviceagent.EnsurePrivilegedExecutable(exe)
	if err != nil {
		return fmt.Errorf("cannot install privileged executable: %w", err)
	}

	if err := service.WritePlist(
		service.Config{
			ExePath: dest,
			Dir:     dir,
		},
	); err != nil {
		return fmt.Errorf("cannot rewrite launchd plist: %w", err)
	}

	if err := registerTrayAutoStart(dest, deviceagent.EnrollmentRunDir(dir)); err != nil {
		return err
	}

	if err := service.ScheduleReload(); err != nil {
		return err
	}

	return deviceagent.ErrRestartRequired
}
