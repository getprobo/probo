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

package deviceagent

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"
	"time"
)

func platformString() string {
	return "LINUX"
}

func collectOSVersion() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		if prettyName := parseOSReleasePrettyName(data); prettyName != "" {
			return prettyName
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, _ := runQuiet(ctx, "uname", "-sr")

	return out
}

func collectHardwareUUID() string {
	for _, path := range []string{
		"/sys/class/dmi/id/product_uuid",
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	} {
		if data, err := os.ReadFile(path); err == nil {
			if s := strings.TrimSpace(string(data)); s != "" {
				return s
			}
		}
	}

	return hashFallbackUUID()
}

func collectSerialNumber() string {
	if data, err := os.ReadFile("/sys/class/dmi/id/product_serial"); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
}

func parseOSReleasePrettyName(data []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		if k, v, ok := splitKV(line); ok {
			if k == "PRETTY_NAME" {
				return v
			}
		}
	}

	return ""
}

func splitKV(line string) (string, string, bool) {
	eq := strings.IndexByte(line, '=')
	if eq <= 0 {
		return "", "", false
	}

	k := strings.TrimSpace(line[:eq])
	v := strings.TrimSpace(line[eq+1:])
	v = strings.Trim(v, `"`)

	return k, v, true
}
