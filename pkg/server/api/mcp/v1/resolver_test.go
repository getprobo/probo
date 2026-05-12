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

package mcp_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeExcludedUserNames(t *testing.T) {
	t.Parallel()

	t.Run(
		"appends new names to existing",
		func(t *testing.T) {
			t.Parallel()

			got := mergeExcludedUserNames(
				[]string{"alice@example.com", "bob@example.com"},
				[]string{"charlie@example.com"},
			)

			assert.Equal(t, []string{"alice@example.com", "bob@example.com", "charlie@example.com"}, got)
		},
	)

	t.Run(
		"deduplicates overlapping names",
		func(t *testing.T) {
			t.Parallel()

			got := mergeExcludedUserNames(
				[]string{"alice@example.com", "bob@example.com"},
				[]string{"bob@example.com", "charlie@example.com"},
			)

			assert.Equal(t, []string{"alice@example.com", "bob@example.com", "charlie@example.com"}, got)
		},
	)

	t.Run(
		"empty existing list",
		func(t *testing.T) {
			t.Parallel()

			got := mergeExcludedUserNames(
				nil,
				[]string{"alice@example.com"},
			)

			assert.Equal(t, []string{"alice@example.com"}, got)
		},
	)

	t.Run(
		"empty incoming list preserves existing",
		func(t *testing.T) {
			t.Parallel()

			got := mergeExcludedUserNames(
				[]string{"alice@example.com", "bob@example.com"},
				nil,
			)

			assert.Equal(t, []string{"alice@example.com", "bob@example.com"}, got)
		},
	)

	t.Run(
		"both lists empty",
		func(t *testing.T) {
			t.Parallel()

			got := mergeExcludedUserNames(nil, nil)

			assert.Equal(t, []string{}, got)
		},
	)

	t.Run(
		"result is sorted",
		func(t *testing.T) {
			t.Parallel()

			got := mergeExcludedUserNames(
				[]string{"charlie@example.com"},
				[]string{"alice@example.com"},
			)

			assert.Equal(t, []string{"alice@example.com", "charlie@example.com"}, got)
		},
	)
}
