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

// This file reproduces the isolate/integrate transaction-view tests
// (can_isolate and can_transaction_at from rust/automerge/tests/test.rs and
// update_text_change_at from rust/automerge/tests/text.rs). Isolation pins reads
// and writes to a historical frontier while committed changes still accumulate
// in the full history and become visible after integrate. Each scenario runs on
// the native and reference engines and their observations must agree.

package automerge_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func textString(t *testing.T, text *automerge.Text) string {
	t.Helper()

	value, err := text.String()
	require.NoError(t, err)

	return value
}

func rootInt(t *testing.T, document *automerge.Document, key string) int64 {
	t.Helper()

	value, err := document.Root().Scalar(key)
	require.NoError(t, err)

	return value.Int
}

func putInt64(
	t *testing.T,
	document *automerge.Document,
	key string,
	value int64,
) {
	t.Helper()

	require.NoError(
		t,
		document.Root().PutScalar(

			key,
			automerge.Scalar{Type: automerge.ScalarTypeInt, Int: value},
		),
	)
}

// canIsolateObservations records the values observed at each checkpoint of the
// can_isolate scenario so the native and reference runs can be compared.
type canIsolateObservations struct {
	pinnedText     string
	pinnedSize     int64
	editedText     string
	editedSize     int64
	afterMergeSize int64
	afterMergeHas  bool
	rePinnedText   string
	rePinnedSize   int64
	integratedText string
	integratedHas  bool
	finalText      string
	finalSize      int64
}

func runCanIsolate(t *testing.T, engine rustParityEngine) canIsolateObservations {
	t.Helper()

	doc1, err := engine.open(actor(0x01))
	require.NoError(t, err)
	closeDocument(t, doc1)

	text, err := doc1.CreateText("text")
	require.NoError(t, err)

	putInt64(t, doc1, "size", 100)
	require.NoError(t, text.Splice(0, 0, "aaabbbccc"))
	_, err = doc1.Commit("seed", commitTime)
	require.NoError(t, err)

	heads1, err := doc1.Heads()
	require.NoError(t, err)

	putInt64(t, doc1, "size", 150)
	_, err = doc1.Commit("size150", commitTime)
	require.NoError(t, err)

	require.NoError(t, doc1.Isolate(heads1))

	doc2, err := doc1.Fork(actor(0x02))
	require.NoError(t, err)
	closeDocument(t, doc2)

	putInt64(t, doc2, "other", 999)
	text2, err := doc2.Text("text")
	require.NoError(t, err)
	require.NoError(t, text2.Splice(9, 0, "111"))
	_, err = doc2.Commit("doc2", commitTime)
	require.NoError(t, err)

	observations := canIsolateObservations{
		pinnedText: textString(t, text),
		pinnedSize: rootInt(t, doc1, "size"),
	}

	require.NoError(t, text.Splice(3, 3, "QQQ"))
	putInt64(t, doc1, "size", 200)

	observations.editedText = textString(t, text)
	observations.editedSize = rootInt(t, doc1, "size")

	_, err = doc1.Commit("qqq", commitTime)
	require.NoError(t, err)

	_, err = doc1.Merge(doc2)
	require.NoError(t, err)

	observations.afterMergeSize = rootInt(t, doc1, "size")
	observations.afterMergeHas = rootHasKey(t, doc1, "other")

	require.NoError(t, doc1.Isolate(heads1))

	observations.rePinnedText = textString(t, text)
	observations.rePinnedSize = rootInt(t, doc1, "size")

	require.NoError(t, text.Splice(3, 3, "ZZZ"))
	putInt64(t, doc1, "size", 300)
	_, err = doc1.Commit("zzz", commitTime)
	require.NoError(t, err)

	require.NoError(t, doc1.Integrate())

	observations.integratedText = textString(t, text)
	observations.integratedHas = rootHasKey(t, doc1, "other")

	require.NoError(t, doc1.Isolate(heads1))
	require.NoError(t, text.Splice(3, 3, "TTT"))
	putInt64(t, doc1, "size", 400)
	_, err = doc1.Commit("ttt", commitTime)
	require.NoError(t, err)

	require.NoError(t, doc1.Integrate())

	observations.finalText = textString(t, text)
	observations.finalSize = rootInt(t, doc1, "size")

	return observations
}

// TestRustText_IncorrectPatchesProducedWhenIsolatingAndIntegrating reproduces
// incorrect_patches_produced_when_isolating_and_integrating: a diff across an
// isolate/integrate cycle with a conflicting object put must reset to the isolate
// frontier and rebuild, producing deletes for the prior keys, conflicting puts,
// and a splice only for each winning object.
func TestRustText_IncorrectPatchesProducedWhenIsolatingAndIntegrating(t *testing.T) {
	t.Parallel()

	run := func(engine rustParityEngine) []automerge.Patch {
		doc, err := engine.open(actor(0xaa))
		require.NoError(t, err)
		closeDocument(t, doc)

		beginning, err := doc.Heads()
		require.NoError(t, err)

		name, err := doc.CreateText("name")
		require.NoError(t, err)

		newName := strings.Repeat("a", 100)
		require.NoError(t, name.Splice(0, 0, newName))

		require.NoError(t, doc.Isolate(beginning))
		color, err := doc.CreateText("color")
		require.NoError(t, err)
		require.NoError(t, color.Splice(0, 0, "red"))
		require.NoError(t, doc.Integrate())

		_, err = doc.DiffIncremental()
		require.NoError(t, err)

		require.NoError(t, doc.Isolate(beginning))
		color2, err := doc.CreateText("color")
		require.NoError(t, err)
		require.NoError(t, color2.Splice(0, 0, "unset"))
		require.NoError(t, doc.Integrate())

		patches, err := doc.DiffIncremental()
		require.NoError(t, err)

		return patches
	}

	native := run(rustParityEngines()[0])
	reference := run(rustParityEngines()[1])

	require.Equal(t, reference, native)

	require.Len(t, reference, 6)
	assert.Equal(t, automerge.PatchDeleteMap, reference[0].Action)
	assert.Equal(t, "color", reference[0].Key)
	assert.Equal(t, automerge.PatchDeleteMap, reference[1].Action)
	assert.Equal(t, "name", reference[1].Key)
	assert.Equal(t, automerge.PatchPutMap, reference[2].Action)
	assert.Equal(t, "color", reference[2].Key)
	assert.True(t, reference[2].Conflict)
	assert.Equal(t, automerge.PatchPutMap, reference[3].Action)
	assert.Equal(t, "name", reference[3].Key)
	assert.False(t, reference[3].Conflict)
	assert.Equal(t, automerge.PatchSpliceText, reference[4].Action)
	assert.Equal(t, strings.Repeat("a", 100), reference[4].Text)
	assert.Equal(t, automerge.PatchSpliceText, reference[5].Action)
	assert.Equal(t, "unset", reference[5].Text)
}

