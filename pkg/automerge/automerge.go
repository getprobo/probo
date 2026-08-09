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

// Package automerge provides a no-CGO Go API for Automerge documents.
//
// The initial backend embeds the official Rust Automerge engine as a WASI
// module. A native Go engine can implement the private backend contract and be
// checked against this reference implementation without changing callers.
package automerge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.probo.inc/probo/pkg/automerge/internal/native"
	"go.probo.inc/probo/pkg/automerge/internal/reference"
)

type (
	// ActorID identifies one writer in an Automerge history.
	ActorID [16]byte

	// Hash identifies an Automerge change.
	Hash [32]byte

	// Cursor is a stable position in an Automerge sequence.
	Cursor []byte

	// Change is one immutable encoded Automerge change.
	Change struct {
		Hash  Hash
		Bytes []byte
	}

	// Document is a concurrency-safe Automerge document.
	Document struct {
		mu      sync.Mutex
		backend backend
		closed  bool
	}

	// Text is a collaborative UTF-16-indexed text object.
	Text struct {
		document *Document
		handle   uint32
	}

	// SyncState tracks one document's synchronization with one remote peer.
	SyncState struct {
		document *Document
		handle   uint32
		closed   bool
	}

	backend interface {
		Close(context.Context) error
		Save(context.Context) ([]byte, error)
		Stats(context.Context) ([]byte, error)
		CurrentState(context.Context) ([]byte, error)
		Diff(context.Context, [][32]byte, [][32]byte) ([]byte, error)
		UpdateDiffCursor(context.Context) error
		DiffIncremental(context.Context) ([]byte, error)
		SaveIncremental(context.Context) ([]byte, error)
		LoadIncremental(context.Context, []byte) (uint64, error)
		SetActor(context.Context, []byte) error
		PutString(context.Context, uint32, string, string) error
		GetString(context.Context, uint32, string) (string, error)
		PutScalar(context.Context, uint32, string, []byte) error
		GetScalar(context.Context, uint32, string) ([]byte, error)
		GetScalarAtHeads(context.Context, uint32, string, [][32]byte) ([]byte, error)
		GetAllScalars(context.Context, uint32, string) ([]byte, error)
		GetAllScalarsAt(context.Context, uint32, uint64) ([]byte, error)
		PutObject(context.Context, uint32, string, string) (uint32, error)
		GetObject(context.Context, uint32, string) (uint32, string, error)
		InsertObject(context.Context, uint32, uint64, string) (uint32, error)
		PutObjectAt(context.Context, uint32, uint64, string) (uint32, error)
		GetObjectAt(context.Context, uint32, uint64) (uint32, string, error)
		InsertScalar(context.Context, uint32, uint64, []byte) error
		PutScalarAt(context.Context, uint32, uint64, []byte) error
		GetScalarAt(context.Context, uint32, uint64) ([]byte, error)
		DeleteMap(context.Context, uint32, string) error
		DeleteSequence(context.Context, uint32, uint64) error
		Increment(context.Context, uint32, string, int64) error
		IncrementAt(context.Context, uint32, uint64, int64) error
		Keys(context.Context, uint32) ([]string, error)
		Length(context.Context, uint32) (uint64, error)
		PutText(context.Context, uint32, string) (uint32, error)
		GetText(context.Context, uint32, string) (uint32, error)
		SpliceText(context.Context, uint32, uint32, int32, string) error
		UpdateText(context.Context, uint32, string) error
		MarkText(context.Context, uint32, uint32, uint32, string, []byte, string) error
		SplitBlock(context.Context, uint32, uint32) (uint32, error)
		JoinBlock(context.Context, uint32, uint32) error
		ReplaceBlock(context.Context, uint32, uint32) (uint32, error)
		Text(context.Context, uint32) (string, error)
		TextAt(context.Context, uint32, [][32]byte) (string, error)
		TextSpans(context.Context, uint32) ([]byte, error)
		Marks(context.Context, uint32) ([]byte, error)
		MarksAt(context.Context, uint32, [][32]byte) ([]byte, error)
		TextCursor(context.Context, uint32, uint32) ([]byte, error)
		TextCursorMoving(context.Context, uint32, uint32, bool) ([]byte, error)
		TextCursorPosition(context.Context, uint32, []byte) (uint32, error)
		Commit(context.Context, string, time.Time) ([32]byte, error)
		EmptyCommit(context.Context, string, time.Time) ([32]byte, error)
		Rollback(context.Context) (uint64, error)
		Heads(context.Context) ([][32]byte, error)
		HasHeads(context.Context, [][32]byte) (bool, error)
		MissingDependencies(context.Context, [][32]byte) ([][32]byte, error)
		Merge(context.Context, []byte) ([][32]byte, error)
		NewSyncState(context.Context) (uint32, error)
		CloseSyncState(context.Context, uint32) error
		GenerateSyncMessage(context.Context, uint32) ([]byte, bool, error)
		ReceiveSyncMessage(context.Context, uint32, []byte) error
		SetSyncReadOnly(context.Context, uint32, bool) error
		SyncPeerReadOnly(context.Context, uint32) (bool, error)
		SaveSyncState(context.Context, uint32) ([]byte, error)
		LoadSyncState(context.Context, []byte) (uint32, error)
	}

	changeBackend interface {
		ChangesSince(context.Context, [][32]byte) ([][]byte, [][32]byte, error)
	}

	changeApplier interface {
		ApplyChanges(context.Context, [][]byte) error
	}
)

