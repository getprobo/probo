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
// The initial engine embeds the official Rust Automerge engine as a WASI
// module. A native Go engine can implement the private engine contract and be
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
		mu     sync.Mutex
		engine engine
		closed bool
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

	engine interface {
		Close(context.Context) error
		Save(context.Context) ([]byte, error)
		SaveWithOptions(context.Context, bool) ([]byte, error)
		SaveNoCompress(context.Context) ([]byte, error)
		Isolate(context.Context, [][32]byte) error
		Integrate(context.Context) error
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
		UpdateSpans(context.Context, uint32, []byte, []byte) error
		MarkText(context.Context, uint32, uint32, uint32, string, []byte, string) error
		SplitBlock(context.Context, uint32, uint32) (uint32, error)
		JoinBlock(context.Context, uint32, uint32) error
		ReplaceBlock(context.Context, uint32, uint32) (uint32, error)
		Text(context.Context, uint32) (string, error)
		TextAt(context.Context, uint32, [][32]byte) (string, error)
		TextSpans(context.Context, uint32) ([]byte, error)
		TextSpansAt(context.Context, uint32, [][32]byte) ([]byte, error)
		Marks(context.Context, uint32) ([]byte, error)
		MarksAt(context.Context, uint32, [][32]byte) ([]byte, error)
		TextCursor(context.Context, uint32, uint32) ([]byte, error)
		TextCursorMoving(context.Context, uint32, uint32, bool) ([]byte, error)
		TextCursorMovingAt(context.Context, uint32, uint32, bool, [][32]byte) ([]byte, error)
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

	documentSaver interface {
		SaveDocument(context.Context) ([]byte, error)
	}
)

