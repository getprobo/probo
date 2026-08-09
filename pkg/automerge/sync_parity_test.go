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

// The tests in this file reproduce upstream Rust synchronization tests from
// automerge 0.10 (rust/automerge/src/sync.rs) against both the native Go engine
// and the Rust/WASM reference engine. Every scenario drives the same read-only
// and read-write sync protocol through the public Go API and asserts the same
// observable outcomes (which changes each peer receives, peer read-only
// discovery, and quiescence) that the upstream tests assert.

package automerge_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func readWriteSyncState(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
) *automerge.SyncState {
	t.Helper()

	state, err := document.NewSyncState(ctx)
	require.NoError(t, err)
	closeSyncState(t, state)

	return state
}

func readOnlySyncState(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
) *automerge.SyncState {
	t.Helper()

	state := readWriteSyncState(t, ctx, document)
	require.NoError(t, state.SetReadOnly(ctx, true))

	return state
}

// syncQuiescent exchanges messages in both directions until neither peer has
// anything to send, failing if the session does not converge. It mirrors the
// upstream sync() test helper.
func syncQuiescent(
	t *testing.T,
	ctx context.Context,
	left *automerge.SyncState,
	right *automerge.SyncState,
) {
	t.Helper()

	const maxRounds = 50

	for range maxRounds {
		leftMessage, leftOK, err := left.GenerateMessage(ctx)
		require.NoError(t, err)
		rightMessage, rightOK, err := right.GenerateMessage(ctx)
		require.NoError(t, err)

		if !leftOK && !rightOK {
			return
		}

		if leftOK {
			require.NoError(t, right.ReceiveMessage(ctx, leftMessage))
		}

		if rightOK {
			require.NoError(t, left.ReceiveMessage(ctx, rightMessage))
		}
	}

	t.Fatalf("sync did not converge within %d rounds", maxRounds)
}

func rootHasKey(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
	key string,
) bool {
	t.Helper()

	_, err := document.Root().Scalar(ctx, key)

	return err == nil
}

func putRoot(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
	key string,
	value string,
	message string,
	when time.Time,
) {
	t.Helper()

	require.NoError(t, document.Root().PutScalar(
		ctx,
		key,
		automerge.Scalar{Type: automerge.ScalarTypeString, String: value},
	))
	_, err := document.Commit(ctx, message, when)
	require.NoError(t, err)
}

func putInt(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
	key string,
	value int64,
	message string,
	when time.Time,
) {
	t.Helper()

	require.NoError(t, document.Root().PutScalar(
		ctx,
		key,
		automerge.Scalar{Type: automerge.ScalarTypeInt, Int: value},
	))
	_, err := document.Commit(ctx, message, when)
	require.NoError(t, err)
}

// TestRustSync_BranchingAndMerging reproduces
// should_handle_lots_of_branching_and_merging: two peers exchange many
// concurrent changes, a third peer's change is merged into one of them, and a
// final synchronization must converge both peers to identical heads.
func TestRustSync_BranchingAndMerging(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(0x01))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(0x89))
			require.NoError(t, err)
			closeDocument(t, doc2)

			doc3, err := engine.open(ctx, actor(0xfe))
			require.NoError(t, err)
			closeDocument(t, doc3)

			putInt(t, ctx, doc1, "x", 0, "x0", commitTime)
			_, err = doc2.Merge(ctx, doc1)
			require.NoError(t, err)
			_, err = doc3.Merge(ctx, doc1)
			require.NoError(t, err)

			putInt(t, ctx, doc3, "x", 1, "x1", commitTime.Add(time.Second))

			for i := int64(1); i < 20; i++ {
				when := commitTime.Add(time.Duration(i+1) * time.Second)
				putInt(t, ctx, doc1, "n1", i, "n1", when)
				putInt(t, ctx, doc2, "n2", i, "n2", when)
				_, err = doc1.Merge(ctx, doc2)
				require.NoError(t, err)
				_, err = doc2.Merge(ctx, doc1)
				require.NoError(t, err)
			}

			s1 := readWriteSyncState(t, ctx, doc1)
			s2 := readWriteSyncState(t, ctx, doc2)
			syncQuiescent(t, ctx, s1, s2)

			// doc3's change is concurrent to the last sync heads, forcing the
			// slower reconciliation path on the next synchronization.
			_, err = doc2.Merge(ctx, doc3)
			require.NoError(t, err)

			putInt(t, ctx, doc1, "n1", 100, "n1 final", commitTime.Add(time.Hour))
			putInt(t, ctx, doc2, "n1", 100, "n1 final", commitTime.Add(time.Hour))

			s1 = readWriteSyncState(t, ctx, doc1)
			s2 = readWriteSyncState(t, ctx, doc2)
			syncQuiescent(t, ctx, s1, s2)

			assert.Equal(t, sortedHeadHex(t, ctx, doc1), sortedHeadHex(t, ctx, doc2))
		})
	}
}

