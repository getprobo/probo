// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"os/exec"
	"strings"
)

type (
	// HostInfo is the device identity reported by the agent.
	HostInfo struct {
		Hostname     string
		Platform     string
		OSVersion    string
		HardwareUUID string
		SerialNumber *string
	}
)

// CollectHostInfo gathers host identity using best-effort probes.
func CollectHostInfo() HostInfo {
	info := HostInfo{
		Platform: platformString(),
	}

	if h, err := os.Hostname(); err == nil {
		info.Hostname = h
	}

	if info.Hostname == "" {
		info.Hostname = "unknown-host"
	}

	info.OSVersion = collectOSVersion()

	info.HardwareUUID = collectHardwareUUID()
	if sn := collectSerialNumber(); sn != "" {
		info.SerialNumber = &sn
	}

	return info
}

// hashFallbackUUID derives a stable fallback from hostname and MAC.
func hashFallbackUUID() string {
	hostname, _ := os.Hostname()
	mac := firstStableMAC()
	h := sha256.New()
	h.Write([]byte(hostname))
	h.Write([]byte{0})
	h.Write([]byte(mac))

	return hex.EncodeToString(h.Sum(nil))
}

func firstStableMAC() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 {
			continue
		}

		if len(ifc.HardwareAddr) == 0 {
			continue
		}

		return ifc.HardwareAddr.String()
	}

	return ""
}

// runQuiet runs a command and returns trimmed stdout.
func runQuiet(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()

	return strings.TrimSpace(string(out)), err
}
