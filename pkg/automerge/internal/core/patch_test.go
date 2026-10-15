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
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestOrderedObjectsInState_IncludesVisibleSequenceObjects(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte{1}))

	items, err := engine.PutObject(0, "items", "list")
	require.NoError(t, err)
	_, err = engine.InsertObject(items, 0, "map")
	require.NoError(t, err)
	sibling, err := engine.InsertObject(items, 1, "map")
	require.NoError(t, err)
	replacement, err := engine.PutObjectAt(items, 0, "map")
	require.NoError(t, err)

	other, err := engine.PutObject(0, "other", "list")
	require.NoError(t, err)
	_, err = engine.InsertObject(other, 0, "text")
	require.NoError(t, err)
	updated, err := engine.PutObjectAt(other, 0, "map")
	require.NoError(t, err)

	assert.Equal(
		t,
		[]opset.ObjectID{
			opset.RootObject(),
			engineObject(t, engine, items),
			engineObject(t, engine, sibling),
			engineObject(t, engine, replacement),
			engineObject(t, engine, other),
			engineObject(t, engine, updated),
		},
		orderedObjectsInState(engine.state),
	)
}

func TestOrderedObjectsInState_OmitsLosingSequenceObject(t *testing.T) {
	t.Parallel()

	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte{1}))

	items, err := base.PutObject(0, "items", "list")
	require.NoError(t, err)
	_, err = base.InsertObject(items, 0, "map")
	require.NoError(t, err)
	sibling, err := base.InsertObject(items, 1, "map")
	require.NoError(t, err)
	siblingID := engineObject(t, base, sibling)
	_, err = base.Commit("base", time.Time{})
	require.NoError(t, err)

	left, err := base.Fork([]byte{2})
	require.NoError(t, err)
	right, err := base.Fork([]byte{3})
	require.NoError(t, err)

	leftItems, _, err := left.GetObject(0, "items")
	require.NoError(t, err)
	leftChild, err := left.PutObjectAt(leftItems, 0, "map")
	require.NoError(t, err)
	_, err = left.Commit("left", time.Time{})
	require.NoError(t, err)

	rightItems, _, err := right.GetObject(0, "items")
	require.NoError(t, err)
	rightChild, err := right.PutObjectAt(rightItems, 0, "map")
	require.NoError(t, err)
	_, err = right.Commit("right", time.Time{})
	require.NoError(t, err)

	leftData, err := left.Save(true, false)
	require.NoError(t, err)
	_, err = right.Merge(leftData)
	require.NoError(t, err)

	leftID := engineObject(t, left, leftChild)
	rightID := engineObject(t, right, rightChild)
	require.NotEqual(t, leftID, rightID)
	assert.False(t, right.state.isSuperseded(leftID.OpID))
	assert.False(t, right.state.isSuperseded(rightID.OpID))

	winnerHandle, _, err := right.GetObjectAt(rightItems, 0)
	require.NoError(t, err)
	winner := engineObject(t, right, winnerHandle)
	loser := leftID
	if winner == leftID {
		loser = rightID
	}

	objects := orderedObjectsInState(right.state)
	assert.Equal(
		t,
		[]opset.ObjectID{
			opset.RootObject(),
			engineObject(t, right, rightItems),
			siblingID,
			winner,
		},
		objects,
	)
	assert.NotContains(t, objects, loser)
}

func TestOrderedObjectsInState_OmitsLosingMapObject(t *testing.T) {
	t.Parallel()

	base, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, base.SetActor([]byte{1}))
	_, err = base.PutObject(0, "item", "map")
	require.NoError(t, err)
	_, err = base.Commit("base", time.Time{})
	require.NoError(t, err)

	left, err := base.Fork([]byte{2})
	require.NoError(t, err)
	right, err := base.Fork([]byte{3})
	require.NoError(t, err)

	leftChild, err := left.PutObject(0, "item", "map")
	require.NoError(t, err)
	_, err = left.Commit("left", time.Time{})
	require.NoError(t, err)

	rightChild, err := right.PutObject(0, "item", "map")
	require.NoError(t, err)
	_, err = right.Commit("right", time.Time{})
	require.NoError(t, err)

	leftData, err := left.Save(true, false)
	require.NoError(t, err)
	_, err = right.Merge(leftData)
	require.NoError(t, err)

	leftID := engineObject(t, left, leftChild)
	rightID := engineObject(t, right, rightChild)
	require.NotEqual(t, leftID, rightID)
	assert.False(t, right.state.isSuperseded(leftID.OpID))
	assert.False(t, right.state.isSuperseded(rightID.OpID))

	winnerHandle, _, err := right.GetObject(0, "item")
	require.NoError(t, err)
	winner := engineObject(t, right, winnerHandle)

	assert.Equal(
		t,
		[]opset.ObjectID{opset.RootObject(), winner},
		orderedObjectsInState(right.state),
	)
}

func engineObject(t *testing.T, engine *Engine, handle uint32) opset.ObjectID {
	t.Helper()

	object, err := engine.object(handle)
	require.NoError(t, err)

	return object
}
