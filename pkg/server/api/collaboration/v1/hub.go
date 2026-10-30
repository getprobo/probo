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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/realtime"
)

type (
	Wake struct {
		Refresh bool
	}

	RoomPeer struct {
		connectionID string
		wake         chan Wake
		// ephemeral carries opaque automerge-repo gossip frames (presence,
		// cursors) destined for this peer. Delivery is best-effort: a full
		// buffer drops the frame.
		ephemeral chan []byte
	}

	presenceEntry struct {
		frame []byte
		at    time.Time
	}

	Documents interface {
		Open(
			context.Context,
			coredata.Scoper,
			gid.GID,
		) (*collaboration.DocumentCollaboration, error)
		Persist(
			context.Context,
			coredata.Scoper,
			gid.GID,
			*automerge.Document,
		) (int64, error)
	}

	seedReleaser interface {
		ReleaseSeed(context.Context, coredata.Scoper, gid.GID, time.Time) error
	}

	Hub struct {
		mu         sync.Mutex
		documents  Documents
		rooms      map[gid.GID]*Room
		instanceID string
	}

	Room struct {
		mu            sync.Mutex
		docMu         sync.Mutex
		collaboration *collaboration.DocumentCollaboration
		documents     Documents
		scope         coredata.Scoper
		documentID    gid.GID
		revision      atomic.Int64
		peers         map[uint64]RoomPeer
		presence      map[string]presenceEntry
		nextPeerID    uint64
		dirty         chan struct{}
		stop          chan struct{}
		done          chan struct{}
		persistErr    error
		opening       chan struct{}
		seedWait      chan struct{}
		seedOnce      sync.Once
		seedReady     bool
		seedReleased  bool
		seedClaimedAt time.Time
	}

	RoomLease struct {
		hub          *Hub
		documentID   gid.GID
		room         *Room
		peerID       uint64
		remotePeerID string
		WakeCh       <-chan Wake
		EphemeralCh  <-chan []byte
		seedOwner    bool
		once         sync.Once
	}
)

const EphemeralBuffer = 64

const (
	PersistDebounce = 50 * time.Millisecond
	PersistTimeout  = 5 * time.Second
	PresenceTTL     = 30 * time.Second
)

func NewConnectionID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("cannot generate collaboration connection ID: %w", err)
	}

	return hex.EncodeToString(value[:]), nil
}

func NewHub(documents Documents, events *realtime.Events) *Hub {
	instanceID, err := NewConnectionID()
	if err != nil {
		instanceID = ""
	}

	hub := &Hub{
		documents:  documents,
		rooms:      make(map[gid.GID]*Room),
		instanceID: instanceID,
	}
	if events != nil {
		events.Subscribe(hub.notifyExternal)
	}

	return hub
}

func (h *Hub) InstanceID() string {
	return h.instanceID
}

func (h *Hub) Acquire(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	connectionID string,
) (*RoomLease, error) {
	for {
		h.mu.Lock()
		if room := h.rooms[documentID]; room != nil {
			opening := room.opening
			seedWait := room.seedWait
			seedReady := room.seedReady
			h.mu.Unlock()

			if opening != nil {
				select {
				case <-opening:
				case <-ctx.Done():
					return nil, fmt.Errorf("cannot wait for collaboration room to open: %w", ctx.Err())
				}

				continue
			}

			if !seedReady && seedWait != nil {
				select {
				case <-seedWait:
				case <-ctx.Done():
					return nil, fmt.Errorf("cannot wait for collaboration seed: %w", ctx.Err())
				}

				continue
			}

			h.mu.Lock()

			current := h.rooms[documentID]
			if current != room || current.opening != nil || (!current.seedReady && current.seedWait != nil) {
				h.mu.Unlock()

				continue
			}

			lease := h.addPeerLocked(documentID, current, connectionID)
			h.mu.Unlock()

			return lease, nil
		}

		room := &Room{
			documents:  h.documents,
			scope:      scope,
			documentID: documentID,
			peers:      make(map[uint64]RoomPeer),
			presence:   make(map[string]presenceEntry),
			dirty:      make(chan struct{}, 1),
			stop:       make(chan struct{}),
			done:       make(chan struct{}),
			opening:    make(chan struct{}),
		}
		h.rooms[documentID] = room
		h.mu.Unlock()

		opened, err := h.documents.Open(ctx, scope, documentID)
		if err != nil {
			h.finishOpening(documentID, room)

			return nil, fmt.Errorf("cannot open collaboration room: %w", err)
		}

		h.mu.Lock()
		room.collaboration = opened
		room.revision.Store(opened.Revision)

		if opened.SeedClaimedAt != nil {
			room.seedClaimedAt = *opened.SeedClaimedAt
		}

		if opened.NeedsSeed {
			room.seedWait = make(chan struct{})
		} else {
			room.seedReady = true
		}

		lease := h.addPeerLocked(documentID, room, connectionID)
		lease.seedOwner = opened.NeedsSeed

		go room.run()
		h.mu.Unlock()
		h.finishOpening(documentID, room)

		return lease, nil
	}
}

