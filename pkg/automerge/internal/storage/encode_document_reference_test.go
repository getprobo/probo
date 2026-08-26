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

package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/testsupport/reference"
)

func boolScalar() []byte    { return []byte(`{"type":"boolean","bool":true}`) }
func nullScalar() []byte    { return []byte(`{"type":"null"}`) }
func counterScalar() []byte { return []byte(`{"type":"counter","int":5}`) }

func stringScalar(value string) []byte {
	return []byte(`{"type":"string","string":"` + value + `"}`)
}

func newReference(t *testing.T, actor byte) *reference.Engine {
	t.Helper()

	engine, err := reference.New()
	require.NoError(t, err)
	require.NoError(
		t,
		engine.SetActor(
			[]byte{
				actor, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
			},
		),
	)

	return engine
}

// assertSnapshotReencodes decodes a reference-written snapshot and requires the
// encoder to reproduce it byte for byte.
func assertSnapshotReencodes(t *testing.T, saved []byte) {
	t.Helper()

	document, err := Decode(saved)
	require.NoError(t, err)

	encoded, err := EncodeDocument(document, storedOperationOrder(t, saved), true)
	require.NoError(t, err)

	assert.Equal(t, saved, encoded)
}

// TestEncodeDocument_MatchesReferenceSnapshots pins snapshot writing against the
// reference implementation across the shapes whose storage form differs most
// from a change: deletes that survive only as successors, marks whose expand
// column is shared, counters, nested objects and several actors.
func TestEncodeDocument_MatchesReferenceSnapshots(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		build func(t *testing.T, engine *reference.Engine) []byte
	}{
		{
			name: "linear text history",
			build: func(t *testing.T, engine *reference.Engine) []byte {
				handle, err := engine.PutText(0, "body")
				require.NoError(t, err)

				for i := range 5 {
					require.NoError(t, engine.SpliceText(handle, uint32(i), 0, "x"))
					_, err = engine.Commit("edit", time.Unix(int64(i+1), 0))
					require.NoError(t, err)
				}

				saved, err := engine.Save(true, true)
				require.NoError(t, err)

				return saved
			},
		},
		{
			name: "marks and unmarks",
			build: func(t *testing.T, engine *reference.Engine) []byte {
				handle, err := engine.PutText(0, "body")
				require.NoError(t, err)
				require.NoError(t, engine.SpliceText(handle, 0, 0, "hello brave world"))
				_, err = engine.Commit("write", time.Unix(1, 0))
				require.NoError(t, err)

				require.NoError(t, engine.MarkText(handle, 0, 5, "strong", boolScalar(), "both"))
				_, err = engine.Commit("mark", time.Unix(2, 0))
				require.NoError(t, err)

				require.NoError(t, engine.MarkText(handle, 1, 3, "strong", nullScalar(), "none"))
				_, err = engine.Commit("unmark", time.Unix(3, 0))
				require.NoError(t, err)

				saved, err := engine.Save(true, true)
				require.NoError(t, err)

				return saved
			},
		},
		{
			name: "deletes and overwrites",
			build: func(t *testing.T, engine *reference.Engine) []byte {
				require.NoError(t, engine.PutString(0, "title", "first"))
				require.NoError(t, engine.PutString(0, "keep", "value"))
				_, err := engine.Commit("one", time.Unix(1, 0))
				require.NoError(t, err)

				require.NoError(t, engine.PutString(0, "title", "second"))
				_, err = engine.Commit("two", time.Unix(2, 0))
				require.NoError(t, err)

				require.NoError(t, engine.DeleteMap(0, "title"))
				_, err = engine.Commit("three", time.Unix(3, 0))
				require.NoError(t, err)

				saved, err := engine.Save(true, true)
				require.NoError(t, err)

				return saved
			},
		},
		{
			name: "list with deletion and counter",
			build: func(t *testing.T, engine *reference.Engine) []byte {
				list, err := engine.PutObject(0, "items", "list")
				require.NoError(t, err)
				require.NoError(t, engine.InsertScalar(list, 0, stringScalar("a")))
				require.NoError(t, engine.InsertScalar(list, 1, stringScalar("b")))
				require.NoError(t, engine.InsertScalar(list, 2, stringScalar("c")))
				require.NoError(t, engine.PutScalar(0, "counter", counterScalar()))
				_, err = engine.Commit("build", time.Unix(1, 0))
				require.NoError(t, err)

				require.NoError(t, engine.DeleteSequence(list, 1))
				require.NoError(t, engine.Increment(0, "counter", 3))
				_, err = engine.Commit("trim", time.Unix(2, 0))
				require.NoError(t, err)

				saved, err := engine.Save(true, true)
				require.NoError(t, err)

				return saved
			},
		},
		{
			name: "nested objects",
			build: func(t *testing.T, engine *reference.Engine) []byte {
				outer, err := engine.PutObject(0, "outer", "map")
				require.NoError(t, err)

				inner, err := engine.PutObject(outer, "inner", "list")
				require.NoError(t, err)
				require.NoError(t, engine.InsertScalar(inner, 0, stringScalar("deep")))

				text, err := engine.PutText(outer, "note")
				require.NoError(t, err)
				require.NoError(t, engine.SpliceText(text, 0, 0, "nested"))

				_, err = engine.Commit("nest", time.Unix(1, 0))
				require.NoError(t, err)

				saved, err := engine.Save(true, true)
				require.NoError(t, err)

				return saved
			},
		},
	} {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				assertSnapshotReencodes(t, testCase.build(t, newReference(t, 0x20)))
			},
		)
	}
}

// TestEncodeDocument_MatchesReferenceConcurrentSnapshot covers a merged history,
// where several actors share the actor table and a change has more than one
// dependency.
func TestEncodeDocument_MatchesReferenceConcurrentSnapshot(t *testing.T) {
	t.Parallel()

	first := newReference(t, 0x20)

	require.NoError(t, first.PutString(0, "title", "one"))

	body, err := first.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, first.SpliceText(body, 0, 0, "hello"))
	_, err = first.Commit("first", time.Unix(1, 0))
	require.NoError(t, err)

	shared, err := first.Save(true, true)
	require.NoError(t, err)

	second, err := reference.Load(shared)
	require.NoError(t, err)
	require.NoError(
		t,
		second.SetActor(
			[]byte{
				0x10, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
			},
		),
	)

	secondBody, _, err := second.GetObject(0, "body")
	require.NoError(t, err)
	require.NoError(t, second.SpliceText(secondBody, 5, 0, " there"))
	require.NoError(t, second.PutString(0, "title", "two"))
	_, err = second.Commit("second", time.Unix(2, 0))
	require.NoError(t, err)

	secondSave, err := second.Save(true, true)
	require.NoError(t, err)

	_, err = first.Merge(secondSave)
	require.NoError(t, err)

	require.NoError(t, first.SpliceText(body, 0, 1, ""))
	_, err = first.Commit("third", time.Unix(3, 0))
	require.NoError(t, err)

	saved, err := first.Save(true, true)
	require.NoError(t, err)

	assertSnapshotReencodes(t, saved)
}
