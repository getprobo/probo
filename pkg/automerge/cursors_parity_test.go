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

// The tests in this file reproduce the historical-cursor behavior from the
// upstream JavaScript cursor suite (javascript/test/cursors.ts), asserting the
// native Go and Rust/WASM reference engines resolve cursors created at a past
// frontier identically.

package automerge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// TestJSCursors_GetCursorRespectsHeads reproduces "getCursor should respect
// heads": cursors created against a historical view resolve to the expected
// positions in the current document.
func TestJSCursors_GetCursorRespectsHeads(t *testing.T) {
	t.Parallel()

	positions := make(map[string][]uint32)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "aaa@bbb")

		frontier, err := document.Heads()
		require.NoError(t, err)

		require.NoError(t, text.Splice(3, 1, "~~~"))

		_, err = document.Commit("replace", commitTime.Add(1))
		require.NoError(t, err)

		before, err := text.CursorForAt(3, automerge.CursorMoveBefore, frontier)
		require.NoError(t, err)
		after, err := text.CursorForAt(3, automerge.CursorMoveAfter, frontier)
		require.NoError(t, err)

		start, err := text.CursorPosition(automerge.StartCursor())
		require.NoError(t, err)
		beforePosition, err := text.CursorPosition(before)
		require.NoError(t, err)
		afterPosition, err := text.CursorPosition(after)
		require.NoError(t, err)
		end, err := text.CursorPosition(automerge.EndCursor())
		require.NoError(t, err)

		positions[engine.name] = []uint32{start, beforePosition, afterPosition, end}
	}

	assert.Equal(t, []uint32{0, 2, 6, 9}, positions["reference"])
	assert.Equal(t, positions["reference"], positions["native"])
}