var (
	ErrClosed          = errors.New("automerge document is closed")
	ErrSameDocument    = errors.New("cannot merge an Automerge document into itself")
	ErrSyncStateClosed = errors.New("automerge sync state is closed")

	_ backend = (*reference.Backend)(nil)
	_ backend = (*native.Backend)(nil)
)

const rootObject uint32 = 0

// NewActorID returns a cryptographically random actor ID.
func NewActorID() (ActorID, error) {
	var actorID ActorID
	if _, err := rand.Read(actorID[:]); err != nil {
		return ActorID{}, fmt.Errorf("cannot generate Automerge actor ID: %w", err)
	}

	return actorID, nil
}

// New creates an empty document using the native Go engine.
func New(ctx context.Context, actorID ActorID) (*Document, error) {
	return NewPureGo(ctx, actorID)
}

// NewReference creates an empty document using the official WASM reference
// engine. It is retained as a differential oracle for the native backend.
func NewReference(ctx context.Context, actorID ActorID) (*Document, error) {
	b, err := reference.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge backend: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot initialize Automerge actor: %w", err)
	}

	return &Document{backend: b}, nil
}

// NewPureGo creates an empty document using the experimental native Go engine.
//
// The native engine is intended for differential testing until its complete
// feature surface reaches parity with the reference backend.
func NewPureGo(ctx context.Context, actorID ActorID) (*Document, error) {
	b, err := native.NewBackend(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create native Automerge backend: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot initialize native Automerge actor: %w", err)
	}

	return &Document{backend: b}, nil
}

// Load creates a document using the native Go engine and assigns a new writer.
func Load(ctx context.Context, data []byte, actorID ActorID) (*Document, error) {
	return LoadPureGo(ctx, data, actorID)
}

// LoadReference loads a document using the official WASM reference engine.
func LoadReference(
	ctx context.Context,
	data []byte,
	actorID ActorID,
) (*Document, error) {
	b, err := reference.Load(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge backend: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot assign loaded Automerge actor: %w", err)
	}

	return &Document{backend: b}, nil
}

// LoadPureGo loads Automerge data using the experimental native Go engine.
func LoadPureGo(
	ctx context.Context,
	data []byte,
	actorID ActorID,
) (*Document, error) {
	b, err := native.LoadBackend(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("cannot load native Automerge backend: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot assign native Automerge actor: %w", err)
	}

	return &Document{backend: b}, nil
}

// Close releases the document's WASM module instance.
func (d *Document) Close(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	if err := d.backend.Close(ctx); err != nil {
		return fmt.Errorf("cannot close Automerge document: %w", err)
	}

	return nil
}

// Save serializes the complete Automerge history.
func (d *Document) Save(ctx context.Context) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.backend.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge document: %w", err)
	}

	return data, nil
}

// Stats reports aggregate document statistics.
type Stats struct {
	NumChanges uint64 `json:"numChanges"`
	NumOps     uint64 `json:"numOps"`
	NumActors  uint64 `json:"numActors"`
}

