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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestSequenceIndex_RangeMatchesFlatReference(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "A😀BC👋D"))

	object := engine.objects[text].OpID
	for index := uint32(0); index <= 8; index++ {
		for deleteCount := uint32(0); deleteCount <= 10; deleteCount++ {
			elements := engine.state.sequenceElements(object)
			offsets := engine.state.sequenceOffsets(object, elements)
			wantStart, wantEnd, wantPrevious, wantErr := sequenceRange(
				elements,
				offsets,
				index,
				deleteCount,
			)
			got, gotErr := engine.state.sequenceIndex(object).rangeAt(index, deleteCount)

			if wantErr != nil {
				require.Error(t, gotErr)
				continue
			}

			require.NoError(t, gotErr)
			assert.Equal(t, wantStart, got.start)
			assert.Equal(t, wantEnd, got.end)
			assert.Equal(t, wantPrevious, got.previous)
			assert.Len(t, got.targets, wantEnd-wantStart)
			if wantEnd > wantStart {
				assert.Equal(t, elements[wantStart:wantEnd], got.targets)
			}
		}
	}
}

func TestSequenceIndex_LocalTextEditsCopyOnlyTouchedChunks(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, strings.Repeat("a", 600)))

	object := engine.objects[text].OpID
	index := engine.state.sequenceIndex(object)
	require.Greater(t, len(index.chunks), 3)
	firstChunk := index.chunks[0]
	lastChunk := index.chunks[len(index.chunks)-1]

	require.NoError(t, engine.SpliceText(text, 300, 1, "😀"))

	assert.Same(t, index, engine.state.sequenceIndexes[object])
	assert.Same(t, firstChunk, index.chunks[0])
	assert.Same(t, lastChunk, index.chunks[len(index.chunks)-1])
	assert.Equal(t, uint32(601), index.utf16Width)
	assert.NotContains(t, engine.state.sequenceElementsCache, object)
	assert.NotContains(t, engine.state.sequenceValuesCache, object)
	assert.NotContains(t, engine.state.sequenceOffsetCache, object)

	value, err := engine.Text(text)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", 300)+"😀"+strings.Repeat("a", 299), value)
}

func TestSequenceIndex_RandomEditsMatchReloadedState(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "seed😀text"))

	random := rand.New(rand.NewPCG(7, 11))
	for range 100 {
		current, err := engine.Text(text)
		require.NoError(t, err)
		width := utf16Width(current)
		index := random.IntN(width + 1)
		deleteCount := 0
		if index < width {
			deleteCount = random.IntN(3)
		}
		insertions := []string{"", "x", "😀", "yz"}
		value := insertions[random.IntN(len(insertions))]

		require.NoError(t, engine.SpliceText(
			text,
			uint32(index),
			int32(deleteCount),
			value,
		))
	}

	want, err := engine.Text(text)
	require.NoError(t, err)
	_, err = engine.Commit("edits", time.Time{})
	require.NoError(t, err)
	data, err := engine.Save(true, false)
	require.NoError(t, err)
	reloaded, err := LoadEngine(data)
	require.NoError(t, err)
	reloadedText, err := reloaded.GetText(0, "body")
	require.NoError(t, err)
	got, err := reloaded.Text(reloadedText)
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestSequenceIndex_RollbackDropsLocalizedState(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "before"))
	_, err = engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	object := engine.objects[text].OpID
	engine.state.sequenceIndex(object)
	require.NoError(t, engine.SpliceText(text, 3, 2, "XYZ"))
	assert.Contains(t, engine.state.sequenceIndexes, object)

	_, err = engine.Rollback()
	require.NoError(t, err)
	assert.NotContains(t, engine.state.sequenceIndexes, object)

	value, err := engine.Text(text)
	require.NoError(t, err)
	assert.Equal(t, "before", value)
}