func (h *Hub) finishOpening(documentID gid.GID, room *Room) {
	h.mu.Lock()
	if room.collaboration == nil && h.rooms[documentID] == room {
		delete(h.rooms, documentID)
	}

	opening := room.opening
	room.opening = nil
	h.mu.Unlock()

	if opening != nil {
		close(opening)
	}
}

func (h *Hub) addPeerLocked(
	documentID gid.GID,
	room *Room,
	connectionID string,
) *RoomLease {
	room.mu.Lock()
	peerID := room.nextPeerID
	room.nextPeerID++
	wake := make(chan Wake, 8)
	ephemeral := make(chan []byte, EphemeralBuffer)
	room.peers[peerID] = RoomPeer{
		connectionID: connectionID,
		wake:         wake,
		ephemeral:    ephemeral,
	}
	room.mu.Unlock()

	return &RoomLease{
		hub:         h,
		documentID:  documentID,
		room:        room,
		peerID:      peerID,
		WakeCh:      wake,
		EphemeralCh: ephemeral,
	}
}

func (l *RoomLease) Collaboration() *collaboration.DocumentCollaboration {
	return l.room.collaboration
}

// LockDocument serializes mutations of the shared room document across every
// connection in the room. It must wrap any read-modify-write on the document
// (such as the comment-only snapshot/validate/restore) so a concurrent editor's
// change on another socket cannot interleave and be rolled back.
func (l *RoomLease) LockDocument() {
	l.room.docMu.Lock()
}

func (l *RoomLease) UnlockDocument() {
	l.room.docMu.Unlock()
}

func (l *RoomLease) Revision() int64 {
	return l.room.revision.Load()
}

func (l *RoomLease) SeedOwner() bool {
	return l.seedOwner
}

func (l *RoomLease) SeedClaimedAt() time.Time {
	if l == nil || l.room == nil {
		return time.Time{}
	}

	return l.room.seedClaimedAt
}

func (l *RoomLease) SetRevision(revision int64) {
	l.room.revision.Store(revision)
}

func (l *RoomLease) NotifyPeers() {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	for peerID, peer := range l.room.peers {
		if peerID == l.peerID {
			continue
		}

		select {
		case peer.wake <- Wake{}:
		default:
		}
	}
}

func (l *RoomLease) BroadcastEphemeral(frame []byte) {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	l.room.rememberEphemeralLocked(frame)

	for peerID, peer := range l.room.peers {
		if peerID == l.peerID {
			continue
		}

		select {
		case peer.ephemeral <- frame:
		default:
		}
	}
}

func (l *RoomLease) SetRemotePeerID(remotePeerID string) {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	l.remotePeerID = remotePeerID
}

func (l *RoomLease) ReplayPresence() {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	peer, ok := l.room.peers[l.peerID]
	if !ok {
		return
	}

	now := time.Now()

	for senderID, entry := range l.room.presence {
		if senderID == l.remotePeerID {
			continue
		}

		if now.Sub(entry.at) > PresenceTTL {
			delete(l.room.presence, senderID)

			continue
		}

		select {
		case peer.ephemeral <- entry.frame:
		default:
		}
	}
}

func (r *Room) rememberEphemeralLocked(frame []byte) {
	message, err := collabpkg.DecodeMessage(frame)
	if err != nil || message.Type != collabpkg.MessageEphemeral {
		return
	}

	if collabpkg.IsDepartureEphemeral(message) {
		delete(r.presence, message.SenderID)

		return
	}

	if r.presence == nil {
		r.presence = make(map[string]presenceEntry)
	}

	cached := append([]byte(nil), frame...)
	r.presence[message.SenderID] = presenceEntry{
		frame: cached,
		at:    time.Now(),
	}
}