// Stats returns the number of changes, operations, and actors in the document.
func (d *Document) Stats(ctx context.Context) (Stats, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Stats{}, ErrClosed
	}

	data, err := d.backend.Stats(ctx)
	if err != nil {
		return Stats{}, fmt.Errorf("cannot read Automerge stats: %w", err)
	}

	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		return Stats{}, fmt.Errorf("cannot decode Automerge stats: %w", err)
	}

	return stats, nil
}

// Fork creates an independent writer with the same document history.
func (d *Document) Fork(
	ctx context.Context,
	actorID ActorID,
) (*Document, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil, ErrClosed
	}

	data, err := d.backend.Save(ctx)
	if err != nil {
		d.mu.Unlock()
		return nil, fmt.Errorf("cannot save Automerge fork source: %w", err)
	}

	_, referenceBackend := d.backend.(*reference.Backend)
	d.mu.Unlock()

	if referenceBackend {
		return LoadReference(ctx, data, actorID)
	}

	return Load(ctx, data, actorID)
}

// SaveIncremental serializes changes since the previous save operation.
func (d *Document) SaveIncremental(ctx context.Context) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.backend.SaveIncremental(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot save incremental Automerge changes: %w", err)
	}

	return data, nil
}

// LoadIncremental applies incrementally encoded Automerge changes.
func (d *Document) LoadIncremental(ctx context.Context, data []byte) (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return 0, ErrClosed
	}

	applied, err := d.backend.LoadIncremental(ctx, data)
	if err != nil {
		return 0, fmt.Errorf("cannot load incremental Automerge changes: %w", err)
	}

	return applied, nil
}

// PutString assigns a string at a key in the root map.
func (d *Document) PutString(ctx context.Context, key, value string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.backend.PutString(ctx, rootObject, key, value); err != nil {
		return fmt.Errorf("cannot put Automerge string: %w", err)
	}

	return nil
}

// String returns a string value from a key in the root map.
func (d *Document) String(ctx context.Context, key string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return "", ErrClosed
	}

	value, err := d.backend.GetString(ctx, rootObject, key)
	if err != nil {
		return "", fmt.Errorf("cannot get Automerge string: %w", err)
	}

	return value, nil
}

// CreateText creates collaborative text at a key in the root map.
func (d *Document) CreateText(ctx context.Context, key string) (*Text, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.backend.PutText(ctx, rootObject, key)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge text: %w", err)
	}

	return &Text{document: d, handle: handle}, nil
}

// Text returns existing collaborative text from the root map.
func (d *Document) Text(ctx context.Context, key string) (*Text, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.backend.GetText(ctx, rootObject, key)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge text: %w", err)
	}

	return &Text{document: d, handle: handle}, nil
}

// Commit records pending operations as one Automerge change.
func (d *Document) Commit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) (Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Hash{}, ErrClosed
	}

	hash, err := d.backend.Commit(ctx, message, timestamp)
	if err != nil {
		return Hash{}, fmt.Errorf("cannot commit Automerge document: %w", err)
	}

	return Hash(hash), nil
}

// CommitNow records pending operations using the current Unix timestamp.
func (d *Document) CommitNow(ctx context.Context, message string) (Hash, error) {
	return d.Commit(ctx, message, time.Now())
}

// EmptyCommit records a change without document operations.
func (d *Document) EmptyCommit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) (Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Hash{}, ErrClosed
	}

	hash, err := d.backend.EmptyCommit(ctx, message, timestamp)
	if err != nil {
		return Hash{}, fmt.Errorf("cannot commit empty Automerge change: %w", err)
	}

	return Hash(hash), nil
}

// EmptyCommitNow records an empty change using the current Unix timestamp.
func (d *Document) EmptyCommitNow(ctx context.Context, message string) (Hash, error) {
	return d.EmptyCommit(ctx, message, time.Now())
}

// Rollback discards every operation pending in the current change.
func (d *Document) Rollback(ctx context.Context) (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return 0, ErrClosed
	}

	cancelled, err := d.backend.Rollback(ctx)
	if err != nil {
		return 0, fmt.Errorf("cannot roll back Automerge document: %w", err)
	}

	return cancelled, nil
}

