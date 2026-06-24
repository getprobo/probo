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
