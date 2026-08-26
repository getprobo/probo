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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func readWriteSyncState(
	t *testing.T,
	document *automerge.Document,
) *automerge.SyncState {
	t.Helper()

	state, err := document.NewSyncState()
	require.NoError(t, err)
	closeSyncState(t, state)

	return state
}

func readOnlySyncState(
	t *testing.T,
	document *automerge.Document,
) *automerge.SyncState {
	t.Helper()

	state := readWriteSyncState(t, document)
	require.NoError(t, state.SetReadOnly(true))

	return state
}

// syncQuiescent exchanges messages in both directions until neither peer has
// anything to send, failing if the session does not converge. It mirrors the
// upstream sync() test helper.
func syncQuiescent(
	t *testing.T,
	left *automerge.SyncState,
	right *automerge.SyncState,
) {
	t.Helper()

	const maxRounds = 50

	for range maxRounds {
		leftMessage, leftOK, err := left.GenerateMessage()
		require.NoError(t, err)
		rightMessage, rightOK, err := right.GenerateMessage()
		require.NoError(t, err)

		if !leftOK && !rightOK {
			return
		}

		if leftOK {
			require.NoError(t, right.ReceiveMessage(leftMessage))
		}

		if rightOK {
			require.NoError(t, left.ReceiveMessage(rightMessage))
		}
	}

	t.Fatalf("sync did not converge within %d rounds", maxRounds)
}

func rootHasKey(
	t *testing.T,
	document *automerge.Document,
	key string,
) bool {
	t.Helper()

	_, err := document.Root().Scalar(key)

	return err == nil
}

func putRoot(
	t *testing.T,
	document *automerge.Document,
	key string,
	value string,
	message string,
	when time.Time,
) {
	t.Helper()

	require.NoError(
		t,
		document.Root().PutScalar(

			key,
			automerge.Scalar{Type: automerge.ScalarTypeString, String: value},
		),
	)
	_, err := document.Commit(message, when)
	require.NoError(t, err)
}

func putInt(
	t *testing.T,
	document *automerge.Document,
	key string,
	value int64,
	message string,
	when time.Time,
) {
	t.Helper()

	require.NoError(
		t,
		document.Root().PutScalar(

			key,
			automerge.Scalar{Type: automerge.ScalarTypeInt, Int: value},
		),
	)
	_, err := document.Commit(message, when)
	require.NoError(t, err)
}

// TestRustSync_FirstMessageNoHeadsSendsWholeDoc reproduces
// if_first_message_has_no_heads_and_supports_v2_message_send_whole_doc: when a
// peer starts empty, the other peer's first response carries the entire document
// so the empty peer converges after a single response.
func TestRustSync_FirstMessageNoHeadsSendsWholeDoc(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				empty, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, empty)

				populated, err := engine.open(actor(2))
				require.NoError(t, err)
				closeDocument(t, populated)
				putRoot(t, populated, "foo", "bar", "seed", commitTime)

				emptyState := readWriteSyncState(t, empty)
				populatedState := readWriteSyncState(t, populated)

				request, ok, err := emptyState.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, populatedState.ReceiveMessage(request))

				response, ok, err := populatedState.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, emptyState.ReceiveMessage(response))

				assert.True(
					t,
					rootHasKey(t, empty, "foo"),
					"empty peer should receive the whole document in the first response",
				)

				value, err := empty.Root().Scalar("foo")
				require.NoError(t, err)
				assert.Equal(t, "bar", value.String)
			},
		)
	}
}

