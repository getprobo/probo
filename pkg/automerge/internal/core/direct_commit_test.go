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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestDirectCommit_OrdinaryCommitsAvoidCompatibilityWork(t *testing.T) {
	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutObject(0, "body", "text")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "hello"))

	before := ReadRuntimeMetrics()
	_, err = engine.Commit("typing", time.Time{})
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 2, 1, "i"))
	_, err = engine.Commit("middle", time.Time{})
	require.NoError(t, err)
	_, err = engine.EmptyCommit("empty", time.Time{})
	require.NoError(t, err)

	after := ReadRuntimeMetrics()

	assert.Equal(t, before.GeneralReconciles, after.GeneralReconciles)
	assert.Equal(t, before.GlobalOrderFallbacks, after.GlobalOrderFallbacks)
	assert.Equal(t, before.SnapshotReplacements, after.SnapshotReplacements)
	assert.Equal(t, before.FullColumnEncodings, after.FullColumnEncodings)
	assert.Equal(t, before.DirectColumnBatches+3, after.DirectColumnBatches)
	requireColumnarStateEquivalent(t, engine)
}

func TestDirectCommit_FailureKeepsPendingOverlayAndColumns(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "pending"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "key", value))
	columns := engine.columns
	heads := engine.currentHeads()
	engine.directColumnFailure = func() error {
		return errors.New("injected")
	}

	_, err = engine.Commit("failure", time.Time{})
	require.ErrorContains(t, err, "injected")
	assert.Same(t, columns, engine.columns)
	assert.Equal(t, heads, engine.currentHeads())
	assert.Len(t, engine.pending, 1)
	assert.Equal(t, 0, engine.state.changeCount())

	engine.directColumnFailure = nil
	_, err = engine.Commit("success", time.Time{})
	require.NoError(t, err)
	requireColumnarStateEquivalent(t, engine)
}

func TestDirectCommit_ForkUsesCopyOnWriteColumns(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "base"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "base", value))
	_, err = engine.Commit("base", time.Time{})
	require.NoError(t, err)
	parentBytes, err := engine.columns.snapshot.Encode(nil, false)
	require.NoError(t, err)

	fork, err := engine.Fork([]byte{42})
	require.NoError(t, err)
	forkValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "fork"},
	)
	require.NoError(t, err)
	require.NoError(t, fork.PutScalar(0, "fork", forkValue))
	_, err = fork.Commit("fork", time.Time{})
	require.NoError(t, err)

	after, err := engine.columns.snapshot.Encode(nil, false)
	require.NoError(t, err)
	assert.Equal(t, parentBytes, after)
	assert.NotSame(t, engine.columns.snapshot, fork.columns.snapshot)
	requireColumnarStateEquivalent(t, engine)
	requireColumnarStateEquivalent(t, fork)
}

func TestDirectCommit_IsolationPreservesHiddenCanonicalRows(t *testing.T) {
	t.Parallel()

	engine, first := isolationFixture(t)
	require.NoError(t, engine.Isolate([][32]byte{first}))

	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "isolated"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "isolated", value))
	_, err = engine.Commit("isolated", time.Time{})
	require.NoError(t, err)
	require.NoError(t, engine.Integrate())

	keys, err := engine.Keys(0)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"isolated", "text", "value"}, keys)
	requireColumnarStateEquivalent(t, engine)
}
