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
	"time"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/realtime"
)

type (
	documentCollaborationWake struct {
		refresh bool
	}

	documentCollaborationDocuments interface {
		OpenCollaboration(
			context.Context,
			coredata.Scoper,
			gid.GID,
		) (*probo.DocumentCollaboration, error)
		PersistCollaboration(
			context.Context,
			coredata.Scoper,
			gid.GID,
			*automerge.Document,
		) (int64, error)
	}

	documentCollaborationHub struct {
		mu        sync.Mutex
		documents documentCollaborationDocuments
		rooms     map[gid.GID]*documentCollaborationRoom
	}

	documentCollaborationRoom struct {
		mu            sync.Mutex
		collaboration *probo.DocumentCollaboration
		documents     documentCollaborationDocuments
		scope         coredata.Scoper
		versionID     gid.GID
		revision      atomic.Int64
		peers         map[uint64]chan documentCollaborationWake
		nextPeerID    uint64
		dirty         chan struct{}
		stop          chan struct{}
		done          chan struct{}
		persistErr    error
	}

	documentCollaborationRoomLease struct {
		hub               *documentCollaborationHub
		documentVersionID gid.GID
		room              *documentCollaborationRoom
		peerID            uint64
		Wake              <-chan documentCollaborationWake
		seedOwner         bool
		once              sync.Once
	}
)

const (
	documentCollaborationPersistDebounce = 50 * time.Millisecond
	documentCollaborationPersistTimeout  = 30 * time.Second
)

func newDocumentCollaborationHub(
	documents documentCollaborationDocuments,
	events *realtime.Events,
) *documentCollaborationHub {
	hub := &documentCollaborationHub{
		documents: documents,
		rooms:     make(map[gid.GID]*documentCollaborationRoom),
	}
	if events != nil {
		events.Subscribe(hub.notifyExternal)
	}

	return hub
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
		documents:     h.documents,
		scope:         scope,
		versionID:     documentVersionID,
		peers:         make(map[uint64]chan documentCollaborationWake),
		dirty:         make(chan struct{}, 1),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
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

	go room.run()
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
	wake := make(chan documentCollaborationWake, 1)
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
		case wake <- documentCollaborationWake{}:
		default:
		}
	}
}

func (h *documentCollaborationHub) notifyExternal(payload string) {
	documentVersionID, err := gid.ParseGID(payload)
	if err != nil || documentVersionID.EntityType() != coredata.DocumentVersionEntityType {
		return
	}

	h.mu.Lock()
	room := h.rooms[documentVersionID]
	h.mu.Unlock()

	if room == nil {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for _, wake := range room.peers {
		select {
		case wake <- documentCollaborationWake{refresh: true}:
		default:
		}
	}
}

func (l *documentCollaborationRoomLease) SchedulePersist() {
	select {
	case l.room.dirty <- struct{}{}:
	default:
	}
}

func (l *documentCollaborationRoomLease) PersistError() error {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	return l.room.persistErr
}

func (l *documentCollaborationRoomLease) Close() {
	l.once.Do(func() {
		l.hub.mu.Lock()
		l.room.mu.Lock()
		delete(l.room.peers, l.peerID)
		empty := len(l.room.peers) == 0
		l.room.mu.Unlock()

		if !empty {
			l.hub.mu.Unlock()

			return
		}

		delete(l.hub.rooms, l.documentVersionID)
		l.hub.mu.Unlock()

		if l.room.stop != nil {
			close(l.room.stop)
			<-l.room.done
		}

		_ = l.room.collaboration.Document.Close(context.Background())
	})
}

func (r *documentCollaborationRoom) run() {
	defer close(r.done)

	var timer *time.Timer

	dirty := false

	for {
		var timerChannel <-chan time.Time
		if timer != nil {
			timerChannel = timer.C
		}

		select {
		case <-r.dirty:
			dirty = true

			if timer == nil {
				timer = time.NewTimer(documentCollaborationPersistDebounce)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}

				timer.Reset(documentCollaborationPersistDebounce)
			}
		case <-timerChannel:
			if dirty {
				r.persist()

				dirty = false
			}

			timer = nil
		case <-r.stop:
			if timer != nil {
				timer.Stop()
			}

			if dirty {
				r.persist()
			}

			return
		}
	}
}

func (r *documentCollaborationRoom) persist() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		documentCollaborationPersistTimeout,
	)
	defer cancel()

	revision, err := r.documents.PersistCollaboration(
		ctx,
		r.scope,
		r.versionID,
		r.collaboration.Document,
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.persistErr = err
	if err == nil {
		r.revision.Store(revision)
	}
}
