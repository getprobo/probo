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

package agentrun

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeError(t *testing.T) {
	t.Parallel()

	t.Run(
		"short message unchanged",
		func(t *testing.T) {
			t.Parallel()

			err := errors.New("short")
			assert.Equal(t, "short", sanitizeError(err))
		},
	)

	t.Run(
		"boundary length unchanged",
		func(t *testing.T) {
			t.Parallel()

			msg := strings.Repeat("a", errorMessageMaxLen)
			assert.Equal(t, msg, sanitizeError(errors.New(msg)))
		},
	)

	t.Run(
		"long utf8 message is rune safe and suffixed",
		func(t *testing.T) {
			t.Parallel()

			msg := strings.Repeat("é", errorMessageMaxLen)
			sanitized := sanitizeError(errors.New(msg))

			assert.True(t, strings.HasSuffix(sanitized, "…"))
			assert.True(t, len(sanitized) <= errorMessageMaxLen+len("…"))
			assert.True(t, strings.HasPrefix(msg, strings.TrimSuffix(sanitized, "…")))
		},
	)
}

func TestWorkerOptions(t *testing.T) {
	t.Parallel()

	t.Run(
		"interval updates only when positive",
		func(t *testing.T) {
			t.Parallel()

			cfg := workerConfig{interval: 3 * time.Second}

			WithWorkerInterval(0)(&cfg)
			assert.Equal(t, 3*time.Second, cfg.interval)

			WithWorkerInterval(7 * time.Second)(&cfg)
			assert.Equal(t, 7*time.Second, cfg.interval)
		},
	)

	t.Run(
		"max concurrency updates only when positive",
		func(t *testing.T) {
			t.Parallel()

			cfg := workerConfig{maxConcurrency: 2}

			WithWorkerMaxConcurrency(0)(&cfg)
			assert.Equal(t, 2, cfg.maxConcurrency)

			WithWorkerMaxConcurrency(9)(&cfg)
			assert.Equal(t, 9, cfg.maxConcurrency)
		},
	)
}
