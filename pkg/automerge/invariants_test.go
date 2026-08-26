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

func TestDocument_AppliesDependentChangesInAnyOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base, err := automerge.New(ctx, actor(100))
	require.NoError(t, err)
	closeDocument(t, base)
	baseText, err := base.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, baseText.Splice(ctx, 0, 0, "A"))
	_, err = base.Commit(ctx, "base", commitTime)
	require.NoError(t, err)
	baseHeads, err := base.Heads(ctx)
	require.NoError(t, err)
	baseData, err := base.Save(ctx)
	require.NoError(t, err)

	source, err := automerge.Load(ctx, baseData, actor(101))
	require.NoError(t, err)
	closeDocument(t, source)
	sourceText, err := source.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, sourceText.Splice(ctx, 1, 0, "B"))
	parentHash, err := source.Commit(ctx, "parent", commitTime.Add(time.Second))
	require.NoError(t, err)
	require.NoError(t, sourceText.Splice(ctx, 2, 0, "C"))
	childHash, err := source.Commit(ctx, "child", commitTime.Add(2*time.Second))
	require.NoError(t, err)

	changes, err := source.ChangesSince(ctx, baseHeads)
	require.NoError(t, err)
	require.Len(t, changes, 2)
	assert.ElementsMatch(
		t,
		[]automerge.Hash{parentHash, childHash},
		[]automerge.Hash{changes[0].Hash, changes[1].Hash},
	)

	target, err := automerge.Load(ctx, baseData, actor(102))
	require.NoError(t, err)
	closeDocument(t, target)
	require.NoError(t, target.ApplyChanges(ctx, []automerge.Change{changes[1]}))
	missing, err := target.MissingDependencies(
		ctx,
		[]automerge.Hash{childHash},
	)
	require.NoError(t, err)
	assert.Equal(t, []automerge.Hash{parentHash}, missing)
	require.NoError(t, target.ApplyChanges(ctx, []automerge.Change{changes[0]}))
	missing, err = target.MissingDependencies(
		ctx,
		[]automerge.Hash{childHash},
	)
	require.NoError(t, err)
	assert.Empty(t, missing)
	require.NoError(
		t,
		target.ApplyChanges(
			ctx,
			[]automerge.Change{
				changes[0],
				changes[1],
			},
		),
	)

	targetText, err := target.Text(ctx, "body")
	require.NoError(t, err)
	value, err := targetText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ABC", value)

	sourceHeads, err := source.Heads(ctx)
	require.NoError(t, err)
	targetHeads, err := target.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, sourceHeads, targetHeads)
}

func TestDocument_InvalidChangesDoNotMutateState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(119))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Stable"))
	_, err = document.Commit(ctx, "initial", commitTime)
	require.NoError(t, err)
	headsBefore, err := document.Heads(ctx)
	require.NoError(t, err)

	err = document.ApplyChanges(
		ctx,
		[]automerge.Change{
			{Bytes: []byte("invalid")},
		},
	)
	require.Error(t, err)

	value, err := text.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Stable", value)

	headsAfter, err := document.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, headsBefore, headsAfter)

	// An unknown baseline head excludes nothing, matching Rust's get_changes,
	// which takes have_deps by value and never errors: the whole history is
	// returned rather than failing. This is what keeps collaboration persistence
	// working when a frontier references a change that is no longer retrievable.
	var unknown automerge.Hash

	unknown[0] = 1
	changes, err := document.ChangesSince(ctx, []automerge.Hash{unknown})
	require.NoError(t, err)
	require.Len(t, changes, 1)
}

