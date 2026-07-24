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
	"fmt"
	"os"
	"strings"
)

var windowsCommandPaths = map[string][]string{
	"manage-bde": {`%s\System32\manage-bde.exe`},
	"net":        {`%s\System32\net.exe`},
	"netsh":      {`%s\System32\netsh.exe`},
	"powershell": {`%s\System32\WindowsPowerShell\v1.0\powershell.exe`},
	"sc":         {`%s\System32\sc.exe`},
	"w32tm":      {`%s\System32\w32tm.exe`},
}

func commandCandidates(cmd string) []string {
	templates := windowsCommandPaths[normalizeWindowsCommand(cmd)]
	if len(templates) == 0 {
		return nil
	}

	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}

	paths := make([]string, len(templates))
	for i, template := range templates {
		paths[i] = fmt.Sprintf(template, systemRoot)
	}

	return paths
}

func normalizeWindowsCommand(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	return strings.TrimSuffix(name, ".exe")
}