func TestSequenceIndex_ForkAndMergedUpdatesStayIndependent(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte{1}))
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "abc"))
	_, err = engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	appendFork, err := engine.Fork([]byte{2})
	require.NoError(t, err)
	appendText, err := appendFork.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, appendFork.SpliceText(appendText, 3, 0, "d"))
	_, err = appendFork.Commit("append", time.Time{})
	require.NoError(t, err)
	appendData, err := appendFork.Save(true, false)
	require.NoError(t, err)

	object := engine.objects[text].OpID
	index := engine.state.sequenceIndex(object)
	_, err = engine.Merge(appendData)
	require.NoError(t, err)
	assert.Same(t, index, engine.state.sequenceIndexes[object])

	middleFork, err := engine.Fork([]byte{3})
	require.NoError(t, err)
	middleText, err := middleFork.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, middleFork.SpliceText(middleText, 1, 0, "X"))
	_, err = middleFork.Commit("middle", time.Time{})
	require.NoError(t, err)
	middleData, err := middleFork.Save(true, false)
	require.NoError(t, err)

	engine.state.sequenceIndex(object)
	_, err = engine.Merge(middleData)
	require.NoError(t, err)
	assert.Same(t, index, engine.state.sequenceIndexes[object])
	assert.False(t, engine.state.sequenceRescue[object])
	assert.NotContains(t, engine.state.sequenceElementsCache, object)
	assert.NotContains(t, engine.state.sequenceOffsetCache, object)

	value, err := engine.Text(text)
	require.NoError(t, err)
	assert.Equal(t, "aXbcd", value)
	appendValue, err := appendFork.Text(appendText)
	require.NoError(t, err)
	assert.Equal(t, "abcd", appendValue)
}

func TestSequenceIndex_RandomRichTextEditsAvoidFlatFallback(t *testing.T) {
	t.Parallel()

	for _, expand := range []string{"before", "after", "both", "none"} {
		t.Run(expand, func(t *testing.T) {
			t.Parallel()

			engine, err := NewEngine()
			require.NoError(t, err)
			require.NoError(t, engine.SetActor([]byte("rich-"+expand)))
			text, err := engine.PutText(0, "body")
			require.NoError(t, err)
			require.NoError(t, engine.SpliceText(text, 0, 0, "A😀BC👋D"))
			encoded, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarTrue})
			require.NoError(t, err)
			require.NoError(t, engine.MarkText(text, 1, 7, "bold", encoded, expand))
			_, err = engine.Commit("seed rich text", time.Time{})
			require.NoError(t, err)

			object := engine.objects[text].OpID
			index := engine.state.sequenceIndex(object)
			random := rand.New(rand.NewPCG(31, uint64(len(expand))))
			insertions := []string{"", "x", "😀", "é", "👩‍💻"}
			for range 75 {
				current, textErr := engine.Text(text)
				require.NoError(t, textErr)
				runes := []rune(current)
				runeIndex := random.IntN(len(runes) + 1)
				position := utf16Width(string(runes[:runeIndex]))
				deleteCount := 0
				if runeIndex < len(runes) && random.IntN(2) == 1 {
					deleteCount = utf16Width(string(runes[runeIndex]))
				}
				require.NoError(t, engine.SpliceText(
					text,
					uint32(position),
					int32(deleteCount),
					insertions[random.IntN(len(insertions))],
				))

				assert.Same(t, index, engine.state.sequenceIndexes[object])
				assert.False(t, engine.state.sequenceRescue[object])
				assert.NotContains(t, engine.state.sequenceElementsCache, object)
				assert.NotContains(t, engine.state.sequenceOffsetCache, object)
			}

			wantText, err := engine.Text(text)
			require.NoError(t, err)
			wantSpans, err := engine.state.RichTextSpans(object)
			require.NoError(t, err)
			_, err = engine.Commit("random rich edits", time.Time{})
			require.NoError(t, err)
			data, err := engine.Save(true, false)
			require.NoError(t, err)
			reloaded, err := LoadEngine(data)
			require.NoError(t, err)
			reloadedText, err := reloaded.GetText(0, "body")
			require.NoError(t, err)
			gotText, err := reloaded.Text(reloadedText)
			require.NoError(t, err)
			gotSpans, err := reloaded.state.RichTextSpans(
				reloaded.objects[reloadedText].OpID,
			)
			require.NoError(t, err)

			assert.Equal(t, wantText, gotText)
			assert.Equal(t, wantSpans, gotSpans)
		})
	}
}

