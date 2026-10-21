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

package collaboration_v1

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/realtime"
)

type fakeDocuments struct {
	opened      *probo.DocumentCollaboration
	persisted   atomic.Int64
	persistedCh chan struct{}
}

func (f *fakeDocuments) Open(
	context.Context,
	coredata.Scoper,
	gid.GID,
) (*probo.DocumentCollaboration, error) {
	return f.opened, nil
}

func (f *fakeDocuments) Persist(
	context.Context,
	coredata.Scoper,
	gid.GID,
	*automerge.Document,
) (int64, error) {
	revision := f.persisted.Add(1) + 1

	select {
	case f.persistedCh <- struct{}{}:
	default:
	}

	return revision, nil
}

func TestRoom_NotifiesOtherPeers(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{1})
	require.NoError(t, err)

	room := &Room{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers: make(map[uint64]RoomPeer),
	}
	room.revision.Store(1)
	hub := &Hub{
		rooms: map[gid.GID]*Room{
			documentID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(documentID, room, "first")
	second := hub.addPeerLocked(documentID, room, "second")
	hub.mu.Unlock()

	first.SetRevision(2)
	first.NotifyPeers()

	assert.Equal(t, int64(2), second.Revision())

	select {
	case <-second.WakeCh:
	default:
		require.Fail(t, "second peer was not notified")
	}

	select {
	case <-first.WakeCh:
		require.Fail(t, "originating peer must not notify itself")
	default:
	}

	hub.notifyExternal(documentID.String())

	for _, wakeChannel := range []<-chan Wake{
		first.WakeCh,
		second.WakeCh,
	} {
		select {
		case wake := <-wakeChannel:
			assert.True(t, wake.Refresh)
		default:
			require.Fail(t, "external notification did not wake peer")
		}
	}

	first.Close()
	assert.Contains(t, hub.rooms, documentID)
	second.Close()
	assert.NotContains(t, hub.rooms, documentID)
}

func TestRoom_ReplaysPresenceOnJoin(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{7})
	require.NoError(t, err)

	room := &Room{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers:    make(map[uint64]RoomPeer),
		presence: make(map[string]presenceEntry),
	}
	room.revision.Store(1)
	hub := &Hub{
		rooms: map[gid.GID]*Room{
			documentID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(documentID, room, "first")
	hub.mu.Unlock()
	first.SetRemotePeerID("peer-a")

	frame, err := collabpkg.EncodeMessage(
		collabpkg.Message{
			Type:       collabpkg.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "gateway",
			DocumentID: "doc",
			SessionID:  "session-a",
			Count:      1,
			Data:       []byte{0x01},
		},
	)
	require.NoError(t, err)

	first.BroadcastEphemeral(frame)

	hub.mu.Lock()
	second := hub.addPeerLocked(documentID, room, "second")
	hub.mu.Unlock()
	second.SetRemotePeerID("peer-b")
	second.ReplayPresence()

	select {
	case got := <-second.EphemeralCh:
		assert.Equal(t, frame, got)
	default:
		require.Fail(t, "joining peer did not receive the cached presence frame")
	}

	secondFrame, err := collabpkg.EncodeMessage(
		collabpkg.Message{
			Type:       collabpkg.MessageEphemeral,
			SenderID:   "peer-b",
			TargetID:   "gateway",
			DocumentID: "doc",
			SessionID:  "session-b",
			Count:      1,
			Data:       []byte{0x02},
		},
	)
	require.NoError(t, err)

	second.BroadcastEphemeral(secondFrame)

	select {
	case got := <-first.EphemeralCh:
		assert.Equal(t, secondFrame, got)
	default:
		require.Fail(t, "existing peer did not receive the new presence frame")
	}

	first.Close()
	assert.NotContains(t, room.presence, "peer-a")
	assert.Contains(t, room.presence, "peer-b")

	second.ReplayPresence()

	select {
	case <-second.EphemeralCh:
		require.Fail(t, "joining peer must not receive its own cached presence")
	default:
	}

	require.NoError(t, document.Close())
}

func TestRoom_DepartureDropsCachedPresence(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{8})
	require.NoError(t, err)

	room := &Room{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers:    make(map[uint64]RoomPeer),
		presence: make(map[string]presenceEntry),
	}
	room.revision.Store(1)
	hub := &Hub{
		rooms: map[gid.GID]*Room{
			documentID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(documentID, room, "first")
	second := hub.addPeerLocked(documentID, room, "second")
	hub.mu.Unlock()
	first.SetRemotePeerID("peer-a")

	selection, err := collabpkg.EncodeMessage(
		collabpkg.Message{
			Type:       collabpkg.MessageEphemeral,
			SenderID:   "peer-a",
			TargetID:   "gateway",
			DocumentID: "doc",
			SessionID:  "session-a",
			Count:      1,
			Data:       []byte{0x01},
		},
	)
	require.NoError(t, err)

	first.BroadcastEphemeral(selection)

	select {
	case got := <-second.EphemeralCh:
		assert.Equal(t, selection, got)
	default:
		require.Fail(t, "peer did not receive the selection frame")
	}

	require.Contains(t, room.presence, "peer-a")

	departure, err := collabpkg.EncodeDepartureEphemeral(
		"peer-a",
		"gateway",
		"doc",
		"departure-session",
	)
	require.NoError(t, err)

	first.BroadcastEphemeral(departure)

	select {
	case got := <-second.EphemeralCh:
		assert.Equal(t, departure, got)
	default:
		require.Fail(t, "peer did not receive the departure frame")
	}

	assert.NotContains(t, room.presence, "peer-a")

	hub.mu.Lock()
	third := hub.addPeerLocked(documentID, room, "third")
	hub.mu.Unlock()
	third.SetRemotePeerID("peer-c")
	third.ReplayPresence()

	select {
	case <-third.EphemeralCh:
		require.Fail(t, "replay must not include a departed sender")
	default:
	}

	require.NoError(t, document.Close())
}

func TestRoom_BroadcastsEphemeralToOtherPeers(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{3})
	require.NoError(t, err)

	room := &Room{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers: make(map[uint64]RoomPeer),
	}
	room.revision.Store(1)
	hub := &Hub{
		rooms: map[gid.GID]*Room{
			documentID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(documentID, room, "first")
	second := hub.addPeerLocked(documentID, room, "second")
	third := hub.addPeerLocked(documentID, room, "third")
	hub.mu.Unlock()

	frame := []byte{0x01, 0x02, 0x03}
	first.BroadcastEphemeral(frame)

	for _, lease := range []*RoomLease{second, third} {
		select {
		case got := <-lease.EphemeralCh:
			assert.Equal(t, frame, got)
		default:
			require.Fail(t, "peer did not receive the ephemeral frame")
		}
	}

	select {
	case <-first.EphemeralCh:
		require.Fail(t, "originating peer must not receive its own ephemeral frame")
	default:
	}

	require.NoError(t, document.Close())
}

func TestHub_DeliversExternalEphemeral(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{5})
	require.NoError(t, err)

	room := &Room{
		collaboration: &probo.DocumentCollaboration{Document: document, Revision: 1},
		peers:         make(map[uint64]RoomPeer),
	}
	room.revision.Store(1)
	hub := &Hub{
		rooms:      map[gid.GID]*Room{documentID: room},
		instanceID: "local-instance",
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(documentID, room, "first")
	second := hub.addPeerLocked(documentID, room, "second")
	hub.mu.Unlock()

	frame := []byte{0x0a, 0x0b, 0x0c}
	remote, err := realtime.EncodeCollaborationEphemeral(realtime.CollaborationEphemeral{
		DocumentID: documentID.String(),
		InstanceID: "remote-instance",
		Frame:      frame,
	})
	require.NoError(t, err)

	hub.notifyExternal(remote)

	for _, lease := range []*RoomLease{first, second} {
		select {
		case got := <-lease.EphemeralCh:
			assert.Equal(t, frame, got)
		default:
			require.Fail(t, "peer did not receive the external ephemeral frame")
		}
	}

	own, err := realtime.EncodeCollaborationEphemeral(realtime.CollaborationEphemeral{
		DocumentID: documentID.String(),
		InstanceID: "local-instance",
		Frame:      []byte{0xff},
	})
	require.NoError(t, err)

	hub.notifyExternal(own)

	select {
	case <-first.EphemeralCh:
		require.Fail(t, "the hub must ignore its own ephemeral echo")
	default:
	}

	require.NoError(t, document.Close())
}

func TestRoom_DropsEphemeralWhenBufferFull(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{4})
	require.NoError(t, err)

	room := &Room{
		collaboration: &probo.DocumentCollaboration{Document: document, Revision: 1},
		peers:         make(map[uint64]RoomPeer),
	}
	room.revision.Store(1)
	hub := &Hub{
		rooms: map[gid.GID]*Room{documentID: room},
	}

	hub.mu.Lock()
	sender := hub.addPeerLocked(documentID, room, "sender")
	_ = hub.addPeerLocked(documentID, room, "slow")
	hub.mu.Unlock()

	for range EphemeralBuffer + 10 {
		sender.BroadcastEphemeral([]byte{0xff})
	}

	require.NoError(t, document.Close())
}

func TestRoom_DebouncesPersistence(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	documentID := gid.New(tenantID, coredata.DocumentEntityType)
	document, err := automerge.New(automerge.ActorID{2})
	require.NoError(t, err)

	documents := &fakeDocuments{
		opened: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		persistedCh: make(chan struct{}, 1),
	}
	room := &Room{
		collaboration: documents.opened,
		documents:     documents,
		scope:         coredata.NewScope(tenantID),
		documentID:    documentID,
		peers:         make(map[uint64]RoomPeer),
		dirty:         make(chan struct{}, 1),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}

	room.revision.Store(1)
	go room.run()

	lease := &RoomLease{room: room}
	for range 20 {
		lease.SchedulePersist()
	}

	select {
	case <-documents.persistedCh:
	case <-time.After(time.Second):
		require.Fail(t, "collaboration room did not persist")
	}

	assert.Equal(t, int64(1), documents.persisted.Load())
	assert.Equal(t, int64(2), room.revision.Load())

	close(room.stop)
	<-room.done
	require.NoError(t, document.Close())
}
