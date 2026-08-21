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

package console_v1

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/realtime"
)

type fakeDocumentCollaborationDocuments struct {
	collaboration *probo.DocumentCollaboration
	persisted     atomic.Int64
	persistedCh   chan struct{}
}

func (f *fakeDocumentCollaborationDocuments) OpenCollaboration(
	context.Context,
	coredata.Scoper,
	gid.GID,
) (*probo.DocumentCollaboration, error) {
	return f.collaboration, nil
}

func (f *fakeDocumentCollaborationDocuments) PersistCollaboration(
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

func TestDocumentCollaborationRoom_NotifiesOtherPeers(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)
	document, err := automerge.New(context.Background(), automerge.ActorID{1})
	require.NoError(t, err)

	room := &documentCollaborationRoom{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers: make(map[uint64]documentCollaborationRoomPeer),
	}
	room.revision.Store(1)
	hub := &documentCollaborationHub{
		rooms: map[gid.GID]*documentCollaborationRoom{
			versionID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(versionID, room, "first")
	second := hub.addPeerLocked(versionID, room, "second")
	hub.mu.Unlock()

	first.SetRevision(2)
	first.NotifyPeers()

	assert.Equal(t, int64(2), second.Revision())

	select {
	case <-second.Wake:
	default:
		require.Fail(t, "second peer was not notified")
	}

	select {
	case <-first.Wake:
		require.Fail(t, "originating peer must not notify itself")
	default:
	}

	hub.notifyExternal(versionID.String())

	for _, wakeChannel := range []<-chan documentCollaborationWake{
		first.Wake,
		second.Wake,
	} {
		select {
		case wake := <-wakeChannel:
			assert.True(t, wake.refresh)
		default:
			require.Fail(t, "external notification did not wake peer")
		}
	}

	first.Close()
	assert.Contains(t, hub.rooms, versionID)
	second.Close()
	assert.NotContains(t, hub.rooms, versionID)
}

func TestDocumentCollaborationRoom_BroadcastsEphemeralToOtherPeers(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)
	document, err := automerge.New(context.Background(), automerge.ActorID{3})
	require.NoError(t, err)

	room := &documentCollaborationRoom{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		peers: make(map[uint64]documentCollaborationRoomPeer),
	}
	room.revision.Store(1)
	hub := &documentCollaborationHub{
		rooms: map[gid.GID]*documentCollaborationRoom{
			versionID: room,
		},
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(versionID, room, "first")
	second := hub.addPeerLocked(versionID, room, "second")
	third := hub.addPeerLocked(versionID, room, "third")
	hub.mu.Unlock()

	frame := []byte{0x01, 0x02, 0x03}
	first.BroadcastEphemeral(frame)

	for _, lease := range []*documentCollaborationRoomLease{second, third} {
		select {
		case got := <-lease.Ephemeral:
			assert.Equal(t, frame, got)
		default:
			require.Fail(t, "peer did not receive the ephemeral frame")
		}
	}

	select {
	case <-first.Ephemeral:
		require.Fail(t, "originating peer must not receive its own ephemeral frame")
	default:
	}

	require.NoError(t, document.Close(context.Background()))
}

func TestDocumentCollaborationHub_DeliversExternalEphemeral(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)
	document, err := automerge.New(context.Background(), automerge.ActorID{5})
	require.NoError(t, err)

	room := &documentCollaborationRoom{
		collaboration: &probo.DocumentCollaboration{Document: document, Revision: 1},
		peers:         make(map[uint64]documentCollaborationRoomPeer),
	}
	room.revision.Store(1)
	hub := &documentCollaborationHub{
		rooms:      map[gid.GID]*documentCollaborationRoom{versionID: room},
		instanceID: "local-instance",
	}

	hub.mu.Lock()
	first := hub.addPeerLocked(versionID, room, "first")
	second := hub.addPeerLocked(versionID, room, "second")
	hub.mu.Unlock()

	frame := []byte{0x0a, 0x0b, 0x0c}
	remote, err := realtime.EncodeCollaborationEphemeral(realtime.CollaborationEphemeral{
		VersionID:  versionID.String(),
		InstanceID: "remote-instance",
		Frame:      frame,
	})
	require.NoError(t, err)

	hub.notifyExternal(remote)

	// A frame from another instance reaches every local peer.
	for _, lease := range []*documentCollaborationRoomLease{first, second} {
		select {
		case got := <-lease.Ephemeral:
			assert.Equal(t, frame, got)
		default:
			require.Fail(t, "peer did not receive the external ephemeral frame")
		}
	}

	// This instance's own echo is ignored: it already delivered locally.
	own, err := realtime.EncodeCollaborationEphemeral(realtime.CollaborationEphemeral{
		VersionID:  versionID.String(),
		InstanceID: "local-instance",
		Frame:      []byte{0xff},
	})
	require.NoError(t, err)

	hub.notifyExternal(own)

	select {
	case <-first.Ephemeral:
		require.Fail(t, "the hub must ignore its own ephemeral echo")
	default:
	}

	require.NoError(t, document.Close(context.Background()))
}

func TestDocumentCollaborationRoom_DropsEphemeralWhenBufferFull(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)
	document, err := automerge.New(context.Background(), automerge.ActorID{4})
	require.NoError(t, err)

	room := &documentCollaborationRoom{
		collaboration: &probo.DocumentCollaboration{Document: document, Revision: 1},
		peers:         make(map[uint64]documentCollaborationRoomPeer),
	}
	room.revision.Store(1)
	hub := &documentCollaborationHub{
		rooms: map[gid.GID]*documentCollaborationRoom{versionID: room},
	}

	hub.mu.Lock()
	sender := hub.addPeerLocked(versionID, room, "sender")
	_ = hub.addPeerLocked(versionID, room, "slow")
	hub.mu.Unlock()

	for range documentCollaborationEphemeralBuffer + 10 {
		sender.BroadcastEphemeral([]byte{0xff})
	}

	require.NoError(t, document.Close(context.Background()))
}

func TestDocumentCollaborationRoom_DebouncesPersistence(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	versionID := gid.New(tenantID, coredata.DocumentVersionEntityType)
	document, err := automerge.New(context.Background(), automerge.ActorID{2})
	require.NoError(t, err)

	documents := &fakeDocumentCollaborationDocuments{
		collaboration: &probo.DocumentCollaboration{
			Document: document,
			Revision: 1,
		},
		persistedCh: make(chan struct{}, 1),
	}
	room := &documentCollaborationRoom{
		collaboration: documents.collaboration,
		documents:     documents,
		scope:         coredata.NewScope(tenantID),
		versionID:     versionID,
		peers:         make(map[uint64]documentCollaborationRoomPeer),
		dirty:         make(chan struct{}, 1),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}

	room.revision.Store(1)
	go room.run()

	lease := &documentCollaborationRoomLease{room: room}
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
	require.NoError(t, document.Close(context.Background()))
}