// TestRustSync_BranchingAndMerging reproduces
// should_handle_lots_of_branching_and_merging: two peers exchange many
// concurrent changes, a third peer's change is merged into one of them, and a
// final synchronization must converge both peers to identical heads.
func TestRustSync_BranchingAndMerging(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0x01))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(0x89))
				require.NoError(t, err)
				closeDocument(t, doc2)

				doc3, err := engine.open(actor(0xfe))
				require.NoError(t, err)
				closeDocument(t, doc3)

				putInt(t, doc1, "x", 0, "x0", commitTime)
				_, err = doc2.Merge(doc1)
				require.NoError(t, err)
				_, err = doc3.Merge(doc1)
				require.NoError(t, err)

				putInt(t, doc3, "x", 1, "x1", commitTime.Add(time.Second))

				for i := int64(1); i < 20; i++ {
					when := commitTime.Add(time.Duration(i+1) * time.Second)
					putInt(t, doc1, "n1", i, "n1", when)
					putInt(t, doc2, "n2", i, "n2", when)
					_, err = doc1.Merge(doc2)
					require.NoError(t, err)
					_, err = doc2.Merge(doc1)
					require.NoError(t, err)
				}

				s1 := readWriteSyncState(t, doc1)
				s2 := readWriteSyncState(t, doc2)
				syncQuiescent(t, s1, s2)

				// doc3's change is concurrent to the last sync heads, forcing the
				// slower reconciliation path on the next synchronization.
				_, err = doc2.Merge(doc3)
				require.NoError(t, err)

				putInt(t, doc1, "n1", 100, "n1 final", commitTime.Add(time.Hour))
				putInt(t, doc2, "n1", 100, "n1 final", commitTime.Add(time.Hour))

				s1 = readWriteSyncState(t, doc1)
				s2 = readWriteSyncState(t, doc2)
				syncQuiescent(t, s1, s2)

				assert.Equal(t, sortedHeadHex(t, doc1), sortedHeadHex(t, doc2))
			},
		)
	}
}

// TestRustSync_FirstResponseIsSomeEvenIfNoChanges reproduces
// first_response_is_some_even_if_no_changes.
func TestRustSync_FirstResponseIsSomeEvenIfNoChanges(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc1)
				putRoot(t, doc1, "key", "value", "put", commitTime)

				doc2, err := doc1.Fork(actor(2))
				require.NoError(t, err)
				closeDocument(t, doc2)

				s1 := readWriteSyncState(t, doc1)
				s2 := readWriteSyncState(t, doc2)

				message, ok, err := s1.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, s2.ReceiveMessage(message))

				_, ok, err = s2.GenerateMessage()
				require.NoError(t, err)
				assert.True(t, ok, "first response must be sent even with equal heads")
			},
		)
	}
}

// TestRustSync_ShouldNotReplyIfNoDataAfterFirstRound reproduces
// should_not_reply_if_we_have_no_data_after_first_round.
func TestRustSync_ShouldNotReplyIfNoDataAfterFirstRound(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(2))
				require.NoError(t, err)
				closeDocument(t, doc2)

				s1 := readWriteSyncState(t, doc1)
				s2 := readWriteSyncState(t, doc2)

				message, ok, err := s1.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, s2.ReceiveMessage(message))

				_, ok, err = s2.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok, "first round response expected")

				_, ok, err = s1.GenerateMessage()
				require.NoError(t, err)
				assert.False(t, ok)

				_, ok, err = s2.GenerateMessage()
				require.NoError(t, err)
				assert.False(t, ok)
			},
		)
	}
}

// TestRustSync_AllowSimultaneousMessages reproduces
// should_allow_simultaneous_messages_during_synchronisation.
func TestRustSync_AllowSimultaneousMessages(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, doc2)

				for i := range 5 {
					require.NoError(
						t,
						doc1.Root().PutScalar(

							"x",
							automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
						),
					)
					_, err = doc1.Commit("x", commitTime.Add(time.Duration(i)*time.Second))
					require.NoError(t, err)
					require.NoError(
						t,
						doc2.Root().PutScalar(

							"y",
							automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
						),
					)
					_, err = doc2.Commit("y", commitTime.Add(time.Duration(i)*time.Second))
					require.NoError(t, err)
				}

				s1 := readWriteSyncState(t, doc1)
				s2 := readWriteSyncState(t, doc2)
				syncQuiescent(t, s1, s2)

				assert.Equal(t, sortedHeadHex(t, doc1), sortedHeadHex(t, doc2))
				assert.True(t, rootHasKey(t, doc1, "y"))
				assert.True(t, rootHasKey(t, doc2, "x"))
			},
		)
	}
}

