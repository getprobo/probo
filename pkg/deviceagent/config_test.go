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
