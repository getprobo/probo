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

package core

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

// objectOperationIDs lists the operations eachObjectOperation yields for an
// object alongside the ones a full scan attributes to it.
func objectOperationIDs(
	state *State,
	object opset.ObjectID,
) (indexed, scanned []opset.OpID) {
	state.eachObjectOperation(object, func(operation opset.Operation) bool {
		indexed = append(indexed, operation.ID)
		return true
	})

	state.eachOperation(func(operation opset.Operation) bool {
		if operation.Object == object {
			scanned = append(scanned, operation.ID)
		}

		return true
	})

	compare := func(left, right opset.OpID) int { return left.Compare(right) }
	slices.SortFunc(indexed, compare)
	slices.SortFunc(scanned, compare)

	return indexed, scanned
}

func assertObjectOperationsMatchScan(t *testing.T, state *State) {
	t.Helper()

	objects := state.documentObjects()
	require.Greater(t, len(objects), 3)

	for _, object := range objects {
		indexed, scanned := objectOperationIDs(state, object)

		assert.Equal(t, scanned, indexed, "object %v", object)
	}
}

func TestState_EachObjectOperation(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte("object-operations")))
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "first block second block"))

	encoded, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarTrue})
	require.NoError(t, err)
	require.NoError(t, engine.MarkText(text, 2, 9, "bold", encoded, "after"))

	for _, index := range []uint32{0, 12} {
		block, splitErr := engine.SplitBlock(text, index)
		require.NoError(t, splitErr)
		require.NoError(t, engine.PutString(0, "title", "register"))

		_, putErr := engine.PutObject(block, "attrs", "map")
		require.NoError(t, putErr)
		_, putErr = engine.PutObject(block, "parents", "list")
		require.NoError(t, putErr)
	}

	t.Run(
		"open transaction",
		func(t *testing.T) {
			assertObjectOperationsMatchScan(t, engine.state)
		},
	)

	_, err = engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	t.Run(
		"committed columns",
		func(t *testing.T) {
			assertObjectOperationsMatchScan(t, engine.state)
		},
	)

	// Deletes leave no row in the wire operation table, so they exercise the
	// non-canonical side of the lookup across several commits.
	random := rand.New(rand.NewPCG(3, 5))

	for range 20 {
		current, textErr := engine.Text(text)
		require.NoError(t, textErr)

		position := random.IntN(utf16Width(current) - 1)
		require.NoError(t, engine.SpliceText(text, uint32(position), 1, "x"))
		require.NoError(t, engine.PutString(0, "title", strconv.Itoa(position)))

		if random.IntN(2) == 0 {
			_, err = engine.Commit("edit", time.Time{})
			require.NoError(t, err)
		}

		assertObjectOperationsMatchScan(t, engine.state)
	}

	require.NoError(t, engine.SpliceText(text, 0, 2, "rolled back"))
	_, err = engine.SplitBlock(text, 4)
	require.NoError(t, err)
	_, err = engine.Rollback()
	require.NoError(t, err)

	t.Run(
		"after rollback",
		func(t *testing.T) {
			assertObjectOperationsMatchScan(t, engine.state)
			assert.Empty(t, engine.state.overlayObjects)
		},
	)

	require.NoError(t, engine.DeleteMap(0, "title"))
	_, err = engine.Commit("delete", time.Time{})
	require.NoError(t, err)

	t.Run(
		"after deletes",
		func(t *testing.T) {
			assertObjectOperationsMatchScan(t, engine.state)
		},
	)

	data, err := engine.Save(true, false)
	require.NoError(t, err)
	reloaded, err := LoadEngine(data)
	require.NoError(t, err)

	t.Run(
		"reloaded snapshot",
		func(t *testing.T) {
			assertObjectOperationsMatchScan(t, reloaded.state)
		},
	)

	fork, err := engine.Fork([]byte("object-operations-fork"))
	require.NoError(t, err)
	forkText, err := fork.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, fork.SpliceText(forkText, 0, 3, "forked"))
	_, err = fork.Commit("fork edit", time.Time{})
	require.NoError(t, err)

	t.Run(
		"fork leaves the parent index intact",
		func(t *testing.T) {
			assertObjectOperationsMatchScan(t, fork.state)
			assertObjectOperationsMatchScan(t, engine.state)
		},
	)
}

func TestMarkSweep_MatchesPerIndexScan(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(17, 23))

	for range 200 {
		length := 1 + random.IntN(40)
		marks := make([]richTextMark, random.IntN(12))

		for i := range marks {
			start := random.IntN(length)
			marks[i] = richTextMark{
				start: start,
				end:   start + 1 + random.IntN(length-start),
			}
		}

		sweep := newMarkSweep(marks)
		previous := []int(nil)

		// Skipping indexes mirrors a walk that only consults the sweep on some
		// elements.
		for index := 0; index < length; index += 1 + random.IntN(3) {
			want := make([]int, 0)

			for rank, mark := range marks {
				if index >= mark.start && index < mark.end {
					want = append(want, rank)
				}
			}

			changed := sweep.advance(index)

			require.ElementsMatch(t, want, sweep.covering)
			require.True(t, slices.IsSorted(sweep.covering))

			if !changed {
				require.Equal(t, previous, slices.Clone(sweep.covering))
			}

			previous = slices.Clone(sweep.covering)
		}
	}
}