func TestDocument_IncrementalSaveLoadParity(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source func(
			context.Context,
			automerge.ActorID,
		) (*automerge.Document, error)
		target func(
			context.Context,
			automerge.ActorID,
		) (*automerge.Document, error)
	}{
		"native to reference": {
			source: automerge.New,
			target: automerge.NewReference,
		},
		"reference to native": {
			source: automerge.NewReference,
			target: automerge.New,
		},
	}

	for name, test := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				source, err := test.source(ctx, actor(151))
				require.NoError(t, err)
				closeDocument(t, source)

				target, err := test.target(ctx, actor(152))
				require.NoError(t, err)
				closeDocument(t, target)

				text, err := source.CreateText(ctx, "body")
				require.NoError(t, err)
				require.NoError(t, text.Splice(ctx, 0, 0, "A"))
				_, err = source.Commit(ctx, "first", commitTime)
				require.NoError(t, err)
				first, err := source.SaveIncremental(ctx)
				require.NoError(t, err)
				require.NotEmpty(t, first)

				empty, err := source.SaveIncremental(ctx)
				require.NoError(t, err)
				assert.Empty(t, empty)

				applied, err := target.LoadIncremental(ctx, first)
				require.NoError(t, err)
				assert.Positive(t, applied)

				targetText, err := target.Text(ctx, "body")
				require.NoError(t, err)
				value, err := targetText.String(ctx)
				require.NoError(t, err)
				assert.Equal(t, "A", value)

				applied, err = target.LoadIncremental(ctx, first)
				require.NoError(t, err)
				assert.Zero(t, applied)

				require.NoError(t, text.Splice(ctx, 1, 0, "B"))
				_, err = source.Commit(ctx, "second", commitTime.Add(time.Second))
				require.NoError(t, err)
				second, err := source.SaveIncremental(ctx)
				require.NoError(t, err)
				require.NotEmpty(t, second)
				applied, err = target.LoadIncremental(ctx, second)
				require.NoError(t, err)
				assert.Positive(t, applied)

				value, err = targetText.String(ctx)
				require.NoError(t, err)
				assert.Equal(t, "AB", value)

				sourceHeads, err := source.Heads(ctx)
				require.NoError(t, err)
				targetHeads, err := target.Heads(ctx)
				require.NoError(t, err)
				assert.ElementsMatch(t, sourceHeads, targetHeads)

				_, err = source.Save(ctx)
				require.NoError(t, err)
				empty, err = source.SaveIncremental(ctx)
				require.NoError(t, err)
				assert.Empty(t, empty)
			},
		)
	}
}

func TestDocument_IncrementalLoadIgnoresCorruptTail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	source, err := automerge.NewReference(ctx, actor(169))
	require.NoError(t, err)
	closeDocument(t, source)
	require.NoError(t, source.PutString(ctx, "key", "value"))
	_, err = source.Commit(ctx, "value", commitTime)
	require.NoError(t, err)
	data, err := source.Save(ctx)
	require.NoError(t, err)

	data = append(data, 1, 2, 3, 4)

	factories := map[string]func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error){
		"native":    automerge.New,
		"reference": automerge.NewReference,
	}
	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(ctx, actor(170))
				require.NoError(t, err)
				closeDocument(t, document)
				applied, err := document.LoadIncremental(ctx, data)
				require.NoError(t, err)
				assert.Positive(t, applied)

				value, err := document.String(ctx, "key")
				require.NoError(t, err)
				assert.Equal(t, "value", value)
			},
		)
	}
}

func TestDocument_MergedChangesForwardToThirdPeer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := newBaseDocument(t)

	relay, err := automerge.Load(ctx, base, actor(103))
	require.NoError(t, err)
	closeDocument(t, relay)

	source, err := automerge.Load(ctx, base, actor(104))
	require.NoError(t, err)
	closeDocument(t, source)
	sourceText, err := source.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, sourceText.Splice(ctx, 5, 0, " forwarded"))
	_, err = source.Commit(ctx, "source edit", commitTime.Add(time.Second))
	require.NoError(t, err)

	_, err = relay.Merge(ctx, source)
	require.NoError(t, err)

	third, err := automerge.Load(ctx, base, actor(105))
	require.NoError(t, err)
	closeDocument(t, third)

	relaySync, err := relay.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, relaySync)

	thirdSync, err := third.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, thirdSync)

	synchronize(t, relaySync, thirdSync)

	thirdText, err := third.Text(ctx, "body")
	require.NoError(t, err)
	value, err := thirdText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello forwarded", value)

	relayHeads, err := relay.Heads(ctx)
	require.NoError(t, err)
	thirdHeads, err := third.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, relayHeads, thirdHeads)
}

