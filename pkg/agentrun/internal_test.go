// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
