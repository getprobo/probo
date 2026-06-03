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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_applyDefaults(t *testing.T) {
	t.Parallel()

	t.Run(
		"uses defaults when unset",
		func(t *testing.T) {
			t.Parallel()

			cfg := &Config{}
			cfg.applyDefaults()

			assert.Equal(t, DefaultHeartbeatInterval, cfg.HeartbeatInterval)
			assert.Equal(t, DefaultPostureInterval, cfg.PostureInterval)
		},
	)

	t.Run(
		"clamps values below minimum floors",
		func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				HeartbeatInterval: 10 * time.Second,
				PostureInterval:   1 * time.Minute,
			}
			cfg.applyDefaults()

			assert.Equal(t, MinHeartbeatInterval, cfg.HeartbeatInterval)
			assert.Equal(t, MinPostureInterval, cfg.PostureInterval)
		},
	)

	t.Run(
		"keeps values above floors",
		func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				HeartbeatInterval: 3 * time.Minute,
				PostureInterval:   2 * time.Hour,
			}
			cfg.applyDefaults()

			assert.Equal(t, 3*time.Minute, cfg.HeartbeatInterval)
			assert.Equal(t, 2*time.Hour, cfg.PostureInterval)
		},
	)
}