// Heads returns the hashes at the document's current frontier.
func (d *Document) Heads(ctx context.Context) ([]Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	backendHeads, err := d.backend.Heads(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge heads: %w", err)
	}

	heads := make([]Hash, len(backendHeads))
	for i := range backendHeads {
		heads[i] = Hash(backendHeads[i])
	}

	return heads, nil
}

// HasHeads reports whether every hash exists in the document history.
func (d *Document) HasHeads(ctx context.Context, heads []Hash) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return false, ErrClosed
	}

	hasHeads, err := d.backend.HasHeads(ctx, backendHashes(heads))
	if err != nil {
		return false, fmt.Errorf("cannot inspect Automerge heads: %w", err)
	}

	return hasHeads, nil
}

// MissingDependencies returns unknown hashes required to reach heads.
func (d *Document) MissingDependencies(
	ctx context.Context,
	heads []Hash,
) ([]Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	missing, err := d.backend.MissingDependencies(
		ctx,
		backendHashes(heads),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge missing dependencies: %w", err)
	}

	result := make([]Hash, len(missing))
	for i, hash := range missing {
		result[i] = Hash(hash)
	}

	return result, nil
}

// ChangesSince returns encoded changes not covered by heads.
func (d *Document) ChangesSince(
	ctx context.Context,
	heads []Hash,
) ([]Change, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	changeSource, ok := d.backend.(changeBackend)
	if !ok {
		return nil, fmt.Errorf("automerge backend does not expose incremental changes")
	}

	backendHeads := make([][32]byte, len(heads))
	for i, head := range heads {
		backendHeads[i] = [32]byte(head)
	}

	raw, hashes, err := changeSource.ChangesSince(ctx, backendHeads)
	if err != nil {
		return nil, fmt.Errorf("cannot get incremental Automerge changes: %w", err)
	}

	if len(raw) != len(hashes) {
		return nil, fmt.Errorf(
			"cannot get incremental Automerge changes: %d changes for %d hashes",
			len(raw),
			len(hashes),
		)
	}

	changes := make([]Change, len(raw))
	for i := range raw {
		changes[i] = Change{
			Hash:  Hash(hashes[i]),
			Bytes: append([]byte(nil), raw[i]...),
		}
	}

	return changes, nil
}

// ApplyChanges applies encoded changes whose dependencies may already exist in
// the document.
func (d *Document) ApplyChanges(
	ctx context.Context,
	changes [][]byte,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	applier, ok := d.backend.(changeApplier)
	if !ok {
		return fmt.Errorf("automerge backend does not accept incremental changes")
	}

	if err := applier.ApplyChanges(ctx, changes); err != nil {
		return fmt.Errorf("cannot apply incremental Automerge changes: %w", err)
	}

	return nil
}

// Merge applies all changes from another document.
func (d *Document) Merge(ctx context.Context, other *Document) ([]Hash, error) {
	if d == other {
		return nil, ErrSameDocument
	}

	otherData, err := other.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge merge source: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	backendHeads, err := d.backend.Merge(ctx, otherData)
	if err != nil {
		return nil, fmt.Errorf("cannot merge Automerge document: %w", err)
	}

	heads := make([]Hash, len(backendHeads))
	for i := range backendHeads {
		heads[i] = Hash(backendHeads[i])
	}

	return heads, nil
}

// NewSyncState starts synchronization with a remote peer.
func (d *Document) NewSyncState(ctx context.Context) (*SyncState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.backend.NewSyncState(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge sync state: %w", err)
	}

	return &SyncState{document: d, handle: handle}, nil
}

// LoadSyncState resumes a previously serialized remote-peer session.
func (d *Document) LoadSyncState(ctx context.Context, data []byte) (*SyncState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.backend.LoadSyncState(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge sync state: %w", err)
	}

	return &SyncState{document: d, handle: handle}, nil
}

// Splice replaces deleteCount UTF-16 code units at index with value.
func (t *Text) Splice(ctx context.Context, index uint32, deleteCount int32, value string) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if err := t.document.backend.SpliceText(ctx, t.handle, index, deleteCount, value); err != nil {
		return fmt.Errorf("cannot splice Automerge text: %w", err)
	}

	return nil
}