func TestSequenceIndex_MarkedConcurrentBranchesStayLocalized(t *testing.T) {
	t.Parallel()

	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte("base")))
	text, err := base.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(text, 0, 0, "ab😀cd"))
	encoded, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarTrue})
	require.NoError(t, err)
	require.NoError(t, base.MarkText(text, 1, 5, "bold", encoded, "both"))
	_, err = base.Commit("base", time.Time{})
	require.NoError(t, err)

	left, err := base.Fork([]byte("left"))
	require.NoError(t, err)
	right, err := base.Fork([]byte("right"))
	require.NoError(t, err)
	leftText, err := left.GetText(0, "body")
	require.NoError(t, err)
	rightText, err := right.GetText(0, "body")
	require.NoError(t, err)
	require.NoError(t, left.SpliceText(leftText, 3, 0, "L"))
	require.NoError(t, right.SpliceText(rightText, 3, 0, "R"))
	_, err = left.Commit("left", time.Time{})
	require.NoError(t, err)
	_, err = right.Commit("right", time.Time{})
	require.NoError(t, err)
	rightData, err := right.Save(true, false)
	require.NoError(t, err)

	object := left.objects[leftText].OpID
	index := left.state.sequenceIndex(object)
	_, err = left.Merge(rightData)
	require.NoError(t, err)

	assert.Same(t, index, left.state.sequenceIndexes[object])
	assert.False(t, left.state.sequenceRescue[object])
	assert.NotContains(t, left.state.sequenceElementsCache, object)
	assert.NotContains(t, left.state.sequenceOffsetCache, object)
	value, err := left.Text(leftText)
	require.NoError(t, err)
	assert.True(t, slices.Contains([]string{"ab😀LRcd", "ab😀RLcd"}, value))
}

func TestDirectTextOverlay_EmitsLocalizedWireRuns(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte("overlay")))
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, strings.Repeat("a", 600)))
	_, err = engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	for edit := range 40 {
		require.NoError(t, engine.SpliceText(text, uint32(edit*10), 1, "😀"))
	}
	object := engine.objects[text]
	hash := opset.ChangeHash{1}
	change := &opset.Change{
		Hash:         &hash,
		Actor:        engine.actor,
		Dependencies: engine.currentHeads(),
		Operations:   append([]opset.Operation(nil), engine.pending...),
	}
	batch, err := newColumnMutationBatch(
		engine.columns,
		engine.state,
		[]*opset.Change{change},
		false,
	)
	require.NoError(t, err)

	require.Len(t, batch.touchedObjects[object], 2)
	assert.LessOrEqual(t, len(batch.operationSplices), 80)
	assert.False(t, engine.state.sequenceRescue[object.OpID])
	assert.NotContains(t, engine.state.sequenceElementsCache, object.OpID)
	assert.NotContains(t, engine.state.sequenceOffsetCache, object.OpID)
}

func TestDirectTextOverlay_MiddleInsertionReloads(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "seed"))
	_, err = engine.Commit("seed", time.Time{})
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 2, 0, "x"))
	_, err = engine.Commit("middle", time.Time{})
	require.NoError(t, err)
	data, err := engine.Save(true, false)
	require.NoError(t, err)
	reloaded, err := LoadEngine(data)
	require.NoError(t, err)
	reloadedText, err := reloaded.GetText(0, "body")
	require.NoError(t, err)
	value, err := reloaded.Text(reloadedText)
	require.NoError(t, err)
	assert.Equal(t, "sexed", value)
}

func TestDirectTextOverlay_MovingMarkedTextMatchesCanonicalState(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "bold text"))
	encoded, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarTrue})
	require.NoError(t, err)
	require.NoError(t, engine.MarkText(text, 0, 4, "bold", encoded, "after"))
	_, err = engine.Commit("setup", time.Time{})
	require.NoError(t, err)
	require.NoError(t, engine.UpdateSpans(
		text,
		[]byte(`[{"type":"text","text":"text "},{"type":"text","text":"bold","marks":{"bold":{"type":"boolean","bool":true}}}]`),
		[]byte(`{"defaultExpand":"after"}`),
	))
	_, err = engine.Commit("move", time.Time{})
	require.NoError(t, err)

	requireColumnarStateEquivalent(t, engine)
	value, err := engine.Text(text)
	require.NoError(t, err)
	assert.Equal(t, "text bold", value)
}
