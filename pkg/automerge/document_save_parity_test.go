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

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// documentSaveScenario applies the same history to whichever engine it is given
// so a native snapshot can be held against the reference's byte for byte.
type documentSaveScenario struct {
	name  string
	apply func(t *testing.T, ctx context.Context, document *automerge.Document)
}

func documentSaveScenarios() []documentSaveScenario {
	base := time.Unix(1786147200, 0).UTC()

	return []documentSaveScenario{
		{
			name: "linear text",
			apply: func(t *testing.T, ctx context.Context, document *automerge.Document) {
				text, err := document.CreateText(ctx, "body")
				require.NoError(t, err)

				for i := range 5 {
					require.NoError(t, text.Splice(ctx, uint32(i), 0, "x"))
					_, err = document.Commit(ctx, "edit", base.Add(time.Duration(i)*time.Second))
					require.NoError(t, err)
				}
			},
		},
		{
			name: "map puts and delete",
			apply: func(t *testing.T, ctx context.Context, document *automerge.Document) {
				require.NoError(t, document.PutScalar(ctx, "title",
					automerge.Scalar{Type: automerge.ScalarTypeString, String: "first"}))
				require.NoError(t, document.PutScalar(ctx, "keep",
					automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"}))
				_, err := document.Commit(ctx, "one", base)
				require.NoError(t, err)

				require.NoError(t, document.PutScalar(ctx, "title",
					automerge.Scalar{Type: automerge.ScalarTypeString, String: "second"}))
				_, err = document.Commit(ctx, "two", base.Add(time.Second))
				require.NoError(t, err)

				require.NoError(t, document.Root().DeleteKey(ctx, "title"))
				_, err = document.Commit(ctx, "three", base.Add(2*time.Second))
				require.NoError(t, err)
			},
		},
		{
			name: "counter increment",
			apply: func(t *testing.T, ctx context.Context, document *automerge.Document) {
				require.NoError(t, document.PutScalar(ctx, "counter",
					automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5}))
				_, err := document.Commit(ctx, "create", base)
				require.NoError(t, err)

				require.NoError(t, document.Root().Increment(ctx, "counter", 3))
				_, err = document.Commit(ctx, "bump", base.Add(time.Second))
				require.NoError(t, err)
			},
		},
		{
			name: "marks and unmarks",
			apply: func(t *testing.T, ctx context.Context, document *automerge.Document) {
				text, err := document.CreateText(ctx, "body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(ctx, 0, 0, "hello brave world"))
				_, err = document.Commit(ctx, "write", base)
				require.NoError(t, err)

				require.NoError(t, text.Mark(ctx, 0, 5, "strong",
					automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
					automerge.MarkExpandBoth))
				_, err = document.Commit(ctx, "mark", base.Add(time.Second))
				require.NoError(t, err)

				require.NoError(t, text.Unmark(ctx, 1, 3, "strong", automerge.MarkExpandNone))
				_, err = document.Commit(ctx, "unmark", base.Add(2*time.Second))
				require.NoError(t, err)
			},
		},
		{
			name: "text with deletion",
			apply: func(t *testing.T, ctx context.Context, document *automerge.Document) {
				text, err := document.CreateText(ctx, "body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(ctx, 0, 0, "hello brave world"))
				_, err = document.Commit(ctx, "write", base)
				require.NoError(t, err)

				require.NoError(t, text.Splice(ctx, 5, 6, ""))
				_, err = document.Commit(ctx, "trim", base.Add(time.Second))
				require.NoError(t, err)
			},
		},
	}
}

// TestDocumentSave_MatchesReferenceBytes requires the compacted snapshot the
// native engine writes to be the file the reference writes for the same
// history. Byte identity is the only check that proves the operation-set order,
// the column layout and the frontier all agree.
func TestDocumentSave_MatchesReferenceBytes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, scenario := range documentSaveScenarios() {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			native, err := automerge.New(ctx, actor(41))
			require.NoError(t, err)
			closeDocument(t, native)

			reference, err := automerge.NewReference(ctx, actor(41))
			require.NoError(t, err)
			closeDocument(t, reference)

			scenario.apply(t, ctx, native)
			scenario.apply(t, ctx, reference)

			expected, err := reference.Save(ctx)
			require.NoError(t, err)

			actual, err := native.Save(ctx)
			require.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	}
}

// TestDocumentSave_ReloadsIntoTheSameHistory checks a compacted snapshot is a
// complete history rather than only the right bytes: it must reload with every
// change reachable and be accepted by the reference.
func TestDocumentSave_ReloadsIntoTheSameHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, scenario := range documentSaveScenarios() {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			document, err := automerge.New(ctx, actor(42))
			require.NoError(t, err)
			closeDocument(t, document)

			scenario.apply(t, ctx, document)

			heads, err := document.Heads(ctx)
			require.NoError(t, err)

			original, err := document.ChangesSince(ctx, nil)
			require.NoError(t, err)

			snapshot, err := document.Save(ctx)
			require.NoError(t, err)

			reloaded, err := automerge.Load(ctx, snapshot, actor(43))
			require.NoError(t, err)
			closeDocument(t, reloaded)

			reloadedHeads, err := reloaded.Heads(ctx)
			require.NoError(t, err)
			assert.Equal(t, heads, reloadedHeads)

			changes, err := reloaded.ChangesSince(ctx, nil)
			require.NoError(t, err)
			require.Len(t, changes, len(original))

			for i := range changes {
				assert.Equal(t, original[i].Hash, changes[i].Hash, "change %d", i)
				assert.Equal(t, original[i].Bytes, changes[i].Bytes, "change %d bytes", i)
			}

			adopted, err := automerge.LoadReference(ctx, snapshot, actor(44))
			require.NoError(t, err)
			closeDocument(t, adopted)

			adoptedHeads, err := adopted.Heads(ctx)
			require.NoError(t, err)
			assert.Equal(t, heads, adoptedHeads)
		})
	}
}