// TestRustSync_FirstResponseIsSomeEvenIfNoChanges reproduces
// first_response_is_some_even_if_no_changes.
func TestRustSync_FirstResponseIsSomeEvenIfNoChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(1))
			require.NoError(t, err)
			closeDocument(t, doc1)
			putRoot(t, ctx, doc1, "key", "value", "put", commitTime)

			doc2, err := doc1.Fork(ctx, actor(2))
			require.NoError(t, err)
			closeDocument(t, doc2)

			s1 := readWriteSyncState(t, ctx, doc1)
			s2 := readWriteSyncState(t, ctx, doc2)

			message, ok, err := s1.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok)
			require.NoError(t, s2.ReceiveMessage(ctx, message))

			_, ok, err = s2.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.True(t, ok, "first response must be sent even with equal heads")
		})
	}
}

// TestRustSync_ShouldNotReplyIfNoDataAfterFirstRound reproduces
// should_not_reply_if_we_have_no_data_after_first_round.
func TestRustSync_ShouldNotReplyIfNoDataAfterFirstRound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(1))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(2))
			require.NoError(t, err)
			closeDocument(t, doc2)

			s1 := readWriteSyncState(t, ctx, doc1)
			s2 := readWriteSyncState(t, ctx, doc2)

			message, ok, err := s1.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok)
			require.NoError(t, s2.ReceiveMessage(ctx, message))

			_, ok, err = s2.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok, "first round response expected")

			_, ok, err = s1.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.False(t, ok)

			_, ok, err = s2.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.False(t, ok)
		})
	}
}

// TestRustSync_AllowSimultaneousMessages reproduces
// should_allow_simultaneous_messages_during_synchronisation.
func TestRustSync_AllowSimultaneousMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, doc2)

			for i := range 5 {
				require.NoError(t, doc1.Root().PutScalar(
					ctx,
					"x",
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
				))
				_, err = doc1.Commit(ctx, "x", commitTime.Add(time.Duration(i)*time.Second))
				require.NoError(t, err)
				require.NoError(t, doc2.Root().PutScalar(
					ctx,
					"y",
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
				))
				_, err = doc2.Commit(ctx, "y", commitTime.Add(time.Duration(i)*time.Second))
				require.NoError(t, err)
			}

			s1 := readWriteSyncState(t, ctx, doc1)
			s2 := readWriteSyncState(t, ctx, doc2)
			syncQuiescent(t, ctx, s1, s2)

			assert.Equal(t, sortedHeadHex(t, ctx, doc1), sortedHeadHex(t, ctx, doc2))
			assert.True(t, rootHasKey(t, ctx, doc1, "y"))
			assert.True(t, rootHasKey(t, ctx, doc2, "x"))
		})
	}
}

// TestRustSync_BothReadOnlyOneMakesLocalChanges reproduces
// both_read_only_one_makes_local_changes.
func TestRustSync_BothReadOnlyOneMakesLocalChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(1))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(2))
			require.NoError(t, err)
			closeDocument(t, doc2)

			s1 := readOnlySyncState(t, ctx, doc1)
			s2 := readOnlySyncState(t, ctx, doc2)
			syncQuiescent(t, ctx, s1, s2)

			putRoot(t, ctx, doc1, "key", "value1", "value1", commitTime)
			syncQuiescent(t, ctx, s1, s2)
			assert.False(t, rootHasKey(t, ctx, doc2, "key"))

			putRoot(t, ctx, doc1, "key", "value2", "value2", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, s1, s2)
			assert.False(t, rootHasKey(t, ctx, doc2, "key"))

			_, ok, err := s1.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.False(t, ok)
			_, ok, err = s2.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.False(t, ok)
		})
	}
}

