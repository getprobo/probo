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

// The tests in this file reproduce the mark-in-patch behaviors from the upstream
// JavaScript mark suite (javascript/test/marks.ts), asserting the native Go and
// Rust/WASM reference engines report identical Mark patches through the diff
// cursor.

package automerge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// TestJSMarks_MarksSeenInPatches reproduces "should allow marks that can be seen
// in patches": marking and then partially unmarking each emit a single Mark
// patch through the diff cursor, the unmark reporting its literal range with a
// null value rather than the resulting split.
func TestJSMarks_MarksSeenInPatches(t *testing.T) {
	t.Parallel()

	markPatches := make(map[string][]automerge.Patch)
	unmarkPatches := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "the quick fox jumps over the lazy dog")

		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, text.Mark(5, 10, "font-weight", markStr("bold"), automerge.MarkExpandNone))
		_, err := document.Commit("mark", commitTime.Add(time.Second))
		require.NoError(t, err)
		marked, err := document.DiffIncremental()
		require.NoError(t, err)

		markPatches[engine.name] = marked

		require.NoError(t, document.UpdateDiffCursor())
		require.NoError(t, text.Unmark(7, 9, "font-weight", automerge.MarkExpandNone))
		_, err = document.Commit("unmark", commitTime.Add(2*time.Second))
		require.NoError(t, err)
		unmarked, err := document.DiffIncremental()
		require.NoError(t, err)

		unmarkPatches[engine.name] = unmarked
	}

	require.Len(t, markPatches["reference"], 1)
	assert.Equal(t, automerge.PatchMark, markPatches["reference"][0].Action)
	require.Len(t, markPatches["reference"][0].Marks, 1)
	assert.Equal(t, uint32(5), markPatches["reference"][0].Marks[0].Start)
	assert.Equal(t, uint32(10), markPatches["reference"][0].Marks[0].End)
	assert.Equal(t, "bold", markPatches["reference"][0].Marks[0].Value.String)
	assert.Equal(t, markPatches["reference"], markPatches["native"])

	require.Len(t, unmarkPatches["reference"], 1)
	require.Len(t, unmarkPatches["reference"][0].Marks, 1)
	assert.Equal(t, uint32(7), unmarkPatches["reference"][0].Marks[0].Start)
	assert.Equal(t, uint32(9), unmarkPatches["reference"][0].Marks[0].End)
	assert.Equal(t, automerge.ScalarTypeNull, unmarkPatches["reference"][0].Marks[0].Value.Type)
	assert.Equal(t, unmarkPatches["reference"], unmarkPatches["native"])
}
