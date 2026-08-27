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

// Package testsupport provides native/reference Automerge differential helpers.
//
// It mirrors the public document API so parity tests can execute the same
// scenario against the Go core and the isolated Rust/WASM oracle.
package testsupport

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.probo.inc/probo/pkg/automerge/internal/core"
	"go.probo.inc/probo/pkg/automerge/internal/testsupport/reference"
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
		Close() error
		Save(bool, bool) ([]byte, error)
		Isolate([][32]byte) error
		Integrate() error
		Stats() ([]byte, error)
		CurrentState() ([]byte, error)
		Diff([][32]byte, [][32]byte) ([]byte, error)
		UpdateDiffCursor() error
		DiffIncremental() ([]byte, error)
		SaveIncremental() ([]byte, error)
		LoadIncremental([]byte) (uint64, error)
		SetActor([]byte) error
		PutString(uint32, string, string) error
		GetString(uint32, string) (string, error)
		PutScalar(uint32, string, []byte) error
		GetScalar(uint32, string) ([]byte, error)
		GetScalarAtHeads(uint32, string, [][32]byte) ([]byte, error)
		GetAllScalars(uint32, string) ([]byte, error)
		GetAllScalarsAt(uint32, uint64) ([]byte, error)
		PutObject(uint32, string, string) (uint32, error)
		GetObject(uint32, string) (uint32, string, error)
		InsertObject(uint32, uint64, string) (uint32, error)
		PutObjectAt(uint32, uint64, string) (uint32, error)
		GetObjectAt(uint32, uint64) (uint32, string, error)
		InsertScalar(uint32, uint64, []byte) error
		PutScalarAt(uint32, uint64, []byte) error
		GetScalarAt(uint32, uint64) ([]byte, error)
		DeleteMap(uint32, string) error
		DeleteSequence(uint32, uint64) error
		Increment(uint32, string, int64) error
		IncrementAt(uint32, uint64, int64) error
		Keys(uint32) ([]string, error)
		Length(uint32) (uint64, error)
		PutText(uint32, string) (uint32, error)
		GetText(uint32, string) (uint32, error)
		SpliceText(uint32, uint32, int32, string) error
		UpdateText(uint32, string) error
		UpdateSpans(uint32, []byte, []byte) error
		MarkText(uint32, uint32, uint32, string, []byte, string) error
		SplitBlock(uint32, uint32) (uint32, error)
		JoinBlock(uint32, uint32) error
		ReplaceBlock(uint32, uint32) (uint32, error)
		Text(uint32) (string, error)
		TextAt(uint32, [][32]byte) (string, error)
		TextSpans(uint32) ([]byte, error)
		TextSpansAt(uint32, [][32]byte) ([]byte, error)
		Marks(uint32) ([]byte, error)
		MarksAt(uint32, [][32]byte) ([]byte, error)
		TextCursor(uint32, uint32) ([]byte, error)
		TextCursorMoving(uint32, uint32, bool) ([]byte, error)
		TextCursorMovingAt(uint32, uint32, bool, [][32]byte) ([]byte, error)
		TextCursorPosition(uint32, []byte) (uint32, error)
		Commit(string, time.Time) ([32]byte, error)
		EmptyCommit(string, time.Time) ([32]byte, error)
		Rollback() (uint64, error)
		Heads() ([][32]byte, error)
		HasHeads([][32]byte) (bool, error)
		MissingDependencies([][32]byte) ([][32]byte, error)
		Merge([]byte) ([][32]byte, error)
		NewSyncState() (uint32, error)
		CloseSyncState(uint32) error
		GenerateSyncMessage(uint32) ([]byte, bool, error)
		ReceiveSyncMessage(uint32, []byte) error
		SetSyncReadOnly(uint32, bool) error
		SyncPeerReadOnly(uint32) (bool, error)
		SaveSyncState(uint32) ([]byte, error)
		LoadSyncState([]byte) (uint32, error)
	}

	changeBackend interface {
		ChangesSince([][32]byte) ([][]byte, [][32]byte, error)
	}

	changeApplier interface {
		ApplyChanges([][]byte) error
	}
)

