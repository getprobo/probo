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

package accessreview

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNameSyncBackoff(t *testing.T) {
	t.Parallel()

	// The schedule a source walks as it fails: doubling from one minute, so
	// the budget of maxNameSyncAttempts spans tens of minutes rather than the
	// milliseconds an ungated claim would take. Every delay must exceed the
	// worker's 10s poll interval, or a backed-off row is claimed again on the
	// very next cycle.
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Minute},
		{2, 2 * time.Minute},
		{3, 4 * time.Minute},
		{4, 8 * time.Minute},
		{5, 16 * time.Minute},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, nameSyncBackoff(tc.attempt), "attempt %d", tc.attempt)
		assert.Greater(t, nameSyncBackoff(tc.attempt), 10*time.Second, "attempt %d must outlast a poll cycle", tc.attempt)
	}
}

func TestNameSyncBackoffIsBounded(t *testing.T) {
	t.Parallel()

	// The budget caps attempts long before these, but the schedule must stay
	// sane rather than overflow into a negative duration if it ever does not.
	assert.Equal(t, nameSyncMaxBackoff, nameSyncBackoff(100))
	assert.Equal(t, nameSyncMaxBackoff, nameSyncBackoff(1_000_000))
	assert.Equal(t, time.Minute, nameSyncBackoff(0))
	assert.Equal(t, time.Minute, nameSyncBackoff(-1))
}
