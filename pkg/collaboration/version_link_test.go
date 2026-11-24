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

package collaboration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func paragraphDocument(text string) string {
	return fmt.Sprintf(
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":%q}]}]}`,
		text,
	)
}

func TestMatchVersionChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	document, err := newAutomergeDocument(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	base := time.Unix(1_000_000, 0).UTC()
	texts := []string{"first", "second", "third"}

	for i, text := range texts {
		require.NoError(t, writeSpans(document, paragraphDocument(text), i == 0))

		_, err := document.Commit("edit", base.Add(time.Duration(i)*time.Hour))
		require.NoError(t, err)
	}

	changes, err := document.ChangeHistory()
	require.NoError(t, err)
	require.Len(t, changes, len(texts))

	render := func(heads []automerge.Hash) (string, error) {
		return materializeAcceptedAtHeads(document, heads)
	}

	contentOf := func(text string) string {
		content, err := materializeAcceptedAtHeads(document, []automerge.Hash{changes[indexOf(texts, text)].Hash})
		require.NoError(t, err)

		return content
	}

	t.Run(
		"matches the state the version was taken at",
		func(t *testing.T) {
			t.Parallel()

			hash, ok, err := matchVersionChange(
				contentOf("second"),
				base.Add(time.Hour+time.Minute),
				changes,
				render,
			)

			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, changes[1].Hash, hash)
		},
	)

	t.Run(
		"matches a state older than the version",
		func(t *testing.T) {
			t.Parallel()

			hash, ok, err := matchVersionChange(
				contentOf("first"),
				base.Add(2*time.Hour+time.Minute),
				changes,
				render,
			)

			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, changes[0].Hash, hash)
		},
	)

	t.Run(
		"matches a state created after the version was",
		func(t *testing.T) {
			t.Parallel()

			hash, ok, err := matchVersionChange(
				contentOf("third"),
				base.Add(-time.Hour),
				changes,
				render,
			)

			require.NoError(t, err)
			require.True(t, ok)
			assert.Equal(t, changes[2].Hash, hash)
		},
	)

	t.Run(
		"reports content that never existed",
		func(t *testing.T) {
			t.Parallel()

			_, ok, err := matchVersionChange(
				paragraphDocument("never"),
				base.Add(time.Hour),
				changes,
				render,
			)

			require.NoError(t, err)
			assert.False(t, ok)
		},
	)

	t.Run(
		"reports an empty history",
		func(t *testing.T) {
			t.Parallel()

			_, ok, err := matchVersionChange(contentOf("first"), base, nil, render)

			require.NoError(t, err)
			assert.False(t, ok)
		},
	)
}

func indexOf(values []string, value string) int {
	for i, candidate := range values {
		if candidate == value {
			return i
		}
	}

	return -1
}
