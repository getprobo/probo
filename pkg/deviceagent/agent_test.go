// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
