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
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestDirectRemote_InvalidLateChangeIsAtomic(t *testing.T) {
	base, first, second := directRemoteFixture(t)
	document, err := storage.DecodePartial(second)
	require.NoError(t, err)
	require.Len(t, document.Changes, 1)
	document.Changes[0].Sequence++
	invalid, err := storage.EncodeChange(&document.Changes[0])
	require.NoError(t, err)

	target, err := LoadEngine(base)
	require.NoError(t, err)
	beforeHeads, err := target.Heads()
	require.NoError(t, err)

	beforeChanges := target.state.changeCount()
	beforeColumns := target.columns

	err = target.ApplyChanges([][]byte{first, invalid})
	require.Error(t, err)
	assert.Equal(t, beforeHeads, mustHeads(t, target))
	assert.Equal(t, beforeChanges, target.state.changeCount())
	assert.Same(t, beforeColumns, target.columns)
	assert.Empty(t, target.queuedChanges)
}

func TestDirectRemote_FailedColumnBatchIsAtomic(t *testing.T) {
	base, first, _ := directRemoteFixture(t)
	target, err := LoadEngine(base)
	require.NoError(t, err)
	beforeHeads := mustHeads(t, target)
	beforeColumns := target.columns
	target.directColumnFailure = func() error {
		return errors.New("injected remote failure")
	}

	err = target.ApplyChanges([][]byte{first})
	require.ErrorContains(t, err, "injected remote failure")
	assert.Equal(t, beforeHeads, mustHeads(t, target))
	assert.Same(t, beforeColumns, target.columns)
	assert.Empty(t, target.queuedChanges)
}

func TestDirectRemote_OutOfOrderSetPublishesOneBatch(t *testing.T) {
	base, first, second := directRemoteFixture(t)
	target, err := LoadEngine(base)
	require.NoError(t, err)
	ResetRuntimeMetrics()

	require.NoError(t, target.ApplyChanges([][]byte{second, first}))

	metrics := ReadRuntimeMetrics()
	assert.Equal(t, uint64(1), metrics.DirectColumnBatches)
	assert.Zero(t, metrics.GeneralReconciles)
	assert.Zero(t, metrics.GlobalOrderFallbacks)
	assert.Zero(t, metrics.SnapshotReplacements)
	assert.Zero(t, metrics.FullColumnEncodings)
	assert.Empty(t, target.queuedChanges)

	firstDocument, err := storage.DecodePartial(first)
	require.NoError(t, err)
	require.Len(t, firstDocument.Changes, 1)
	require.NotNil(t, firstDocument.Changes[0].Hash)

	secondDocument, err := storage.DecodePartial(second)
	require.NoError(t, err)
	require.Len(t, secondDocument.Changes, 1)
	require.NotNil(t, secondDocument.Changes[0].Hash)
	require.GreaterOrEqual(t, len(target.columns.changes), 2)
	assert.Equal(
		t,
		*firstDocument.Changes[0].Hash,
		*target.columns.changes[len(target.columns.changes)-2].Hash,
	)
	assert.Equal(
		t,
		*secondDocument.Changes[0].Hash,
		*target.columns.changes[len(target.columns.changes)-1].Hash,
	)
}

func TestDirectRemote_HeadBranchBeforeFirstRowSurvivesFollowUp(t *testing.T) {
	t.Parallel()

	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte{0x01}))
	text, err := base.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(text, 0, 0, "ab"))
	_, err = base.Commit("base", time.Time{})
	require.NoError(t, err)
	baseHeads := mustHeads(t, base)
	baseData, err := base.Save(true, false)
	require.NoError(t, err)

	remote, err := LoadEngine(baseData)
	require.NoError(t, err)
	require.NoError(t, remote.SetActor([]byte{0xff}))
	remoteText, err := remote.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, remote.SpliceText(remoteText, 0, 0, "R"))
	_, err = remote.Commit("remote", time.Time{})
	require.NoError(t, err)
	changes, _, err := remote.ChangesSince(baseHeads)
	require.NoError(t, err)
	require.Len(t, changes, 1)

	target, err := LoadEngine(baseData)
	require.NoError(t, err)
	require.NoError(t, target.SetActor([]byte{0x02}))
	require.NoError(t, target.ApplyChanges(changes))

	targetText, err := target.GetText(0, "body")
	require.NoError(t, err)
	inserted, err := target.Text(targetText)
	require.NoError(t, err)
	require.Equal(t, "Rab", inserted)

	object := target.objects[targetText]
	span, ok := target.columns.objectSpans[object]
	require.True(t, ok)
	require.Greater(t, span.end, span.start)
	require.NotEmpty(t, remote.columns.changes)
	remoteChange := remote.columns.changes[len(remote.columns.changes)-1]
	require.NotEmpty(t, remoteChange.Operations)
	require.Equal(t, remoteChange.Operations[0].ID, target.columns.operations[span.start].ID)

	require.NoError(t, target.SpliceText(targetText, 2, 0, "Z"))
	_, err = target.Commit("follow-up", time.Time{})
	require.NoError(t, err)

	value, err := target.Text(targetText)
	require.NoError(t, err)
	assert.Equal(t, "RaZb", value)
	requireColumnarStateEquivalent(t, target)
}

