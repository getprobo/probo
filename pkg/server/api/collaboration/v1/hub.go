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
	"errors"
	"fmt"
	"sync"
	"time"

	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	"go.probo.inc/probo/pkg/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/realtime"
)

type (
	// Wake tells a connection that the room document changed, so it sends
	// its peer the sync frames that change produces.
	Wake struct{}

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
		// FetchChanges reads what the canonical document gained after a
		// known position, without touching the live document, and returns
		// nil when it did not move.
		FetchChanges(
			context.Context,
			coredata.Scoper,
			gid.GID,
			collaboration.Position,
		) (*collaboration.RefreshDelta, error)
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
		// position is how far the room document has caught up with the
		// canonical document. Only the room goroutine reads or writes it
		// once the room runs.
		position collaboration.Position
		// refresh asks the room goroutine to catch up with the canonical
		// document now, instead of at its next poll.
		refresh    chan struct{}
		peers      map[uint64]RoomPeer
		presence   map[string]presenceEntry
		nextPeerID uint64
		dirty      chan struct{}
		stop       chan struct{}
		done       chan struct{}
		persistErr error
		opening    chan struct{}
	}

	RoomLease struct {
		hub          *Hub
		documentID   gid.GID
		room         *Room
		peerID       uint64
		remotePeerID string
		WakeCh       <-chan Wake
		EphemeralCh  <-chan []byte
		once         sync.Once
	}
)

const EphemeralBuffer = 64

const (
	PersistDebounce = 50 * time.Millisecond
	PersistTimeout  = 5 * time.Second
	PersistRetryMax = 5 * time.Second
	PresenceTTL     = 30 * time.Second
	RoomOpenTimeout = 20 * time.Second
	// RefreshInterval is how often a room polls for changes another server
	// instance persisted, in case their notification was missed.
	RefreshInterval = 500 * time.Millisecond
	// RefreshTimeout bounds one room refresh, so a slow read cannot keep the
	// room from persisting or catching up.
	RefreshTimeout = 10 * time.Second
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
			h.mu.Unlock()

			if opening != nil {
				select {
				case <-opening:
				case <-ctx.Done():
					return nil, fmt.Errorf("cannot wait for collaboration room to open: %w", ctx.Err())
				}

				continue
			}

			h.mu.Lock()

			current := h.rooms[documentID]
			if current != room || current.opening != nil {
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
			refresh:    make(chan struct{}, 1),
			stop:       make(chan struct{}),
			done:       make(chan struct{}),
			opening:    make(chan struct{}),
		}
		h.rooms[documentID] = room
		h.mu.Unlock()

		openCtx, openCancel := context.WithTimeout(ctx, RoomOpenTimeout)
		opened, err := h.documents.Open(openCtx, scope, documentID)

		openCancel()

		if err != nil {
			h.finishOpening(documentID, room)

			return nil, fmt.Errorf("cannot open collaboration room: %w", err)
		}

		h.mu.Lock()
		room.collaboration = opened
		room.position = opened.Position()

		lease := h.addPeerLocked(documentID, room, connectionID)

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

	// The room catches up once for all its peers, then wakes those whose
	// document changed. A pending request already covers this one.
	select {
	case room.refresh <- struct{}{}:
	default:
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

		if empty && l.hub.rooms[l.documentID] == l.room {
			delete(l.hub.rooms, l.documentID)
		}

		l.hub.mu.Unlock()

		if !empty {
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

func (r *Room) run() {
	defer close(r.done)

	var timer *time.Timer

	dirty := false
	failures := 0

	poll := time.NewTicker(RefreshInterval)
	defer poll.Stop()

	for {
		var timerChannel <-chan time.Time
		if timer != nil {
			timerChannel = timer.C
		}

		select {
		case <-r.refresh:
			r.refreshOnce()
		case <-poll.C:
			r.refreshOnce()
		case <-r.dirty:
			dirty = true

			// A failing persist is already scheduled with a backoff. New
			// edits must not pull that retry forward.
			if failures > 0 {
				continue
			}

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
					failures++
					timer.Reset(persistRetryDelay(failures))

					continue
				}

				failures = 0
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

// persistRetryDelay doubles the wait after each consecutive failure so a
// persist that cannot succeed stops contending for the document row.
func persistRetryDelay(failures int) time.Duration {
	delay := PersistDebounce

	for range failures {
		delay *= 2
		if delay >= PersistRetryMax {
			return PersistRetryMax
		}
	}

	return delay
}

func (r *Room) persist() error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		PersistTimeout,
	)
	defer cancel()

	// The room position is left alone: the next refresh reads the rows this
	// persist appended, which the document already holds, so applying them
	// moves nothing and wakes no peer.
	_, err := r.documents.Persist(
		ctx,
		r.scope,
		r.documentID,
		r.collaboration.Document,
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.persistErr = err

	return err
}

// refreshOnce catches the room document up with changes other writers
// persisted, once for all its peers. The canonical read runs without the
// document lock, so a joining peer's request never waits behind it. The lock
// covers only applying the delta, which reads no database and, for the rows
// after the room's position, replays only those. A failed refresh is retried at
// the next poll.
func (r *Room) refreshOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), RefreshTimeout)
	defer cancel()

	delta, err := r.documents.FetchChanges(ctx, r.scope, r.documentID, r.position)
	if err != nil || delta == nil {
		return
	}

	defer delta.Close()

	r.docMu.Lock()
	changed, err := delta.Apply(r.collaboration.Document)
	r.docMu.Unlock()

	if errors.Is(err, collaboration.ErrRefreshIncomplete) {
		// The change log no longer holds a change the room lacks. The zero
		// position makes the next refresh rebuild from the snapshot.
		r.position = collaboration.Position{}

		return
	}

	if err != nil {
		return
	}

	r.position = delta.Position

	if !changed {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, peer := range r.peers {
		select {
		case peer.wake <- Wake{}:
		default:
		}
	}
}
