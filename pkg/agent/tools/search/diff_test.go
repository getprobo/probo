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

package search

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDiff(t *testing.T) {
	t.Parallel()

	t.Run(
		"identical documents have no changes",
		func(t *testing.T) {
			t.Parallel()

			lines := []string{"line one", "line two", "line three"}
			diff := computeDiff(lines, lines, "a", "b")

			assert.Equal(t, 0, diff.added)
			assert.Equal(t, 0, diff.removed)
		},
	)

	t.Run(
		"completely different documents",
		func(t *testing.T) {
			t.Parallel()

			linesA := []string{"alpha", "beta"}
			linesB := []string{"gamma", "delta"}
			diff := computeDiff(linesA, linesB, "a", "b")

			assert.Equal(t, 2, diff.added)
			assert.Equal(t, 2, diff.removed)
			assert.Contains(t, diff.output, "- alpha")
			assert.Contains(t, diff.output, "- beta")
			assert.Contains(t, diff.output, "+ gamma")
			assert.Contains(t, diff.output, "+ delta")
		},
	)

	t.Run(
		"added lines only",
		func(t *testing.T) {
			t.Parallel()

			linesA := []string{"line one"}
			linesB := []string{"line one", "line two", "line three"}
			diff := computeDiff(linesA, linesB, "a", "b")

			assert.Equal(t, 2, diff.added)
			assert.Equal(t, 0, diff.removed)
			assert.Contains(t, diff.output, "+ line two")
			assert.Contains(t, diff.output, "+ line three")
		},
	)

	t.Run(
		"removed lines only",
		func(t *testing.T) {
			t.Parallel()

			linesA := []string{"line one", "line two", "line three"}
			linesB := []string{"line one"}
			diff := computeDiff(linesA, linesB, "a", "b")

			assert.Equal(t, 0, diff.added)
			assert.Equal(t, 2, diff.removed)
			assert.Contains(t, diff.output, "- line two")
			assert.Contains(t, diff.output, "- line three")
		},
	)

	t.Run(
		"mixed changes",
		func(t *testing.T) {
			t.Parallel()

			linesA := []string{"keep", "remove me", "also keep"}
			linesB := []string{"keep", "add me", "also keep"}
			diff := computeDiff(linesA, linesB, "a", "b")

			assert.Equal(t, 1, diff.added)
			assert.Equal(t, 1, diff.removed)
			assert.Contains(t, diff.output, "- remove me")
			assert.Contains(t, diff.output, "+ add me")
		},
	)

	t.Run(
		"both inputs empty",
		func(t *testing.T) {
			t.Parallel()

			diff := computeDiff([]string{}, []string{}, "a", "b")

			assert.Equal(t, 0, diff.added)
			assert.Equal(t, 0, diff.removed)
		},
	)

	t.Run(
		"first input empty",
		func(t *testing.T) {
			t.Parallel()

			linesB := []string{"new line"}
			diff := computeDiff([]string{}, linesB, "a", "b")

			assert.Equal(t, 1, diff.added)
			assert.Equal(t, 0, diff.removed)
			assert.Contains(t, diff.output, "+ new line")
		},
	)

	t.Run(
		"second input empty",
		func(t *testing.T) {
			t.Parallel()

			linesA := []string{"old line"}
			diff := computeDiff(linesA, []string{}, "a", "b")

			assert.Equal(t, 0, diff.added)
			assert.Equal(t, 1, diff.removed)
			assert.Contains(t, diff.output, "- old line")
		},
	)

	t.Run(
		"single line documents identical",
		func(t *testing.T) {
			t.Parallel()

			diff := computeDiff([]string{"same"}, []string{"same"}, "a", "b")

			assert.Equal(t, 0, diff.added)
			assert.Equal(t, 0, diff.removed)
		},
	)

	t.Run(
		"single line documents different",
		func(t *testing.T) {
			t.Parallel()

			diff := computeDiff([]string{"old"}, []string{"new"}, "a", "b")

			assert.Equal(t, 1, diff.added)
			assert.Equal(t, 1, diff.removed)
		},
	)

	t.Run(
		"output contains labels",
		func(t *testing.T) {
			t.Parallel()

			diff := computeDiff(
				[]string{"a"},
				[]string{"b"},
				"current version",
				"archived version",
			)

			assert.True(t, strings.HasPrefix(diff.output, "--- current version\n+++ archived version\n"))
		},
	)

	t.Run(
		"documents too large returns bounded message",
		func(t *testing.T) {
			t.Parallel()

			large := make([]string, 5001)
			for i := range large {
				large[i] = "line"
			}

			diff := computeDiff(large, []string{"small"}, "a", "b")

			assert.Equal(t, 0, diff.added)
			assert.Equal(t, 0, diff.removed)
			assert.Contains(t, diff.output, "too large")
		},
	)
}