// Update replaces the text content with value using a minimal splice so that
// concurrent edits to unaffected regions merge cleanly. It mirrors the Rust
// AutoCommit::update_text and JavaScript updateText helpers.
func (t *Text) Update(ctx context.Context, value string) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if err := t.document.backend.UpdateText(ctx, t.handle, value); err != nil {
		return fmt.Errorf("cannot update Automerge text: %w", err)
	}

	return nil
}

// String returns the current materialized text.
func (t *Text) String(ctx context.Context) (string, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return "", ErrClosed
	}

	value, err := t.document.backend.Text(ctx, t.handle)
	if err != nil {
		return "", fmt.Errorf("cannot read Automerge text: %w", err)
	}

	return value, nil
}

// StringAt returns text at a historical causal frontier.
func (t *Text) StringAt(ctx context.Context, heads []Hash) (string, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return "", ErrClosed
	}

	value, err := t.document.backend.TextAt(
		ctx,
		t.handle,
		backendHashes(heads),
	)
	if err != nil {
		return "", fmt.Errorf("cannot read historical Automerge text: %w", err)
	}

	return value, nil
}

// Cursor returns a stable address for the UTF-16 position at index.
func (t *Text) Cursor(ctx context.Context, index uint32) (Cursor, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	cursor, err := t.document.backend.TextCursor(ctx, t.handle, index)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge text cursor: %w", err)
	}

	return Cursor(cursor), nil
}

// CursorPosition resolves a stable cursor in the current document.
func (t *Text) CursorPosition(ctx context.Context, cursor Cursor) (uint32, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return 0, ErrClosed
	}

	position, err := t.document.backend.TextCursorPosition(ctx, t.handle, cursor)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve Automerge text cursor: %w", err)
	}

	return position, nil
}

// Close releases the peer-specific synchronization state.
func (s *SyncState) Close(ctx context.Context) error {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	if s.document.closed {
		return nil
	}

	if err := s.document.backend.CloseSyncState(ctx, s.handle); err != nil {
		return fmt.Errorf("cannot close Automerge sync state: %w", err)
	}

	return nil
}

// GenerateMessage returns the next message for the remote peer.
func (s *SyncState) GenerateMessage(ctx context.Context) ([]byte, bool, error) {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return nil, false, ErrClosed
	}

	if s.closed {
		return nil, false, ErrSyncStateClosed
	}

	message, ok, err := s.document.backend.GenerateSyncMessage(ctx, s.handle)
	if err != nil {
		return nil, false, fmt.Errorf("cannot generate Automerge sync message: %w", err)
	}

	return message, ok, nil
}

// ReceiveMessage applies a message received from the remote peer.
func (s *SyncState) ReceiveMessage(ctx context.Context, message []byte) error {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return ErrClosed
	}

	if s.closed {
		return ErrSyncStateClosed
	}

	if err := s.document.backend.ReceiveSyncMessage(ctx, s.handle, message); err != nil {
		return fmt.Errorf("cannot receive Automerge sync message: %w", err)
	}

	return nil
}

// SetReadOnly controls whether incoming changes are applied by this peer.
func (s *SyncState) SetReadOnly(ctx context.Context, readOnly bool) error {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return ErrClosed
	}

	if s.closed {
		return ErrSyncStateClosed
	}

	if err := s.document.backend.SetSyncReadOnly(
		ctx,
		s.handle,
		readOnly,
	); err != nil {
		return fmt.Errorf("cannot set Automerge sync read-only mode: %w", err)
	}

	return nil
}

// PeerReadOnly reports whether the remote peer advertised read-only mode.
func (s *SyncState) PeerReadOnly(ctx context.Context) (bool, error) {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return false, ErrClosed
	}

	if s.closed {
		return false, ErrSyncStateClosed
	}

	readOnly, err := s.document.backend.SyncPeerReadOnly(ctx, s.handle)
	if err != nil {
		return false, fmt.Errorf("cannot get Automerge peer read-only mode: %w", err)
	}

	return readOnly, nil
}

// Save serializes the peer-specific synchronization state.
func (s *SyncState) Save(ctx context.Context) ([]byte, error) {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return nil, ErrClosed
	}

	if s.closed {
		return nil, ErrSyncStateClosed
	}

	data, err := s.document.backend.SaveSyncState(ctx, s.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge sync state: %w", err)
	}

	return data, nil
}

// String returns the lowercase hexadecimal change hash.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