func (h *Hub) notifyExternal(payload string) {
	if ephemeral, ok := realtime.DecodeCollaborationEphemeral(payload); ok {
		h.notifyExternalEphemeral(ephemeral)
		return
	}

	documentID, err := gid.ParseGID(payload)
	if err != nil || documentID.EntityType() != coredata.DocumentEntityType {
		return
	}

	h.mu.Lock()
	room := h.rooms[documentID]
	h.mu.Unlock()

	if room == nil {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	for _, peer := range room.peers {
		select {
		case peer.wake <- Wake{Refresh: true}:
		default:
		}
	}
}

func (h *Hub) notifyExternalEphemeral(ephemeral realtime.CollaborationEphemeral) {
	if ephemeral.InstanceID != "" && ephemeral.InstanceID == h.instanceID {
		return
	}

	documentID, err := gid.ParseGID(ephemeral.DocumentID)
	if err != nil || documentID.EntityType() != coredata.DocumentEntityType {
		return
	}

	h.mu.Lock()
	room := h.rooms[documentID]
	h.mu.Unlock()

	if room == nil {
		return
	}

	room.fanoutEphemeral(ephemeral.Frame)
}

func (r *Room) fanoutEphemeral(frame []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rememberEphemeralLocked(frame)

	for _, peer := range r.peers {
		select {
		case peer.ephemeral <- frame:
		default:
		}
	}
}

func (l *RoomLease) SchedulePersist() {
	select {
	case l.room.dirty <- struct{}{}:
	default:
	}
}

func (l *RoomLease) PersistError() error {
	l.room.mu.Lock()
	defer l.room.mu.Unlock()

	return l.room.persistErr
}

func (l *RoomLease) Close() {
	l.once.Do(func() {
		l.hub.mu.Lock()
		l.room.mu.Lock()
		delete(l.room.peers, l.peerID)

		if l.remotePeerID != "" && l.room.presence != nil {
			delete(l.room.presence, l.remotePeerID)
		}

		empty := len(l.room.peers) == 0
		l.room.mu.Unlock()

		abortSeed := l.seedOwner && !l.room.seedReady

		remove := empty || abortSeed
		if remove && l.hub.rooms[l.documentID] == l.room {
			delete(l.hub.rooms, l.documentID)
		}

		l.hub.mu.Unlock()

		if abortSeed {
			releaseCtx, cancel := context.WithTimeout(context.Background(), PersistTimeout)
			_ = l.releaseUnseeded(releaseCtx)

			cancel()
			l.room.signalSeed()
		}

		if !remove {
			return
		}

		if l.room.stop != nil {
			close(l.room.stop)
			<-l.room.done
		}

		if l.room.collaboration != nil && l.room.collaboration.Document != nil {
			_ = l.room.collaboration.Document.Close()
		}
	})
}

func (l *RoomLease) completeSeed() {
	if l.hub == nil || l.room == nil {
		return
	}

	l.hub.mu.Lock()
	l.room.seedReady = true
	l.hub.mu.Unlock()
	l.room.signalSeed()
}

func (l *RoomLease) releaseUnseeded(ctx context.Context) error {
	if l.hub == nil || l.room == nil || !l.seedOwner {
		return nil
	}

	l.hub.mu.Lock()
	if l.room.seedReady || l.room.seedReleased {
		l.hub.mu.Unlock()

		return nil
	}

	l.hub.mu.Unlock()

	releaser, ok := l.room.documents.(seedReleaser)
	if !ok {
		l.markSeedReleased()

		return nil
	}

	if err := releaser.ReleaseSeed(
		ctx,
		l.room.scope,
		l.room.documentID,
		l.room.seedClaimedAt,
	); err != nil {
		return err
	}

	l.markSeedReleased()

	return nil
}

func (l *RoomLease) markSeedReleased() {
	if l.hub == nil || l.room == nil {
		return
	}

	l.hub.mu.Lock()
	l.room.seedReleased = true
	l.hub.mu.Unlock()
}

func (r *Room) signalSeed() {
	r.seedOnce.Do(func() {
		if r.seedWait != nil {
			close(r.seedWait)
		}
	})
}

func (r *Room) run() {
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
				timer = time.NewTimer(PersistDebounce)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}

				timer.Reset(PersistDebounce)
			}
		case <-timerChannel:
			if dirty {
				if err := r.persist(); err != nil {
					timer.Reset(PersistDebounce)

					continue
				}

				dirty = false
			}

			timer = nil
		case <-r.stop:
			if timer != nil {
				timer.Stop()
			}

			if dirty {
				_ = r.persist()
			}

			return
		}
	}
}

func (r *Room) persist() error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		PersistTimeout,
	)
	defer cancel()

	revision, err := r.documents.Persist(
		ctx,
		r.scope,
		r.documentID,
		r.collaboration.Document,
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.persistErr = err
	if err != nil {
		return err
	}

	r.revision.Store(revision)

	return nil
}
