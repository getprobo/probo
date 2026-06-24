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
	"context"
	"time"
)

// Check keys shared across OS implementations.
const (
	KeyDiskEncryption    = "DISK_ENCRYPTION"
	KeyScreenLock        = "SCREEN_LOCK"
	KeyFirewallEnabled   = "FIREWALL_ENABLED"
	KeyTimeSync          = "TIME_SYNC"
	KeyOSVersion         = "OS_VERSION"
	KeyAutoUpdate        = "AUTO_UPDATE"
	KeyPasswordPolicy    = "PASSWORD_POLICY"
	KeyRemoteLogin       = "REMOTE_LOGIN"
	KeyMalwareProtection = "MALWARE_PROTECTION"
)

type funcCheck struct {
	key string
	run func(ctx context.Context) Result
}

func (c funcCheck) Key() string { return c.key }

func (c funcCheck) Run(ctx context.Context) Result {
	r := c.run(ctx)
	if r.CheckKey == "" {
		r.CheckKey = c.key
	}

	if r.ObservedAt.IsZero() {
		r.ObservedAt = time.Now().UTC()
	}

	return r
}

func pass(ev map[string]any) Result {
	return Result{Status: StatusPass, Evidence: ev}
}

func fail(ev map[string]any) Result {
	return Result{Status: StatusFail, Evidence: ev}
}

func unknown(ev map[string]any) Result {
	return Result{Status: StatusUnknown, Evidence: ev}
}

func notApplicable(ev map[string]any) Result {
	return Result{Status: StatusNotApplicable, Evidence: ev}
}

// truncate limits oversized evidence values.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n] + "…"
}

// errString returns "" for a nil error.
func errString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
