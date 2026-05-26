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
