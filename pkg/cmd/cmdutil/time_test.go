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

package cmdutil_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func TestNormalizeDatetime_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"date only anchors to midnight UTC", "2026-01-01", "2026-01-01T00:00:00Z"},
		{"date only end of year", "2026-12-31", "2026-12-31T00:00:00Z"},
		{"rfc3339 passthrough", "2026-01-01T15:04:05Z", "2026-01-01T15:04:05Z"},
		{"rfc3339 with offset passthrough", "2026-01-01T15:04:05+02:00", "2026-01-01T15:04:05+02:00"},
		{"rfc3339 nano passthrough", "2026-01-01T15:04:05.123456789Z", "2026-01-01T15:04:05.123456789Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := cmdutil.NormalizeDatetime(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			// The result must be accepted by the same RFC3339 parser the
			// GraphQL Datetime scalar uses, or the server would reject it.
			_, parseErr := time.Parse(time.RFC3339, got)
			assert.NoError(t, parseErr)
		})
	}
}

func TestNormalizeDatetime_Invalid(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "not-a-date", "01/02/2026", "2026-13-01", "2026-01-01 15:04:05"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			_, err := cmdutil.NormalizeDatetime(in)
			assert.Error(t, err)
		})
	}
}