// TestRustSync_BothReadOnlyOneMakesLocalChanges reproduces
// both_read_only_one_makes_local_changes.
func TestRustSync_BothReadOnlyOneMakesLocalChanges(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(1))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(2))
				require.NoError(t, err)
				closeDocument(t, doc2)

				s1 := readOnlySyncState(t, doc1)
				s2 := readOnlySyncState(t, doc2)
				syncQuiescent(t, s1, s2)

				putRoot(t, doc1, "key", "value1", "value1", commitTime)
				syncQuiescent(t, s1, s2)
				assert.False(t, rootHasKey(t, doc2, "key"))

				putRoot(t, doc1, "key", "value2", "value2", commitTime.Add(time.Second))
				syncQuiescent(t, s1, s2)
				assert.False(t, rootHasKey(t, doc2, "key"))

				_, ok, err := s1.GenerateMessage()
				require.NoError(t, err)
				assert.False(t, ok)
				_, ok, err = s2.GenerateMessage()
				require.NoError(t, err)
				assert.False(t, ok)
			},
		)
	}
}

// TestRustSync_BothReadOnlySimultaneousChanges reproduces
// both_read_only_simultaneous_changes_during_sync.
func TestRustSync_BothReadOnlySimultaneousChanges(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, doc2)

				s1 := readOnlySyncState(t, doc1)
				s2 := readOnlySyncState(t, doc2)

				putRoot(t, doc1, "x", "1", "x1", commitTime)
				putRoot(t, doc2, "y", "2", "y2", commitTime)
				syncQuiescent(t, s1, s2)

				putRoot(t, doc1, "x", "3", "x3", commitTime.Add(time.Second))
				putRoot(t, doc2, "y", "4", "y4", commitTime.Add(time.Second))
				syncQuiescent(t, s1, s2)

				_, ok, err := s1.GenerateMessage()
				require.NoError(t, err)
				assert.False(t, ok)
				_, ok, err = s2.GenerateMessage()
				require.NoError(t, err)
				assert.False(t, ok)

				assert.False(t, rootHasKey(t, doc1, "y"))
				assert.False(t, rootHasKey(t, doc2, "x"))
			},
		)
	}
}

// TestRustSync_ReadOnlyPeerNewChangesBetweenRounds reproduces
// read_only_peer_new_changes_between_sync_rounds.
func TestRustSync_ReadOnlyPeerNewChangesBetweenRounds(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, doc2)

				putRoot(t, doc1, "round1", "from_doc1", "r1a", commitTime)
				putRoot(t, doc2, "round1", "from_doc2", "r1b", commitTime)

				s1 := readOnlySyncState(t, doc1)
				s2 := readWriteSyncState(t, doc2)
				syncQuiescent(t, s1, s2)
				assert.True(t, rootHasKey(t, doc2, "round1"))

				putRoot(t, doc1, "round2", "new_from_doc1", "r2a", commitTime.Add(time.Second))
				putRoot(t, doc2, "round2", "new_from_doc2", "r2b", commitTime.Add(time.Second))
				syncQuiescent(t, s1, s2)

				values, err := doc2.Root().Scalars("round2")
				require.NoError(t, err)

				found := make(map[string]bool)
				for _, value := range values {
					found[value.String] = true
				}

				assert.True(t, found["new_from_doc1"])
				assert.True(t, found["new_from_doc2"])

				doc1Values, err := doc1.Root().Scalars("round2")
				require.NoError(t, err)
				require.Len(t, doc1Values, 1)
				assert.Equal(t, "new_from_doc1", doc1Values[0].String)
				assert.False(t, rootHasKey(t, doc1, "from_doc2"))
			},
		)
	}
}

// TestRustSync_ReadOnlyPeerConcurrentChanges reproduces
// read_only_peer_concurrent_changes_during_sync.
func TestRustSync_ReadOnlyPeerConcurrentChanges(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, doc2)

				s1 := readOnlySyncState(t, doc1)
				s2 := readWriteSyncState(t, doc2)
				syncQuiescent(t, s1, s2)

				require.NoError(
					t,
					doc2.Root().PutScalar(

						"x",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 0},
					),
				)
				_, err = doc2.Commit("x", commitTime.Add(time.Second))
				require.NoError(t, err)

				message, ok, err := s2.GenerateMessage()
				require.NoError(t, err)
				require.True(t, ok)
				require.NoError(t, s1.ReceiveMessage(message))

				require.NoError(
					t,
					doc1.Root().PutScalar(

						"y",
						automerge.Scalar{Type: automerge.ScalarTypeInt, Int: 1},
					),
				)
				_, err = doc1.Commit("y", commitTime.Add(2*time.Second))
				require.NoError(t, err)

				syncQuiescent(t, s1, s2)

				assert.True(t, rootHasKey(t, doc2, "y"))
				assert.False(t, rootHasKey(t, doc1, "x"))
			},
		)
	}
}

