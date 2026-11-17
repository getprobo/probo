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
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

// concurrentHistory builds a document two actors edited concurrently: text
// inserted and deleted, marks, a counter, and map keys written and removed.
// It returns the engine, which has never been saved, and the heads recorded
// after each merge.
func concurrentHistory(t *testing.T, seed uint64) (*Engine, [][][32]byte) {
	t.Helper()

	random := rand.New(rand.NewPCG(seed, 101))

	first, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, first.SetActor([]byte("\x20first-actor")))

	_, err = first.PutText(0, "body")
	require.NoError(t, err)

	counter, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarCounter, Int: 1})
	require.NoError(t, err)
	require.NoError(t, first.PutScalar(0, "count", counter))
	require.NoError(t, first.PutString(0, "title", "start"))

	_, err = first.Commit("seed", time.Unix(1, 0))
	require.NoError(t, err)

	// The second actor sorts before the first, so actor order and the order
	// the actors joined in disagree.
	second, err := first.Fork([]byte("\x10second-actor"))
	require.NoError(t, err)

	bold, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarTrue})
	require.NoError(t, err)

	clock := int64(1)

	edit := func(engine *Engine) {
		body, _, err := engine.GetObject(0, "body")
		require.NoError(t, err)

		text, err := engine.Text(body)
		require.NoError(t, err)

		length := len(text)

		switch choice := random.IntN(7); {
		case choice < 2 || length < 4:
			require.NoError(
				t,
				engine.SpliceText(
					body,
					uint32(random.IntN(length+1)),
					0,
					[]string{"a", "bc", "def "}[random.IntN(3)],
				),
			)
		case choice == 2:
			require.NoError(
				t,
				engine.SpliceText(body, uint32(random.IntN(length-2)), 2, ""),
			)
		case choice == 3:
			start := uint32(random.IntN(length - 2))
			require.NoError(t, engine.MarkText(body, start, start+2, "bold", bold, "after"))
		case choice == 4:
			require.NoError(t, engine.Increment(0, "count", int64(1+random.IntN(3))))
		case choice == 5:
			require.NoError(
				t,
				engine.PutString(0, "title", fmt.Sprintf("title %d", random.IntN(100))),
			)
		default:
			key := fmt.Sprintf("scratch-%d", random.IntN(3))

			require.NoError(t, engine.PutString(0, key, "temporary"))
			require.NoError(t, engine.DeleteMap(0, key))
		}

		clock++

		_, err = engine.Commit("edit", time.Unix(clock, 0))
		require.NoError(t, err)
	}

	var checkpoints [][][32]byte

	for range 5 {
		for range 1 + random.IntN(4) {
			edit(first)
		}

		for range 1 + random.IntN(4) {
			edit(second)
		}

		saved, err := second.Save(true, false)
		require.NoError(t, err)

		_, err = first.Merge(saved)
		require.NoError(t, err)

		heads, err := first.Heads()
		require.NoError(t, err)

		checkpoints = append(checkpoints, heads)
	}

	return first, checkpoints
}

// withoutSuccessors drops what an operation's row gains after it is written,
// leaving what its change recorded.
func withoutSuccessors(operations []opset.Operation) []opset.Operation {
	stripped := slices.Clone(operations)
	for i := range stripped {
		stripped[i].Successors = nil
	}

	return stripped
}

