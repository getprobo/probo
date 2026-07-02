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