// TestRustSync_SwitchReadWriteToReadOnlyMidSession reproduces
// switch_read_write_to_read_only_mid_session.
func TestRustSync_SwitchReadWriteToReadOnlyMidSession(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				docA, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, docA)

				docB, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, docB)

				putRoot(t, docA, "from_a", "hello", "a", commitTime)
				putRoot(t, docB, "from_b", "world", "b", commitTime)

				sa := readWriteSyncState(t, docA)
				sb := readWriteSyncState(t, docB)
				syncQuiescent(t, sa, sb)
				assert.Equal(t, sortedHeadHex(t, docA), sortedHeadHex(t, docB))

				require.NoError(t, sa.SetReadOnly(true))

				putRoot(t, docB, "new_from_b", "secret", "nb", commitTime.Add(time.Second))
				putRoot(t, docA, "new_from_a", "published", "na", commitTime.Add(time.Second))
				syncQuiescent(t, sa, sb)

				assert.True(t, rootHasKey(t, docB, "new_from_a"))
				assert.False(t, rootHasKey(t, docA, "new_from_b"))
			},
		)
	}
}

// TestRustSync_SwitchReadOnlyToReadWriteMultipleRounds reproduces
// switch_read_only_to_read_write_with_multiple_rounds.
func TestRustSync_SwitchReadOnlyToReadWriteMultipleRounds(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				docA, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, docA)

				docB, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, docB)

				putRoot(t, docA, "from_a", "initial", "a", commitTime)

				sa := readOnlySyncState(t, docA)
				sb := readWriteSyncState(t, docB)

				for round := 1; round <= 3; round++ {
					putRoot(
						t,
						docB,
						roundKey(round),
						"from_b",
						"b",
						commitTime.Add(time.Duration(round)*time.Second),
					)
					syncQuiescent(t, sa, sb)
					assert.False(t, rootHasKey(t, docA, roundKey(round)))
				}

				require.NoError(t, sa.SetReadOnly(false))
				syncQuiescent(t, sa, sb)

				for round := 1; round <= 3; round++ {
					assert.True(t, rootHasKey(t, docA, roundKey(round)))
				}

				assert.Equal(t, sortedHeadHex(t, docA), sortedHeadHex(t, docB))
			},
		)
	}
}

// TestRustSync_ToggleReadOnlyMultipleTimes reproduces
// toggle_read_only_multiple_times.
func TestRustSync_ToggleReadOnlyMultipleTimes(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				docA, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, docA)

				docB, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, docB)

				sa := readOnlySyncState(t, docA)
				sb := readWriteSyncState(t, docB)

				putRoot(t, docB, "b1", "val", "b1", commitTime)
				putRoot(t, docA, "a1", "val", "a1", commitTime)
				syncQuiescent(t, sa, sb)
				assert.True(t, rootHasKey(t, docB, "a1"))
				assert.False(t, rootHasKey(t, docA, "b1"))

				require.NoError(t, sa.SetReadOnly(false))
				putRoot(t, docB, "b2", "val", "b2", commitTime.Add(time.Second))
				putRoot(t, docA, "a2", "val", "a2", commitTime.Add(time.Second))
				syncQuiescent(t, sa, sb)
				assert.True(t, rootHasKey(t, docA, "b1"))
				assert.True(t, rootHasKey(t, docA, "b2"))
				assert.True(t, rootHasKey(t, docB, "a2"))

				require.NoError(t, sa.SetReadOnly(true))
				putRoot(t, docB, "b3", "val", "b3", commitTime.Add(2*time.Second))
				putRoot(t, docA, "a3", "val", "a3", commitTime.Add(2*time.Second))
				syncQuiescent(t, sa, sb)
				assert.True(t, rootHasKey(t, docB, "a3"))
				assert.False(t, rootHasKey(t, docA, "b3"))

				require.NoError(t, sa.SetReadOnly(false))
				syncQuiescent(t, sa, sb)
				assert.True(t, rootHasKey(t, docA, "b3"))
				assert.Equal(t, sortedHeadHex(t, docA), sortedHeadHex(t, docB))
			},
		)
	}
}