func TestLoadEngine_KeepsEachOperationOnce(t *testing.T) {
	t.Parallel()

	source, _ := concurrentHistory(t, 7)

	saved, err := source.Save(true, true)
	require.NoError(t, err)

	decoded, err := storage.Decode(saved)
	require.NoError(t, err)

	loaded, err := LoadEngine(saved)
	require.NoError(t, err)

	require.Nil(t, loaded.columns.decodedChanges)
	require.Len(t, loaded.columns.changes, len(decoded.Changes))
	require.NotEmpty(t, loaded.columns.nonCanonicalOperations)

	assertOperations := func(t *testing.T, sameSuccessors bool) {
		t.Helper()

		for i := range decoded.Changes {
			want := &decoded.Changes[i]

			change, ok := loaded.state.change(*want.Hash)
			require.True(t, ok, "change %d", i)

			// The change itself holds nothing: its operations are the
			// snapshot's rows.
			assert.True(t, change.OperationsDeferred, "change %d", i)
			assert.Empty(t, change.Operations, "change %d", i)

			got, err := loaded.state.withOperations(change)
			require.NoError(t, err)

			assert.False(t, got.OperationsDeferred, "change %d", i)
			assert.Equal(t, want.Raw, got.Raw, "change %d", i)

			switch {
			case len(want.Operations) == 0:
				assert.Empty(t, got.Operations, "change %d", i)
			case sameSuccessors:
				assert.Equal(t, want.Operations, got.Operations, "change %d", i)
			default:
				assert.Equal(
					t,
					withoutSuccessors(want.Operations),
					withoutSuccessors(got.Operations),
					"change %d",
					i,
				)
			}
		}
	}

	t.Run(
		"as loaded",
		func(t *testing.T) {
			assertOperations(t, true)
		},
	)

	// Later edits supersede loaded operations and move their rows. A loaded
	// change still names the same operations, however its rows have moved.
	// There are enough commits for the row index to be flattened on the way,
	// which replaces the lookup the load built.
	body, _, err := loaded.GetObject(0, "body")
	require.NoError(t, err)

	for i := range 6 {
		require.NoError(t, loaded.SpliceText(body, uint32(i), 1, "xy"))
		require.NoError(t, loaded.PutString(0, "title", fmt.Sprintf("later %d", i)))
		require.NoError(t, loaded.DeleteMap(0, "title"))

		_, err = loaded.Commit("later", time.Unix(int64(100+i), 0))
		require.NoError(t, err)
	}

	for i := range maxOperationRowIndexDepth {
		require.NoError(t, loaded.SpliceText(body, 0, 0, "z"))

		_, err = loaded.Commit("more", time.Unix(int64(200+i), 0))
		require.NoError(t, err)
	}

	t.Run(
		"after further edits",
		func(t *testing.T) {
			require.Nil(t, loaded.columns.operationRows.decoded)
			assertOperations(t, false)
		},
	)

	t.Run(
		"rebuilding the omitted operations keeps every delete",
		func(t *testing.T) {
			rebuilt := loaded.columns.clone()
			rebuilt.rebuildNonCanonicalOperations()

			assert.Equal(
				t,
				loaded.columns.nonCanonicalOperations,
				rebuilt.nonCanonicalOperations,
			)
			assert.Equal(
				t,
				loaded.columns.nonCanonicalObjects,
				rebuilt.nonCanonicalObjects,
			)
		},
	)
}

func TestLoadEngine_ReadsHistoryLikeAnUnsavedEngine(t *testing.T) {
	t.Parallel()

	for _, seed := range []uint64{11, 12, 13} {
		t.Run(
			fmt.Sprintf("seed %d", seed),
			func(t *testing.T) {
				t.Parallel()

				// source holds every change's operations in the change;
				// loaded holds them once, in the snapshot's rows.
				source, checkpoints := concurrentHistory(t, seed)

				saved, err := source.Save(true, true)
				require.NoError(t, err)

				loaded, err := LoadEngine(saved)
				require.NoError(t, err)

				sourceBody, _, err := source.GetObject(0, "body")
				require.NoError(t, err)

				loadedBody, _, err := loaded.GetObject(0, "body")
				require.NoError(t, err)

				for i, heads := range checkpoints {
					want, err := source.TextAt(sourceBody, heads)
					require.NoError(t, err)

					got, err := loaded.TextAt(loadedBody, heads)
					require.NoError(t, err)
					assert.Equal(t, want, got, "text at checkpoint %d", i)

					wantSpans, err := source.TextSpansAt(sourceBody, heads)
					require.NoError(t, err)

					gotSpans, err := loaded.TextSpansAt(loadedBody, heads)
					require.NoError(t, err)
					assert.JSONEq(t, string(wantSpans), string(gotSpans), "spans at checkpoint %d", i)

					if i == 0 {
						continue
					}

					wantDiff, err := source.Diff(checkpoints[i-1], heads)
					require.NoError(t, err)

					gotDiff, err := loaded.Diff(checkpoints[i-1], heads)
					require.NoError(t, err)
					assert.JSONEq(t, string(wantDiff), string(gotDiff), "diff to checkpoint %d", i)
				}

				// An isolated view replays the pinned changes into columns of
				// its own, and a write made there lands on top of them.
				pinned := checkpoints[1]

				for name, engine := range map[string]*Engine{"source": source, "loaded": loaded} {
					isolated, err := engine.Fork([]byte("\x30isolated-actor"))
					require.NoError(t, err, name)
					require.NoError(t, isolated.Isolate(pinned), name)

					body, _, err := isolated.GetObject(0, "body")
					require.NoError(t, err, name)

					text, err := isolated.Text(body)
					require.NoError(t, err, name)

					want, err := source.TextAt(sourceBody, pinned)
					require.NoError(t, err, name)
					assert.Equal(t, want, text, name)

					require.NoError(t, isolated.SpliceText(body, 0, 0, "isolated "), name)

					_, err = isolated.Commit("isolated", time.Unix(500, 0))
					require.NoError(t, err, name)

					text, err = isolated.Text(body)
					require.NoError(t, err, name)
					assert.Equal(t, "isolated "+want, text, name)
				}

				anonymizedSource, err := source.Anonymize()
				require.NoError(t, err)

				anonymizedLoaded, err := loaded.Anonymize()
				require.NoError(t, err)

				wantHistory, err := anonymizedSource.ChangeHistory()
				require.NoError(t, err)

				gotHistory, err := anonymizedLoaded.ChangeHistory()
				require.NoError(t, err)
				assert.Len(t, gotHistory, len(wantHistory))

				// Anonymizing replaces every name and value, so the two results
				// are compared by shape.
				wantKeys, err := anonymizedSource.Keys(0)
				require.NoError(t, err)

				gotKeys, err := anonymizedLoaded.Keys(0)
				require.NoError(t, err)
				assert.Len(t, gotKeys, len(wantKeys))
				assert.Equal(
					t,
					anonymizedSource.state.operationCount(),
					anonymizedLoaded.state.operationCount(),
				)

				// The same edits made by the same actor on both engines must
				// produce the same document, byte for byte.
				for name, engine := range map[string]*Engine{"source": source, "loaded": loaded} {
					require.NoError(t, engine.SetActor([]byte("\x25later-actor")), name)

					body, _, err := engine.GetObject(0, "body")
					require.NoError(t, err, name)

					for i := range 4 {
						require.NoError(t, engine.SpliceText(body, uint32(i), 1, "zz"), name)
						require.NoError(t, engine.Increment(0, "count", 2), name)
						require.NoError(t, engine.PutString(0, "scratch-0", "again"), name)
						require.NoError(t, engine.DeleteMap(0, "scratch-0"), name)

						_, err = engine.Commit("later", time.Unix(int64(600+i), 0))
						require.NoError(t, err, name)
					}
				}

				wantSaved, err := source.Save(true, true)
				require.NoError(t, err)

				gotSaved, err := loaded.Save(true, true)
				require.NoError(t, err)
				assert.Equal(t, wantSaved, gotSaved)
			},
		)
	}
}

