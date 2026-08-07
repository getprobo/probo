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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

type documentFactory func(context.Context, automerge.ActorID) (*automerge.Document, error)

func exerciseMapAndText(
	t *testing.T,
	factory documentFactory,
) (*automerge.Document, automerge.Hash) {
	t.Helper()

	ctx := context.Background()
	document, err := factory(ctx, actor(21))
	require.NoError(t, err)
	closeDocument(t, document)

	require.NoError(t, document.PutString(ctx, "title", "Policy"))
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "A😀B"))
	require.NoError(t, text.Splice(ctx, 1, 2, "é"))
	hash, err := document.Commit(ctx, "Create policy", commitTime)
	require.NoError(t, err)

	return document, hash
}

func readBody(t *testing.T, document *automerge.Document) string {
	t.Helper()

	text, err := document.Text(context.Background(), "body")
	require.NoError(t, err)
	value, err := text.String(context.Background())
	require.NoError(t, err)
	return value
}

func TestPureGo_DifferentialMapAndTextChange(t *testing.T) {
	t.Parallel()

	referenceDocument, referenceHash := exerciseMapAndText(t, automerge.New)
	nativeDocument, nativeHash := exerciseMapAndText(t, automerge.NewPureGo)

	assert.Equal(t, referenceHash, nativeHash)

	referenceHeads, err := referenceDocument.Heads(context.Background())
	require.NoError(t, err)
	nativeHeads, err := nativeDocument.Heads(context.Background())
	require.NoError(t, err)
	assert.Equal(t, referenceHeads, nativeHeads)
	assert.Equal(t, readBody(t, referenceDocument), readBody(t, nativeDocument))

	nativeData, err := nativeDocument.Save(context.Background())
	require.NoError(t, err)
	loadedByReference, err := automerge.Load(context.Background(), nativeData, actor(22))
	require.NoError(t, err)
	closeDocument(t, loadedByReference)
	assert.Equal(t, "AéB", readBody(t, loadedByReference))
}

func TestPureGo_LoadsAndExtendsReferenceHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	referenceDocument, _ := exerciseMapAndText(t, automerge.New)
	referenceData, err := referenceDocument.Save(ctx)
	require.NoError(t, err)

	nativeDocument, err := automerge.LoadPureGo(ctx, referenceData, actor(23))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	text, err := nativeDocument.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 1, 0, " native"))
	nativeHash, err := nativeDocument.Commit(ctx, "Edit in pure Go", commitTime.Add(time.Second))
	require.NoError(t, err)

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	loadedByReference, err := automerge.Load(ctx, nativeData, actor(24))
	require.NoError(t, err)
	closeDocument(t, loadedByReference)
	assert.Equal(t, "A nativeéB", readBody(t, loadedByReference))
	heads, err := loadedByReference.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, []automerge.Hash{nativeHash}, heads)
}

func TestPureGo_ConcurrentChangesConvergeWithOracle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := newBaseDocument(t)

	left, err := automerge.LoadPureGo(ctx, base, actor(25))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, left.PutString(ctx, "winner", "left"))
	require.NoError(t, leftText.Splice(ctx, 5, 0, " left"))
	_, err = left.Commit(ctx, "Edit left", commitTime.Add(time.Second))
	require.NoError(t, err)

	right, err := automerge.LoadPureGo(ctx, base, actor(26))
	require.NoError(t, err)
	closeDocument(t, right)
	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, right.PutString(ctx, "winner", "right"))
	require.NoError(t, rightText.Splice(ctx, 5, 0, " right"))
	_, err = right.Commit(ctx, "Edit right", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	leftData, err := left.Save(ctx)
	require.NoError(t, err)
	rightData, err := right.Save(ctx)
	require.NoError(t, err)
	leftFirst, err := automerge.LoadPureGo(ctx, leftData, actor(27))
	require.NoError(t, err)
	closeDocument(t, leftFirst)
	rightFirst, err := automerge.LoadPureGo(ctx, rightData, actor(28))
	require.NoError(t, err)
	closeDocument(t, rightFirst)

	_, err = leftFirst.Merge(ctx, right)
	require.NoError(t, err)
	_, err = rightFirst.Merge(ctx, left)
	require.NoError(t, err)
	assert.Equal(t, readBody(t, leftFirst), readBody(t, rightFirst))

	leftHeads, err := leftFirst.Heads(ctx)
	require.NoError(t, err)
	rightHeads, err := rightFirst.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, leftHeads, rightHeads)

	mergedData, err := leftFirst.Save(ctx)
	require.NoError(t, err)
	oracle, err := automerge.Load(ctx, mergedData, actor(29))
	require.NoError(t, err)
	closeDocument(t, oracle)
	assert.Equal(t, readBody(t, leftFirst), readBody(t, oracle))
	oracleHeads, err := oracle.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, leftHeads, oracleHeads)
}

func TestPureGo_CursorUsesUTF16Positions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.NewPureGo(ctx, actor(30))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "A😀B"))

	cursor, err := text.Cursor(ctx, 3)
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "😀"))
	position, err := text.CursorPosition(ctx, cursor)
	require.NoError(t, err)
	assert.Equal(t, uint32(5), position)
}
