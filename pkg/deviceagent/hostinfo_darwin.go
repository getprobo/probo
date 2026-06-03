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

package deviceagent

import (
	"context"
	"strings"
	"time"
)

func platformString() string {
	return "DARWIN"
}

func collectOSVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, _ := runQuiet(ctx, "sw_vers", "-productVersion")
	if out != "" {
		return out
	}

	out, _ = runQuiet(ctx, "uname", "-sr")

	return out
}

func collectHardwareUUID() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, _ := runQuiet(ctx, "/usr/sbin/ioreg", "-d2", "-c", "IOPlatformExpertDevice")
	if uuid := extractValue(out, "IOPlatformUUID"); uuid != "" {
		return uuid
	}

	return hashFallbackUUID()
}

func collectSerialNumber() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, _ := runQuiet(ctx, "/usr/sbin/ioreg", "-d2", "-c", "IOPlatformExpertDevice")

	return extractValue(out, "IOPlatformSerialNumber")
}

// extractValue parses ioreg key/value output.
func extractValue(s, key string) string {
	idx := strings.Index(s, "\""+key+"\"")
	if idx < 0 {
		return ""
	}

	rest := s[idx:]

	eq := strings.Index(rest, "=")
	if eq < 0 {
		return ""
	}

	rest = strings.TrimSpace(rest[eq+1:])
	rest = strings.TrimPrefix(rest, "<")
	rest = strings.TrimPrefix(rest, ">")

	if strings.HasPrefix(rest, "\"") {
		rest = rest[1:]

		before, _, ok := strings.Cut(rest, "\"")
		if !ok {
			return ""
		}

		return strings.TrimSpace(before)
	}

	end := strings.IndexAny(rest, "\r\n")
	if end < 0 {
		return strings.TrimSpace(rest)
	}

	return strings.TrimSpace(rest[:end])
}