func TestSyncState_DuplicateMessagesAreIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	source, err := automerge.New(ctx, actor(106))
	require.NoError(t, err)
	closeDocument(t, source)
	sourceText, err := source.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, sourceText.Splice(ctx, 0, 0, "Once"))
	_, err = source.Commit(ctx, "create body", commitTime)
	require.NoError(t, err)

	target, err := automerge.New(ctx, actor(107))
	require.NoError(t, err)
	closeDocument(t, target)

	sourceSync, err := source.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, sourceSync)

	targetSync, err := target.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, targetSync)

	message, ok, err := sourceSync.GenerateMessage(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, targetSync.ReceiveMessage(ctx, message))
	require.NoError(t, targetSync.ReceiveMessage(ctx, message))
	synchronize(t, sourceSync, targetSync)

	targetText, err := target.Text(ctx, "body")
	require.NoError(t, err)
	value, err := targetText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Once", value)

	sourceHeads, err := source.Heads(ctx)
	require.NoError(t, err)
	targetHeads, err := target.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, sourceHeads, targetHeads)
}

func TestSyncState_ResumesPersistedSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	left, err := automerge.New(ctx, actor(108))
	require.NoError(t, err)
	closeDocument(t, left)
	leftText, err := left.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, leftText.Splice(ctx, 0, 0, "A"))
	_, err = left.Commit(ctx, "initial", commitTime)
	require.NoError(t, err)

	right, err := automerge.New(ctx, actor(109))
	require.NoError(t, err)
	closeDocument(t, right)

	leftSync, err := left.NewSyncState(ctx)
	require.NoError(t, err)

	rightSync, err := right.NewSyncState(ctx)
	require.NoError(t, err)
	synchronize(t, leftSync, rightSync)

	leftState, err := leftSync.Save(ctx)
	require.NoError(t, err)
	rightState, err := rightSync.Save(ctx)
	require.NoError(t, err)
	require.NoError(t, leftSync.Close(ctx))
	require.NoError(t, rightSync.Close(ctx))

	resumedLeft, err := left.LoadSyncState(ctx, leftState)
	require.NoError(t, err)
	closeSyncState(t, resumedLeft)

	resumedRight, err := right.LoadSyncState(ctx, rightState)
	require.NoError(t, err)
	closeSyncState(t, resumedRight)

	require.NoError(t, leftText.Splice(ctx, 1, 0, "B"))
	_, err = left.Commit(ctx, "incremental", commitTime.Add(time.Second))
	require.NoError(t, err)
	synchronize(t, resumedLeft, resumedRight)

	rightText, err := right.Text(ctx, "body")
	require.NoError(t, err)
	value, err := rightText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AB", value)
}

func TestSyncState_ResendsInFlightMessageAfterRestore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	source, err := automerge.New(ctx, actor(120))
	require.NoError(t, err)
	closeDocument(t, source)
	sourceText, err := source.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, sourceText.Splice(ctx, 0, 0, "Recovered"))
	_, err = source.Commit(ctx, "initial", commitTime)
	require.NoError(t, err)

	target, err := automerge.New(ctx, actor(121))
	require.NoError(t, err)
	closeDocument(t, target)

	sourceSync, err := source.NewSyncState(ctx)
	require.NoError(t, err)
	_, ok, err := sourceSync.GenerateMessage(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	sourceState, err := sourceSync.Save(ctx)
	require.NoError(t, err)
	require.NoError(t, sourceSync.Close(ctx))

	targetSync, err := target.NewSyncState(ctx)
	require.NoError(t, err)
	targetState, err := targetSync.Save(ctx)
	require.NoError(t, err)
	require.NoError(t, targetSync.Close(ctx))

	resumedSource, err := source.LoadSyncState(ctx, sourceState)
	require.NoError(t, err)
	closeSyncState(t, resumedSource)

	resumedTarget, err := target.LoadSyncState(ctx, targetState)
	require.NoError(t, err)
	closeSyncState(t, resumedTarget)

	message, ok, err := resumedSource.GenerateMessage(ctx)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotEmpty(t, message)
	require.NoError(t, resumedTarget.ReceiveMessage(ctx, message))
	synchronize(t, resumedSource, resumedTarget)

	targetText, err := target.Text(ctx, "body")
	require.NoError(t, err)
	value, err := targetText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Recovered", value)
}

