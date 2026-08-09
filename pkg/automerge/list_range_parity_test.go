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

// The tests in this file reproduce the list-range behaviors from upstream Rust
// automerge 0.10 (rust/automerge/src/iter/list_range.rs), asserting the native
// Go and Rust/WASM reference engines expose the same list values and per-element
// conflict flags.

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

// TestRust_ReproduceClockCacheBug reproduces reproduce_clock_cache_bug: after
// merging many branches authored by distinct actors, no change lies outside the
// merged frontier, so ChangesSince(heads) is empty. A clock-caching defect would
// omit some ancestors and report spurious changes.
func TestRust_ReproduceClockCacheBug(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	base, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, base)

	for i := 0; i < 20; i++ {
		require.NoError(t, base.Root().PutScalar(
			ctx,
			"initial_commit",
			automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
		))
		_, err := base.Commit(ctx, "initial", commitTime.Add(time.Duration(i)))
		require.NoError(t, err)
	}

	const branches = 20

	for branch := 0; branch < branches; branch++ {
		fork, err := base.Fork(ctx, actor(byte(30+branch)))
		require.NoError(t, err)
		closeDocument(t, fork)

		for commit := 0; commit < 2; commit++ {
			require.NoError(t, fork.Root().PutScalar(
				ctx,
				"branch_value",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(branch*10 + commit)},
			))
			_, err := fork.Commit(ctx, "branch", commitTime.Add(time.Duration(branch*10+commit)))
			require.NoError(t, err)
		}

		_, err = base.Merge(ctx, fork)
		require.NoError(t, err)
	}

	heads, err := base.Heads(ctx)
	require.NoError(t, err)

	changes, err := base.ChangesSince(ctx, heads)
	require.NoError(t, err)
	assert.Empty(t, changes)
}

// TestRustListRange_Bounds reproduces list_range_bounds: reading the list yields
// its values in order.
func TestRustListRange_Bounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]int64)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		list, err := document.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
		require.NoError(t, err)

		for index, value := range []int64{1, 2, 3, 4, 5} {
			require.NoError(t, list.InsertScalar(
				ctx,
				uint64(index),
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: value},
			))
		}

		_, err = document.Commit(ctx, "list", commitTime)
		require.NoError(t, err)

		length, err := list.Len(ctx)
		require.NoError(t, err)

		values := make([]int64, 0, length)
		for index := uint64(0); index < length; index++ {
			scalar, err := list.ScalarAt(ctx, index)
			require.NoError(t, err)
			values = append(values, scalar.Int)
		}

		result[engine.name] = values
	}

	assert.Equal(t, []int64{1, 2, 3, 4, 5}, result["reference"])
	assert.Equal(t, result["reference"], result["native"])
}

// TestRustListRange_Conflict reproduces list_range_conflict: a concurrently
// overwritten element is reported as conflicted with the winning value.
func TestRustListRange_Conflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	values := make(map[string][]int64)
	conflicts := make(map[string][]bool)

	for _, engine := range rustParityEngines() {
		document, err := engine.open(ctx, actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, document)

		list, err := document.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
		require.NoError(t, err)

		for index, value := range []int64{1, 2, 3, 4, 5} {
			require.NoError(t, list.InsertScalar(
				ctx,
				uint64(index),
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: value},
			))
		}

		_, err = document.Commit(ctx, "list", commitTime)
		require.NoError(t, err)

		other, err := document.Fork(ctx, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, other)

		otherList, err := other.Root().Object(ctx, "list")
		require.NoError(t, err)
		require.NoError(t, otherList.PutScalarAt(ctx, 3, automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 11}))
		_, err = other.Commit(ctx, "other", commitTime.Add(1))
		require.NoError(t, err)

		require.NoError(t, list.PutScalarAt(ctx, 3, automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 10}))
		_, err = document.Commit(ctx, "mine", commitTime.Add(1))
		require.NoError(t, err)

		_, err = other.Merge(ctx, document)
		require.NoError(t, err)

		length, err := otherList.Len(ctx)
		require.NoError(t, err)

		rowValues := make([]int64, 0, length)
		rowConflicts := make([]bool, 0, length)

		for index := uint64(0); index < length; index++ {
			scalar, err := otherList.ScalarAt(ctx, index)
			require.NoError(t, err)
			rowValues = append(rowValues, scalar.Int)

			all, err := otherList.ScalarsAt(ctx, index)
			require.NoError(t, err)
			rowConflicts = append(rowConflicts, len(all) > 1)
		}

		values[engine.name] = rowValues
		conflicts[engine.name] = rowConflicts
	}

	assert.Equal(t, []bool{false, false, false, true, false}, conflicts["reference"])
	assert.Equal(t, values["reference"], values["native"])
	assert.Equal(t, conflicts["reference"], conflicts["native"])
}
