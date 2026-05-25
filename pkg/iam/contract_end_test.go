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

package iam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContractEndDateHasPassed(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.May, 25, 17, 30, 0, 0, time.UTC)

	t.Run(
		"nil end date",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, contractEndDateHasPassed(nil, now))
		},
	)

	t.Run(
		"past date",
		func(t *testing.T) {
			t.Parallel()

			endDate := time.Date(2026, time.May, 24, 0, 0, 0, 0, time.UTC)

			assert.True(t, contractEndDateHasPassed(&endDate, now))
		},
	)

	t.Run(
		"same date",
		func(t *testing.T) {
			t.Parallel()

			endDate := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)

			assert.False(t, contractEndDateHasPassed(&endDate, now))
		},
	)

	t.Run(
		"future date",
		func(t *testing.T) {
			t.Parallel()

			endDate := time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC)

			assert.False(t, contractEndDateHasPassed(&endDate, now))
		},
	)
}
