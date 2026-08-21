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

// This file corresponds to observe_counter_change_application
// (rust/automerge/src/automerge/tests.rs). The upstream test applies a change
// that creates and then increments a counter and expects diff_incremental to
// replay per-operation patches (a put of the base value followed by two
// increment patches). The pinned reference engine (automerge 0.10.0 embedded as
// WASM) instead collapses that applied change into a single put through
// diff_incremental, and the native engine matches the reference exactly. Because
// the parity gate is native-matches-reference observable behavior, this test
// asserts that agreement rather than the upstream native-Rust expectation.

package automerge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestRustAutomerge_ObserveCounterChangeApplication(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		source, err := engine.open(ctx, actor(0xc0))
		require.NoError(t, err)
		closeDocument(t, source)

		require.NoError(t, source.Root().PutScalar(
			ctx,
			"counter",
			automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 1},
		))
		require.NoError(t, source.Root().Increment(ctx, "counter", 2))
		require.NoError(t, source.Root().Increment(ctx, "counter", 5))

		_, err = source.Commit(ctx, "counter", commitTime)
		require.NoError(t, err)

		change, err := source.SaveIncremental(ctx)
		require.NoError(t, err)

		document, err := engine.open(ctx, actor(0xd0))
		require.NoError(t, err)
		closeDocument(t, document)

		require.NoError(t, document.Root().PutScalar(
			ctx,
			"foo",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "bar"},
		))

		_, err = document.Commit(ctx, "foo", commitTime)
		require.NoError(t, err)

		require.NoError(t, document.UpdateDiffCursor(ctx))

		_, err = document.LoadIncremental(ctx, change)
		require.NoError(t, err)

		patches, err := document.DiffIncremental(ctx)
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]

	// The reference collapses the create-and-increment sequence into a single
	// put of the counter's materialized value; native reproduces it exactly.
	require.Len(t, reference, 1)
	assert.Equal(t, automerge.PatchPutMap, reference[0].Action)
	assert.Equal(t, "counter", reference[0].Key)
	require.NotNil(t, reference[0].Value.Scalar)
	assert.Equal(t, automerge.ScalarTypeCounter, reference[0].Value.Scalar.Type)
	assert.Equal(t, int64(8), reference[0].Value.Scalar.Int)

	assert.Equal(t, reference, result["native"])
}