// TestRustSync_BothReadOnlySimultaneousChanges reproduces
// both_read_only_simultaneous_changes_during_sync.
func TestRustSync_BothReadOnlySimultaneousChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, doc2)

			s1 := readOnlySyncState(t, ctx, doc1)
			s2 := readOnlySyncState(t, ctx, doc2)

			putRoot(t, ctx, doc1, "x", "1", "x1", commitTime)
			putRoot(t, ctx, doc2, "y", "2", "y2", commitTime)
			syncQuiescent(t, ctx, s1, s2)

			putRoot(t, ctx, doc1, "x", "3", "x3", commitTime.Add(time.Second))
			putRoot(t, ctx, doc2, "y", "4", "y4", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, s1, s2)

			_, ok, err := s1.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.False(t, ok)
			_, ok, err = s2.GenerateMessage(ctx)
			require.NoError(t, err)
			assert.False(t, ok)

			assert.False(t, rootHasKey(t, ctx, doc1, "y"))
			assert.False(t, rootHasKey(t, ctx, doc2, "x"))
		})
	}
}

// TestRustSync_ReadOnlyPeerNewChangesBetweenRounds reproduces
// read_only_peer_new_changes_between_sync_rounds.
func TestRustSync_ReadOnlyPeerNewChangesBetweenRounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, doc2)

			putRoot(t, ctx, doc1, "round1", "from_doc1", "r1a", commitTime)
			putRoot(t, ctx, doc2, "round1", "from_doc2", "r1b", commitTime)

			s1 := readOnlySyncState(t, ctx, doc1)
			s2 := readWriteSyncState(t, ctx, doc2)
			syncQuiescent(t, ctx, s1, s2)
			assert.True(t, rootHasKey(t, ctx, doc2, "round1"))

			putRoot(t, ctx, doc1, "round2", "new_from_doc1", "r2a", commitTime.Add(time.Second))
			putRoot(t, ctx, doc2, "round2", "new_from_doc2", "r2b", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, s1, s2)

			values, err := doc2.Root().Scalars(ctx, "round2")
			require.NoError(t, err)

			found := make(map[string]bool)
			for _, value := range values {
				found[value.String] = true
			}

			assert.True(t, found["new_from_doc1"])
			assert.True(t, found["new_from_doc2"])

			doc1Values, err := doc1.Root().Scalars(ctx, "round2")
			require.NoError(t, err)
			require.Len(t, doc1Values, 1)
			assert.Equal(t, "new_from_doc1", doc1Values[0].String)
			assert.False(t, rootHasKey(t, ctx, doc1, "from_doc2"))
		})
	}
}

// TestRustSync_ReadOnlyPeerConcurrentChanges reproduces
// read_only_peer_concurrent_changes_during_sync.
func TestRustSync_ReadOnlyPeerConcurrentChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, doc2)

			s1 := readOnlySyncState(t, ctx, doc1)
			s2 := readWriteSyncState(t, ctx, doc2)
			syncQuiescent(t, ctx, s1, s2)

			require.NoError(t, doc2.Root().PutScalar(
				ctx,
				"x",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 0},
			))
			_, err = doc2.Commit(ctx, "x", commitTime.Add(time.Second))
			require.NoError(t, err)

			message, ok, err := s2.GenerateMessage(ctx)
			require.NoError(t, err)
			require.True(t, ok)
			require.NoError(t, s1.ReceiveMessage(ctx, message))

			require.NoError(t, doc1.Root().PutScalar(
				ctx,
				"y",
				automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
			))
			_, err = doc1.Commit(ctx, "y", commitTime.Add(2*time.Second))
			require.NoError(t, err)

			syncQuiescent(t, ctx, s1, s2)

			assert.True(t, rootHasKey(t, ctx, doc2, "y"))
			assert.False(t, rootHasKey(t, ctx, doc1, "x"))
		})
	}
}

// TestRustSync_SwitchReadWriteToReadOnlyMidSession reproduces
// switch_read_write_to_read_only_mid_session.
func TestRustSync_SwitchReadWriteToReadOnlyMidSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			docA, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, docA)

			docB, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, docB)

			putRoot(t, ctx, docA, "from_a", "hello", "a", commitTime)
			putRoot(t, ctx, docB, "from_b", "world", "b", commitTime)

			sa := readWriteSyncState(t, ctx, docA)
			sb := readWriteSyncState(t, ctx, docB)
			syncQuiescent(t, ctx, sa, sb)
			assert.Equal(t, sortedHeadHex(t, ctx, docA), sortedHeadHex(t, ctx, docB))

			require.NoError(t, sa.SetReadOnly(ctx, true))

			putRoot(t, ctx, docB, "new_from_b", "secret", "nb", commitTime.Add(time.Second))
			putRoot(t, ctx, docA, "new_from_a", "published", "na", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, sa, sb)

			assert.True(t, rootHasKey(t, ctx, docB, "new_from_a"))
			assert.False(t, rootHasKey(t, ctx, docA, "new_from_b"))
		})
	}
}

