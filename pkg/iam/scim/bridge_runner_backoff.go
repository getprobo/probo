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

package scim

import "time"

const (
	// DefaultMaxConsecutiveFailures is the maximum number of consecutive failures
	// before a bridge is disabled.
	DefaultMaxConsecutiveFailures = 10

	// DefaultMaxBackoff is the maximum backoff duration between retries.
	DefaultMaxBackoff = 24 * time.Hour

	// DefaultStaleSyncThreshold is the time after which a SYNCING bridge is
	// considered stale and can be recovered by another runner.
	DefaultStaleSyncThreshold = 10 * time.Minute
)

func (r *BridgeRunner) calculateBackoff(consecutiveFailures int) time.Duration {
	if consecutiveFailures <= 0 {
		return r.cfg.Interval
	}

	// Cap the shift exponent to prevent integer overflow from the shift itself.
	// Bit 63 is the sign bit, so shifting by 63+ produces negative or zero values.
	const maxShift = 62
	shiftAmount := consecutiveFailures
	if shiftAmount > maxShift {
		shiftAmount = maxShift
	}

	backoff := r.cfg.Interval * time.Duration(1<<shiftAmount)

	// Detect multiplication overflow: if result is non-positive or less than the
	// base interval, overflow occurred. Return MaxBackoff in this case.
	if backoff <= 0 || backoff < r.cfg.Interval {
		return r.cfg.MaxBackoff
	}

	if backoff > r.cfg.MaxBackoff {
		return r.cfg.MaxBackoff
	}

	return backoff
}

func (r *BridgeRunner) shouldDisable(consecutiveFailures int) bool {
	return consecutiveFailures >= r.cfg.MaxConsecutiveFailures
}
