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

package changefilter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/changefilter"
)

func TestBuildChangeEventFilter(t *testing.T) {
	t.Parallel()

	t.Run("empty args yields empty filter", func(t *testing.T) {
		t.Parallel()

		filter, err := changefilter.BuildChangeEventFilter(changefilter.ChangeEventFilterArgs{})
		require.NoError(t, err)
		assert.Empty(t, filter)
	})

	t.Run("valid observed-at bounds and change types", func(t *testing.T) {
		t.Parallel()

		filter, err := changefilter.BuildChangeEventFilter(
			changefilter.ChangeEventFilterArgs{
				SourceID:             "source-1",
				ChangeTypes:          []string{"APPEARED", "DISAPPEARED"},
				ObservedAtOnOrAfter:  "2026-01-02T03:04:05Z",
				ObservedAtOnOrBefore: "2026-02-03T04:05:06Z",
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "source-1", filter["accessReviewSourceId"])
		assert.Equal(t, []string{"APPEARED", "DISAPPEARED"}, filter["changeTypes"])
		assert.Equal(t, "2026-01-02T03:04:05Z", filter["observedAtOnOrAfter"])
		assert.Equal(t, "2026-02-03T04:05:06Z", filter["observedAtOnOrBefore"])
	})

	t.Run("invalid observed-at-on-or-after", func(t *testing.T) {
		t.Parallel()

		_, err := changefilter.BuildChangeEventFilter(
			changefilter.ChangeEventFilterArgs{ObservedAtOnOrAfter: "not-a-timestamp"},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--observed-at-on-or-after")
	})

	t.Run("invalid observed-at-on-or-before", func(t *testing.T) {
		t.Parallel()

		_, err := changefilter.BuildChangeEventFilter(
			changefilter.ChangeEventFilterArgs{ObservedAtOnOrBefore: "2026-01-02"},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--observed-at-on-or-before")
	})
}