// TestRustSync_SwitchReadOnlyToReadWriteMultipleRounds reproduces
// switch_read_only_to_read_write_with_multiple_rounds.
func TestRustSync_SwitchReadOnlyToReadWriteMultipleRounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			docA, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, docA)

			docB, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, docB)

			putRoot(t, ctx, docA, "from_a", "initial", "a", commitTime)

			sa := readOnlySyncState(t, ctx, docA)
			sb := readWriteSyncState(t, ctx, docB)

			for round := 1; round <= 3; round++ {
				putRoot(
					t,
					ctx,
					docB,
					roundKey(round),
					"from_b",
					"b",
					commitTime.Add(time.Duration(round)*time.Second),
				)
				syncQuiescent(t, ctx, sa, sb)
				assert.False(t, rootHasKey(t, ctx, docA, roundKey(round)))
			}

			require.NoError(t, sa.SetReadOnly(ctx, false))
			syncQuiescent(t, ctx, sa, sb)

			for round := 1; round <= 3; round++ {
				assert.True(t, rootHasKey(t, ctx, docA, roundKey(round)))
			}

			assert.Equal(t, sortedHeadHex(t, ctx, docA), sortedHeadHex(t, ctx, docB))
		})
	}
}

// TestRustSync_ToggleReadOnlyMultipleTimes reproduces
// toggle_read_only_multiple_times.
func TestRustSync_ToggleReadOnlyMultipleTimes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			docA, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, docA)

			docB, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, docB)

			sa := readOnlySyncState(t, ctx, docA)
			sb := readWriteSyncState(t, ctx, docB)

			putRoot(t, ctx, docB, "b1", "val", "b1", commitTime)
			putRoot(t, ctx, docA, "a1", "val", "a1", commitTime)
			syncQuiescent(t, ctx, sa, sb)
			assert.True(t, rootHasKey(t, ctx, docB, "a1"))
			assert.False(t, rootHasKey(t, ctx, docA, "b1"))

			require.NoError(t, sa.SetReadOnly(ctx, false))
			putRoot(t, ctx, docB, "b2", "val", "b2", commitTime.Add(time.Second))
			putRoot(t, ctx, docA, "a2", "val", "a2", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, sa, sb)
			assert.True(t, rootHasKey(t, ctx, docA, "b1"))
			assert.True(t, rootHasKey(t, ctx, docA, "b2"))
			assert.True(t, rootHasKey(t, ctx, docB, "a2"))

			require.NoError(t, sa.SetReadOnly(ctx, true))
			putRoot(t, ctx, docB, "b3", "val", "b3", commitTime.Add(2*time.Second))
			putRoot(t, ctx, docA, "a3", "val", "a3", commitTime.Add(2*time.Second))
			syncQuiescent(t, ctx, sa, sb)
			assert.True(t, rootHasKey(t, ctx, docB, "a3"))
			assert.False(t, rootHasKey(t, ctx, docA, "b3"))

			require.NoError(t, sa.SetReadOnly(ctx, false))
			syncQuiescent(t, ctx, sa, sb)
			assert.True(t, rootHasKey(t, ctx, docA, "b3"))
			assert.Equal(t, sortedHeadHex(t, ctx, docA), sortedHeadHex(t, ctx, docB))
		})
	}
}

