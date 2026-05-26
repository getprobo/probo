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

package checks

import (
	"os"
	"path/filepath"
	"strings"
)

func commandCandidates(cmd string) []string {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}
	system32 := filepath.Join(systemRoot, "System32")

	switch strings.ToLower(cmd) {
	case "powershell", "powershell.exe":
		return []string{
			filepath.Join(system32, "WindowsPowerShell", "v1.0", "powershell.exe"),
		}
	case "manage-bde", "manage-bde.exe":
		return []string{filepath.Join(system32, "manage-bde.exe")}
	case "netsh", "netsh.exe":
		return []string{filepath.Join(system32, "netsh.exe")}
	case "w32tm", "w32tm.exe":
		return []string{filepath.Join(system32, "w32tm.exe")}
	case "sc", "sc.exe":
		return []string{filepath.Join(system32, "sc.exe")}
	case "net", "net.exe":
		return []string{filepath.Join(system32, "net.exe")}
	default:
		return nil
	}
}