func TestState_ApplyChangeRejectsChangeWithoutItsOperations(t *testing.T) {
	t.Parallel()

	source, _ := concurrentHistory(t, 17)

	saved, err := source.Save(true, true)
	require.NoError(t, err)

	loaded, err := LoadEngine(saved)
	require.NoError(t, err)

	changes, ok := loaded.state.allChanges()
	require.True(t, ok)
	require.NotEmpty(t, changes)
	require.True(t, changes[0].OperationsDeferred)

	// Applying such a change as it stands would record a change that wrote
	// nothing, so it is refused rather than silently emptied.
	err = NewState().ApplyChange(changes[0])
	require.ErrorContains(t, err, "change operations are not loaded")

	withOperations, err := loaded.state.withOperations(changes[0])
	require.NoError(t, err)
	require.NoError(t, NewState().ApplyChange(withOperations))
}

func TestLoadEngine_MergesLikeAnUnsavedEngine(t *testing.T) {
	t.Parallel()

	source, _ := concurrentHistory(t, 19)

	saved, err := source.Save(true, true)
	require.NoError(t, err)

	loaded, err := LoadEngine(saved)
	require.NoError(t, err)

	// A peer that shares the history writes on, and both engines merge what
	// it saved: once as a whole snapshot, once as the changes alone.
	peer, err := source.Fork([]byte("\x40peer-actor"))
	require.NoError(t, err)

	body, _, err := peer.GetObject(0, "body")
	require.NoError(t, err)
	require.NoError(t, peer.SpliceText(body, 0, 2, "merged "))
	require.NoError(t, peer.PutString(0, "title", "from the peer"))

	_, err = peer.Commit("peer", time.Unix(900, 0))
	require.NoError(t, err)

	snapshot, err := peer.Save(true, true)
	require.NoError(t, err)

	for name, engine := range map[string]*Engine{"source": source, "loaded": loaded} {
		_, err := engine.Merge(snapshot)
		require.NoError(t, err, name)
	}

	wantSaved, err := source.Save(true, true)
	require.NoError(t, err)

	gotSaved, err := loaded.Save(true, true)
	require.NoError(t, err)
	assert.Equal(t, wantSaved, gotSaved)

	// The other direction: an empty engine merges what the loaded one saves,
	// which makes the loaded engine hand over every change it holds.
	empty, err := NewEngine()
	require.NoError(t, err)

	_, err = empty.Merge(gotSaved)
	require.NoError(t, err)

	emptySaved, err := empty.Save(true, true)
	require.NoError(t, err)
	assert.Equal(t, gotSaved, emptySaved)

	// Handing the changes over one by one reads each loaded change's bytes,
	// which must be the ones an engine that never saved would send.
	wantBatch, err := source.PrepareMerge(nil)
	require.NoError(t, err)

	gotBatch, err := loaded.PrepareMerge(nil)
	require.NoError(t, err)
	assert.Equal(t, wantBatch.changes, gotBatch.changes)

	fresh, err := NewEngine()
	require.NoError(t, err)

	_, err = fresh.ApplyMerge(gotBatch)
	require.NoError(t, err)

	wantHeads, err := loaded.Heads()
	require.NoError(t, err)

	gotHeads, err := fresh.Heads()
	require.NoError(t, err)
	assert.Equal(t, wantHeads, gotHeads)

	loadedBody, _, err := loaded.GetObject(0, "body")
	require.NoError(t, err)

	freshBody, _, err := fresh.GetObject(0, "body")
	require.NoError(t, err)

	wantText, err := loaded.Text(loadedBody)
	require.NoError(t, err)

	gotText, err := fresh.Text(freshBody)
	require.NoError(t, err)
	assert.Equal(t, wantText, gotText)
}