// TestRustSync_BothToggleAfterMultipleReadOnlyRounds reproduces
// both_toggle_after_multiple_read_only_rounds.
func TestRustSync_BothToggleAfterMultipleReadOnlyRounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			doc1, err := engine.open(ctx, actor(0xab))
			require.NoError(t, err)
			closeDocument(t, doc1)

			doc2, err := engine.open(ctx, actor(0xcd))
			require.NoError(t, err)
			closeDocument(t, doc2)

			s1 := readOnlySyncState(t, ctx, doc1)
			s2 := readOnlySyncState(t, ctx, doc2)

			for i := range 5 {
				putRoot(
					t,
					ctx,
					doc1,
					doc1Round(i),
					"v",
					"d1",
					commitTime.Add(time.Duration(i)*time.Second),
				)
				putRoot(
					t,
					ctx,
					doc2,
					doc2Round(i),
					"v",
					"d2",
					commitTime.Add(time.Duration(i)*time.Second),
				)
				syncQuiescent(t, ctx, s1, s2)
			}

			for i := range 5 {
				assert.False(t, rootHasKey(t, ctx, doc1, doc2Round(i)))
				assert.False(t, rootHasKey(t, ctx, doc2, doc1Round(i)))
			}

			require.NoError(t, s1.SetReadOnly(ctx, false))
			require.NoError(t, s2.SetReadOnly(ctx, false))
			syncQuiescent(t, ctx, s1, s2)

			for i := range 5 {
				assert.True(t, rootHasKey(t, ctx, doc1, doc2Round(i)))
				assert.True(t, rootHasKey(t, ctx, doc2, doc1Round(i)))
			}

			assert.Equal(t, sortedHeadHex(t, ctx, doc1), sortedHeadHex(t, ctx, doc2))
		})
	}
}

// TestRustSync_ReadOnlyPublisherToMultipleConsumers reproduces
// read_only_publisher_to_multiple_consumers.
func TestRustSync_ReadOnlyPublisherToMultipleConsumers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			r, err := engine.open(ctx, actor(0xaa))
			require.NoError(t, err)
			closeDocument(t, r)

			a, err := engine.open(ctx, actor(0xbb))
			require.NoError(t, err)
			closeDocument(t, a)

			b, err := engine.open(ctx, actor(0xcc))
			require.NoError(t, err)
			closeDocument(t, b)

			putRoot(t, ctx, r, "from_r", "hello", "r", commitTime)

			srA := readOnlySyncState(t, ctx, r)
			saR := readWriteSyncState(t, ctx, a)
			syncQuiescent(t, ctx, srA, saR)
			assert.True(t, rootHasKey(t, ctx, a, "from_r"))

			putRoot(t, ctx, a, "from_a", "world", "a", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, srA, saR)
			assert.False(t, rootHasKey(t, ctx, r, "from_a"))

			srB := readOnlySyncState(t, ctx, r)
			sbR := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, srB, sbR)
			assert.True(t, rootHasKey(t, ctx, b, "from_r"))
			assert.False(t, rootHasKey(t, ctx, b, "from_a"))
		})
	}
}

// TestRustSync_ReadOnlyFullyConnectedTriangle reproduces
// read_only_fully_connected_triangle.
func TestRustSync_ReadOnlyFullyConnectedTriangle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			r, err := engine.open(ctx, actor(0xaa))
			require.NoError(t, err)
			closeDocument(t, r)

			a, err := engine.open(ctx, actor(0xbb))
			require.NoError(t, err)
			closeDocument(t, a)

			b, err := engine.open(ctx, actor(0xcc))
			require.NoError(t, err)
			closeDocument(t, b)

			putRoot(t, ctx, r, "from_r", "r_val", "r", commitTime)
			putRoot(t, ctx, a, "from_a", "a_val", "a", commitTime)
			putRoot(t, ctx, b, "from_b", "b_val", "b", commitTime)
			rHeads := sortedHeadHex(t, ctx, r)

			srA := readOnlySyncState(t, ctx, r)
			saR := readWriteSyncState(t, ctx, a)
			syncQuiescent(t, ctx, srA, saR)

			srB := readOnlySyncState(t, ctx, r)
			sbR := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, srB, sbR)

			assert.True(t, rootHasKey(t, ctx, a, "from_r"))
			assert.True(t, rootHasKey(t, ctx, b, "from_r"))

			saB := readWriteSyncState(t, ctx, a)
			sbA := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, saB, sbA)

			for _, document := range []*automerge.Document{a, b} {
				assert.True(t, rootHasKey(t, ctx, document, "from_a"))
				assert.True(t, rootHasKey(t, ctx, document, "from_b"))
				assert.True(t, rootHasKey(t, ctx, document, "from_r"))
			}

			assert.Equal(t, sortedHeadHex(t, ctx, a), sortedHeadHex(t, ctx, b))
			assert.Equal(t, rHeads, sortedHeadHex(t, ctx, r))
			assert.False(t, rootHasKey(t, ctx, r, "from_a"))
			assert.False(t, rootHasKey(t, ctx, r, "from_b"))
		})
	}
}

