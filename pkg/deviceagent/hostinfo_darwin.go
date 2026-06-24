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