// TestRustText_UpdateTextChangeAt reproduces update_text_change_at: an isolated
// update_text branches from the initial heads and integrates alongside the
// concurrent update.
func TestRustText_UpdateTextChangeAt(t *testing.T) {
	t.Parallel()

	run := func(engine rustParityEngine) string {
		doc, err := engine.open(actor(0x01))
		require.NoError(t, err)
		closeDocument(t, doc)

		text, err := doc.CreateText("text")
		require.NoError(t, err)

		require.NoError(t, text.Update("a\n"))
		_, err = doc.Commit("a", commitTime)
		require.NoError(t, err)

		heads, err := doc.Heads()
		require.NoError(t, err)

		require.NoError(t, text.Update("a\nb\n"))
		_, err = doc.Commit("b", commitTime)
		require.NoError(t, err)

		require.NoError(t, doc.Isolate(heads))
		require.NoError(t, text.Update("a\nc\n"))
		_, err = doc.Commit("c", commitTime)
		require.NoError(t, err)
		require.NoError(t, doc.Integrate())

		return textString(t, text)
	}

	native := run(rustParityEngines()[0])
	reference := run(rustParityEngines()[1])

	require.Equal(t, reference, native)
	require.Equal(t, "a\nc\nb\n", reference)
}

// TestRustTest_CanTransactionAt reproduces can_transaction_at using the
// isolate/integrate equivalent of transaction_at: writes based at a historical
// frontier merge with the concurrent writes made since.
func TestRustTest_CanTransactionAt(t *testing.T) {
	t.Parallel()

	type observation struct {
		firstText  string
		firstSize  int64
		secondText string
		secondSize int64
	}

	run := func(engine rustParityEngine) observation {
		doc, err := engine.open(actor(0x01))
		require.NoError(t, err)
		closeDocument(t, doc)

		text, err := doc.CreateText("text")
		require.NoError(t, err)

		putInt64(t, doc, "size", 100)
		require.NoError(t, text.Splice(0, 0, "aaabbbccc"))
		_, err = doc.Commit("seed", commitTime)
		require.NoError(t, err)

		heads1, err := doc.Heads()
		require.NoError(t, err)

		require.NoError(t, text.Splice(3, 3, "QQQ"))
		putInt64(t, doc, "size", 200)
		_, err = doc.Commit("qqq", commitTime)
		require.NoError(t, err)

		require.NoError(t, doc.Isolate(heads1))
		require.NoError(t, text.Splice(3, 3, "ZZZ"))
		putInt64(t, doc, "size", 300)
		_, err = doc.Commit("zzz", commitTime)
		require.NoError(t, err)
		require.NoError(t, doc.Integrate())

		result := observation{
			firstText: textString(t, text),
			firstSize: rootInt(t, doc, "size"),
		}

		require.NoError(t, doc.Isolate(heads1))
		require.NoError(t, text.Splice(3, 3, "TTT"))
		putInt64(t, doc, "size", 400)
		_, err = doc.Commit("ttt", commitTime)
		require.NoError(t, err)
		require.NoError(t, doc.Integrate())

		result.secondText = textString(t, text)
		result.secondSize = rootInt(t, doc, "size")

		return result
	}

	native := run(rustParityEngines()[0])
	reference := run(rustParityEngines()[1])

	require.Equal(t, reference, native)
	require.Equal(t, "aaaZZZQQQccc", reference.firstText)
	require.Equal(t, int64(300), reference.firstSize)
	require.Equal(t, "aaaTTTZZZQQQccc", reference.secondText)
	require.Equal(t, int64(400), reference.secondSize)
}

func TestRustTest_CanIsolate(t *testing.T) {
	t.Parallel()

	native := runCanIsolate(t, rustParityEngines()[0])
	reference := runCanIsolate(t, rustParityEngines()[1])

	require.Equal(t, reference, native)

	// Anchor the reference observations to the upstream expectations so the
	// differential is also an absolute correctness check.
	require.Equal(t, "aaabbbccc", reference.pinnedText)
	require.Equal(t, int64(100), reference.pinnedSize)
	require.Equal(t, "aaaQQQccc", reference.editedText)
	require.Equal(t, int64(200), reference.editedSize)
	require.Equal(t, int64(200), reference.afterMergeSize)
	require.False(t, reference.afterMergeHas)
	require.Equal(t, "aaabbbccc", reference.rePinnedText)
	require.Equal(t, "aaaZZZQQQccc111", reference.integratedText)
	require.True(t, reference.integratedHas)
	require.Equal(t, "aaaTTTZZZQQQccc111", reference.finalText)
	require.Equal(t, int64(400), reference.finalSize)
}