func TestDocument_ThreeWayMergeIsAssociative(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := newBaseDocument(t)
	documents := make([]*automerge.Document, 3)

	for i := range documents {
		document, err := automerge.Load(ctx, base, actor(byte(110+i)))
		require.NoError(t, err)
		closeDocument(t, document)
		text, err := document.Text(ctx, "body")
		require.NoError(t, err)
		require.NoError(t, text.Splice(ctx, 5, 0, string(rune('A'+i))))
		_, err = document.Commit(
			ctx,
			"concurrent edit",
			commitTime.Add(time.Duration(i+1)*time.Second),
		)
		require.NoError(t, err)

		documents[i] = document
	}

	first, err := automerge.Load(ctx, base, actor(113))
	require.NoError(t, err)
	closeDocument(t, first)
	_, err = first.Merge(ctx, documents[0])
	require.NoError(t, err)
	_, err = first.Merge(ctx, documents[1])
	require.NoError(t, err)
	_, err = first.Merge(ctx, documents[2])
	require.NoError(t, err)

	second, err := automerge.Load(ctx, base, actor(114))
	require.NoError(t, err)
	closeDocument(t, second)
	_, err = second.Merge(ctx, documents[2])
	require.NoError(t, err)
	_, err = second.Merge(ctx, documents[0])
	require.NoError(t, err)
	_, err = second.Merge(ctx, documents[1])
	require.NoError(t, err)

	firstText, err := first.Text(ctx, "body")
	require.NoError(t, err)
	firstValue, err := firstText.String(ctx)
	require.NoError(t, err)
	secondText, err := second.Text(ctx, "body")
	require.NoError(t, err)
	secondValue, err := secondText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, firstValue, secondValue)

	firstHeads, err := first.Heads(ctx)
	require.NoError(t, err)
	secondHeads, err := second.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, firstHeads, secondHeads)

	firstData, err := first.Save(ctx)
	require.NoError(t, err)
	reference, err := automerge.LoadReference(ctx, firstData, actor(115))
	require.NoError(t, err)
	closeDocument(t, reference)
	referenceText, err := reference.Text(ctx, "body")
	require.NoError(t, err)
	referenceValue, err := referenceText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, firstValue, referenceValue)
}

func TestSyncState_ThreePeerRelayConvergesWithReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first, err := automerge.New(ctx, actor(116))
	require.NoError(t, err)
	closeDocument(t, first)
	firstText, err := first.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, firstText.Splice(ctx, 0, 0, "A"))
	_, err = first.Commit(ctx, "initial", commitTime)
	require.NoError(t, err)

	second, err := automerge.NewReference(ctx, actor(117))
	require.NoError(t, err)
	closeDocument(t, second)

	firstSecond, err := first.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, firstSecond)

	secondFirst, err := second.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, secondFirst)
	synchronize(t, firstSecond, secondFirst)

	third, err := automerge.New(ctx, actor(118))
	require.NoError(t, err)
	closeDocument(t, third)

	secondThird, err := second.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, secondThird)

	thirdSecond, err := third.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, thirdSecond)
	synchronize(t, secondThird, thirdSecond)

	secondText, err := second.Text(ctx, "body")
	require.NoError(t, err)
	thirdText, err := third.Text(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, firstText.Splice(ctx, 1, 0, "1"))
	_, err = first.Commit(ctx, "first edit", commitTime.Add(time.Second))
	require.NoError(t, err)
	require.NoError(t, secondText.Splice(ctx, 1, 0, "2"))
	_, err = second.Commit(ctx, "second edit", commitTime.Add(2*time.Second))
	require.NoError(t, err)
	require.NoError(t, thirdText.Splice(ctx, 1, 0, "3"))
	_, err = third.Commit(ctx, "third edit", commitTime.Add(3*time.Second))
	require.NoError(t, err)

	synchronize(t, firstSecond, secondFirst)
	synchronize(t, secondThird, thirdSecond)

	thirdFirst, err := third.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, thirdFirst)

	firstThird, err := first.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, firstThird)
	synchronize(t, thirdFirst, firstThird)
	synchronize(t, firstSecond, secondFirst)
	synchronize(t, secondThird, thirdSecond)

	firstValue, err := firstText.String(ctx)
	require.NoError(t, err)
	secondValue, err := secondText.String(ctx)
	require.NoError(t, err)
	thirdValue, err := thirdText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, firstValue, secondValue)
	assert.Equal(t, firstValue, thirdValue)

	firstHeads, err := first.Heads(ctx)
	require.NoError(t, err)
	secondHeads, err := second.Heads(ctx)
	require.NoError(t, err)
	thirdHeads, err := third.Heads(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, firstHeads, secondHeads)
	assert.ElementsMatch(t, firstHeads, thirdHeads)
}
