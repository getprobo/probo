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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// documentSaveScenario applies the same history to whichever engine it is given
// so a native snapshot can be held against the reference's byte for byte.
type documentSaveScenario struct {
	name  string
	apply func(t *testing.T, document *automerge.Document)
}

func documentSaveScenarios() []documentSaveScenario {
	base := time.Unix(1786147200, 0).UTC()

	return []documentSaveScenario{
		{
			name: "linear text",
			apply: func(t *testing.T, document *automerge.Document) {
				text, err := document.CreateText("body")
				require.NoError(t, err)

				for i := range 5 {
					require.NoError(t, text.Splice(uint32(i), 0, "x"))
					_, err = document.Commit("edit", base.Add(time.Duration(i)*time.Second))
					require.NoError(t, err)
				}
			},
		},
		{
			name: "map puts and delete",
			apply: func(t *testing.T, document *automerge.Document) {
				require.NoError(
					t,
					document.PutScalar(

						"title",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "first"},
					),
				)
				require.NoError(
					t,
					document.PutScalar(

						"keep",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "value"},
					),
				)
				_, err := document.Commit("one", base)
				require.NoError(t, err)

				require.NoError(
					t,
					document.PutScalar(

						"title",
						automerge.Scalar{Type: automerge.ScalarTypeString, String: "second"},
					),
				)
				_, err = document.Commit("two", base.Add(time.Second))
				require.NoError(t, err)

				require.NoError(t, document.Root().DeleteKey("title"))
				_, err = document.Commit("three", base.Add(2*time.Second))
				require.NoError(t, err)
			},
		},
		{
			name: "counter increment",
			apply: func(t *testing.T, document *automerge.Document) {
				require.NoError(
					t,
					document.PutScalar(

						"counter",
						automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 5},
					),
				)
				_, err := document.Commit("create", base)
				require.NoError(t, err)

				require.NoError(t, document.Root().Increment("counter", 3))
				_, err = document.Commit("bump", base.Add(time.Second))
				require.NoError(t, err)
			},
		},
		{
			name: "marks and unmarks",
			apply: func(t *testing.T, document *automerge.Document) {
				text, err := document.CreateText("body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(0, 0, "hello brave world"))

				_, err = document.Commit("write", base)
				require.NoError(t, err)

				require.NoError(
					t,
					text.Mark(

						0,
						5,
						"strong",
						automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
						automerge.MarkExpandBoth,
					),
				)

				_, err = document.Commit("mark", base.Add(time.Second))
				require.NoError(t, err)

				require.NoError(t, text.Unmark(1, 3, "strong", automerge.MarkExpandNone))

				_, err = document.Commit("unmark", base.Add(2*time.Second))
				require.NoError(t, err)
			},
		},
		{
			name: "text with deletion",
			apply: func(t *testing.T, document *automerge.Document) {
				text, err := document.CreateText("body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(0, 0, "hello brave world"))

				_, err = document.Commit("write", base)
				require.NoError(t, err)

				require.NoError(t, text.Splice(5, 6, ""))

				_, err = document.Commit("trim", base.Add(time.Second))
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

	for _, scenario := range documentSaveScenarios() {
		t.Run(
			scenario.name,
			func(t *testing.T) {
				t.Parallel()

				native, err := automerge.New(actor(41))
				require.NoError(t, err)
				closeDocument(t, native)

				reference, err := automerge.NewReference(actor(41))
				require.NoError(t, err)
				closeDocument(t, reference)

				scenario.apply(t, native)
				scenario.apply(t, reference)

				expected, err := reference.Save()
				require.NoError(t, err)

				actual, err := native.Save()
				require.NoError(t, err)

				assert.Equal(t, expected, actual)
			},
		)
	}
}

// TestDocumentSave_ReloadsIntoTheSameHistory checks a compacted snapshot is a
// complete history rather than only the right bytes: it must reload with every
// change reachable and be accepted by the reference.
func TestDocumentSave_ReloadsIntoTheSameHistory(t *testing.T) {
	t.Parallel()

	for _, scenario := range documentSaveScenarios() {
		t.Run(
			scenario.name,
			func(t *testing.T) {
				t.Parallel()

				document, err := automerge.New(actor(42))
				require.NoError(t, err)
				closeDocument(t, document)

				scenario.apply(t, document)

				heads, err := document.Heads()
				require.NoError(t, err)

				original, err := document.ChangesSince(nil)
				require.NoError(t, err)

				snapshot, err := document.Save()
				require.NoError(t, err)

				reloaded, err := automerge.Load(snapshot, actor(43))
				require.NoError(t, err)
				closeDocument(t, reloaded)

				reloadedHeads, err := reloaded.Heads()
				require.NoError(t, err)
				assert.Equal(t, heads, reloadedHeads)

				changes, err := reloaded.ChangesSince(nil)
				require.NoError(t, err)
				require.Len(t, changes, len(original))

				for i := range changes {
					assert.Equal(t, original[i].Hash, changes[i].Hash, "change %d", i)
					assert.Equal(t, original[i].Bytes, changes[i].Bytes, "change %d bytes", i)
				}

				adopted, err := automerge.LoadReference(snapshot, actor(44))
				require.NoError(t, err)
				closeDocument(t, adopted)

				adoptedHeads, err := adopted.Heads()
				require.NoError(t, err)
				assert.Equal(t, heads, adoptedHeads)
			},
		)
	}
}