// TestRustSync_BothToggleAfterMultipleReadOnlyRounds reproduces
// both_toggle_after_multiple_read_only_rounds.
func TestRustSync_BothToggleAfterMultipleReadOnlyRounds(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				doc1, err := engine.open(actor(0xab))
				require.NoError(t, err)
				closeDocument(t, doc1)

				doc2, err := engine.open(actor(0xcd))
				require.NoError(t, err)
				closeDocument(t, doc2)

				s1 := readOnlySyncState(t, doc1)
				s2 := readOnlySyncState(t, doc2)

				for i := range 5 {
					putRoot(
						t,
						doc1,
						doc1Round(i),
						"v",
						"d1",
						commitTime.Add(time.Duration(i)*time.Second),
					)
					putRoot(
						t,
						doc2,
						doc2Round(i),
						"v",
						"d2",
						commitTime.Add(time.Duration(i)*time.Second),
					)
					syncQuiescent(t, s1, s2)
				}

				for i := range 5 {
					assert.False(t, rootHasKey(t, doc1, doc2Round(i)))
					assert.False(t, rootHasKey(t, doc2, doc1Round(i)))
				}

				require.NoError(t, s1.SetReadOnly(false))
				require.NoError(t, s2.SetReadOnly(false))
				syncQuiescent(t, s1, s2)

				for i := range 5 {
					assert.True(t, rootHasKey(t, doc1, doc2Round(i)))
					assert.True(t, rootHasKey(t, doc2, doc1Round(i)))
				}

				assert.Equal(t, sortedHeadHex(t, doc1), sortedHeadHex(t, doc2))
			},
		)
	}
}

// TestRustSync_ReadOnlyPublisherToMultipleConsumers reproduces
// read_only_publisher_to_multiple_consumers.
func TestRustSync_ReadOnlyPublisherToMultipleConsumers(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				r, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, r)

				a, err := engine.open(actor(0xbb))
				require.NoError(t, err)
				closeDocument(t, a)

				b, err := engine.open(actor(0xcc))
				require.NoError(t, err)
				closeDocument(t, b)

				putRoot(t, r, "from_r", "hello", "r", commitTime)

				srA := readOnlySyncState(t, r)
				saR := readWriteSyncState(t, a)
				syncQuiescent(t, srA, saR)
				assert.True(t, rootHasKey(t, a, "from_r"))

				putRoot(t, a, "from_a", "world", "a", commitTime.Add(time.Second))
				syncQuiescent(t, srA, saR)
				assert.False(t, rootHasKey(t, r, "from_a"))

				srB := readOnlySyncState(t, r)
				sbR := readWriteSyncState(t, b)
				syncQuiescent(t, srB, sbR)
				assert.True(t, rootHasKey(t, b, "from_r"))
				assert.False(t, rootHasKey(t, b, "from_a"))
			},
		)
	}
}

// TestRustSync_ReadOnlyFullyConnectedTriangle reproduces
// read_only_fully_connected_triangle.
func TestRustSync_ReadOnlyFullyConnectedTriangle(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				r, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, r)

				a, err := engine.open(actor(0xbb))
				require.NoError(t, err)
				closeDocument(t, a)

				b, err := engine.open(actor(0xcc))
				require.NoError(t, err)
				closeDocument(t, b)

				putRoot(t, r, "from_r", "r_val", "r", commitTime)
				putRoot(t, a, "from_a", "a_val", "a", commitTime)
				putRoot(t, b, "from_b", "b_val", "b", commitTime)
				rHeads := sortedHeadHex(t, r)

				srA := readOnlySyncState(t, r)
				saR := readWriteSyncState(t, a)
				syncQuiescent(t, srA, saR)

				srB := readOnlySyncState(t, r)
				sbR := readWriteSyncState(t, b)
				syncQuiescent(t, srB, sbR)

				assert.True(t, rootHasKey(t, a, "from_r"))
				assert.True(t, rootHasKey(t, b, "from_r"))

				saB := readWriteSyncState(t, a)
				sbA := readWriteSyncState(t, b)
				syncQuiescent(t, saB, sbA)

				for _, document := range []*automerge.Document{a, b} {
					assert.True(t, rootHasKey(t, document, "from_a"))
					assert.True(t, rootHasKey(t, document, "from_b"))
					assert.True(t, rootHasKey(t, document, "from_r"))
				}

				assert.Equal(t, sortedHeadHex(t, a), sortedHeadHex(t, b))
				assert.Equal(t, rHeads, sortedHeadHex(t, r))
				assert.False(t, rootHasKey(t, r, "from_a"))
				assert.False(t, rootHasKey(t, r, "from_b"))
			},
		)
	}
}

