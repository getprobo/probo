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
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/internal/core"
)

// drainSyncMessages mirrors the server's sendAvailableSyncMessages: it generates
// messages until the protocol quiesces, bounded so a non-quiescing state fails
// loudly instead of looping forever.
func drainSyncMessages(t *testing.T, state *automerge.SyncState) int {
	t.Helper()

	for count := range 100 {
		_, ok, err := state.GenerateMessage()
		require.NoError(t, err)

		if !ok {
			return count
		}
	}

	t.Fatal("sync protocol did not quiesce")

	return 0
}

// TestSyncState_QuiescesWithOrphanedChange reproduces the production livelock: a
// document holding an orphaned change (its base never arrived) recomputes a Need
// for the missing base on every received message. Because that Need never
// changed, regenerating it must not keep producing messages, or the server's
// bounded send loop aborts the whole collaboration connection.
func TestSyncState_QuiescesWithOrphanedChange(t *testing.T) {
	t.Parallel()

	// Build two dependent changes, then apply only the second so its base is
	// missing and it stays queued as an orphan.
	source, err := automerge.New(actor(1))
	require.NoError(t, err)
	closeDocument(t, source)

	text, err := source.CreateText("body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(0, 0, "first"))

	base, err := source.Commit("base", commitTime)
	require.NoError(t, err)

	require.NoError(t, text.Splice(5, 0, " second"))
	_, err = source.Commit("child", commitTime)
	require.NoError(t, err)

	childChanges, err := source.ChangesSince([]automerge.Hash{base})
	require.NoError(t, err)
	require.Len(t, childChanges, 1)

	orphanHost, err := automerge.New(actor(2))
	require.NoError(t, err)
	closeDocument(t, orphanHost)

	// Give the host its own history first so the incoming orphan takes the
	// change-queue path rather than initializing an empty document.
	hostText, err := orphanHost.CreateText("host")
	require.NoError(t, err)
	require.NoError(t, hostText.Splice(0, 0, "local"))
	_, err = orphanHost.Commit("host", commitTime)
	require.NoError(t, err)

	// The child change depends on the base the host never received, so it is
	// retained as an orphan rather than applied.
	require.NoError(t, orphanHost.ApplyChanges([]automerge.Change{childChanges[0]}))

	// A fresh peer with an empty document handshakes with the orphan host.
	peer, err := automerge.New(actor(3))
	require.NoError(t, err)
	closeDocument(t, peer)

	hostState, err := orphanHost.NewSyncState()
	require.NoError(t, err)

	peerState, err := peer.NewSyncState()
	require.NoError(t, err)

	// Exchange a few rounds. Each round the host recomputes the same Need for the
	// missing base; the send loop must still quiesce every time.
	for round := range 5 {
		peerMessage, ok, err := peerState.GenerateMessage()
		require.NoError(t, err)

		if !ok {
			break
		}

		require.NoError(t, hostState.ReceiveMessage(peerMessage))

		sent := drainSyncMessages(t, hostState)
		require.LessOrEqualf(
			t,
			sent,
			1,
			"round %d: host sent %d messages for an unchanging Need",
			round,
			sent,
		)

		hostMessage, ok, err := hostState.GenerateMessage()
		require.NoError(t, err)

		if ok {
			require.NoError(t, peerState.ReceiveMessage(hostMessage))
		}
	}
}

// TestSyncState_QuiescesWhenReadOnlyPeerRequestsChanges pins a second
// non-quiescing state found by the model-based chaos test: a peer requests a
// missing change and marks itself read-only in the same message. The sender
// cannot service that request, so retaining it must not make generation loop.
// Once the peer becomes writable it advertises its missing heads again.
func TestSyncState_QuiescesWhenReadOnlyPeerRequestsChanges(t *testing.T) {
	t.Parallel()

	source, err := automerge.New(actor(4))
	require.NoError(t, err)
	closeDocument(t, source)

	sourceState, err := source.NewSyncState()
	require.NoError(t, err)

	var requested [32]byte

	requested[0] = 1

	// Model the message found by the chaos test: a peer advertises read-only and
	// still carries a stale Need from its previous writable mode.
	message, err := (core.SyncMessage{
		Version: core.SyncMessageVersion2,
		Need:    [][32]byte{requested},
		Flags:   []byte{2, 0x80 | 0x02 | 0x04},
	}).Encode()
	require.NoError(t, err)
	require.NoError(t, sourceState.ReceiveMessage(message))

	sent := drainSyncMessages(t, sourceState)
	require.LessOrEqual(t, sent, 1)
}
