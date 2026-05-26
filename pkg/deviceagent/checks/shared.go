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