var (
	ErrClosed          = errors.New("automerge document is closed")
	ErrNilDocument     = errors.New("cannot merge a nil Automerge document")
	ErrSameDocument    = errors.New("cannot merge an Automerge document into itself")
	ErrSyncStateClosed = errors.New("automerge sync state is closed")

	_ engine = (*reference.Engine)(nil)
	_ engine = (*core.Engine)(nil)
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
func New(actorID ActorID) (*Document, error) {
	b, err := core.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("cannot create native Automerge engine: %w", err)
	}

	if err := b.SetActor(actorID[:]); err != nil {
		_ = b.Close()
		return nil, fmt.Errorf("cannot initialize native Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// NewReference creates an empty document using the official WASM reference
// engine. It exists as a differential oracle for the native engine and is
// intended for tests, not production use.
func NewReference(actorID ActorID) (*Document, error) {
	b, err := reference.New()
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge engine: %w", err)
	}

	if err := b.SetActor(actorID[:]); err != nil {
		_ = b.Close()
		return nil, fmt.Errorf("cannot initialize Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// LoadOption configures how Load interprets stored data.
type LoadOption func(*loadConfig)

type loadConfig struct {
	convertStringsToText bool
}

// ConvertStringsToText converts every string scalar stored in a map or list
// into a text object as the document loads, mirroring Rust's
// StringMigration::ConvertToText load option.
func ConvertStringsToText() LoadOption {
	return func(c *loadConfig) { c.convertStringsToText = true }
}

// Load creates a document from stored data using the native Go engine and
// assigns a new writer.
func Load(
	data []byte,
	actorID ActorID,
	options ...LoadOption,
) (*Document, error) {
	config := loadConfig{}
	for _, option := range options {
		option(&config)
	}

	b, err := core.LoadEngine(data)
	if err != nil {
		return nil, fmt.Errorf("cannot load native Automerge engine: %w", err)
	}

	if err := b.SetActor(actorID[:]); err != nil {
		_ = b.Close()
		return nil, fmt.Errorf("cannot assign native Automerge actor: %w", err)
	}

	document := &Document{engine: b}

	if config.convertStringsToText {
		if err := document.convertStringsToText(); err != nil {
			_ = document.Close()

			return nil, err
		}
	}

	return document, nil
}

// LoadReference loads a document using the official WASM reference engine. Like
// NewReference it is intended for tests, not production use.
func LoadReference(
	data []byte,
	actorID ActorID,
	options ...LoadOption,
) (*Document, error) {
	config := loadConfig{}
	for _, option := range options {
		option(&config)
	}

	load := reference.Load
	if config.convertStringsToText {
		// The reference applies the migration during load through its own WASM
		// entry point rather than as a post-load pass.
		load = reference.LoadConvertingStrings
	}

	b, err := load(data)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge engine: %w", err)
	}

	if err := b.SetActor(actorID[:]); err != nil {
		_ = b.Close()
		return nil, fmt.Errorf("cannot assign loaded Automerge actor: %w", err)
	}

	return &Document{engine: b}, nil
}

// convertStringsToText replaces every string scalar reachable from the root in
// a map or list with a text object holding that string, then commits the
// conversion when anything changed. It backs the string-to-text load migration.
func (d *Document) convertStringsToText() error {
	changed, err := convertObjectStrings(d.Root())
	if err != nil {
		return fmt.Errorf("cannot migrate strings to text: %w", err)
	}

	if !changed {
		return nil
	}

	if _, err := d.Commit("convert strings to text", time.Unix(0, 0)); err != nil {
		return fmt.Errorf("cannot commit string migration: %w", err)
	}

	return nil
}

func convertObjectStrings(object *Object) (bool, error) {
	switch object.Type {
	case ObjectTypeMap, ObjectTypeTable:
		return convertMapStrings(object)
	case ObjectTypeList:
		return convertListStrings(object)
	default:
		return false, nil
	}
}

func convertMapStrings(object *Object) (bool, error) {
	keys, err := object.Keys()
	if err != nil {
		return false, err
	}

	changed := false

	for _, key := range keys {
		if scalar, err := object.Scalar(key); err == nil {
			if scalar.Type != ScalarTypeString {
				continue
			}

			text, err := object.CreateObject(key, ObjectTypeText)
			if err != nil {
				return false, err
			}

			handle, err := text.Text()
			if err != nil {
				return false, err
			}

			if err := handle.Splice(0, 0, scalar.String); err != nil {
				return false, err
			}

			changed = true

			continue
		}

		child, err := object.Object(key)
		if err != nil {
			continue
		}

		childChanged, err := convertObjectStrings(child)
		if err != nil {
			return false, err
		}

		changed = changed || childChanged
	}

	return changed, nil
}

func convertListStrings(object *Object) (bool, error) {
	length, err := object.Len()
	if err != nil {
		return false, err
	}

	changed := false

	for index := range length {
		if scalar, err := object.ScalarAt(index); err == nil {
			if scalar.Type != ScalarTypeString {
				continue
			}

			text, err := object.PutObjectAt(index, ObjectTypeText)
			if err != nil {
				return false, err
			}

			handle, err := text.Text()
			if err != nil {
				return false, err
			}

			if err := handle.Splice(0, 0, scalar.String); err != nil {
				return false, err
			}

			changed = true

			continue
		}

		child, err := object.ObjectAt(index)
		if err != nil {
			continue
		}

		childChanged, err := convertObjectStrings(child)
		if err != nil {
			return false, err
		}

		changed = changed || childChanged
	}

	return changed, nil
}

// Close releases the engine resources held by the document.
func (d *Document) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	if err := d.engine.Close(); err != nil {
		return fmt.Errorf("cannot close Automerge document: %w", err)
	}

	return nil
}

// SaveOption configures how Save serializes a document.
type SaveOption func(*saveConfig)

type saveConfig struct {
	retainOrphans bool
	compress      bool
}

// NoCompress disables DEFLATE compression of the saved document. The default is
// to compress, which the reference's save_nocompress also opts out of; the
// uncompressed form is mainly useful for comparing sizes or debugging.
func NoCompress() SaveOption {
	return func(c *saveConfig) { c.compress = false }
}

// DiscardOrphans drops orphan changes (changes whose dependencies are missing)
// instead of retaining them. Retaining them, the default, preserves them across
// a save/load round trip so they resolve once their dependencies arrive;
// discarding drops them permanently. It mirrors Rust's SaveOptions.retain_orphans.
func DiscardOrphans() SaveOption {
	return func(c *saveConfig) { c.retainOrphans = false }
}

// Save serializes the complete Automerge history as a compacted document. By
// default it compresses and retains orphan changes; pass NoCompress or
// DiscardOrphans to change that.
func (d *Document) Save(options ...SaveOption) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	config := saveConfig{retainOrphans: true, compress: true}
	for _, option := range options {
		option(&config)
	}

	data, err := d.engine.Save(config.retainOrphans, config.compress)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge document: %w", err)
	}

	return data, nil
}

// Isolate pins the document to the given heads so that subsequent reads reflect
// that frontier plus writes made while isolated, and new changes branch from it.
// Isolated changes still accumulate in the full history and become visible after
// Integrate. It mirrors the Rust AutoCommit::isolate API.
func (d *Document) Isolate(heads []Hash) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.engine.Isolate(engineHashes(heads)); err != nil {
		return fmt.Errorf("cannot isolate Automerge document: %w", err)
	}

	return nil
}

