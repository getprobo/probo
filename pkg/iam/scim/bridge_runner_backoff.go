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

	shiftAmount := min(consecutiveFailures, maxShift)

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
