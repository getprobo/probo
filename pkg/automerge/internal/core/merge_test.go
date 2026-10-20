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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackendMerge_AppliesReversedDependentChanges(t *testing.T) {
	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte{1}))
	text, err := base.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(text, 0, 0, "A"))
	_, err = base.Commit("base", time.Unix(1, 0))
	require.NoError(t, err)
	baseData, err := base.Save(true, true)
	require.NoError(t, err)

	source, err := LoadEngine(baseData)
	require.NoError(t, err)
	require.NoError(t, source.SetActor([]byte{2}))
	sourceText, err := source.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, source.SpliceText(sourceText, 1, 0, "B"))
	_, err = source.Commit("parent", time.Unix(2, 0))
	require.NoError(t, err)

	parent := append([]byte(nil), source.appended[len(source.appended)-1]...)

	require.NoError(t, source.SpliceText(sourceText, 2, 0, "C"))
	_, err = source.Commit("child", time.Unix(3, 0))
	require.NoError(t, err)

	child := append([]byte(nil), source.appended[len(source.appended)-1]...)

	retainedData := append([]byte(nil), baseData...)
	retainedData = append(retainedData, child...)
	retained, err := LoadEngine(retainedData)
	require.NoError(t, err)
	require.Len(t, retained.queuedChanges, 1)

	for _, queued := range retained.queuedChanges {
		for _, operation := range queued.Operations {
			assert.NotContains(t, retained.state.operations, operation.ID)
			_, ok := retained.lookupOperation(operation.ID)
			assert.False(t, ok)
		}
	}

	requireCanonicalRowsOwnedByColumns(t, retained)

	_, err = retained.Merge(parent)
	require.NoError(t, err)
	assert.Empty(t, retained.queuedChanges)
	retainedText, err := retained.GetText(0, "body")
	require.NoError(t, err)
	retainedValue, err := retained.Text(retainedText)
	require.NoError(t, err)
	assert.Equal(t, "ABC", retainedValue)

	target, err := LoadEngine(baseData)
	require.NoError(t, err)
	_, err = target.Merge(append(child, parent...))
	require.NoError(t, err)
	targetText, err := target.GetText(0, "body")
	require.NoError(t, err)
	value, err := target.Text(targetText)
	require.NoError(t, err)
	assert.Equal(t, "ABC", value)

	separate, err := LoadEngine(baseData)
	require.NoError(t, err)
	_, err = separate.Merge(child)
	require.NoError(t, err)
	assert.Len(t, separate.queuedChanges, 1)

	for _, queued := range separate.queuedChanges {
		for _, operation := range queued.Operations {
			assert.NotContains(t, separate.state.operations, operation.ID)
			_, ok := separate.lookupOperation(operation.ID)
			assert.False(t, ok)
		}
	}

	before := ReadRuntimeMetrics()
	_, err = separate.Merge(parent)
	require.NoError(t, err)

	after := ReadRuntimeMetrics()

	assert.Empty(t, separate.queuedChanges)
	assert.Equal(t, before.GeneralReconciles, after.GeneralReconciles)
	assert.Equal(t, before.GlobalOrderFallbacks, after.GlobalOrderFallbacks)
	assert.Equal(t, before.SnapshotReplacements, after.SnapshotReplacements)
	assert.Equal(t, before.FullColumnEncodings, after.FullColumnEncodings)
	assert.Equal(t, before.QueryIndexRescues, after.QueryIndexRescues)
	assert.Equal(t, before.DirectOperationCopies, after.DirectOperationCopies)
	assert.Equal(t, before.DirectColumnBatches+1, after.DirectColumnBatches)
	requireCanonicalRowsOwnedByColumns(t, separate)

	separateText, err := separate.GetText(0, "body")
	require.NoError(t, err)
	separateValue, err := separate.Text(separateText)
	require.NoError(t, err)
	assert.Equal(t, "ABC", separateValue)

	expected, err := target.Save(true, false)
	require.NoError(t, err)
	actual, err := separate.Save(true, false)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