// Integrate ends isolation, returning reads and writes to the full history that
// includes every isolated and merged change. It mirrors AutoCommit::integrate.
func (d *Document) Integrate() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.engine.Integrate(); err != nil {
		return fmt.Errorf("cannot integrate Automerge document: %w", err)
	}

	return nil
}

// Stats reports aggregate document statistics.
type Stats struct {
	NumChanges uint64 `json:"numChanges"`
	NumOps     uint64 `json:"numOps"`
	NumActors  uint64 `json:"numActors"`
}

// Stats returns the number of changes, operations, and actors in the document.
func (d *Document) Stats() (Stats, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Stats{}, ErrClosed
	}

	data, err := d.engine.Stats()
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
	actorID ActorID,
) (*Document, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil, ErrClosed
	}

	data, err := d.engine.Save(true, true)
	if err != nil {
		d.mu.Unlock()
		return nil, fmt.Errorf("cannot save Automerge fork source: %w", err)
	}

	_, referenceBackend := d.engine.(*reference.Engine)
	d.mu.Unlock()

	if referenceBackend {
		return LoadReference(data, actorID)
	}

	return Load(data, actorID)
}

// SaveIncremental serializes changes since the previous save operation.
func (d *Document) SaveIncremental() ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	data, err := d.engine.SaveIncremental()
	if err != nil {
		return nil, fmt.Errorf("cannot save incremental Automerge changes: %w", err)
	}

	return data, nil
}

// LoadIncremental applies incrementally encoded Automerge changes.
func (d *Document) LoadIncremental(data []byte) (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return 0, ErrClosed
	}

	applied, err := d.engine.LoadIncremental(data)
	if err != nil {
		return 0, fmt.Errorf("cannot load incremental Automerge changes: %w", err)
	}

	return applied, nil
}

// PutString assigns a string at a key in the root map.
func (d *Document) PutString(key, value string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrClosed
	}

	if err := d.engine.PutString(rootObject, key, value); err != nil {
		return fmt.Errorf("cannot put Automerge string: %w", err)
	}

	return nil
}