// TestRustSync_StaleSharedHeadsAfterReadOnlySync reproduces
// stale_shared_heads_after_read_only_sync.
func TestRustSync_StaleSharedHeadsAfterReadOnlySync(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			r, err := engine.open(ctx, actor(0xaa))
			require.NoError(t, err)
			closeDocument(t, r)

			a, err := engine.open(ctx, actor(0xbb))
			require.NoError(t, err)
			closeDocument(t, a)

			b, err := engine.open(ctx, actor(0xcc))
			require.NoError(t, err)
			closeDocument(t, b)

			for i := range 10 {
				require.NoError(t, r.Root().PutScalar(
					ctx,
					"counter",
					automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
				))
				_, err = r.Commit(ctx, "counter", commitTime.Add(time.Duration(i)*time.Second))
				require.NoError(t, err)
			}

			putRoot(t, ctx, a, "from_a", "a_val", "a", commitTime)

			srA := readOnlySyncState(t, ctx, r)
			saR := readWriteSyncState(t, ctx, a)
			syncQuiescent(t, ctx, srA, saR)
			assert.True(t, rootHasKey(t, ctx, a, "counter"))

			saB := readWriteSyncState(t, ctx, a)
			sbA := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, saB, sbA)
			assert.True(t, rootHasKey(t, ctx, b, "counter"))
			assert.True(t, rootHasKey(t, ctx, b, "from_a"))

			srB := readOnlySyncState(t, ctx, r)
			sbR := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, srB, sbR)

			assert.False(t, rootHasKey(t, ctx, r, "from_a"))
			assert.True(t, rootHasKey(t, ctx, b, "counter"))
			assert.True(t, rootHasKey(t, ctx, b, "from_a"))
		})
	}
}

// TestRustSync_ReadOnlyPeerReceivesSameChangesFromTwoPeers reproduces
// read_only_peer_receives_same_changes_from_two_peers.
func TestRustSync_ReadOnlyPeerReceivesSameChangesFromTwoPeers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, engine := range rustParityEngines() {
		t.Run(engine.name, func(t *testing.T) {
			t.Parallel()

			r, err := engine.open(ctx, actor(0xaa))
			require.NoError(t, err)
			closeDocument(t, r)

			a, err := engine.open(ctx, actor(0xbb))
			require.NoError(t, err)
			closeDocument(t, a)

			b, err := engine.open(ctx, actor(0xcc))
			require.NoError(t, err)
			closeDocument(t, b)

			putRoot(t, ctx, r, "from_r", "r_val", "r", commitTime)
			putRoot(t, ctx, a, "from_a", "a_val", "a", commitTime)
			putRoot(t, ctx, b, "from_b", "b_val", "b", commitTime)

			saB := readWriteSyncState(t, ctx, a)
			sbA := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, saB, sbA)
			assert.Equal(t, sortedHeadHex(t, ctx, a), sortedHeadHex(t, ctx, b))

			rHeads := sortedHeadHex(t, ctx, r)

			srA := readOnlySyncState(t, ctx, r)
			saR := readWriteSyncState(t, ctx, a)
			syncQuiescent(t, ctx, srA, saR)
			assert.True(t, rootHasKey(t, ctx, a, "from_r"))
			assert.Equal(t, rHeads, sortedHeadHex(t, ctx, r))

			srB := readOnlySyncState(t, ctx, r)
			sbR := readWriteSyncState(t, ctx, b)
			syncQuiescent(t, ctx, srB, sbR)
			assert.True(t, rootHasKey(t, ctx, b, "from_r"))

			assert.Equal(t, rHeads, sortedHeadHex(t, ctx, r))
			assert.False(t, rootHasKey(t, ctx, r, "from_a"))
			assert.False(t, rootHasKey(t, ctx, r, "from_b"))

			putRoot(t, ctx, r, "from_r_2", "new", "r2", commitTime.Add(time.Second))
			syncQuiescent(t, ctx, srA, saR)
			assert.True(t, rootHasKey(t, ctx, a, "from_r_2"))
			syncQuiescent(t, ctx, srB, sbR)
			assert.True(t, rootHasKey(t, ctx, b, "from_r_2"))
		})
	}
}

func roundKey(round int) string {
	return "round" + string(rune('0'+round))
}

func doc1Round(index int) string {
	return "doc1_r" + string(rune('0'+index))
}

func doc2Round(index int) string {
	return "doc2_r" + string(rune('0'+index))
}
