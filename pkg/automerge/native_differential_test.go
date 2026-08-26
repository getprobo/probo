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

package automerge_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func TestPureGoDocument_RandomTextParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 20
		steps     = 30
	)

	characters := []rune("abcXYZ😀é")

	for history := range histories {
		random := rand.New(rand.NewSource(int64(history + 1)))
		nativeDocument, err := automerge.New(actor(byte(80 + history)))
		require.NoError(t, err)
		closeDocument(t, nativeDocument)
		nativeText, err := nativeDocument.CreateText("body")
		require.NoError(t, err)

		referenceDocument, err := automerge.NewReference(actor(byte(80 + history)))
		require.NoError(t, err)
		closeDocument(t, referenceDocument)
		referenceText, err := referenceDocument.CreateText("body")
		require.NoError(t, err)

		var model []rune
		for step := range steps {
			offsets := utf16Offsets(model)

			var (
				index       uint32
				deleteCount int32
				insert      string
			)

			if len(model) > 0 && random.Intn(3) == 0 {
				position := random.Intn(len(model))
				index = offsets[position]
				deleteCount = int32(offsets[position+1] - offsets[position])
				model = append(model[:position], model[position+1:]...)
			} else {
				position := random.Intn(len(model) + 1)
				index = offsets[position]
				character := characters[random.Intn(len(characters))]
				insert = string(character)

				model = append(model, 0)
				copy(model[position+1:], model[position:])
				model[position] = character
			}

			require.NoError(t, nativeText.Splice(index, deleteCount, insert))
			require.NoError(t, referenceText.Splice(index, deleteCount, insert))

			message := fmt.Sprintf("history %d step %d", history, step)
			_, err = nativeDocument.Commit(message, commitTime)
			require.NoError(t, err)
			_, err = referenceDocument.Commit(message, commitTime)
			require.NoError(t, err)

			nativeValue, err := nativeText.String()
			require.NoError(t, err)
			referenceValue, err := referenceText.String()
			require.NoError(t, err)
			assert.Equal(t, string(model), nativeValue)
			assert.Equal(t, referenceValue, nativeValue)
		}

		data, err := nativeDocument.Save()
		require.NoError(t, err)
		loaded, err := automerge.LoadReference(data, actor(byte(120+history)))
		require.NoError(t, err)
		closeDocument(t, loaded)
		loadedText, err := loaded.Text("body")
		require.NoError(t, err)
		loadedValue, err := loadedText.String()
		require.NoError(t, err)
		assert.Equal(t, string(model), loadedValue)
	}
}

func TestPureGoDocument_RandomConcurrentSyncParity(t *testing.T) {
	t.Parallel()

	const (
		histories = 10
		rounds    = 20
	)

	characters := []rune("abcXYZ😀é")

	for history := range histories {
		random := rand.New(rand.NewSource(int64(10_000 + history)))
		nativeDocument, err := automerge.New(

			actor(byte(140 + history)),
		)
		require.NoError(t, err)
		closeDocument(t, nativeDocument)
		nativeText, err := nativeDocument.CreateText("body")
		require.NoError(t, err)
		_, err = nativeDocument.Commit("create body", commitTime)
		require.NoError(t, err)

		referenceDocument, err := automerge.NewReference(

			actor(byte(160 + history)),
		)
		require.NoError(t, err)
		closeDocument(t, referenceDocument)

		nativeSync, err := nativeDocument.NewSyncState()
		require.NoError(t, err)
		closeSyncState(t, nativeSync)

		referenceSync, err := referenceDocument.NewSyncState()
		require.NoError(t, err)
		closeSyncState(t, referenceSync)
		synchronize(t, nativeSync, referenceSync)

		referenceText, err := referenceDocument.Text("body")
		require.NoError(t, err)

		for round := range rounds {
			randomTextMutation(
				t,
				random,
				nativeText,
				characters,
			)
			_, err = nativeDocument.Commit(

				fmt.Sprintf("native history %d round %d", history, round),
				commitTime.Add(time.Duration(round+1)*time.Second),
			)
			require.NoError(t, err)

			randomTextMutation(
				t,
				random,
				referenceText,
				characters,
			)
			_, err = referenceDocument.Commit(

				fmt.Sprintf("reference history %d round %d", history, round),
				commitTime.Add(time.Duration(round+1)*time.Second),
			)
			require.NoError(t, err)

			synchronize(t, nativeSync, referenceSync)

			nativeValue, err := nativeText.String()
			require.NoError(t, err)
			referenceValue, err := referenceText.String()
			require.NoError(t, err)
			assert.Equal(
				t,
				referenceValue,
				nativeValue,
				"history %d round %d",
				history,
				round,
			)

			nativeHeads, err := nativeDocument.Heads()
			require.NoError(t, err)
			referenceHeads, err := referenceDocument.Heads()
			require.NoError(t, err)
			assert.ElementsMatch(t, referenceHeads, nativeHeads)
		}
	}
}

func randomTextMutation(
	t *testing.T,
	random *rand.Rand,
	text *automerge.Text,
	characters []rune,
) {
	t.Helper()

	value, err := text.String()
	require.NoError(t, err)

	runes := []rune(value)
	offsets := utf16Offsets(runes)

	if len(runes) > 0 && random.Intn(3) == 0 {
		position := random.Intn(len(runes))
		require.NoError(
			t,
			text.Splice(

				offsets[position],
				int32(offsets[position+1]-offsets[position]),
				"",
			),
		)

		return
	}

	position := random.Intn(len(runes) + 1)
	require.NoError(
		t,
		text.Splice(

			offsets[position],
			0,
			string(characters[random.Intn(len(characters))]),
		),
	)
}

func utf16Offsets(value []rune) []uint32 {
	offsets := make([]uint32, len(value)+1)
	for i, character := range value {
		offsets[i+1] = offsets[i] + uint32(len(utf16.Encode([]rune{character})))
	}

	return offsets
}