// String returns a string value from a key in the root map.
func (d *Document) String(key string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return "", ErrClosed
	}

	value, err := d.engine.GetString(rootObject, key)
	if err != nil {
		return "", fmt.Errorf("cannot get Automerge string: %w", err)
	}

	return value, nil
}

// CreateText creates collaborative text at a key in the root map.
func (d *Document) CreateText(key string) (*Text, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.engine.PutText(rootObject, key)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge text: %w", err)
	}

	return &Text{document: d, handle: handle}, nil
}

// Text returns existing collaborative text from the root map.
func (d *Document) Text(key string) (*Text, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.engine.GetText(rootObject, key)
	if err != nil {
		return nil, fmt.Errorf("cannot get Automerge text: %w", err)
	}

	return &Text{document: d, handle: handle}, nil
}

// Commit records pending operations as one Automerge change.
func (d *Document) Commit(
	message string,
	timestamp time.Time,
) (Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Hash{}, ErrClosed
	}

	hash, err := d.engine.Commit(message, timestamp)
	if err != nil {
		return Hash{}, fmt.Errorf("cannot commit Automerge document: %w", err)
	}

	return Hash(hash), nil
}

// CommitNow records pending operations using the current Unix timestamp.
func (d *Document) CommitNow(message string) (Hash, error) {
	return d.Commit(message, time.Now())
}

// EmptyCommit records a change without document operations.
func (d *Document) EmptyCommit(
	message string,
	timestamp time.Time,
) (Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return Hash{}, ErrClosed
	}

	hash, err := d.engine.EmptyCommit(message, timestamp)
	if err != nil {
		return Hash{}, fmt.Errorf("cannot commit empty Automerge change: %w", err)
	}

	return Hash(hash), nil
}

// EmptyCommitNow records an empty change using the current Unix timestamp.
func (d *Document) EmptyCommitNow(message string) (Hash, error) {
	return d.EmptyCommit(message, time.Now())
}

// Rollback discards every operation pending in the current change.
func (d *Document) Rollback() (uint64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return 0, ErrClosed
	}

	cancelled, err := d.engine.Rollback()
	if err != nil {
		return 0, fmt.Errorf("cannot roll back Automerge document: %w", err)
	}

	return cancelled, nil
}

// Heads returns the hashes at the document's current frontier.
func (d *Document) Heads() ([]Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	engineHeads, err := d.engine.Heads()
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

	return oracle.BloomContains([32]byte(target), seedArrays)
}

// HasHeads reports whether every hash exists in the document history.
func (d *Document) HasHeads(heads []Hash) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return false, ErrClosed
	}

	hasHeads, err := d.engine.HasHeads(engineHashes(heads))
	if err != nil {
		return false, fmt.Errorf("cannot inspect Automerge heads: %w", err)
	}

	return hasHeads, nil
}

// MissingDependencies returns unknown hashes required to reach heads.
func (d *Document) MissingDependencies(
	heads []Hash,
) ([]Hash, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	missing, err := d.engine.MissingDependencies(engineHashes(heads))
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

	raw, hashes, err := changeSource.ChangesSince(engineHeads)
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
	changes []Change,
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

	raw := make([][]byte, len(changes))
	for i, change := range changes {
		raw[i] = change.Bytes
	}

	if err := applier.ApplyChanges(raw); err != nil {
		return fmt.Errorf("cannot apply incremental Automerge changes: %w", err)
	}

	return nil
}

// Merge applies all changes from another document.
func (d *Document) Merge(other *Document) ([]Hash, error) {
	if other == nil {
		return nil, ErrNilDocument
	}

	if d == other {
		return nil, ErrSameDocument
	}

	otherData, err := other.Save()
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge merge source: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	engineHeads, err := d.engine.Merge(otherData)
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
func (d *Document) NewSyncState() (*SyncState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.engine.NewSyncState()
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge sync state: %w", err)
	}

	return &SyncState{document: d, handle: handle}, nil
}

// LoadSyncState resumes a previously serialized remote-peer session.
func (d *Document) LoadSyncState(data []byte) (*SyncState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrClosed
	}

	handle, err := d.engine.LoadSyncState(data)
	if err != nil {
		return nil, fmt.Errorf("cannot load Automerge sync state: %w", err)
	}

	return &SyncState{document: d, handle: handle}, nil
}

