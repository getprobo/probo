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

	"go.probo.inc/probo/pkg/automerge/internal/reference"
)

type (
	// ActorID identifies one writer in an Automerge history.
	ActorID [16]byte

	// Hash identifies an Automerge change.
	Hash [32]byte

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

	backend interface {
		Close(context.Context) error
		Save(context.Context) ([]byte, error)
		SetActor(context.Context, []byte) error
		PutString(context.Context, uint32, string, string) error
		PutText(context.Context, uint32, string) (uint32, error)
		GetText(context.Context, uint32, string) (uint32, error)
		SpliceText(context.Context, uint32, uint32, int32, string) error
		Text(context.Context, uint32) (string, error)
		Commit(context.Context, string, time.Time) ([32]byte, error)
		Heads(context.Context) ([][32]byte, error)
		Merge(context.Context, []byte) ([][32]byte, error)
	}
)

var (
	ErrClosed       = errors.New("Automerge document is closed")
	ErrSameDocument = errors.New("cannot merge an Automerge document into itself")

	_ backend = (*reference.Backend)(nil)
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

// New creates an empty document using the reference backend.
func New(ctx context.Context, actorID ActorID) (*Document, error) {
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

// Load creates a document from Automerge binary data and assigns a new writer.
func Load(ctx context.Context, data []byte, actorID ActorID) (*Document, error) {
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

// String returns the lowercase hexadecimal change hash.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