func TestEncodeChange_RejectsChangeWithoutItsOperations(t *testing.T) {
	t.Parallel()

	source, _ := concurrentHistory(t, 23)

	saved, err := source.Save(true, true)
	require.NoError(t, err)

	loaded, err := LoadEngine(saved)
	require.NoError(t, err)

	changes, ok := loaded.state.allChanges()
	require.True(t, ok)
	require.NotEmpty(t, changes)

	for i, change := range changes {
		require.True(t, change.OperationsDeferred, "change %d", i)

		// Encoded as it stands, the change would come out empty under a hash
		// no peer knows.
		clone := *change

		_, err := storage.EncodeChange(&clone)
		require.ErrorContains(t, err, "change operations are not loaded", "change %d", i)

		// With its operations collected it encodes to the bytes it was
		// loaded from.
		withOperations, err := loaded.state.withOperations(change)
		require.NoError(t, err)

		clone = *withOperations

		raw, err := storage.EncodeChange(&clone)
		require.NoError(t, err)
		assert.Equal(t, change.Raw, raw, "change %d", i)
		assert.Equal(t, *change.Hash, *clone.Hash, "change %d", i)
	}
}

func TestColumnMutationBatch_LeavesSharedDeleteListsUntouched(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte("delete-lists")))

	body, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(body, 0, 0, "abcdefghijklmnopqrstuvwxyz"))

	_, err = engine.Commit("seed", time.Unix(1, 0))
	require.NoError(t, err)

	require.NoError(t, engine.SpliceText(body, 0, 3, ""))

	_, err = engine.Commit("first deletes", time.Unix(2, 0))
	require.NoError(t, err)

	// The fork reads the same committed columns as the engine it came from.
	fork, err := engine.Fork([]byte("delete-lists-fork"))
	require.NoError(t, err)

	object, err := engine.textObject(body)
	require.NoError(t, err)

	before := slices.Clone(fork.columns.nonCanonicalObjects[object])
	require.Len(t, before, 3)

	// One commit that deletes many elements of one object appends to that
	// object's list repeatedly.
	require.NoError(t, engine.SpliceText(body, 0, 12, ""))

	_, err = engine.Commit("more deletes", time.Unix(3, 0))
	require.NoError(t, err)

	assert.Equal(t, before, fork.columns.nonCanonicalObjects[object])

	// The fork then deletes on its own. Had the engine appended into the
	// list both started from, the fork would now overwrite the engine's
	// entries.
	forkBody, _, err := fork.GetObject(0, "body")
	require.NoError(t, err)
	require.NoError(t, fork.SpliceText(forkBody, 20, 3, ""))

	_, err = fork.Commit("fork deletes", time.Unix(4, 0))
	require.NoError(t, err)

	for name, columns := range map[string]*columnarState{
		"engine": engine.columns,
		"fork":   fork.columns,
	} {
		identifiers := columns.nonCanonicalObjects[object]
		assert.Equal(t, before, identifiers[:3], name)

		for _, identifier := range identifiers {
			_, ok := columns.nonCanonicalOperations[identifier]
			assert.True(t, ok, "%s delete %v", name, identifier)
		}
	}

	assert.Len(t, engine.columns.nonCanonicalObjects[object], 15)
	assert.Len(t, fork.columns.nonCanonicalObjects[object], 6)

	text, err := fork.Text(forkBody)
	require.NoError(t, err)
	assert.Equal(t, "defghijklmnopqrstuvw", text)

	text, err = engine.Text(body)
	require.NoError(t, err)
	assert.Equal(t, "pqrstuvwxyz", text)
}