// Splice replaces deleteCount UTF-16 code units at index with value.
func (t *Text) Splice(index uint32, deleteCount int32, value string) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if err := t.document.engine.SpliceText(t.handle, index, deleteCount, value); err != nil {
		return fmt.Errorf("cannot splice Automerge text: %w", err)
	}

	return nil
}

// Update replaces the text content with value using a minimal splice so that
// concurrent edits to unaffected regions merge cleanly. It mirrors the Rust
// AutoCommit::update_text and JavaScript updateText helpers.
func (t *Text) Update(value string) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if err := t.document.engine.UpdateText(t.handle, value); err != nil {
		return fmt.Errorf("cannot update Automerge text: %w", err)
	}

	return nil
}

// String returns the current materialized text.
func (t *Text) String() (string, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return "", ErrClosed
	}

	value, err := t.document.engine.Text(t.handle)
	if err != nil {
		return "", fmt.Errorf("cannot read Automerge text: %w", err)
	}

	return value, nil
}

// StringAt returns text at a historical causal frontier.
func (t *Text) StringAt(heads []Hash) (string, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return "", ErrClosed
	}

	value, err := t.document.engine.TextAt(
		t.handle,
		engineHashes(heads),
	)
	if err != nil {
		return "", fmt.Errorf("cannot read historical Automerge text: %w", err)
	}

	return value, nil
}

// Cursor returns a stable address for the UTF-16 position at index.
func (t *Text) Cursor(index uint32) (Cursor, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	cursor, err := t.document.engine.TextCursor(t.handle, index)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge text cursor: %w", err)
	}

	return Cursor(cursor), nil
}

// CursorPosition resolves a stable cursor in the current document.
func (t *Text) CursorPosition(cursor Cursor) (uint32, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return 0, ErrClosed
	}

	position, err := t.document.engine.TextCursorPosition(t.handle, cursor)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve Automerge text cursor: %w", err)
	}

	return position, nil
}

// Close releases the peer-specific synchronization state.
func (s *SyncState) Close() error {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	if s.document.closed {
		return nil
	}

	if err := s.document.engine.CloseSyncState(s.handle); err != nil {
		return fmt.Errorf("cannot close Automerge sync state: %w", err)
	}

	return nil
}

// GenerateMessage returns the next message for the remote peer.
func (s *SyncState) GenerateMessage() ([]byte, bool, error) {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return nil, false, ErrClosed
	}

	if s.closed {
		return nil, false, ErrSyncStateClosed
	}

	message, ok, err := s.document.engine.GenerateSyncMessage(s.handle)
	if err != nil {
		return nil, false, fmt.Errorf("cannot generate Automerge sync message: %w", err)
	}

	return message, ok, nil
}

// ReceiveMessage applies a message received from the remote peer.
func (s *SyncState) ReceiveMessage(message []byte) error {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return ErrClosed
	}

	if s.closed {
		return ErrSyncStateClosed
	}

	if err := s.document.engine.ReceiveSyncMessage(s.handle, message); err != nil {
		return fmt.Errorf("cannot receive Automerge sync message: %w", err)
	}

	return nil
}

// SetReadOnly controls whether incoming changes are applied by this peer.
func (s *SyncState) SetReadOnly(readOnly bool) error {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return ErrClosed
	}

	if s.closed {
		return ErrSyncStateClosed
	}

	if err := s.document.engine.SetSyncReadOnly(
		s.handle,
		readOnly,
	); err != nil {
		return fmt.Errorf("cannot set Automerge sync read-only mode: %w", err)
	}

	return nil
}

// PeerReadOnly reports whether the remote peer advertised read-only mode.
func (s *SyncState) PeerReadOnly() (bool, error) {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return false, ErrClosed
	}

	if s.closed {
		return false, ErrSyncStateClosed
	}

	readOnly, err := s.document.engine.SyncPeerReadOnly(s.handle)
	if err != nil {
		return false, fmt.Errorf("cannot get Automerge peer read-only mode: %w", err)
	}

	return readOnly, nil
}

// Save serializes the peer-specific synchronization state.
func (s *SyncState) Save() ([]byte, error) {
	s.document.mu.Lock()
	defer s.document.mu.Unlock()

	if s.document.closed {
		return nil, ErrClosed
	}

	if s.closed {
		return nil, ErrSyncStateClosed
	}

	data, err := s.document.engine.SaveSyncState(s.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot save Automerge sync state: %w", err)
	}

	return data, nil
}

// String returns the lowercase hexadecimal change hash.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// String returns the lowercase hexadecimal actor ID.
func (a ActorID) String() string {
	return hex.EncodeToString(a[:])
}