var (
	ErrClosed          = errors.New("automerge document is closed")
	ErrSameDocument    = errors.New("cannot merge an Automerge document into itself")
	ErrSyncStateClosed = errors.New("automerge sync state is closed")

	_ engine = (*reference.Engine)(nil)
	_ engine = (*native.Engine)(nil)
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
// engine. It is retained as a differential oracle for the native engine.
func NewReference(ctx context.Context, actorID ActorID) (*Document, error) {
	b, err := reference.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge engine: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot initialize Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// NewPureGo creates an empty document using the experimental native Go engine.
//
// The native engine is intended for differential testing until its complete
// feature surface reaches parity with the reference engine.
func NewPureGo(ctx context.Context, actorID ActorID) (*Document, error) {
	b, err := native.NewEngine(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create native Automerge engine: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot initialize native Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
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
		return nil, fmt.Errorf("cannot load Automerge engine: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot assign loaded Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// LoadConvertingStrings loads a document with the native engine, converting
// every string scalar stored in a map or list into a text object. It mirrors
// the Rust StringMigration::ConvertToText load option.
func LoadConvertingStrings(
	ctx context.Context,
	data []byte,
	actorID ActorID,
) (*Document, error) {
	document, err := LoadPureGo(ctx, data, actorID)
	if err != nil {
		return nil, err
	}

	if err := document.convertStringsToText(ctx); err != nil {
		_ = document.Close(ctx)

		return nil, err
	}

	return document, nil
}

// LoadReferenceConvertingStrings loads a document with the reference engine and
// the string-to-text migration applied.
func LoadReferenceConvertingStrings(
	ctx context.Context,
	data []byte,
	actorID ActorID,
) (*Document, error) {
	b, err := reference.LoadConvertingStrings(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge engine: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)

		return nil, fmt.Errorf("cannot assign loaded Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// convertStringsToText replaces every string scalar reachable from the root in
// a map or list with a text object holding that string, then commits the
// conversion when anything changed. It backs the string-to-text load migration.
func (d *Document) convertStringsToText(ctx context.Context) error {
	changed, err := convertObjectStrings(ctx, d.Root())
	if err != nil {
		return fmt.Errorf("cannot migrate strings to text: %w", err)
	}

	if !changed {
		return nil
	}

	if _, err := d.Commit(ctx, "convert strings to text", time.Unix(0, 0)); err != nil {
		return fmt.Errorf("cannot commit string migration: %w", err)
	}

	return nil
}

func convertObjectStrings(ctx context.Context, object *Object) (bool, error) {
	switch object.Type {
	case ObjectTypeMap, ObjectTypeTable:
		return convertMapStrings(ctx, object)
	case ObjectTypeList:
		return convertListStrings(ctx, object)
	default:
		return false, nil
	}
}

func convertMapStrings(ctx context.Context, object *Object) (bool, error) {
	keys, err := object.Keys(ctx)
	if err != nil {
		return false, err
	}

	changed := false

	for _, key := range keys {
		if scalar, err := object.Scalar(ctx, key); err == nil {
			if scalar.Type != ScalarTypeString {
				continue
			}

			text, err := object.CreateObject(ctx, key, ObjectTypeText)
			if err != nil {
				return false, err
			}

			handle, err := text.Text()
			if err != nil {
				return false, err
			}

			if err := handle.Splice(ctx, 0, 0, scalar.String); err != nil {
				return false, err
			}

			changed = true

			continue
		}

		child, err := object.Object(ctx, key)
		if err != nil {
			continue
		}

		childChanged, err := convertObjectStrings(ctx, child)
		if err != nil {
			return false, err
		}

		changed = changed || childChanged
	}

	return changed, nil
}

func convertListStrings(ctx context.Context, object *Object) (bool, error) {
	length, err := object.Len(ctx)
	if err != nil {
		return false, err
	}

	changed := false

	for index := range length {
		if scalar, err := object.ScalarAt(ctx, index); err == nil {
			if scalar.Type != ScalarTypeString {
				continue
			}

			text, err := object.PutObjectAt(ctx, index, ObjectTypeText)
			if err != nil {
				return false, err
			}

			handle, err := text.Text()
			if err != nil {
				return false, err
			}

			if err := handle.Splice(ctx, 0, 0, scalar.String); err != nil {
				return false, err
			}

			changed = true

			continue
		}

		child, err := object.ObjectAt(ctx, index)
		if err != nil {
			continue
		}

		childChanged, err := convertObjectStrings(ctx, child)
		if err != nil {
			return false, err
		}

		changed = changed || childChanged
	}

	return changed, nil
}

// LoadPureGo loads Automerge data using the experimental native Go engine.
func LoadPureGo(
	ctx context.Context,
	data []byte,
	actorID ActorID,
) (*Document, error) {
	b, err := native.LoadEngine(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("cannot load native Automerge engine: %w", err)
	}

	if err := b.SetActor(ctx, actorID[:]); err != nil {
		_ = b.Close(ctx)
		return nil, fmt.Errorf("cannot assign native Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// Close releases the document's WASM module instance.
func (d *Document) Close(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	if err := d.engine.Close(ctx); err != nil {
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

	data, err := d.engine.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge document: %w", err)
	}

	return data, nil
}

// SaveDocument serializes the history as a single compacted document chunk,
// the form save() produces in the Rust and JavaScript implementations. Save
// writes the history as a stream of changes instead, which stays byte-for-byte
// faithful to what was loaded but grows with every commit.
func (d *Document) SaveDocument(ctx context.Context) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	compactor, ok := d.engine.(documentSaver)
	if !ok {
		// The reference engine only ever writes compacted documents.
		return d.engine.Save(ctx)
	}

	data, err := compactor.SaveDocument(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot save compacted Automerge document: %w", err)
	}

	return data, nil
}

// SaveWithOptions serializes the document, choosing whether to retain orphan
// changes (changes whose dependencies are missing). Retaining them, the default
// for Save, preserves them across a save/load round trip so they can be resolved
// once their dependencies arrive; discarding them drops them permanently. It
// mirrors the Rust SaveOptions.retain_orphans flag.
func (d *Document) SaveWithOptions(
	ctx context.Context,
	retainOrphans bool,
) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.engine.SaveWithOptions(ctx, retainOrphans)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge document: %w", err)
	}

	return data, nil
}

// Isolate pins the document to the given heads so that subsequent reads reflect
// that frontier plus writes made while isolated, and new changes branch from it.
// Isolated changes still accumulate in the full history and become visible after
// Integrate. It mirrors the Rust AutoCommit::isolate API.
func (d *Document) Isolate(ctx context.Context, heads []Hash) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.engine.Isolate(ctx, engineHashes(heads)); err != nil {
		return fmt.Errorf("cannot isolate Automerge document: %w", err)
	}

	return nil
}

// Integrate ends isolation, returning reads and writes to the full history that
// includes every isolated and merged change. It mirrors AutoCommit::integrate.
func (d *Document) Integrate(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.engine.Integrate(ctx); err != nil {
		return fmt.Errorf("cannot integrate Automerge document: %w", err)
	}

	return nil
}

// SaveNoCompress serializes the document without DEFLATE-compressing its change
// data. The default Save compresses large change chunks; this variant is useful
// for comparing compressed and uncompressed sizes. It mirrors the Rust
// AutoCommit::save_nocompress API.
func (d *Document) SaveNoCompress(ctx context.Context) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.engine.SaveNoCompress(ctx)
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

	data, err := d.engine.Stats(ctx)
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

	data, err := d.engine.Save(ctx)
	if err != nil {
		d.mu.Unlock()
		return nil, fmt.Errorf("cannot save Automerge fork source: %w", err)
	}

	_, referenceBackend := d.engine.(*reference.Engine)
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

	data, err := d.engine.SaveIncremental(ctx)
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

	applied, err := d.engine.LoadIncremental(ctx, data)
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

	if err := d.engine.PutString(ctx, rootObject, key, value); err != nil {
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

	value, err := d.engine.GetString(ctx, rootObject, key)
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

	handle, err := d.engine.PutText(ctx, rootObject, key)
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

	handle, err := d.engine.GetText(ctx, rootObject, key)
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

	hash, err := d.engine.Commit(ctx, message, timestamp)
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

	hash, err := d.engine.EmptyCommit(ctx, message, timestamp)
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

	cancelled, err := d.engine.Rollback(ctx)
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

	engineHeads, err := d.engine.Heads(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge heads: %w", err)
	}

	heads := make([]Hash, len(engineHeads))
	for i := range engineHeads {
		heads[i] = Hash(engineHeads[i])
	}

	return heads, nil
}

// ReferenceBloomContains reports whether a sync Bloom filter built from the
// seed change hashes (possibly falsely) contains the target hash. It is only
// available on reference (WASM) documents and exists so parity tests can
// reproduce the upstream Bloom false-positive search deterministically; the
// native engine's V2 sync uses exact head comparison rather than Bloom filters,
// so native documents return an error.
func (d *Document) ReferenceBloomContains(
	ctx context.Context,
	seeds []Hash,
	target Hash,
) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return false, ErrClosed
	}

	oracle, ok := d.engine.(*reference.Engine)
	if !ok {
		return false, fmt.Errorf(
			"bloom filter membership is only available on reference documents",
		)
	}

	seedArrays := make([][32]byte, len(seeds))
	for i := range seeds {
		seedArrays[i] = [32]byte(seeds[i])
	}

	return oracle.BloomContains(ctx, [32]byte(target), seedArrays)
}

// HasHeads reports whether every hash exists in the document history.
func (d *Document) HasHeads(ctx context.Context, heads []Hash) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return false, ErrClosed
	}

	hasHeads, err := d.engine.HasHeads(ctx, engineHashes(heads))
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

	missing, err := d.engine.MissingDependencies(
		ctx,
		engineHashes(heads),
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

	changeSource, ok := d.engine.(changeBackend)
	if !ok {
		return nil, fmt.Errorf("automerge engine does not expose incremental changes")
	}

	engineHeads := make([][32]byte, len(heads))
	for i, head := range heads {
		engineHeads[i] = [32]byte(head)
	}

	raw, hashes, err := changeSource.ChangesSince(ctx, engineHeads)
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

	applier, ok := d.engine.(changeApplier)
	if !ok {
		return fmt.Errorf("automerge engine does not accept incremental changes")
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

	engineHeads, err := d.engine.Merge(ctx, otherData)
	if err != nil {
		return nil, fmt.Errorf("cannot merge Automerge document: %w", err)
	}

	heads := make([]Hash, len(engineHeads))
	for i := range engineHeads {
		heads[i] = Hash(engineHeads[i])
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

	handle, err := d.engine.NewSyncState(ctx)
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

	handle, err := d.engine.LoadSyncState(ctx, data)
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

	if err := t.document.engine.SpliceText(ctx, t.handle, index, deleteCount, value); err != nil {
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

	if err := t.document.engine.UpdateText(ctx, t.handle, value); err != nil {
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

	value, err := t.document.engine.Text(ctx, t.handle)
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

	value, err := t.document.engine.TextAt(
		ctx,
		t.handle,
		engineHashes(heads),
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

	cursor, err := t.document.engine.TextCursor(ctx, t.handle, index)
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

	position, err := t.document.engine.TextCursorPosition(ctx, t.handle, cursor)
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

	if err := s.document.engine.CloseSyncState(ctx, s.handle); err != nil {
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

	message, ok, err := s.document.engine.GenerateSyncMessage(ctx, s.handle)
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

	if err := s.document.engine.ReceiveSyncMessage(ctx, s.handle, message); err != nil {
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

	if err := s.document.engine.SetSyncReadOnly(
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

	readOnly, err := s.document.engine.SyncPeerReadOnly(ctx, s.handle)
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

	data, err := s.document.engine.SaveSyncState(ctx, s.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge sync state: %w", err)
	}

	return data, nil
}

// String returns the lowercase hexadecimal change hash.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