// TestRustSync_StaleSharedHeadsAfterReadOnlySync reproduces
// stale_shared_heads_after_read_only_sync.
func TestRustSync_StaleSharedHeadsAfterReadOnlySync(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				r, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, r)

				a, err := engine.open(actor(0xbb))
				require.NoError(t, err)
				closeDocument(t, a)

				b, err := engine.open(actor(0xcc))
				require.NoError(t, err)
				closeDocument(t, b)

				for i := range 10 {
					require.NoError(
						t,
						r.Root().PutScalar(

							"counter",
							automerge.Scalar{Type: automerge.ScalarTypeInt, Int: int64(i)},
						),
					)
					_, err = r.Commit("counter", commitTime.Add(time.Duration(i)*time.Second))
					require.NoError(t, err)
				}

				putRoot(t, a, "from_a", "a_val", "a", commitTime)

				srA := readOnlySyncState(t, r)
				saR := readWriteSyncState(t, a)
				syncQuiescent(t, srA, saR)
				assert.True(t, rootHasKey(t, a, "counter"))

				saB := readWriteSyncState(t, a)
				sbA := readWriteSyncState(t, b)
				syncQuiescent(t, saB, sbA)
				assert.True(t, rootHasKey(t, b, "counter"))
				assert.True(t, rootHasKey(t, b, "from_a"))

				srB := readOnlySyncState(t, r)
				sbR := readWriteSyncState(t, b)
				syncQuiescent(t, srB, sbR)

				assert.False(t, rootHasKey(t, r, "from_a"))
				assert.True(t, rootHasKey(t, b, "counter"))
				assert.True(t, rootHasKey(t, b, "from_a"))
			},
		)
	}
}

// TestRustSync_ReadOnlyPeerReceivesSameChangesFromTwoPeers reproduces
// read_only_peer_receives_same_changes_from_two_peers.
func TestRustSync_ReadOnlyPeerReceivesSameChangesFromTwoPeers(t *testing.T) {
	t.Parallel()

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				r, err := engine.open(actor(0xaa))
				require.NoError(t, err)
				closeDocument(t, r)

				a, err := engine.open(actor(0xbb))
				require.NoError(t, err)
				closeDocument(t, a)

				b, err := engine.open(actor(0xcc))
				require.NoError(t, err)
				closeDocument(t, b)

				putRoot(t, r, "from_r", "r_val", "r", commitTime)
				putRoot(t, a, "from_a", "a_val", "a", commitTime)
				putRoot(t, b, "from_b", "b_val", "b", commitTime)

				saB := readWriteSyncState(t, a)
				sbA := readWriteSyncState(t, b)
				syncQuiescent(t, saB, sbA)
				assert.Equal(t, sortedHeadHex(t, a), sortedHeadHex(t, b))

				rHeads := sortedHeadHex(t, r)

				srA := readOnlySyncState(t, r)
				saR := readWriteSyncState(t, a)
				syncQuiescent(t, srA, saR)
				assert.True(t, rootHasKey(t, a, "from_r"))
				assert.Equal(t, rHeads, sortedHeadHex(t, r))

				srB := readOnlySyncState(t, r)
				sbR := readWriteSyncState(t, b)
				syncQuiescent(t, srB, sbR)
				assert.True(t, rootHasKey(t, b, "from_r"))

				assert.Equal(t, rHeads, sortedHeadHex(t, r))
				assert.False(t, rootHasKey(t, r, "from_a"))
				assert.False(t, rootHasKey(t, r, "from_b"))

				putRoot(t, r, "from_r_2", "new", "r2", commitTime.Add(time.Second))
				syncQuiescent(t, srA, saR)
				assert.True(t, rootHasKey(t, a, "from_r_2"))
				syncQuiescent(t, srB, sbR)
				assert.True(t, rootHasKey(t, b, "from_r_2"))
			},
		)
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
