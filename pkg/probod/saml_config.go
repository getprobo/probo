// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package probod

import (
	"time"
)

type samlConfig struct {
	SessionDuration                   int    `json:"session-duration"`
	CleanupIntervalSeconds            int    `json:"cleanup-interval-seconds"`
	Certificate                       string `json:"certificate"`
	PrivateKey                        string `json:"private-key"`
	DomainVerificationIntervalSeconds int    `json:"domain-verification-interval-seconds"`
	DomainVerificationResolverAddr    string `json:"domain-verification-resolver-addr"`
}

func (c samlConfig) SessionDurationTime() time.Duration {
	if c.SessionDuration == 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(c.SessionDuration) * time.Second
}

func (c samlConfig) CleanupInterval() time.Duration {
	if c.CleanupIntervalSeconds == 0 {
		return 0
	}

	return time.Duration(c.CleanupIntervalSeconds) * time.Second
}

func (c samlConfig) DomainVerificationInterval() time.Duration {
	return time.Duration(c.DomainVerificationIntervalSeconds) * time.Second
}
