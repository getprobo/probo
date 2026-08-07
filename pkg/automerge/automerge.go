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
		SetActor(context.Context, []byte) error
		PutString(context.Context, uint32, string, string) error
		PutText(context.Context, uint32, string) (uint32, error)
		GetText(context.Context, uint32, string) (uint32, error)
		SpliceText(context.Context, uint32, uint32, int32, string) error
		Text(context.Context, uint32) (string, error)
		TextSpans(context.Context, uint32) ([]byte, error)
		TextCursor(context.Context, uint32, uint32) ([]byte, error)
		TextCursorPosition(context.Context, uint32, []byte) (uint32, error)
		Commit(context.Context, string, time.Time) ([32]byte, error)
		Heads(context.Context) ([][32]byte, error)
		Merge(context.Context, []byte) ([][32]byte, error)
		NewSyncState(context.Context) (uint32, error)
		CloseSyncState(context.Context, uint32) error
		GenerateSyncMessage(context.Context, uint32) ([]byte, bool, error)
		ReceiveSyncMessage(context.Context, uint32, []byte) error
		SaveSyncState(context.Context, uint32) ([]byte, error)
		LoadSyncState(context.Context, []byte) (uint32, error)
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
