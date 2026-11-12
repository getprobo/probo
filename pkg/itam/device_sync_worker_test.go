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

package itam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeviceSyncBackoff(t *testing.T) {
	t.Parallel()

	interval := time.Hour

	assert.Equal(t, 5*time.Minute, deviceSyncBackoff(1, interval))
	assert.Equal(t, 10*time.Minute, deviceSyncBackoff(2, interval))
	assert.Equal(t, 20*time.Minute, deviceSyncBackoff(3, interval))
	assert.Equal(t, 40*time.Minute, deviceSyncBackoff(4, interval))
	// Capped at the regular interval: a failing sync is never retried less
	// often than a healthy one runs.
	assert.Equal(t, interval, deviceSyncBackoff(5, interval))
	assert.Equal(t, interval, deviceSyncBackoff(64, interval))
	// A zero attempt is the first.
	assert.Equal(t, 5*time.Minute, deviceSyncBackoff(0, interval))
}
