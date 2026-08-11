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

package testutil

import (
	"testing"
	"time"
)

// Poll runs condition on the calling goroutine until it returns true or waitFor
// elapses, sleeping tick between attempts. Unlike testify's Eventually, this
// does not spawn a background goroutine, so it is safe under the race detector
// when the condition mutates local state.
func Poll(
	t testing.TB,
	waitFor time.Duration,
	tick time.Duration,
	condition func() bool,
) bool {
	t.Helper()

	deadline := time.Now().Add(waitFor)

	for {
		if condition() {
			return true
		}

		if !time.Now().Before(deadline) {
			return false
		}

		time.Sleep(tick)
	}
}
