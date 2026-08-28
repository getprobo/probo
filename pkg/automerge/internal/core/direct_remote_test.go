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