func TestDirectRemote_ConcurrentSingleInsertRoundTrips(t *testing.T) {
	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte("base")))
	text, err := base.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(text, 0, 0, "base"))
	_, err = base.Commit("base", time.Time{})
	require.NoError(t, err)
	baseHeads := mustHeads(t, base)

	left, err := base.Fork([]byte("left"))
	require.NoError(t, err)
	right, err := base.Fork([]byte("right"))
	require.NoError(t, err)
	leftText, err := left.GetText(0, "body")
	require.NoError(t, err)
	rightText, err := right.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, left.SpliceText(leftText, 4, 0, "L"))
	_, err = left.Commit("left", time.Time{})
	require.NoError(t, err)
	require.NoError(t, right.SpliceText(rightText, 4, 0, "R"))
	_, err = right.Commit("right", time.Time{})
	require.NoError(t, err)
	changes, _, err := right.ChangesSince(baseHeads)
	require.NoError(t, err)

	require.NoError(t, left.ApplyChanges(changes))
	saved, err := left.Save(true, false)
	require.NoError(t, err)
	_, err = LoadEngine(saved)
	require.NoError(t, err)
}

func TestDirectRemote_ObjectCreationSplicesOrderedSpans(t *testing.T) {
	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte("zzzz")))
	first, err := base.PutObject(0, "first", "map")
	require.NoError(t, err)
	baseValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "base"},
	)
	require.NoError(t, err)
	require.NoError(t, base.PutScalar(first, "value", baseValue))
	_, err = base.Commit("base object", time.Time{})
	require.NoError(t, err)
	baseHeads := mustHeads(t, base)
	baseData, err := base.Save(true, false)
	require.NoError(t, err)

	source, err := LoadEngine(baseData)
	require.NoError(t, err)
	require.NoError(t, source.SetActor([]byte("aaaa")))
	second, err := source.PutObject(0, "second", "map")
	require.NoError(t, err)
	remoteValue, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "remote"},
	)
	require.NoError(t, err)
	require.NoError(t, source.PutScalar(second, "value", remoteValue))
	_, err = source.Commit("remote object", time.Time{})
	require.NoError(t, err)
	changes, _, err := source.ChangesSince(baseHeads)
	require.NoError(t, err)
	expected, err := source.Save(true, false)
	require.NoError(t, err)

	target, err := LoadEngine(baseData)
	require.NoError(t, err)
	ResetRuntimeMetrics()
	require.NoError(t, target.ApplyChanges(changes))
	actual, err := target.Save(true, false)
	require.NoError(t, err)

	metrics := ReadRuntimeMetrics()
	assert.Equal(t, uint64(1), metrics.DirectColumnBatches)
	assert.Zero(t, metrics.GeneralReconciles)
	assert.Zero(t, metrics.GlobalOrderFallbacks)
	assert.Zero(t, metrics.SnapshotReplacements)
	assert.Zero(t, metrics.FullColumnEncodings)
	assert.Zero(t, metrics.QueryIndexRescues)
	assert.Zero(t, metrics.DirectOperationCopies)
	assert.Equal(t, expected, actual)
	requireColumnarStateEquivalent(t, target)
}

func directRemoteFixture(t *testing.T) ([]byte, []byte, []byte) {
	t.Helper()

	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte("base")))
	text, err := base.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(text, 0, 0, "base"))
	_, err = base.Commit("base", time.Time{})
	require.NoError(t, err)
	baseData, err := base.Save(true, false)
	require.NoError(t, err)

	source, err := LoadEngine(baseData)
	require.NoError(t, err)
	require.NoError(t, source.SetActor([]byte("remote")))
	text, err = source.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, source.SpliceText(text, 4, 0, "1"))
	_, err = source.Commit("first", time.Time{})
	require.NoError(t, err)
	first, err := source.SaveIncremental()
	require.NoError(t, err)

	require.NoError(t, source.SpliceText(text, 5, 0, "2"))
	_, err = source.Commit("second", time.Time{})
	require.NoError(t, err)
	second, err := source.SaveIncremental()
	require.NoError(t, err)

	return baseData, first, second
}

func mustHeads(t *testing.T, engine *Engine) [][32]byte {
	t.Helper()

	heads, err := engine.Heads()
	require.NoError(t, err)

	return heads
}
