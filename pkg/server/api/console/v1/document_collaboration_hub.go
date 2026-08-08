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
	"fmt"
	"sync"
	"sync/atomic"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
)

type (
	documentCollaborationHub struct {
		mu        sync.Mutex
		documents *probo.DocumentService
		rooms     map[gid.GID]*documentCollaborationRoom
	}

	documentCollaborationRoom struct {
		mu            sync.Mutex
		collaboration *probo.DocumentCollaboration
		revision      atomic.Int64
		peers         map[uint64]chan struct{}
		nextPeerID    uint64
	}

	documentCollaborationRoomLease struct {
		hub               *documentCollaborationHub
		documentVersionID gid.GID
		room              *documentCollaborationRoom
		peerID            uint64
		Wake              <-chan struct{}
		seedOwner         bool
		once              sync.Once
	}
)

func newDocumentCollaborationHub(
	documents *probo.DocumentService,
) *documentCollaborationHub {
	return &documentCollaborationHub{
		documents: documents,
		rooms:     make(map[gid.GID]*documentCollaborationRoom),
	}
}

func (h *documentCollaborationHub) acquire(
	ctx context.Context,
	scope coredata.Scoper,
	documentVersionID gid.GID,
) (*documentCollaborationRoomLease, error) {
	h.mu.Lock()
	if room := h.rooms[documentVersionID]; room != nil {
		lease := h.addPeerLocked(documentVersionID, room)
		h.mu.Unlock()

		return lease, nil
	}

	h.mu.Unlock()

	collaboration, err := h.documents.OpenCollaboration(
		ctx,
		scope,
		documentVersionID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open collaboration room: %w", err)
	}

	room := &documentCollaborationRoom{
		collaboration: collaboration,
		peers:         make(map[uint64]chan struct{}),
	}
	room.revision.Store(collaboration.Revision)

	h.mu.Lock()
	if existing := h.rooms[documentVersionID]; existing != nil {
		_ = collaboration.Document.Close(context.Background())
		lease := h.addPeerLocked(documentVersionID, existing)
		h.mu.Unlock()

		return lease, nil
	}

	h.rooms[documentVersionID] = room
	lease := h.addPeerLocked(documentVersionID, room)
	lease.seedOwner = collaboration.NeedsSeed
	h.mu.Unlock()

	return lease, nil
}

func (h *documentCollaborationHub) addPeerLocked(
	documentVersionID gid.GID,
	room *documentCollaborationRoom,
) *documentCollaborationRoomLease {
	room.mu.Lock()
	peerID := room.nextPeerID
	room.nextPeerID++
	wake := make(chan struct{}, 1)
	room.peers[peerID] = wake
	room.mu.Unlock()

	return &documentCollaborationRoomLease{
		hub:               h,
		documentVersionID: documentVersionID,
		room:              room,
		peerID:            peerID,
		Wake:              wake,
	}
}

func (l *documentCollaborationRoomLease) Collaboration() *probo.DocumentCollaboration {
	return l.room.collaboration
}

func (l *documentCollaborationRoomLease) Revision() int64 {
	return l.room.revision.Load()
}

func (l *documentCollaborationRoomLease) SeedOwner() bool {
	return l.seedOwner
}

func (l *documentCollaborationRoomLease) SetRevision(revision int64) {
	l.room.revision.Store(revision)
}

func (l *documentCollaborationRoomLease) NotifyPeers() {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	for peerID, wake := range l.room.peers {
		if peerID == l.peerID {
			continue
		}

		select {
		case wake <- struct{}{}:
		default:
		}
	}
}

func (l *documentCollaborationRoomLease) Close() {
	l.once.Do(func() {
		l.hub.mu.Lock()
		defer l.hub.mu.Unlock()

		l.room.mu.Lock()
		delete(l.room.peers, l.peerID)
		empty := len(l.room.peers) == 0
		l.room.mu.Unlock()

		if !empty {
			return
		}

		delete(l.hub.rooms, l.documentVersionID)
		_ = l.room.collaboration.Document.Close(context.Background())
	})
}
