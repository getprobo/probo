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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgent_currentHostInfo(t *testing.T) {
	t.Parallel()

	t.Run(
		"reuses cached host info before refresh interval",
		func(t *testing.T) {
			t.Parallel()

			count := 0
			a := &Agent{
				collectHostInfo: func() HostInfo {
					count++
					return HostInfo{Hostname: fmt.Sprintf("host-%d", count)}
				},
			}

			now := time.Unix(1_000, 0)
			first := a.currentHostInfo(now)
			second := a.currentHostInfo(now.Add(2 * time.Hour))

			assert.Equal(t, 1, count)
			assert.Equal(t, first, second)
		},
	)

	t.Run(
		"refreshes host info when cache expires",
		func(t *testing.T) {
			t.Parallel()

			count := 0
			a := &Agent{
				collectHostInfo: func() HostInfo {
					count++
					return HostInfo{Hostname: fmt.Sprintf("host-%d", count)}
				},
			}

			now := time.Unix(2_000, 0)
			first := a.currentHostInfo(now)
			second := a.currentHostInfo(now.Add(hostInfoRefreshInterval + time.Minute))

			assert.Equal(t, 2, count)
			assert.NotEqual(t, first, second)
		},
	)
}
