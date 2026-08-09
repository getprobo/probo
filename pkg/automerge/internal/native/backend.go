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

package native

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type Backend struct {
	state         *State
	actor         ActorID
	nextOp        uint64
	base          []byte
	appended      [][]byte
	saveCursor    int
	pending       []Operation
	objects       map[uint32]ObjectID
	nextHandle    uint32
	syncStates    map[uint32]*nativeSyncState
	nextSyncState uint32
	queuedChanges map[ChangeHash]*Change
	queuedBytes   int
	diffCursor    [][32]byte

	// isolation pins reads and writes to a historical frontier. While active,
	// state points at a view built from the isolation heads and fullState keeps
	// the complete history; committed isolated changes are applied to both, while
	// merged changes are applied only to fullState.
	isolationActive bool
	fullState       *State
	baseActor       ActorID

	// isolationDiffTargets records the frontiers isolated to since the diff
	// cursor was last set. When present, an incremental diff replays the
	// transition from the cursor down to each isolation frontier and back up to
	// the current heads, matching the reference's patch-log output across
	// isolate/integrate rather than a direct state comparison.
	isolationDiffTargets [][][32]byte
}

type nativeSyncState struct {
	RemoteHeads       [][32]byte `json:"remoteHeads"`
	LastSentHeads     [][32]byte `json:"lastSentHeads"`
	Need              [][32]byte `json:"need"`
	Requested         [][32]byte `json:"requested"`
	NeedsAck          bool       `json:"needsAck"`
	InFlight          bool       `json:"inFlight"`
	Sent              bool       `json:"sent"`
	ReadOnly          bool       `json:"readOnly"`
	PeerReadOnly      bool       `json:"peerReadOnly"`
	PeerModeChanged   bool       `json:"peerModeChanged"`
	PeerSupportsReset bool       `json:"peerSupportsReset"`
	NeedsReset        bool       `json:"needsReset"`
	ModeChanged       bool       `json:"modeChanged"`
}

type scalarWire struct {
	Type   string `json:"type"`
	Bool   bool   `json:"bool"`
	Uint   uint64 `json:"uint"`
	Int    int64  `json:"int"`
	Float  uint64 `json:"floatBits"`
	String string `json:"string"`
	Bytes  string `json:"bytes"`
}

const (
	maxQueuedChangeBytes = 64 * 1024 * 1024
	maxQueuedChanges     = 100_000

	syncFlagReset         = 1 << 0
	syncFlagReadOnly      = 1 << 1
	syncFlagSupportsReset = 1 << 2
	syncFlagMarker        = 0x80
)

func NewBackend(ctx context.Context) (*Backend, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	actor, err := randomActorID()
	if err != nil {
		return nil, err
	}

	base := encodeEmptyDocument()

	document, err := Decode(base)
	if err != nil {
		return nil, fmt.Errorf("cannot decode native empty document: %w", err)
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize native empty state: %w", err)
	}

	return &Backend{
		state:         state,
		actor:         actor,
		nextOp:        state.maxOpGlobal() + 1,
		base:          base,
		objects:       map[uint32]ObjectID{0: RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*nativeSyncState),
		nextSyncState: 1,
		queuedChanges: make(map[ChangeHash]*Change),
	}, nil
}

func LoadBackend(ctx context.Context, data []byte) (*Backend, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	document, err := Decode(data)
	if err != nil {
		// A document may retain orphan changes (changes whose dependencies are
		// not present) that were preserved across a save. Strict decoding
		// rejects them, so fall back to a tolerant load that applies every
		// change whose dependencies are satisfiable and queues the rest. A load
		// that cannot apply a single change (a bare orphan) still fails.
		if backend, ok, tolerantErr := loadBackendRetainingOrphans(data, err); ok {
			return backend, tolerantErr
		}

		return nil, fmt.Errorf("cannot decode native document: %w", err)
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize native document state: %w", err)
	}

	actor, err := randomActorID()
	if err != nil {
		return nil, err
	}

	return &Backend{
		state:         state,
		actor:         actor,
		nextOp:        state.maxOpGlobal() + 1,
		base:          append([]byte(nil), data...),
		objects:       map[uint32]ObjectID{0: RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*nativeSyncState),
		nextSyncState: 1,
		queuedChanges: make(map[ChangeHash]*Change),
	}, nil
}

// loadBackendRetainingOrphans attempts a tolerant load for documents that carry
// orphan changes. It returns ok=false when the tolerant path does not apply (the
// data is corrupt beyond missing dependencies, or nothing can be applied), so
// the caller reports the original strict error. On success the applied history
// forms the base and the orphan changes are queued for later resolution.
func loadBackendRetainingOrphans(
	data []byte,
	strictErr error,
) (*Backend, bool, error) {
	if !strings.Contains(strictErr.Error(), "missing dependency") {
		return nil, false, nil
	}

	document, err := DecodePartial(data)
	if err != nil {
		return nil, false, nil
	}

	state := NewState()
	queued := make(map[ChangeHash]*Change, len(document.Changes))

	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash == nil || len(change.Raw) == 0 {
			return nil, false, nil
		}

		queued[*change.Hash] = change
	}

	applied := make([]*Change, 0, len(document.Changes))

	for {
		progressed := false

		for _, change := range orderedQueuedChanges(queued) {
			if !state.hasDependencies(change) {
				continue
			}

			if err := state.ApplyChange(change); err != nil {
				return nil, false, nil
			}

			applied = append(applied, change)
			delete(queued, *change.Hash)

			progressed = true
		}

		if !progressed {
			break
		}
	}

	if len(applied) == 0 {
		return nil, false, nil
	}

	actor, err := randomActorID()
	if err != nil {
		return nil, true, err
	}

	base := make([]byte, 0, len(data))
	for _, change := range applied {
		base = append(base, change.Raw...)
	}

	queuedClone := make(map[ChangeHash]*Change, len(queued))
	queuedBytes := 0

	for hash, change := range queued {
		clone := *change
		clone.Raw = append([]byte(nil), change.Raw...)
		queuedClone[hash] = &clone
		queuedBytes += len(clone.Raw)
	}

	return &Backend{
		state:         state,
		actor:         actor,
		nextOp:        state.maxOpGlobal() + 1,
		base:          base,
		objects:       map[uint32]ObjectID{0: RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*nativeSyncState),
		nextSyncState: 1,
		queuedChanges: queuedClone,
		queuedBytes:   queuedBytes,
	}, true, nil
}

// orderedQueuedChanges returns queued changes in a deterministic order (by hash)
// so tolerant loading applies and re-serializes changes reproducibly.
func orderedQueuedChanges(queued map[ChangeHash]*Change) []*Change {
	changes := make([]*Change, 0, len(queued))
	for _, change := range queued {
		changes = append(changes, change)
	}

	sort.Slice(changes, func(i, j int) bool {
		return bytes.Compare(changes[i].Hash[:], changes[j].Hash[:]) < 0
	})

	return changes
}

func (b *Backend) Close(context.Context) error {
	return nil
}

func (b *Backend) Save(ctx context.Context) ([]byte, error) {
	return b.save(ctx, true, true)
}

// SaveNoCompress serializes the document without DEFLATE-compressing any change
// chunks, mirroring Rust's AutoCommit::save_nocompress.
func (b *Backend) SaveNoCompress(ctx context.Context) ([]byte, error) {
	return b.save(ctx, true, false)
}

// SaveWithOptions serializes the document, optionally appending retained orphan
// changes (queued changes whose dependencies are still missing) so they survive
// a save/load round trip. It mirrors the Rust SaveOptions.retain_orphans flag;
// the reference retains orphans by default.
func (b *Backend) SaveWithOptions(
	ctx context.Context,
	retainOrphans bool,
) ([]byte, error) {
	return b.save(ctx, retainOrphans, true)
}

func (b *Backend) save(
	ctx context.Context,
	retainOrphans bool,
	deflate bool,
) ([]byte, error) {
	if len(b.pending) > 0 {
		if _, err := b.Commit(ctx, "", time.Time{}); err != nil {
			return nil, err
		}
	}

	total := len(b.base)
	for _, change := range b.appended {
		total += len(change)
	}

	data := make([]byte, 0, total)
	data = append(data, b.base...)

	for _, change := range b.appended {
		data = append(data, maybeCompressChangeChunk(change, deflate)...)
	}

	b.saveCursor = len(b.appended)

	if retainOrphans {
		for _, change := range orderedQueuedChanges(b.queuedChanges) {
			data = append(data, maybeCompressChangeChunk(change.Raw, deflate)...)
		}
	}

	return data, nil
}

// deflateMinSize matches Rust's change::DEFLATE_MIN_SIZE: change chunks whose
// body is at least this many bytes are worth compressing.
const deflateMinSize = 250

// maybeCompressChangeChunk reframes an uncompressed change chunk as a compressed
// change chunk when compression is requested and the body is large enough to
// benefit. The 4-byte checksum is preserved because the reference (and native
// decoder) recompute the hash from the inflated body. Any other chunk kind, a
// small body, or a non-shrinking result is returned unchanged.
func maybeCompressChangeChunk(raw []byte, deflateEnabled bool) []byte {
	const headerSize = 9 // 4 magic + 4 checksum + 1 type

	if !deflateEnabled || len(raw) <= headerSize || ChunkType(raw[8]) != ChunkChange {
		return raw
	}

	reader := &reader{data: raw, offset: headerSize}

	bodyLength, err := reader.uleb()
	if err != nil || reader.offset+int(bodyLength) > len(raw) {
		return raw
	}

	body := raw[reader.offset : reader.offset+int(bodyLength)]
	if len(body) < deflateMinSize {
		return raw
	}

	compressed, err := deflate(body)
	if err != nil || len(compressed) >= len(body) {
		return raw
	}

	out := make([]byte, 0, headerSize+len(compressed)+8)
	out = append(out, raw[:8]...)
	out = append(out, byte(ChunkCompressedChange))
	out = appendULEB(out, uint64(len(compressed)))
	out = append(out, compressed...)

	return out
}

func (b *Backend) SaveIncremental(ctx context.Context) ([]byte, error) {
	if len(b.pending) > 0 {
		if _, err := b.Commit(ctx, "", time.Time{}); err != nil {
			return nil, err
		}
	}

	if b.saveCursor > len(b.appended) {
		b.saveCursor = len(b.appended)
	}

	total := 0
	for _, change := range b.appended[b.saveCursor:] {
		total += len(change)
	}

	data := make([]byte, 0, total)
	for _, change := range b.appended[b.saveCursor:] {
		data = append(data, change...)
	}

	b.saveCursor = len(b.appended)

	return data, nil
}

func (b *Backend) LoadIncremental(
	ctx context.Context,
	data []byte,
) (uint64, error) {
	_, consumed, err := DecodeIncremental(data)
	if err != nil {
		return 0, err
	}

	before := len(b.state.changes)
	if _, err := b.Merge(ctx, data[:consumed]); err != nil {
		return 0, err
	}

	after := len(b.state.changes)
	if after < before {
		return 0, fmt.Errorf("incremental load reduced the change count")
	}

	return uint64(after - before), nil
}

func (b *Backend) SetActor(ctx context.Context, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	actor, err := NewActorID(value)
	if err != nil {
		return err
	}

	if len(b.pending) > 0 {
		return fmt.Errorf("cannot change actor with pending operations")
	}

	b.actor = actor

	return nil
}

func (b *Backend) PutString(
	ctx context.Context,
	object uint32,
	key string,
	value string,
) error {
	if err := b.requireRoot(ctx, object); err != nil {
		return err
	}

	if existing, ok := b.state.visibleMapOperation(key, ActionSet); ok &&
		existing.Value != nil &&
		existing.Value.Type == ScalarString &&
		existing.Value.String == value {
		return nil
	}

	property := key

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: RootObject(),
		Key:    Key{Property: &property},
		Action: ActionSet,
		Value:  &Scalar{Type: ScalarString, String: value},
	}
	for _, predecessor := range b.state.visibleMapOperations(key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	return b.addPending(operation)
}

func (b *Backend) GetString(
	ctx context.Context,
	object uint32,
	key string,
) (string, error) {
	if err := b.requireRoot(ctx, object); err != nil {
		return "", err
	}

	operation, ok := b.state.visibleMapOperation(key, ActionSet)
	if !ok || operation.Value == nil || operation.Value.Type != ScalarString {
		return "", fmt.Errorf("string property %q does not exist", key)
	}

	return operation.Value.String, nil
}

func (b *Backend) PutScalar(
	ctx context.Context,
	object uint32,
	key string,
	encoded []byte,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return err
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	if existing, ok := b.state.visibleMapObjectValue(objectID, key); ok {
		existingValue, scalar := b.state.scalarValue(existing)
		if scalar && scalarValuesEqual(existingValue, value) {
			return nil
		}
	}

	property := key

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: objectID,
		Key:    Key{Property: &property},
		Action: ActionSet,
		Value:  &value,
	}
	for _, predecessor := range b.state.visibleMapObjectOperations(objectID, key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	return b.addPending(operation)
}

func (b *Backend) GetScalar(
	ctx context.Context,
	object uint32,
	key string,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return nil, err
	}

	operation, ok := b.state.visibleMapObjectValue(objectID, key)
	if !ok {
		return nil, fmt.Errorf("scalar property %q does not exist", key)
	}

	value, ok := b.state.scalarValue(operation)
	if !ok {
		return nil, fmt.Errorf("map value %q is not a scalar", key)
	}

	return encodeScalarWire(value)
}

func (b *Backend) GetScalarAtHeads(
	ctx context.Context,
	object uint32,
	key string,
	heads [][32]byte,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	operation, ok := historical.visibleMapObjectValue(objectID, key)
	if !ok {
		return nil, fmt.Errorf("scalar property %q does not exist", key)
	}

	value, ok := historical.scalarValue(operation)
	if !ok {
		return nil, fmt.Errorf("map value %q is not a scalar", key)
	}

	return encodeScalarWire(value)
}

func (b *Backend) GetAllScalars(
	ctx context.Context,
	object uint32,
	key string,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return nil, err
	}

	var values []json.RawMessage

	for _, operation := range b.state.visibleMapObjectOperations(objectID, key) {
		if operation.Action == ActionIncrement {
			continue
		}

		value, ok := b.state.scalarValue(operation)
		if !ok {
			continue
		}

		encoded, err := encodeScalarWire(value)
		if err != nil {
			return nil, err
		}

		values = append(values, json.RawMessage(encoded))
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("scalar property %q does not exist", key)
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("cannot encode scalar conflicts: %w", err)
	}

	return encoded, nil
}

func (b *Backend) GetAllScalarsAt(
	ctx context.Context,
	object uint32,
	index uint64,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sequenceObject, err := b.sequenceObject(object)
	if err != nil {
		return nil, err
	}

	conflicts, ok := b.state.sequenceConflicts(sequenceObject.OpID, index)
	if !ok {
		return nil, fmt.Errorf("sequence value at index %d does not exist", index)
	}

	var values []json.RawMessage

	for _, operation := range conflicts {
		value, ok := b.state.scalarValue(operation)
		if !ok {
			continue
		}

		encoded, err := encodeScalarWire(value)
		if err != nil {
			return nil, err
		}

		values = append(values, json.RawMessage(encoded))
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("sequence value at index %d is not a scalar", index)
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("cannot encode sequence scalar conflicts: %w", err)
	}

	return encoded, nil
}

func (b *Backend) PutObject(
	ctx context.Context,
	object uint32,
	key string,
	rawType string,
) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return 0, err
	}

	action, err := objectAction(rawType)
	if err != nil {
		return 0, err
	}

	property := key

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: objectID,
		Key:    Key{Property: &property},
		Action: action,
	}
	for _, predecessor := range b.state.visibleMapObjectOperations(objectID, key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Backend) GetObject(
	ctx context.Context,
	object uint32,
	key string,
) (uint32, string, error) {
	if err := ctx.Err(); err != nil {
		return 0, "", err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return 0, "", err
	}

	operation, ok := b.state.visibleMapObjectValue(objectID, key)
	if !ok {
		return 0, "", fmt.Errorf("object property %q does not exist", key)
	}

	rawType, err := actionObjectType(operation.Action)
	if err != nil {
		return 0, "", err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), rawType, nil
}

func (b *Backend) InsertScalar(
	ctx context.Context,
	object uint32,
	index uint64,
	encoded []byte,
) error {
	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	_, err = b.insertSequenceOperation(ctx, object, index, ActionSet, &value)

	return err
}

func (b *Backend) PutScalarAt(
	ctx context.Context,
	object uint32,
	index uint64,
	encoded []byte,
) error {
	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	target, err := b.sequenceOperation(ctx, object, index)
	if err != nil {
		return err
	}

	objectID, err := b.object(object)
	if err != nil {
		return err
	}

	// Assigning the identical value to a list element with a single visible
	// value is a no-op, matching the reference and native's map put. A
	// conflicted element still records the assignment so the conflict resolves.
	if existingValue, scalar := b.state.scalarValue(target.Operation); scalar &&
		scalarValuesEqual(existingValue, value) &&
		len(b.state.visibleSequenceElementOperations(target.Element)) == 1 {
		return nil
	}

	return b.addPending(Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          Key{Element: new(target.Element)},
		Action:       ActionSet,
		Value:        &value,
		Predecessors: b.sequenceElementPredecessors(target.Element),
	})
}

func (b *Backend) InsertObject(
	ctx context.Context,
	object uint32,
	index uint64,
	rawType string,
) (uint32, error) {
	action, err := objectAction(rawType)
	if err != nil {
		return 0, err
	}

	operation, err := b.insertSequenceOperation(
		ctx,
		object,
		index,
		action,
		nil,
	)
	if err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Backend) PutObjectAt(
	ctx context.Context,
	object uint32,
	index uint64,
	rawType string,
) (uint32, error) {
	action, err := objectAction(rawType)
	if err != nil {
		return 0, err
	}

	target, err := b.sequenceOperation(ctx, object, index)
	if err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, err
	}

	operation := Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          Key{Element: new(target.Element)},
		Action:       action,
		Predecessors: b.sequenceElementPredecessors(target.Element),
	}
	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Backend) GetScalarAt(
	ctx context.Context,
	object uint32,
	index uint64,
) ([]byte, error) {
	operation, err := b.sequenceOperation(ctx, object, index)
	if err != nil {
		return nil, err
	}

	value, ok := b.state.scalarValue(operation.Operation)
	if !ok {
		return nil, fmt.Errorf("sequence value at index %d is not a scalar", index)
	}

	return encodeScalarWire(value)
}

func (b *Backend) GetObjectAt(
	ctx context.Context,
	object uint32,
	index uint64,
) (uint32, string, error) {
	operation, err := b.sequenceOperation(ctx, object, index)
	if err != nil {
		return 0, "", err
	}

	rawType, err := actionObjectType(operation.Operation.Action)
	if err != nil {
		return 0, "", err
	}

	return b.pushObject(ObjectID{OpID: operation.Operation.ID}), rawType, nil
}

func (b *Backend) DeleteMap(
	ctx context.Context,
	object uint32,
	key string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return err
	}

	property := key

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: objectID,
		Key:    Key{Property: &property},
		Action: ActionDelete,
	}
	for _, predecessor := range b.state.visibleMapObjectOperations(objectID, key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	if len(operation.Predecessors) == 0 {
		return fmt.Errorf("map property %q does not exist", key)
	}

	return b.addPending(operation)
}

func (b *Backend) DeleteSequence(
	ctx context.Context,
	object uint32,
	index uint64,
) error {
	target, err := b.sequenceOperation(ctx, object, index)
	if err != nil {
		return err
	}

	objectID, err := b.object(object)
	if err != nil {
		return err
	}

	return b.addPending(Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          Key{Element: new(target.Element)},
		Action:       ActionDelete,
		Predecessors: b.sequenceElementPredecessors(target.Element),
	})
}

func (b *Backend) Increment(
	ctx context.Context,
	object uint32,
	key string,
	delta int64,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return err
	}

	visible := b.state.visibleMapObjectOperations(objectID, key)

	hasCounter := false

	for _, operation := range visible {
		if isCounterOperation(operation) {
			hasCounter = true

			break
		}
	}

	if !hasCounter {
		return fmt.Errorf("map property %q is not a counter", key)
	}

	property := key

	predecessors := make([]OpID, 0, len(visible))
	for _, operation := range visible {
		predecessors = append(predecessors, operation.ID)
	}

	return b.addPending(Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          Key{Property: &property},
		Action:       ActionIncrement,
		Value:        &Scalar{Type: ScalarInt, Int: delta},
		Predecessors: predecessors,
	})
}

func (b *Backend) IncrementAt(
	ctx context.Context,
	object uint32,
	index uint64,
	delta int64,
) error {
	target, err := b.sequenceOperation(ctx, object, index)
	if err != nil {
		return err
	}

	visible := b.state.visibleSequenceElementOperations(target.Element)

	hasCounter := false

	for _, operation := range visible {
		if isCounterOperation(operation) {
			hasCounter = true

			break
		}
	}

	if !hasCounter {
		return fmt.Errorf("sequence value at index %d is not a counter", index)
	}

	objectID, err := b.object(object)
	if err != nil {
		return err
	}

	predecessors := make([]OpID, 0, len(visible))
	for _, operation := range visible {
		predecessors = append(predecessors, operation.ID)
	}

	return b.addPending(Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          Key{Element: new(target.Element)},
		Action:       ActionIncrement,
		Value:        &Scalar{Type: ScalarInt, Int: delta},
		Predecessors: predecessors,
	})
}

func (b *Backend) Keys(ctx context.Context, object uint32) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	objectID, err := b.mapObject(object)
	if err != nil {
		return nil, err
	}

	return b.state.mapKeys(objectID), nil
}

func (b *Backend) Length(ctx context.Context, object uint32) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, err
	}

	if objectID.IsRoot {
		return b.state.mapLength(objectID), nil
	}

	operation, ok := b.state.operations[objectID.OpID]
	if !ok {
		return 0, fmt.Errorf("object does not exist")
	}

	if operation.Action == ActionMakeMap ||
		operation.Action == ActionMakeTable {
		return b.state.mapLength(objectID), nil
	}

	if operation.Action != ActionMakeList &&
		operation.Action != ActionMakeText {
		return 0, fmt.Errorf("object does not have a length")
	}

	sequence := b.state.sequenceValues(objectID.OpID)

	if operation.Action == ActionMakeText {
		total := uint64(0)
		for _, value := range sequence {
			total += sequenceValueUTF16Width(value)
		}

		return total, nil
	}

	return uint64(len(sequence)), nil
}

func (b *Backend) PutText(
	ctx context.Context,
	object uint32,
	key string,
) (uint32, error) {
	if err := b.requireRoot(ctx, object); err != nil {
		return 0, err
	}

	property := key

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: RootObject(),
		Key:    Key{Property: &property},
		Action: ActionMakeText,
	}
	for _, predecessor := range b.state.visibleMapOperations(key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Backend) GetText(
	ctx context.Context,
	object uint32,
	key string,
) (uint32, error) {
	if err := b.requireRoot(ctx, object); err != nil {
		return 0, err
	}

	operation, ok := b.state.visibleMapOperation(key, ActionMakeText)
	if !ok {
		return 0, fmt.Errorf("text property %q does not exist", key)
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Backend) SpliceText(
	ctx context.Context,
	handle uint32,
	index uint32,
	deleteCount int32,
	value string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if deleteCount < 0 {
		return fmt.Errorf("negative text deletion is unsupported")
	}

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	// Splice positions share the unified rich-text index space with marks and
	// blocks, so walk the full visible element sequence (text and block markers)
	// rather than the text-only view.
	sequence := b.state.sequenceElements(object.OpID)

	start, end, previous, err := sequenceRange(sequence, index, uint32(deleteCount))
	if err != nil {
		return err
	}

	for _, target := range sequence[start:end] {
		operation := Operation{
			ID:           b.nextOperationID(),
			Object:       object,
			Key:          Key{Element: new(target.ID)},
			Action:       ActionDelete,
			Predecessors: []OpID{target.ID},
		}
		if err := b.addPending(operation); err != nil {
			return err
		}
	}

	for _, character := range value {
		key := Key{IsHead: previous == nil}
		if previous != nil {
			key.Element = new(*previous)
		}

		operation := Operation{
			ID:     b.nextOperationID(),
			Object: object,
			Key:    key,
			Insert: true,
			Action: ActionSet,
			Value:  &Scalar{Type: ScalarString, String: string(character)},
		}
		if err := b.addPending(operation); err != nil {
			return err
		}

		previous = new(operation.ID)
	}

	return nil
}

type (
	updateSpanInput struct {
		Type  string                     `json:"type"`
		Text  string                     `json:"text"`
		Marks map[string]json.RawMessage `json:"marks"`
		Block json.RawMessage            `json:"block"`
	}

	updateSpansConfigInput struct {
		DefaultExpand  string            `json:"defaultExpand"`
		PerMarkExpands map[string]string `json:"perMarkExpands"`
	}

	desiredMark struct {
		name  string
		value Scalar
		start uint32
		end   uint32
	}
)

// UpdateSpans transforms the text object so its spans equal the supplied spans,
// mirroring the Rust AutoCommit::update_spans helper. The text content is
// reconciled with a minimal grapheme diff and the marks are then set to exactly
// the marks named on the spans, honoring the per-mark and default expand config.
// Block spans are not yet supported.
func (b *Backend) UpdateSpans(
	ctx context.Context,
	handle uint32,
	spans []byte,
	config []byte,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	var inputs []updateSpanInput
	if err := json.Unmarshal(spans, &inputs); err != nil {
		return fmt.Errorf("cannot decode update spans: %w", err)
	}

	var configuration updateSpansConfigInput
	if err := json.Unmarshal(config, &configuration); err != nil {
		return fmt.Errorf("cannot decode update spans config: %w", err)
	}

	target, err := targetBlockGraphemes(inputs)
	if err != nil {
		return err
	}

	current := b.currentBlockGraphemes(object)

	hook := &blockDiffHook{
		ctx:     ctx,
		backend: b,
		handle:  handle,
		old:     current,
		new:     target,
	}
	myersDiff(hook, blockTokens(current), blockTokens(target))

	if hook.err != nil {
		return hook.err
	}

	desired, err := desiredMarks(inputs)
	if err != nil {
		return err
	}

	return b.reconcileMarks(ctx, handle, object, desired, configuration)
}

// blockOrGrapheme is one unit of the block-aware span diff: either a block
// marker with its attributes or a single grapheme cluster of text.
type blockOrGrapheme struct {
	block    map[string]any
	isBlock  bool
	grapheme string
}

func (item blockOrGrapheme) width() int {
	if item.isBlock {
		return 1
	}

	return utf16Width(item.grapheme)
}

// blockTokens renders each unit to a comparison token so the grapheme-based
// Myers diff can compare blocks by their attributes and text by its clusters.
func blockTokens(items []blockOrGrapheme) []string {
	tokens := make([]string, len(items))

	for i, item := range items {
		if !item.isBlock {
			tokens[i] = "g" + item.grapheme

			continue
		}

		encoded, _ := json.Marshal(item.block)
		tokens[i] = "b" + string(encoded)
	}

	return tokens
}

// currentBlockGraphemes materializes the text object as the block/grapheme units
// the diff operates on: block markers become blocks and text runs are split into
// grapheme clusters.
func (b *Backend) currentBlockGraphemes(object ObjectID) []blockOrGrapheme {
	items := make([]blockOrGrapheme, 0)

	var run strings.Builder

	flush := func() {
		if run.Len() == 0 {
			return
		}

		for _, grapheme := range graphemeClusters(run.String()) {
			items = append(items, blockOrGrapheme{grapheme: grapheme})
		}

		run.Reset()
	}

	for _, value := range b.state.sequenceValues(object.OpID) {
		operation := value.Operation

		if operation.Action == ActionMakeMap {
			flush()

			attributes, err := b.state.mapValue(operation.ID, make(map[OpID]struct{}))
			if err != nil || attributes == nil {
				attributes = map[string]any{}
			}

			items = append(items, blockOrGrapheme{isBlock: true, block: attributes})

			continue
		}

		if operation.Value != nil && operation.Value.Type == ScalarString {
			run.WriteString(operation.Value.String)
		}
	}

	flush()

	return items
}

// targetBlockGraphemes converts the requested spans into block/grapheme units.
func targetBlockGraphemes(spans []updateSpanInput) ([]blockOrGrapheme, error) {
	items := make([]blockOrGrapheme, 0)

	for _, span := range spans {
		switch span.Type {
		case "text":
			for _, grapheme := range graphemeClusters(span.Text) {
				items = append(items, blockOrGrapheme{grapheme: grapheme})
			}
		case "block":
			attributes := map[string]any{}
			if len(span.Block) > 0 {
				if err := json.Unmarshal(span.Block, &attributes); err != nil {
					return nil, fmt.Errorf("cannot decode block span: %w", err)
				}
			}

			items = append(items, blockOrGrapheme{isBlock: true, block: attributes})
		default:
			return nil, fmt.Errorf("unsupported update span type %q", span.Type)
		}
	}

	return items, nil
}

// blockDiffHook applies the block-aware Myers edit script, splicing text and
// splitting, joining, or rewriting block markers as required.
type blockDiffHook struct {
	ctx     context.Context
	backend *Backend
	handle  uint32
	old     []blockOrGrapheme
	new     []blockOrGrapheme
	idx     int
	err     error
}

func (h *blockDiffHook) failed() bool {
	return h.err != nil
}

func (h *blockDiffHook) equal(oldIndex, _ int, length int) {
	for i := 0; i < length; i++ {
		h.idx += h.old[oldIndex+i].width()
	}
}

func (h *blockDiffHook) delete(oldIndex, oldLen, _ int) {
	for i := 0; i < oldLen && h.err == nil; i++ {
		item := h.old[oldIndex+i]
		if item.isBlock {
			h.err = h.backend.JoinBlock(h.ctx, h.handle, uint32(h.idx))

			continue
		}

		h.err = h.backend.SpliceText(h.ctx, h.handle, uint32(h.idx), int32(item.width()), "")
	}
}

func (h *blockDiffHook) insert(_ int, newIndex, newLen int) {
	var run strings.Builder

	flush := func() {
		if run.Len() == 0 || h.err != nil {
			return
		}

		chars := run.String()
		if err := h.backend.SpliceText(h.ctx, h.handle, uint32(h.idx), 0, chars); err != nil {
			h.err = err

			return
		}

		h.idx += utf16Width(chars)
		run.Reset()
	}

	for i := 0; i < newLen && h.err == nil; i++ {
		item := h.new[newIndex+i]
		if !item.isBlock {
			run.WriteString(item.grapheme)

			continue
		}

		flush()

		if h.err != nil {
			return
		}

		blockHandle, err := h.backend.SplitBlock(h.ctx, h.handle, uint32(h.idx))
		if err != nil {
			h.err = err

			return
		}

		if err := h.backend.setBlockAttributes(h.ctx, blockHandle, item.block); err != nil {
			h.err = err

			return
		}

		h.idx++
	}

	flush()
}

// desiredMarks flattens the marks named on text spans into absolute UTF-16
// ranges in the order they appear.
func desiredMarks(spans []updateSpanInput) ([]desiredMark, error) {
	marks := make([]desiredMark, 0)
	index := uint32(0)

	for _, span := range spans {
		if span.Type == "block" {
			index++

			continue
		}

		width := uint32(utf16Width(span.Text))

		names := make([]string, 0, len(span.Marks))
		for name := range span.Marks {
			names = append(names, name)
		}

		sort.Strings(names)

		for _, name := range names {
			value, err := decodeScalarWire(span.Marks[name])
			if err != nil {
				return nil, fmt.Errorf("cannot decode mark %q value: %w", name, err)
			}

			marks = append(marks, desiredMark{
				name:  name,
				value: value,
				start: index,
				end:   index + width,
			})
		}

		index += width
	}

	return marks, nil
}

// reconcileMarks removes marks that are not desired and adds the ones that are
// missing, matching the two-phase reconciliation upstream performs.
func (b *Backend) reconcileMarks(
	ctx context.Context,
	handle uint32,
	object ObjectID,
	desired []desiredMark,
	config updateSpansConfigInput,
) error {
	for _, current := range b.state.Marks(object.OpID) {
		keep := false

		for _, want := range desired {
			if want.name == current.Name &&
				want.start == current.Start &&
				want.end == current.End &&
				current.Value != nil &&
				scalarValuesEqual(want.value, *current.Value) {
				keep = true

				break
			}
		}

		if keep {
			continue
		}

		if err := b.markRange(
			ctx,
			handle,
			current.Start,
			current.End,
			current.Name,
			Scalar{Type: ScalarNull},
			config.expandFor(current.Name),
		); err != nil {
			return err
		}
	}

	for _, want := range desired {
		exists := false

		for _, current := range b.state.Marks(object.OpID) {
			if want.name == current.Name &&
				want.start == current.Start &&
				want.end == current.End &&
				current.Value != nil &&
				scalarValuesEqual(want.value, *current.Value) {
				exists = true

				break
			}
		}

		if exists {
			continue
		}

		if err := b.markRange(
			ctx,
			handle,
			want.start,
			want.end,
			want.name,
			want.value,
			config.expandFor(want.name),
		); err != nil {
			return err
		}
	}

	return nil
}

// setBlockAttributes writes the attribute map onto a freshly created block
// object, recursing into nested maps and lists.
func (b *Backend) setBlockAttributes(
	ctx context.Context,
	handle uint32,
	attributes map[string]any,
) error {
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if err := b.setMapValue(ctx, handle, key, attributes[key]); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) setMapValue(
	ctx context.Context,
	handle uint32,
	key string,
	value any,
) error {
	switch typed := value.(type) {
	case map[string]any:
		child, err := b.PutObject(ctx, handle, key, "map")
		if err != nil {
			return err
		}

		return b.setBlockAttributes(ctx, child, typed)
	case []any:
		child, err := b.PutObject(ctx, handle, key, "list")
		if err != nil {
			return err
		}

		for index, element := range typed {
			if err := b.insertListValue(ctx, child, uint64(index), element); err != nil {
				return err
			}
		}

		return nil
	default:
		scalar, err := hydrateScalar(value)
		if err != nil {
			return err
		}

		encoded, err := encodeScalarWire(scalar)
		if err != nil {
			return err
		}

		return b.PutScalar(ctx, handle, key, encoded)
	}
}

func (b *Backend) insertListValue(
	ctx context.Context,
	handle uint32,
	index uint64,
	value any,
) error {
	switch typed := value.(type) {
	case map[string]any:
		child, err := b.InsertObject(ctx, handle, index, "map")
		if err != nil {
			return err
		}

		return b.setBlockAttributes(ctx, child, typed)
	case []any:
		child, err := b.InsertObject(ctx, handle, index, "list")
		if err != nil {
			return err
		}

		for offset, element := range typed {
			if err := b.insertListValue(ctx, child, uint64(offset), element); err != nil {
				return err
			}
		}

		return nil
	default:
		scalar, err := hydrateScalar(value)
		if err != nil {
			return err
		}

		encoded, err := encodeScalarWire(scalar)
		if err != nil {
			return err
		}

		return b.InsertScalar(ctx, handle, index, encoded)
	}
}

// hydrateScalar maps a decoded JSON scalar to an Automerge scalar, treating
// integral numbers as integers to match the reference block hydration.
func hydrateScalar(value any) (Scalar, error) {
	switch typed := value.(type) {
	case nil:
		return Scalar{Type: ScalarNull}, nil
	case bool:
		if typed {
			return Scalar{Type: ScalarTrue, Bool: true}, nil
		}

		return Scalar{Type: ScalarFalse}, nil
	case string:
		return Scalar{Type: ScalarString, String: typed}, nil
	case float64:
		if typed == float64(int64(typed)) {
			return Scalar{Type: ScalarInt, Int: int64(typed)}, nil
		}

		return Scalar{Type: ScalarFloat64, Float: typed}, nil
	default:
		return Scalar{}, fmt.Errorf("unsupported block attribute value %T", value)
	}
}

func (b *Backend) markRange(
	ctx context.Context,
	handle uint32,
	start uint32,
	end uint32,
	name string,
	value Scalar,
	expand string,
) error {
	encoded, err := encodeScalarWire(value)
	if err != nil {
		return err
	}

	return b.MarkText(ctx, handle, start, end, name, encoded, expand)
}

func (c updateSpansConfigInput) expandFor(name string) string {
	if expand, ok := c.PerMarkExpands[name]; ok && expand != "" {
		return expand
	}

	if c.DefaultExpand != "" {
		return c.DefaultExpand
	}

	return "after"
}

func (b *Backend) MarkText(
	ctx context.Context,
	handle uint32,
	start uint32,
	end uint32,
	name string,
	encoded []byte,
	expand string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if start > end {
		return fmt.Errorf("mark range is inverted")
	}

	if start == end && expand == "none" {
		return nil
	}

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	startKey, err := b.textMarkKey(object, start)
	if err != nil {
		return err
	}

	endKey, err := b.textMarkKey(object, end)
	if err != nil {
		return err
	}

	expandBefore, expandAfter, err := markExpansion(expand)
	if err != nil {
		return err
	}

	begin := Operation{
		ID:         b.nextOperationID(),
		Object:     object,
		Key:        startKey,
		Insert:     true,
		Action:     ActionMark,
		Value:      &value,
		MarkExpand: &expandBefore,
		MarkName:   &name,
	}
	if err := b.addPending(begin); err != nil {
		return err
	}

	endOperation := Operation{
		ID:         b.nextOperationID(),
		Object:     object,
		Key:        endKey,
		Insert:     true,
		Action:     ActionMark,
		Value:      &Scalar{Type: ScalarNull},
		MarkExpand: &expandAfter,
	}

	return b.addPending(endOperation)
}

func (b *Backend) SplitBlock(
	ctx context.Context,
	handle uint32,
	index uint32,
) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return 0, err
	}

	sequence := b.state.sequenceElements(object.OpID)

	_, previous, err := richTextPosition(sequence, index)
	if err != nil {
		return 0, err
	}

	key := Key{IsHead: previous == nil}
	if previous != nil {
		key.Element = new(*previous)
	}

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: object,
		Key:    key,
		Insert: true,
		Action: ActionMakeMap,
	}
	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(ObjectID{OpID: operation.ID}), nil
}

func (b *Backend) JoinBlock(
	ctx context.Context,
	handle uint32,
	index uint32,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return err
	}

	sequence := b.state.sequenceElements(object.OpID)

	target, _, err := richTextPosition(sequence, index)
	if err != nil {
		return err
	}

	if target == nil || target.Action != ActionMakeMap {
		return fmt.Errorf("text position %d is not a block", index)
	}

	return b.addPending(Operation{
		ID:           b.nextOperationID(),
		Object:       object,
		Key:          Key{Element: new(target.ID)},
		Action:       ActionDelete,
		Predecessors: []OpID{target.ID},
	})
}

func (b *Backend) ReplaceBlock(
	ctx context.Context,
	handle uint32,
	index uint32,
) (uint32, error) {
	if err := b.JoinBlock(ctx, handle, index); err != nil {
		return 0, err
	}

	return b.SplitBlock(ctx, handle, index)
}

func (b *Backend) Text(ctx context.Context, handle uint32) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return "", err
	}

	var output strings.Builder

	// Materialize the winning value of every visible element so that a put over
	// a text position replaces the original character, matching the reference.
	for _, value := range b.state.sequenceValues(object.OpID) {
		operation := value.Operation
		if operation.Value != nil && operation.Value.Type == ScalarString {
			output.WriteString(operation.Value.String)
		}
	}

	return output.String(), nil
}

func (b *Backend) TextAt(
	ctx context.Context,
	handle uint32,
	heads [][32]byte,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return "", err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return "", fmt.Errorf("historical heads are unknown")
	}

	var output strings.Builder

	for _, operation := range historical.sequence(object.OpID) {
		if operation.Value != nil && operation.Value.Type == ScalarString {
			output.WriteString(operation.Value.String)
		}
	}

	return output.String(), nil
}

func (b *Backend) TextSpans(
	ctx context.Context,
	handle uint32,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	spans, err := b.state.RichTextSpans(object.OpID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(spans)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native rich-text spans: %w", err)
	}

	return data, nil
}

func (b *Backend) TextSpansAt(
	ctx context.Context,
	handle uint32,
	heads [][32]byte,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	spans, err := historical.RichTextSpans(object.OpID)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(spans)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native historical rich-text spans: %w", err)
	}

	return data, nil
}

func encodeMarks(marks []MarkRange) ([]byte, error) {
	type markWire struct {
		Start uint32          `json:"start"`
		End   uint32          `json:"end"`
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}

	wire := make([]markWire, 0, len(marks))

	for _, mark := range marks {
		value := &Scalar{Type: ScalarNull}
		if mark.Value != nil {
			value = mark.Value
		}

		encoded, err := encodeScalarWire(*value)
		if err != nil {
			return nil, err
		}

		wire = append(wire, markWire{
			Start: mark.Start,
			End:   mark.End,
			Name:  mark.Name,
			Value: json.RawMessage(encoded),
		})
	}

	data, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native marks: %w", err)
	}

	return data, nil
}

func (b *Backend) Marks(ctx context.Context, handle uint32) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	return encodeMarks(b.state.Marks(object.OpID))
}

func (b *Backend) MarksAt(
	ctx context.Context,
	handle uint32,
	heads [][32]byte,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	return encodeMarks(historical.Marks(object.OpID))
}

func (b *Backend) TextCursor(
	ctx context.Context,
	handle uint32,
	index uint32,
) ([]byte, error) {
	return b.TextCursorMoving(ctx, handle, index, false)
}

func (b *Backend) TextCursorMoving(
	ctx context.Context,
	handle uint32,
	index uint32,
	moveBefore bool,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	sequence := b.state.sequence(object.OpID)

	position := uint32(0)

	for _, operation := range sequence {
		length := uint32(utf16Length(operation))
		if index >= position && index < position+length {
			data := []byte{1, 3}
			data = appendLengthPrefixedNative(data, operation.ID.Actor.Bytes())

			data = appendULEB(data, operation.ID.Counter)
			if moveBefore {
				data = append(data, 1)
			} else {
				data = append(data, 2)
			}

			return data, nil
		}

		position += length
	}

	return nil, fmt.Errorf("text cursor index %d is out of bounds", index)
}

func (b *Backend) TextCursorMovingAt(
	ctx context.Context,
	handle uint32,
	index uint32,
	moveBefore bool,
	heads [][32]byte,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return nil, err
	}

	historical, ok := b.state.at(nativeHashes(heads))
	if !ok {
		return nil, fmt.Errorf("historical heads are unknown")
	}

	position := uint32(0)

	for _, operation := range historical.sequence(object.OpID) {
		length := uint32(utf16Length(operation))
		if index >= position && index < position+length {
			data := []byte{1, 3}
			data = appendLengthPrefixedNative(data, operation.ID.Actor.Bytes())
			data = appendULEB(data, operation.ID.Counter)

			if moveBefore {
				data = append(data, 1)
			} else {
				data = append(data, 2)
			}

			return data, nil
		}

		position += length
	}

	return nil, fmt.Errorf("text cursor index %d is out of bounds", index)
}

func (b *Backend) TextCursorPosition(
	ctx context.Context,
	handle uint32,
	cursor []byte,
) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	object, err := b.textObject(handle)
	if err != nil {
		return 0, err
	}

	if bytes.Equal(cursor, []byte{1, 1}) {
		return 0, nil
	}

	if bytes.Equal(cursor, []byte{1, 2}) {
		var length uint32
		for _, operation := range b.state.sequence(object.OpID) {
			length += uint32(utf16Length(operation))
		}

		return length, nil
	}

	target, move, err := decodeCursor(cursor)
	if err != nil {
		return 0, err
	}

	position := uint32(0)

	for _, operation := range b.state.sequenceAll(object.OpID) {
		if operation.ID == target {
			if b.state.isSuperseded(operation.ID) && move == 1 {
				return b.cursorMoveBeforePosition(object.OpID, operation)
			}

			return position, nil
		}

		if !b.state.isSuperseded(operation.ID) {
			position += uint32(utf16Length(operation))
		}
	}

	return 0, fmt.Errorf("text cursor target does not exist")
}

func (b *Backend) cursorMoveBeforePosition(
	object OpID,
	target Operation,
) (uint32, error) {
	visited := make(map[OpID]struct{})

	for {
		if target.Key.IsHead {
			return 0, nil
		}

		if target.Key.Element == nil {
			return 0, fmt.Errorf("text cursor target has no predecessor")
		}

		if _, ok := visited[*target.Key.Element]; ok {
			return 0, fmt.Errorf("text cursor predecessor cycle")
		}

		visited[*target.Key.Element] = struct{}{}

		var position uint32

		for _, operation := range b.state.sequence(object) {
			if operation.ID == *target.Key.Element {
				return position, nil
			}

			position += uint32(utf16Length(operation))
		}

		predecessor, ok := b.state.operations[*target.Key.Element]
		if !ok {
			return 0, fmt.Errorf("text cursor predecessor does not exist")
		}

		target = predecessor
	}
}

// changeDependencies computes the dependency set for a new change authored by
// this backend's actor at the given sequence number. The dependencies are the
// current heads plus, matching upstream Rust, the actor's own previous change
// hash when it is not already a head (so that direct causal succession from the
// author's prior change is always recorded explicitly).
func (b *Backend) changeDependencies(sequence uint64) []ChangeHash {
	dependencies := b.state.Heads()

	if sequence > 1 {
		last, ok := b.state.hashForActorSequence(b.actor, sequence-1)
		if ok && !containsHash(dependencies, last) {
			dependencies = append(dependencies, last)
			sort.Slice(dependencies, func(i, j int) bool {
				return bytes.Compare(dependencies[i][:], dependencies[j][:]) < 0
			})
		}
	}

	return dependencies
}

func containsHash(hashes []ChangeHash, target ChangeHash) bool {
	for _, hash := range hashes {
		if hash == target {
			return true
		}
	}

	return false
}

// Isolate pins the document to the given heads: subsequent reads reflect that
// frontier plus isolated writes, and new changes branch from it using a derived
// isolation actor so they never collide with the base actor's later history. It
// mirrors Rust's AutoCommit::isolate. Repeated calls re-pin to fresh heads.
func (b *Backend) Isolate(ctx context.Context, heads [][32]byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(b.pending) > 0 {
		if _, err := b.Commit(ctx, "", time.Time{}); err != nil {
			return err
		}
	}

	full := b.fullState
	if !b.isolationActive {
		full = b.state
	}

	nativeHeads := nativeHashes(heads)

	pinned, ok := full.at(nativeHeads)
	if !ok {
		return fmt.Errorf("isolation heads are unknown")
	}

	baseActor := b.baseActor
	if !b.isolationActive {
		baseActor = b.actor
	}

	b.isolationActive = true
	b.fullState = full
	b.baseActor = baseActor
	b.state = pinned
	b.actor = isolationActor(full, pinned, baseActor)
	b.nextOp = full.maxOpGlobal() + 1

	b.isolationDiffTargets = append(
		b.isolationDiffTargets,
		append([][32]byte(nil), nativeToArrayHeads(nativeHeads)...),
	)

	return nil
}

// nativeToArrayHeads converts change hashes to the [32]byte head form used by
// the incremental diff cursor.
func nativeToArrayHeads(heads []ChangeHash) [][32]byte {
	result := make([][32]byte, len(heads))
	for i, hash := range heads {
		result[i] = [32]byte(hash)
	}

	return result
}

// Integrate ends isolation, returning reads and writes to the full history that
// accumulated every isolated and merged change. It mirrors AutoCommit::integrate.
func (b *Backend) Integrate(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if !b.isolationActive {
		return nil
	}

	if len(b.pending) > 0 {
		if _, err := b.Commit(ctx, "", time.Time{}); err != nil {
			return err
		}
	}

	b.state = b.fullState
	b.actor = b.baseActor
	b.fullState = nil
	b.isolationActive = false
	b.nextOp = b.state.maxOpGlobal() + 1

	return nil
}

// isolationActor selects the actor for isolated writes: the base actor when all
// of its operations are already covered by the isolation heads, otherwise the
// lowest-level derived concurrency actor whose operations are covered, matching
// Rust's isolate_actor.
func isolationActor(full, pinned *State, base ActorID) ActorID {
	for level := uint64(0); ; level++ {
		candidate := base.WithConcurrency(level)
		if full.maxOpForActor(candidate) == pinned.maxOpForActor(candidate) {
			return candidate
		}
	}
}

func (b *Backend) Commit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if err := ctx.Err(); err != nil {
		return [32]byte{}, err
	}

	if len(b.pending) == 0 {
		return [32]byte{}, fmt.Errorf("change contains no operations")
	}

	sequence := b.state.sequenceForActor(b.actor) + 1
	dependencies := b.changeDependencies(sequence)

	change := &Change{
		Actor:        b.actor,
		Sequence:     sequence,
		StartOp:      b.pending[0].ID.Counter,
		MaxOp:        b.pending[len(b.pending)-1].ID.Counter,
		Time:         timestamp.Unix(),
		Message:      message,
		Dependencies: dependencies,
		Operations:   append([]Operation(nil), b.pending...),
	}
	if timestamp.IsZero() {
		change.Time = 0
	}

	raw, err := EncodeChange(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf("cannot encode native change: %w", err)
	}

	if err := b.state.recordAppliedChange(change); err != nil {
		return [32]byte{}, err
	}

	// While isolated, the pinned view holds the change for subsequent reads, but
	// the full history must also record it so integration sees every isolated
	// change alongside merges. Decode a fresh copy from the encoded bytes so the
	// two states never share mutable operation state.
	if b.isolationActive && b.fullState != nil {
		document, err := DecodePartial(raw)
		if err != nil || len(document.Changes) == 0 {
			return [32]byte{}, fmt.Errorf("cannot decode isolated change for full history: %w", err)
		}

		fullChange := document.Changes[0]

		fullChange.Raw = append([]byte(nil), raw...)

		if err := b.fullState.ApplyChange(&fullChange); err != nil {
			return [32]byte{}, err
		}

		if next := b.fullState.maxOpGlobal() + 1; next > b.nextOp {
			b.nextOp = next
		}
	}

	b.appended = append(b.appended, raw)
	b.pending = nil

	return [32]byte(*change.Hash), nil
}

func (b *Backend) EmptyCommit(
	ctx context.Context,
	message string,
	timestamp time.Time,
) ([32]byte, error) {
	if err := ctx.Err(); err != nil {
		return [32]byte{}, err
	}

	if len(b.pending) != 0 {
		return [32]byte{}, fmt.Errorf("cannot create empty change with pending operations")
	}

	sequence := b.state.sequenceForActor(b.actor) + 1

	change := &Change{
		Actor:        b.actor,
		Sequence:     sequence,
		StartOp:      b.nextOp,
		MaxOp:        b.nextOp - 1,
		Time:         timestamp.Unix(),
		Message:      message,
		Dependencies: b.changeDependencies(sequence),
	}
	if timestamp.IsZero() {
		change.Time = 0
	}

	raw, err := EncodeChange(change)
	if err != nil {
		return [32]byte{}, fmt.Errorf("cannot encode native empty change: %w", err)
	}

	if err := b.state.recordAppliedChange(change); err != nil {
		return [32]byte{}, err
	}

	b.appended = append(b.appended, raw)

	return [32]byte(*change.Hash), nil
}

func (b *Backend) Rollback(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if len(b.pending) == 0 {
		return 0, nil
	}

	data := append([]byte(nil), b.base...)
	for _, change := range b.appended {
		data = append(data, change...)
	}

	document, err := Decode(data)
	if err != nil {
		return 0, fmt.Errorf("cannot decode committed state during rollback: %w", err)
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return 0, fmt.Errorf("cannot restore committed state during rollback: %w", err)
	}

	cancelled := uint64(len(b.pending))
	b.state = state
	b.nextOp = state.maxOpGlobal() + 1
	b.pending = nil
	b.objects = map[uint32]ObjectID{0: RootObject()}
	b.nextHandle = 1

	return cancelled, nil
}

func (b *Backend) Stats(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	actors := make(map[ActorID]struct{})
	for id := range b.state.operations {
		actors[id.Actor] = struct{}{}
	}

	stats := struct {
		NumChanges uint64 `json:"numChanges"`
		NumOps     uint64 `json:"numOps"`
		NumActors  uint64 `json:"numActors"`
	}{
		NumChanges: uint64(len(b.state.changes)),
		NumOps:     uint64(len(b.state.operations)),
		NumActors:  uint64(len(actors)),
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native stats: %w", err)
	}

	return data, nil
}

type (
	patchOut struct {
		Obj    string         `json:"obj"`
		Action patchActionOut `json:"action"`
	}

	patchActionOut struct {
		Type     string           `json:"type"`
		Key      string           `json:"key,omitempty"`
		Index    uint64           `json:"index"`
		Length   uint64           `json:"length,omitempty"`
		Value    *patchValueOut   `json:"value,omitempty"`
		Values   []patchInsertOut `json:"values,omitempty"`
		Text     string           `json:"text,omitempty"`
		Conflict bool             `json:"conflict"`
		Marks    []markPatchOut   `json:"marks,omitempty"`
	}

	markPatchOut struct {
		Start uint32          `json:"start"`
		End   uint32          `json:"end"`
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}

	patchInsertOut struct {
		Value    patchValueOut `json:"value"`
		Conflict bool          `json:"conflict"`
	}

	patchValueOut struct {
		Scalar json.RawMessage `json:"scalar,omitempty"`
		Object string          `json:"object,omitempty"`
		ID     string          `json:"id,omitempty"`
	}
)

func objectIDString(object ObjectID) string {
	if object.IsRoot {
		return "_root"
	}

	return fmt.Sprintf(
		"%d@%s",
		object.OpID.Counter,
		hex.EncodeToString([]byte(object.OpID.Actor)),
	)
}

func patchValueForOperation(state *State, operation Operation) (patchValueOut, error) {
	if objectType, err := actionObjectType(operation.Action); err == nil {
		return patchValueOut{
			Object: objectType,
			ID:     objectIDString(ObjectID{OpID: operation.ID}),
		}, nil
	}

	value, ok := state.scalarValue(operation)
	if !ok {
		return patchValueOut{}, fmt.Errorf("operation %v has no materializable value", operation.ID)
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return patchValueOut{}, err
	}

	return patchValueOut{Scalar: json.RawMessage(encoded)}, nil
}

// CurrentState returns the patches that materialize the document from empty,
// ordered to match upstream Rust: the root first, then other objects by
// creation operation ID, with map keys in lexical order and sequence elements
// in index order.
func (b *Backend) CurrentState(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	patches := make([]patchOut, 0)

	for _, object := range orderedObjectsInState(b.state) {
		objectPatches, err := materializeObjectPatches(b.state, object)
		if err != nil {
			return nil, err
		}

		patches = append(patches, objectPatches...)
	}

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native patches: %w", err)
	}

	return data, nil
}

// orderedObjectsInState returns the visible objects in a state, the root first
// and then every non-deleted composite object ordered by its creation ID.
func orderedObjectsInState(state *State) []ObjectID {
	objects := []ObjectID{RootObject()}

	makers := make([]Operation, 0)

	for _, operation := range state.operations {
		if _, err := actionObjectType(operation.Action); err != nil {
			continue
		}

		if state.isSuperseded(operation.ID) {
			continue
		}

		// A composite object concurrently assigned to the same map key as another
		// object is a conflict alternative; only the winning value is materialized
		// (its content spliced), matching the reference. Losing alternatives still
		// exist but are surfaced only through the put's conflict flag.
		if !operation.Insert && operation.Key.Property != nil {
			if winner, ok := state.visibleMapObjectValue(operation.Object, *operation.Key.Property); ok &&
				winner.ID != operation.ID {
				continue
			}
		}

		makers = append(makers, operation)
	}

	sort.Slice(makers, func(i, j int) bool {
		return makers[i].ID.Compare(makers[j].ID) < 0
	})

	for _, operation := range makers {
		objects = append(objects, ObjectID{OpID: operation.ID})
	}

	return objects
}

func objectTypeInState(state *State, object ObjectID) (string, error) {
	if object.IsRoot {
		return "map", nil
	}

	operation, ok := state.operations[object.OpID]
	if !ok {
		return "", fmt.Errorf("object %v does not exist", object.OpID)
	}

	return actionObjectType(operation.Action)
}

func objectVisibleInState(state *State, object ObjectID) bool {
	if object.IsRoot {
		return true
	}

	if _, ok := state.operations[object.OpID]; !ok {
		return false
	}

	return !state.isSuperseded(object.OpID)
}

// materializeObjectPatches emits the patches that build an object from empty.
func materializeObjectPatches(state *State, object ObjectID) ([]patchOut, error) {
	objectType, err := objectTypeInState(state, object)
	if err != nil {
		return nil, err
	}

	identifier := objectIDString(object)

	switch objectType {
	case "map", "table":
		patches := make([]patchOut, 0)

		for _, key := range state.mapKeys(object) {
			winner, ok := state.visibleMapObjectValue(object, key)
			if !ok {
				continue
			}

			value, err := patchValueForOperation(state, winner)
			if err != nil {
				return nil, err
			}

			patches = append(patches, patchOut{
				Obj: identifier,
				Action: patchActionOut{
					Type:     "put_map",
					Key:      key,
					Value:    &value,
					Conflict: len(state.visibleMapObjectOperations(object, key)) > 1,
				},
			})
		}

		return patches, nil
	case "list":
		values := state.sequenceValues(object.OpID)
		if len(values) == 0 {
			return nil, nil
		}

		inserts := make([]patchInsertOut, 0, len(values))

		for index := range values {
			value, err := patchValueForOperation(state, values[index].Operation)
			if err != nil {
				return nil, err
			}

			inserts = append(inserts, patchInsertOut{
				Value:    value,
				Conflict: len(state.visibleSequenceElementOperations(values[index].Element)) > 1,
			})
		}

		return []patchOut{{
			Obj:    identifier,
			Action: patchActionOut{Type: "insert", Index: 0, Values: inserts},
		}}, nil
	case "text":
		patches := make([]patchOut, 0)
		position := uint64(0)

		var run strings.Builder

		runStart := uint64(0)

		flush := func() error {
			if run.Len() == 0 {
				return nil
			}

			runs, err := textRunsWithMarks(state, object, runStart, run.String())
			if err != nil {
				return err
			}

			for _, textRun := range runs {
				patches = append(patches, patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:  "splice_text",
						Index: textRun.index,
						Text:  textRun.text,
						Marks: textRun.marks,
					},
				})
			}

			run.Reset()

			return nil
		}

		for _, value := range state.sequenceValues(object.OpID) {
			operation := value.Operation

			if operation.Action == ActionMakeMap {
				if err := flush(); err != nil {
					return nil, err
				}

				blockValue, err := patchValueForOperation(state, operation)
				if err != nil {
					return nil, err
				}

				patches = append(patches, patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:   "insert",
						Index:  position,
						Values: []patchInsertOut{{Value: blockValue}},
					},
				})
				position++

				continue
			}

			if operation.Value != nil && operation.Value.Type == ScalarString {
				if run.Len() == 0 {
					runStart = position
				}

				run.WriteString(operation.Value.String)
				position += uint64(utf16Width(operation.Value.String))
			}
		}

		if err := flush(); err != nil {
			return nil, err
		}

		return patches, nil
	default:
		return nil, fmt.Errorf("unknown object type %q", objectType)
	}
}

// Diff returns the patches that transform the document state at the before heads
// into the state at the after heads.
func (b *Backend) Diff(
	ctx context.Context,
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	patches, err := b.diffPatches(beforeHeads, afterHeads, false)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native diff patches: %w", err)
	}

	return data, nil
}

// UpdateDiffCursor records the current heads as the incremental diff cursor so a
// following DiffIncremental reports only the changes committed since this call.
func (b *Backend) UpdateDiffCursor(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	heads, err := b.Heads(ctx)
	if err != nil {
		return err
	}

	b.diffCursor = heads
	b.isolationDiffTargets = nil

	return nil
}

// DiffIncremental returns the patches for the changes committed since the diff
// cursor (or from an empty document when the cursor is unset) and advances the
// cursor to the current heads, mirroring the reference diff_incremental helper.
func (b *Backend) DiffIncremental(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	heads, err := b.Heads(ctx)
	if err != nil {
		return nil, err
	}

	patches, err := b.incrementalDiffPatches(heads)
	if err != nil {
		return nil, err
	}

	b.diffCursor = heads
	b.isolationDiffTargets = nil

	data, err := json.Marshal(patches)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native incremental diff patches: %w", err)
	}

	return data, nil
}

// incrementalDiffPatches computes the incremental patches from the diff cursor
// to the current heads. When isolation frontiers were recorded in the window,
// the diff is chained through each of them (cursor to each isolation frontier
// and finally to the current heads) so the patch stream matches the reference's
// patch-log output across isolate/integrate.
func (b *Backend) incrementalDiffPatches(heads [][32]byte) ([]patchOut, error) {
	if len(b.isolationDiffTargets) == 0 {
		return b.incrementalPatches(b.diffCursor, heads)
	}

	frontiers := make([][][32]byte, 0, len(b.isolationDiffTargets)+2)
	frontiers = append(frontiers, b.diffCursor)
	frontiers = append(frontiers, b.isolationDiffTargets...)
	frontiers = append(frontiers, heads)

	patches := make([]patchOut, 0)

	for i := 0; i+1 < len(frontiers); i++ {
		segment, err := b.incrementalPatches(frontiers[i], frontiers[i+1])
		if err != nil {
			return nil, err
		}

		patches = append(patches, segment...)
	}

	return patches, nil
}

func (b *Backend) incrementalPatches(
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
) ([]patchOut, error) {
	return b.diffPatches(beforeHeads, afterHeads, true)
}

func (b *Backend) diffPatches(
	beforeHeads [][32]byte,
	afterHeads [][32]byte,
	incremental bool,
) ([]patchOut, error) {
	source, ok := b.state.at(nativeHashes(beforeHeads))
	if !ok {
		return nil, fmt.Errorf("before heads are unknown")
	}

	target, ok := b.state.at(nativeHashes(afterHeads))
	if !ok {
		return nil, fmt.Errorf("after heads are unknown")
	}

	patches := make([]patchOut, 0)

	for _, object := range orderedObjectsInState(target) {
		var (
			objectPatches []patchOut
			err           error
		)

		if objectVisibleInState(source, object) {
			objectPatches, err = diffObjectPatches(source, target, object, incremental)
		} else {
			objectPatches, err = materializeObjectPatches(target, object)
		}

		if err != nil {
			return nil, err
		}

		patches = append(patches, objectPatches...)
	}

	return patches, nil
}

// diffObjectPatches emits patches transforming an object from the source state
// into the target state, for an object present in both.
func diffObjectPatches(source, target *State, object ObjectID, incremental bool) ([]patchOut, error) {
	objectType, err := objectTypeInState(target, object)
	if err != nil {
		return nil, err
	}

	identifier := objectIDString(object)

	switch objectType {
	case "map", "table":
		return diffMapPatches(source, target, object, identifier)
	case "list":
		return diffSequencePatches(source, target, object, objectType, identifier, incremental)
	case "text":
		patches, err := diffSequencePatches(source, target, object, objectType, identifier, incremental)
		if err != nil {
			return nil, err
		}

		return mergeTextMarkPatches(source, target, object, identifier, patches)
	default:
		return nil, fmt.Errorf("unknown object type %q", objectType)
	}
}

// mergeTextMarkPatches folds the mark differences between the source and target
// states into the ordered sequence patches for a text object. Added or changed
// marks carry their new value; removed marks carry a null value. The reference
// emits a single mark patch positioned by the smallest affected index, so the
// combined patch is inserted before the first sequence patch beyond that index.
func mergeTextMarkPatches(
	source, target *State,
	object ObjectID,
	identifier string,
	patches []patchOut,
) ([]patchOut, error) {
	// The reference derives mark patches from the mark operations applied in the
	// window, not from state comparison, so a mark range that merely grew because
	// text was spliced into an expanding mark produces no mark patch (the marks
	// ride on the splice patch instead), and a partial unmark reports the literal
	// operation range rather than the resulting split.
	marks, err := diffMarkPatches(source, target, object)
	if err != nil {
		return nil, err
	}

	if len(marks) == 0 {
		return patches, nil
	}

	anchor := marks[0].Start
	for _, mark := range marks[1:] {
		if mark.Start < anchor {
			anchor = mark.Start
		}
	}

	markPatch := patchOut{
		Obj:    identifier,
		Action: patchActionOut{Type: "mark", Marks: marks},
	}

	insertAt := len(patches)

	for index, patch := range patches {
		if patch.Action.Index > uint64(anchor) {
			insertAt = index

			break
		}
	}

	merged := make([]patchOut, 0, len(patches)+1)
	merged = append(merged, patches[:insertAt]...)
	merged = append(merged, markPatch)
	merged = append(merged, patches[insertAt:]...)

	return merged, nil
}

// diffMarkPatches reports the mark and unmark operations applied in the window
// between the source and target states as Mark patch entries, ordered by the
// operation identifier so they appear in application order. Each entry carries
// the operation's literal UTF-16 range and value (null for an unmark), matching
// the reference's operation-based diff rather than a state comparison.
func diffMarkPatches(source, target *State, object ObjectID) ([]markPatchOut, error) {
	begins := make([]Operation, 0)

	for id, operation := range target.operations {
		if operation.Action != ActionMark ||
			operation.Object != object ||
			operation.MarkName == nil {
			continue
		}

		if _, ok := source.operations[id]; ok {
			continue
		}

		begins = append(begins, operation)
	}

	sort.Slice(begins, func(i, j int) bool {
		return begins[i].ID.Compare(begins[j].ID) < 0
	})

	out := make([]markPatchOut, 0, len(begins))

	for _, begin := range begins {
		end, ok := target.operations[OpID{Actor: begin.ID.Actor, Counter: begin.ID.Counter + 1}]
		if !ok {
			continue
		}

		start, finish, ok := target.markOpUTF16Range(object.OpID, begin, end)
		if !ok {
			continue
		}

		value := Scalar{Type: ScalarNull}
		if begin.Value != nil {
			value = *begin.Value
		}

		encoded, err := encodeScalarWire(value)
		if err != nil {
			return nil, err
		}

		out = append(out, markPatchOut{
			Start: start,
			End:   finish,
			Name:  *begin.MarkName,
			Value: json.RawMessage(encoded),
		})
	}

	return out, nil
}

type textRun struct {
	index uint64
	text  string
	marks []markPatchOut
}

// textRunsWithMarks splits a run of text starting at the given UTF-16 position
// into maximal sub-runs that share the same active mark set, attaching those
// marks (sorted by name) to each sub-run. It mirrors the reference, which emits
// one splice_text patch per mark run.
func textRunsWithMarks(
	state *State,
	object ObjectID,
	startPosition uint64,
	text string,
) ([]textRun, error) {
	ranges := state.Marks(object.OpID)

	activeMarks := func(position uint64) ([]markPatchOut, string, error) {
		marks := make([]markPatchOut, 0)

		for _, candidate := range ranges {
			if uint64(candidate.Start) > position || position >= uint64(candidate.End) {
				continue
			}

			value := Scalar{Type: ScalarNull}
			if candidate.Value != nil {
				value = *candidate.Value
			}

			encoded, err := encodeScalarWire(value)
			if err != nil {
				return nil, "", err
			}

			marks = append(marks, markPatchOut{
				Name:  candidate.Name,
				Value: json.RawMessage(encoded),
			})
		}

		sort.Slice(marks, func(i, j int) bool {
			return marks[i].Name < marks[j].Name
		})

		var key strings.Builder
		for _, mark := range marks {
			key.WriteString(mark.Name)
			key.WriteByte('=')
			key.Write(mark.Value)
			key.WriteByte(';')
		}

		if len(marks) == 0 {
			marks = nil
		}

		return marks, key.String(), nil
	}

	runs := make([]textRun, 0)

	var (
		builder  strings.Builder
		runMarks []markPatchOut
		runKey   string
		runStart = startPosition
		position = startPosition
		haveRun  bool
	)

	for _, character := range text {
		marks, key, err := activeMarks(position)
		if err != nil {
			return nil, err
		}

		if !haveRun {
			runStart = position
			runMarks = marks
			runKey = key
			haveRun = true
		} else if key != runKey {
			runs = append(runs, textRun{index: runStart, text: builder.String(), marks: runMarks})
			builder.Reset()
			runStart = position
			runMarks = marks
			runKey = key
		}

		builder.WriteRune(character)

		if character > 0xFFFF {
			position += 2
		} else {
			position++
		}
	}

	if haveRun {
		runs = append(runs, textRun{index: runStart, text: builder.String(), marks: runMarks})
	}

	return runs, nil
}

func diffMapPatches(
	source, target *State,
	object ObjectID,
	identifier string,
) ([]patchOut, error) {
	keys := make(map[string]struct{})
	for _, key := range source.mapKeys(object) {
		keys[key] = struct{}{}
	}

	for _, key := range target.mapKeys(object) {
		keys[key] = struct{}{}
	}

	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}

	sort.Strings(ordered)

	patches := make([]patchOut, 0)

	for _, key := range ordered {
		targetOp, targetOK := target.visibleMapObjectValue(object, key)
		sourceOp, sourceOK := source.visibleMapObjectValue(object, key)

		switch {
		case targetOK && (!sourceOK || targetOp.ID != sourceOp.ID):
			value, err := patchValueForOperation(target, targetOp)
			if err != nil {
				return nil, err
			}

			patches = append(patches, patchOut{
				Obj: identifier,
				Action: patchActionOut{
					Type:     "put_map",
					Key:      key,
					Value:    &value,
					Conflict: len(target.visibleMapObjectOperations(object, key)) > 1,
				},
			})
		case !targetOK && sourceOK:
			patches = append(patches, patchOut{
				Obj:    identifier,
				Action: patchActionOut{Type: "delete_map", Key: key},
			})
		}
	}

	return patches, nil
}

func diffSequencePatches(
	source, target *State,
	object ObjectID,
	objectType string,
	identifier string,
	incremental bool,
) ([]patchOut, error) {
	sourceValues := source.sequenceValues(object.OpID)
	targetValues := target.sequenceValues(object.OpID)

	sourceElements := make(map[OpID]struct{}, len(sourceValues))
	for _, value := range sourceValues {
		sourceElements[value.Element] = struct{}{}
	}

	targetElements := make(map[OpID]struct{}, len(targetValues))
	for _, value := range targetValues {
		targetElements[value.Element] = struct{}{}
	}

	// Text objects report positions in UTF-16 code units; other sequences use
	// one unit per element.
	width := func(value sequenceValue) uint64 {
		if objectType == "text" {
			return sequenceValueUTF16Width(value)
		}

		return 1
	}

	patches := make([]patchOut, 0)
	position := uint64(0)
	i, j := 0, 0

	for i < len(sourceValues) || j < len(targetValues) {
		if i < len(sourceValues) && j < len(targetValues) &&
			sourceValues[i].Element == targetValues[j].Element {
			if sourceValues[i].Operation.ID == targetValues[j].Operation.ID {
				position += width(targetValues[j])
				i++
				j++

				continue
			}

			// Same element, different winning value. A state-comparison diff of
			// text cannot express an in-place replacement, so it becomes a
			// delete followed by a splice; the incremental patch log and every
			// list report a put_seq instead, mirroring the reference.
			if objectType == "text" && !incremental {
				patches = append(patches, patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:   "delete_seq",
						Index:  position,
						Length: width(sourceValues[i]),
					},
				})

				operation := targetValues[j].Operation
				if operation.Value != nil && operation.Value.Type == ScalarString {
					runs, err := textRunsWithMarks(target, object, position, operation.Value.String)
					if err != nil {
						return nil, err
					}

					for _, run := range runs {
						patches = append(patches, patchOut{
							Obj: identifier,
							Action: patchActionOut{
								Type:  "splice_text",
								Index: run.index,
								Text:  run.text,
								Marks: run.marks,
							},
						})
					}

					position += width(targetValues[j])
				}
			} else {
				value, err := patchValueForOperation(target, targetValues[j].Operation)
				if err != nil {
					return nil, err
				}

				patches = append(patches, patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:     "put_seq",
						Index:    position,
						Value:    &value,
						Conflict: len(target.visibleSequenceElementOperations(targetValues[j].Element)) > 1,
					},
				})
				position += width(targetValues[j])
			}

			i++
			j++

			continue
		}

		if i < len(sourceValues) {
			if _, ok := targetElements[sourceValues[i].Element]; !ok {
				length := uint64(0)

				for i < len(sourceValues) {
					if _, ok := targetElements[sourceValues[i].Element]; ok {
						break
					}

					length += width(sourceValues[i])
					i++
				}

				patches = append(patches, patchOut{
					Obj: identifier,
					Action: patchActionOut{
						Type:   "delete_seq",
						Index:  position,
						Length: length,
					},
				})

				continue
			}
		}

		if objectType == "text" {
			var text strings.Builder

			start := position

			for j < len(targetValues) {
				if _, ok := sourceElements[targetValues[j].Element]; ok {
					break
				}

				operation := targetValues[j].Operation
				if operation.Value == nil || operation.Value.Type != ScalarString {
					break
				}

				text.WriteString(operation.Value.String)
				position += width(targetValues[j])
				j++
			}

			if text.Len() > 0 {
				runs, err := textRunsWithMarks(target, object, start, text.String())
				if err != nil {
					return nil, err
				}

				for _, run := range runs {
					patches = append(patches, patchOut{
						Obj: identifier,
						Action: patchActionOut{
							Type:  "splice_text",
							Index: run.index,
							Text:  run.text,
							Marks: run.marks,
						},
					})
				}

				continue
			}
		}

		inserts := make([]patchInsertOut, 0)
		start := position

		for j < len(targetValues) {
			if _, ok := sourceElements[targetValues[j].Element]; ok {
				break
			}

			value, err := patchValueForOperation(target, targetValues[j].Operation)
			if err != nil {
				return nil, err
			}

			inserts = append(inserts, patchInsertOut{
				Value:    value,
				Conflict: len(target.visibleSequenceElementOperations(targetValues[j].Element)) > 1,
			})
			position++
			j++
		}

		if len(inserts) == 0 {
			break
		}

		patches = append(patches, patchOut{
			Obj: identifier,
			Action: patchActionOut{
				Type:   "insert",
				Index:  start,
				Values: inserts,
			},
		})
	}

	return patches, nil
}

func (b *Backend) Heads(ctx context.Context) ([][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	heads := b.state.Heads()

	result := make([][32]byte, len(heads))
	for i := range heads {
		result[i] = [32]byte(heads[i])
	}

	return result, nil
}

func (b *Backend) HasHeads(
	ctx context.Context,
	heads [][32]byte,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	for _, head := range heads {
		if !b.state.hasChange(ChangeHash(head)) {
			return false, nil
		}
	}

	return true, nil
}

func (b *Backend) MissingDependencies(
	ctx context.Context,
	heads [][32]byte,
) ([][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	missing := make(map[[32]byte]struct{})

	for _, head := range heads {
		_, queued := b.queuedChanges[ChangeHash(head)]
		if !b.state.hasChange(ChangeHash(head)) && !queued {
			missing[head] = struct{}{}
		}
	}

	for _, change := range b.queuedChanges {
		for _, dependency := range change.Dependencies {
			if !b.state.hasChange(dependency) {
				missing[[32]byte(dependency)] = struct{}{}
			}
		}
	}

	result := make([][32]byte, 0, len(missing))
	for dependency := range missing {
		result = append(result, dependency)
	}

	sort.Slice(result, func(i, j int) bool {
		return bytes.Compare(result[i][:], result[j][:]) < 0
	})

	return result, nil
}

func (b *Backend) ChangesSince(
	ctx context.Context,
	heads [][32]byte,
) ([][]byte, [][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	knownHeads := make([]ChangeHash, len(heads))
	for i, head := range heads {
		knownHeads[i] = ChangeHash(head)
	}

	changes, ok := b.state.changesSince(knownHeads)
	if !ok {
		return nil, nil, fmt.Errorf("cannot compute changes from unknown heads")
	}

	raw := make([][]byte, len(changes))

	hashes := make([][32]byte, len(changes))
	for i, change := range changes {
		if change.Hash == nil {
			return nil, nil, fmt.Errorf("change %d has no hash", i)
		}

		raw[i] = append([]byte(nil), change.Raw...)
		hashes[i] = [32]byte(*change.Hash)
	}

	return raw, hashes, nil
}

func (b *Backend) ApplyChanges(
	ctx context.Context,
	changes [][]byte,
) error {
	for i, change := range changes {
		if _, err := b.Merge(ctx, change); err != nil {
			return fmt.Errorf("cannot apply native change %d: %w", i, err)
		}
	}

	return nil
}

func (b *Backend) Merge(ctx context.Context, data []byte) ([][32]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// While isolated, merged changes belong to the full history rather than the
	// pinned view, so operate on the full state and keep the pinned view intact.
	if b.isolationActive && b.fullState != nil {
		b.state, b.fullState = b.fullState, b.state

		defer func() {
			b.state, b.fullState = b.fullState, b.state
			b.nextOp = b.fullState.maxOpGlobal() + 1
		}()
	}

	document, err := Decode(data)
	if err != nil {
		document, err = DecodePartial(data)
	}

	if err != nil {
		return nil, err
	}

	if len(b.state.Heads()) == 0 && len(b.pending) == 0 {
		state, err := NewStateFromDocument(document)
		if err != nil {
			return nil, fmt.Errorf("cannot initialize merged native state: %w", err)
		}

		b.state = state
		b.nextOp = state.maxOpGlobal() + 1

		b.base = append([]byte(nil), data...)
		b.appended = nil
		b.saveCursor = 0

		return b.Heads(ctx)
	}

	if b.requiresSnapshotMerge(document) {
		if err := b.mergeDocumentSnapshot(data, document); err != nil {
			return nil, err
		}

		return b.Heads(ctx)
	}

	if err := b.applyMergedChanges(document.Changes); err != nil {
		return nil, err
	}

	if next := b.state.maxOpGlobal() + 1; next > b.nextOp {
		b.nextOp = next
	}

	return b.Heads(ctx)
}

func (b *Backend) requiresSnapshotMerge(document *Document) bool {
	if len(document.ChunkTypes) == 0 ||
		document.ChunkTypes[0] != ChunkDocument {
		return false
	}

	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash != nil &&
			!b.state.hasChange(*change.Hash) &&
			len(change.Raw) == 0 {
			return true
		}
	}

	return false
}

func (b *Backend) mergeDocumentSnapshot(
	data []byte,
	document *Document,
) error {
	localChanges, ok := b.state.allChanges()
	if !ok {
		return fmt.Errorf("cannot enumerate local changes for snapshot merge")
	}

	state, err := NewStateFromDocument(document)
	if err != nil {
		return fmt.Errorf("cannot initialize merged snapshot state: %w", err)
	}

	for _, change := range localChanges {
		if change.Hash == nil || state.hasChange(*change.Hash) {
			continue
		}

		incoming := documentChangeByActorSequence(
			document,
			change.Actor,
			change.Sequence,
		)
		if incoming != nil {
			state.changes[*change.Hash] = incoming
		}
	}

	appended := make([][]byte, 0)

	for _, change := range localChanges {
		if change.Hash == nil ||
			state.hasChange(*change.Hash) ||
			documentChangeByActorSequence(
				document,
				change.Actor,
				change.Sequence,
			) != nil {
			continue
		}

		if len(change.Raw) == 0 {
			return fmt.Errorf(
				"cannot preserve local change %s during snapshot merge",
				change.Hash,
			)
		}

		if err := state.ApplyChange(change); err != nil {
			return fmt.Errorf("cannot apply local change to merged snapshot: %w", err)
		}

		appended = append(appended, append([]byte(nil), change.Raw...))
	}

	b.state = state

	b.base = append([]byte(nil), data...)
	b.appended = appended
	b.saveCursor = 0
	b.queuedChanges = make(map[ChangeHash]*Change)
	b.queuedBytes = 0
	b.nextOp = state.maxOpGlobal() + 1

	return nil
}

func documentChangeByActorSequence(
	document *Document,
	actor ActorID,
	sequence uint64,
) *Change {
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Actor == actor && change.Sequence == sequence {
			return change
		}
	}

	return nil
}

func (b *Backend) applyMergedChanges(changes []Change) error {
	for i := range changes {
		change := &changes[i]
		if change.Hash == nil || b.state.hasChange(*change.Hash) {
			continue
		}

		if _, queued := b.queuedChanges[*change.Hash]; queued {
			continue
		}

		if len(change.Raw) == 0 {
			return fmt.Errorf(
				"cannot preserve merged change %s: original bytes are unavailable",
				change.Hash,
			)
		}

		if len(b.queuedChanges) >= maxQueuedChanges ||
			b.queuedBytes+len(change.Raw) > maxQueuedChangeBytes {
			return fmt.Errorf("merged change queue exceeds its resource limit")
		}

		clone := *change
		clone.Raw = append([]byte(nil), change.Raw...)
		b.queuedChanges[*change.Hash] = &clone
		b.queuedBytes += len(clone.Raw)
	}

	for len(b.queuedChanges) > 0 {
		progressed := false

		for hash, change := range b.queuedChanges {
			if !b.state.hasDependencies(change) {
				continue
			}

			if err := b.state.ApplyChange(change); err != nil {
				return fmt.Errorf("cannot apply merged native change: %w", err)
			}

			b.appended = append(
				b.appended,
				append([]byte(nil), change.Raw...),
			)
			b.queuedBytes -= len(change.Raw)
			delete(b.queuedChanges, hash)

			progressed = true
		}

		if !progressed {
			break
		}
	}

	return nil
}

func (b *Backend) NewSyncState(ctx context.Context) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	handle := b.nextSyncState
	b.nextSyncState++
	b.syncStates[handle] = &nativeSyncState{}

	return handle, nil
}

func (b *Backend) CloseSyncState(ctx context.Context, handle uint32) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, ok := b.syncStates[handle]; !ok {
		return fmt.Errorf("invalid sync state %d", handle)
	}

	delete(b.syncStates, handle)

	return nil
}

func (b *Backend) SetSyncReadOnly(
	ctx context.Context,
	handle uint32,
	readOnly bool,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return err
	}

	if state.ReadOnly == readOnly {
		return nil
	}

	if state.ReadOnly && !readOnly {
		peerSupportsReset := state.PeerSupportsReset
		*state = nativeSyncState{
			PeerSupportsReset: peerSupportsReset,
			NeedsReset:        true,
			ModeChanged:       true,
		}
	} else {
		state.ReadOnly = true
		state.InFlight = false
		state.ModeChanged = true
	}

	return nil
}

func (b *Backend) SyncPeerReadOnly(
	ctx context.Context,
	handle uint32,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return false, err
	}

	return state.PeerReadOnly, nil
}

func (b *Backend) GenerateSyncMessage(
	ctx context.Context,
	handle uint32,
) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return nil, false, err
	}

	heads, err := b.Heads(ctx)
	if err != nil {
		return nil, false, err
	}

	// A message is only truly in flight while we have nothing new to say. New
	// local changes (heads advanced past the last sent frontier) must be sent
	// even while a previous message awaits acknowledgement, matching upstream
	// Rust, which never withholds local changes during synchronization.
	if state.InFlight &&
		!state.ModeChanged &&
		!state.NeedsReset &&
		len(state.Requested) == 0 &&
		len(state.Need) == 0 &&
		equalHashes(heads, state.LastSentHeads) {
		return nil, false, nil
	}

	if state.ModeChanged || state.NeedsReset {
		state.InFlight = false
	}

	// The first message for a sync state is always sent so the peer learns our
	// heads and capabilities, matching upstream Rust's first_response_is_some
	// behavior. Subsequent messages may be suppressed when nothing is pending.
	if state.Sent {
		if state.PeerReadOnly &&
			!state.PeerModeChanged &&
			!state.ModeChanged &&
			!state.NeedsReset &&
			!state.NeedsAck &&
			len(state.Requested) == 0 &&
			len(state.Need) == 0 &&
			equalHashes(heads, state.LastSentHeads) {
			return nil, false, nil
		}

		if state.ReadOnly &&
			!state.ModeChanged &&
			!state.NeedsReset &&
			!state.NeedsAck &&
			len(state.Requested) == 0 &&
			len(state.Need) == 0 &&
			equalHashes(heads, state.LastSentHeads) {
			return nil, false, nil
		}

		if !state.NeedsAck &&
			!state.ModeChanged &&
			!state.NeedsReset &&
			len(state.Requested) == 0 &&
			len(state.Need) == 0 &&
			equalHashes(heads, state.RemoteHeads) {
			return nil, false, nil
		}
	}

	flags := byte(syncFlagSupportsReset)
	if state.ReadOnly {
		flags |= syncFlagReadOnly
	}

	messageHeads := heads

	if state.NeedsReset {
		if state.PeerSupportsReset {
			flags |= syncFlagReset
		} else {
			messageHeads = nil
		}
	}

	message := SyncMessage{
		Version: SyncMessageVersion2,
		Heads:   messageHeads,
		Need:    append([][32]byte(nil), state.Need...),
		Flags:   []byte{2, syncFlagMarker | flags},
	}
	if !state.NeedsAck && !state.PeerReadOnly {
		switch {
		case len(state.Requested) > 0:
			for _, requested := range state.Requested {
				change, ok := b.state.changes[ChangeHash(requested)]
				if !ok || len(change.Raw) == 0 {
					continue
				}

				message.Changes = append(
					message.Changes,
					append([]byte(nil), change.Raw...),
				)
			}

			state.Requested = nil
		case len(state.Need) == 0:
			remoteHeads := make([]ChangeHash, len(state.RemoteHeads))
			for i, head := range state.RemoteHeads {
				remoteHeads[i] = ChangeHash(head)
			}

			changes, incremental := b.state.changesSince(remoteHeads)
			if incremental {
				for _, change := range changes {
					message.Changes = append(
						message.Changes,
						append([]byte(nil), change.Raw...),
					)
				}
			} else {
				document, err := b.Save(ctx)
				if err != nil {
					return nil, false, err
				}

				message.Changes = [][]byte{document}
			}
		}
	}

	if !state.NeedsAck {
		state.InFlight = true
	}

	state.NeedsAck = false
	state.Sent = true
	state.LastSentHeads = append(state.LastSentHeads[:0], heads...)
	state.PeerModeChanged = false
	state.ModeChanged = false
	state.NeedsReset = false

	data, err := message.Encode()
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

func (b *Backend) ReceiveSyncMessage(
	ctx context.Context,
	handle uint32,
	data []byte,
) error {
	state, err := b.syncState(handle)
	if err != nil {
		return err
	}

	message, err := ParseSyncMessage(data)
	if err != nil {
		return err
	}

	state.InFlight = false

	flags := syncMessageFlagBits(message.Flags)

	peerReadOnly := flags&syncFlagReadOnly != 0
	if peerReadOnly != state.PeerReadOnly {
		state.PeerModeChanged = true
	}

	state.PeerReadOnly = peerReadOnly

	state.PeerSupportsReset = flags&syncFlagSupportsReset != 0
	if flags&syncFlagReset != 0 {
		state.RemoteHeads = nil
		state.Requested = nil
	}

	if !state.ReadOnly {
		for _, change := range message.Changes {
			if _, err := b.Merge(ctx, change); err != nil {
				return fmt.Errorf("cannot merge native sync payload: %w", err)
			}
		}
	}

	state.RemoteHeads = append([][32]byte(nil), message.Heads...)
	state.Requested = append(state.Requested[:0], message.Need...)

	needed := make(map[[32]byte]struct{})

	if !state.ReadOnly {
		for _, head := range message.Heads {
			if _, ok := b.state.changes[ChangeHash(head)]; !ok {
				needed[head] = struct{}{}
			}
		}

		for _, change := range b.queuedChanges {
			for _, dependency := range change.Dependencies {
				if !b.state.hasChange(dependency) {
					needed[[32]byte(dependency)] = struct{}{}
				}
			}
		}
	}

	state.Need = state.Need[:0]
	for dependency := range needed {
		state.Need = append(state.Need, dependency)
	}

	sort.Slice(state.Need, func(i, j int) bool {
		return bytes.Compare(state.Need[i][:], state.Need[j][:]) < 0
	})

	state.NeedsAck = len(message.Changes) > 0 || state.PeerModeChanged

	return nil
}

func (b *Backend) SaveSyncState(
	ctx context.Context,
	handle uint32,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	state, err := b.syncState(handle)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native sync state: %w", err)
	}

	return data, nil
}

func (b *Backend) LoadSyncState(
	ctx context.Context,
	data []byte,
) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var state nativeSyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return 0, fmt.Errorf("cannot decode native sync state: %w", err)
	}

	// A serialized state cannot retain an in-flight transport message. Allow
	// the restored session to regenerate it instead of waiting forever for an
	// acknowledgement that may have been lost with the previous process.
	state.InFlight = false

	handle := b.nextSyncState
	b.nextSyncState++
	b.syncStates[handle] = &state

	return handle, nil
}

func (b *Backend) addPending(operation Operation) error {
	if err := b.state.applyPending([]Operation{operation}); err != nil {
		return err
	}

	b.pending = append(b.pending, operation)

	return nil
}

func (b *Backend) nextOperationID() OpID {
	id := OpID{
		Actor:   b.actor,
		Counter: b.nextOp,
	}
	b.nextOp++

	return id
}

func (b *Backend) requireRoot(ctx context.Context, handle uint32) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	object, err := b.object(handle)
	if err != nil {
		return err
	}

	if !object.IsRoot {
		return fmt.Errorf("object is not the root map")
	}

	return nil
}

func (b *Backend) object(handle uint32) (ObjectID, error) {
	object, ok := b.objects[handle]
	if !ok {
		return ObjectID{}, fmt.Errorf("invalid object handle %d", handle)
	}

	return object, nil
}

func (b *Backend) mapObject(handle uint32) (ObjectID, error) {
	object, err := b.object(handle)
	if err != nil {
		return ObjectID{}, err
	}

	if object.IsRoot {
		return object, nil
	}

	operation, ok := b.state.operations[object.OpID]
	if !ok ||
		(operation.Action != ActionMakeMap &&
			operation.Action != ActionMakeTable) {
		return ObjectID{}, fmt.Errorf("object is not a map")
	}

	return object, nil
}

func (b *Backend) sequenceObject(handle uint32) (ObjectID, error) {
	object, err := b.object(handle)
	if err != nil {
		return ObjectID{}, err
	}

	if object.IsRoot {
		return ObjectID{}, fmt.Errorf("root map is not a sequence")
	}

	operation, ok := b.state.operations[object.OpID]
	if !ok ||
		(operation.Action != ActionMakeList &&
			operation.Action != ActionMakeText) {
		return ObjectID{}, fmt.Errorf("object is not a sequence")
	}

	return object, nil
}

func (b *Backend) textObject(handle uint32) (ObjectID, error) {
	object, err := b.object(handle)
	if err != nil {
		return ObjectID{}, err
	}

	if object.IsRoot {
		return ObjectID{}, fmt.Errorf("root map is not text")
	}

	operation, ok := b.state.operations[object.OpID]
	if !ok || operation.Action != ActionMakeText {
		return ObjectID{}, fmt.Errorf("object is not text")
	}

	return object, nil
}

func (b *Backend) pushObject(object ObjectID) uint32 {
	handle := b.nextHandle
	b.nextHandle++
	b.objects[handle] = object

	return handle
}

func (b *Backend) syncState(handle uint32) (*nativeSyncState, error) {
	state, ok := b.syncStates[handle]
	if !ok {
		return nil, fmt.Errorf("invalid sync state %d", handle)
	}

	return state, nil
}

func (b *Backend) rootTextObjects() map[string]Operation {
	objects := make(map[string]Operation)

	for _, operation := range b.state.operations {
		if operation.Object.IsRoot &&
			operation.Key.Property != nil &&
			operation.Action == ActionMakeText &&
			!b.state.isSuperseded(operation.ID) {
			property := *operation.Key.Property

			current, ok := objects[property]
			if !ok || operation.ID.Compare(current.ID) > 0 {
				objects[property] = operation
			}
		}
	}

	return objects
}

func (b *Backend) insertSequenceOperation(
	ctx context.Context,
	handle uint32,
	index uint64,
	action Action,
	value *Scalar,
) (Operation, error) {
	if err := ctx.Err(); err != nil {
		return Operation{}, err
	}

	object, err := b.sequenceObject(handle)
	if err != nil {
		return Operation{}, err
	}

	sequence := b.state.sequenceValues(object.OpID)

	element, ok := b.resolveSequenceIndex(object, sequence, index)
	if !ok || element > uint64(len(sequence)) {
		return Operation{}, fmt.Errorf(
			"sequence index %d is out of bounds for length %d",
			index,
			len(sequence),
		)
	}

	key := Key{IsHead: element == 0}
	if element > 0 {
		key.Element = new(sequence[element-1].Element)
	}

	operation := Operation{
		ID:     b.nextOperationID(),
		Object: object,
		Key:    key,
		Insert: true,
		Action: action,
		Value:  value,
	}
	if err := b.addPending(operation); err != nil {
		return Operation{}, err
	}

	return operation, nil
}

// sequenceElementPredecessors returns the IDs of every visible operation at the
// list element, in ascending order. A put, delete, or increment must reference
// all of them so that concurrent conflicting values are overwritten identically
// to upstream Rust.
func (b *Backend) sequenceElementPredecessors(element OpID) []OpID {
	visible := b.state.visibleSequenceElementOperations(element)
	predecessors := make([]OpID, 0, len(visible))

	for _, operation := range visible {
		predecessors = append(predecessors, operation.ID)
	}

	return predecessors
}

func (b *Backend) sequenceOperation(
	ctx context.Context,
	handle uint32,
	index uint64,
) (sequenceValue, error) {
	if err := ctx.Err(); err != nil {
		return sequenceValue{}, err
	}

	object, err := b.sequenceObject(handle)
	if err != nil {
		return sequenceValue{}, err
	}

	sequence := b.state.sequenceValues(object.OpID)

	element, ok := b.resolveSequenceIndex(object, sequence, index)
	if !ok || element >= uint64(len(sequence)) {
		return sequenceValue{}, fmt.Errorf(
			"sequence index %d is out of bounds for length %d",
			index,
			len(sequence),
		)
	}

	return sequence[element], nil
}

// resolveSequenceIndex maps a caller-supplied index to a raw element index in
// the visible sequence. Text objects address positions in UTF-16 code units to
// match the reference encoding, so the index is translated to the element that
// begins at that code-unit boundary (a position inside a surrogate pair is
// advanced to the following boundary, as upstream Rust does); other sequences
// use element indices directly. The boolean reports whether the index resolves
// to a boundary at or before the end of the sequence.
func (b *Backend) resolveSequenceIndex(
	object ObjectID,
	sequence []sequenceValue,
	index uint64,
) (uint64, bool) {
	if !b.isTextObject(object) {
		return index, true
	}

	position := uint64(0)

	for i, value := range sequence {
		if position == index {
			return uint64(i), true
		}

		position += sequenceValueUTF16Width(value)
		if position > index {
			return uint64(i + 1), true
		}
	}

	if position == index {
		return uint64(len(sequence)), true
	}

	return 0, false
}

func (b *Backend) isTextObject(object ObjectID) bool {
	if object.IsRoot {
		return false
	}

	operation, ok := b.state.operations[object.OpID]

	return ok && operation.Action == ActionMakeText
}

func sequenceValueUTF16Width(value sequenceValue) uint64 {
	operation := value.Operation
	if operation.Value != nil && operation.Value.Type == ScalarString {
		return uint64(utf16Width(operation.Value.String))
	}

	return 1
}

func objectAction(rawType string) (Action, error) {
	switch rawType {
	case "map":
		return ActionMakeMap, nil
	case "list":
		return ActionMakeList, nil
	case "text":
		return ActionMakeText, nil
	case "table":
		return ActionMakeTable, nil
	default:
		return 0, fmt.Errorf("unknown object type %q", rawType)
	}
}

func actionObjectType(action Action) (string, error) {
	switch action {
	case ActionMakeMap:
		return "map", nil
	case ActionMakeList:
		return "list", nil
	case ActionMakeText:
		return "text", nil
	case ActionMakeTable:
		return "table", nil
	default:
		return "", fmt.Errorf("operation is not an object")
	}
}

func (b *Backend) textMarkKey(
	object ObjectID,
	index uint32,
) (Key, error) {
	// Mark positions share the unified rich-text index space with splice and
	// block operations, so block markers occupy a position (length 1) just like
	// a character. Walk the full element sequence, not the text-only view.
	sequence := b.state.sequenceElements(object.OpID)

	_, previous, err := richTextPosition(sequence, index)
	if err != nil {
		return Key{}, err
	}

	if previous == nil {
		return Key{IsHead: true}, nil
	}

	return Key{Element: new(*previous)}, nil
}

func markExpansion(value string) (bool, bool, error) {
	switch value {
	case "before":
		return true, false, nil
	case "after":
		return false, true, nil
	case "both":
		return true, true, nil
	case "none":
		return false, false, nil
	default:
		return false, false, fmt.Errorf("unknown mark expansion %q", value)
	}
}

func richTextPosition(
	sequence []Operation,
	index uint32,
) (*Operation, *OpID, error) {
	var (
		position uint32
		previous *OpID
	)

	for i := range sequence {
		operation := &sequence[i]
		if position == index {
			return operation, previous, nil
		}

		length := uint32(utf16Length(*operation))
		if operation.Action == ActionMakeMap {
			length = 1
		}

		if position+length > index {
			return nil, nil, fmt.Errorf(
				"rich-text index splits a Unicode character or block",
			)
		}

		position += length
		previous = new(operation.ID)
	}

	if position != index {
		return nil, nil, fmt.Errorf("rich-text index %d is out of bounds", index)
	}

	return nil, previous, nil
}

func sequenceRange(
	sequence []Operation,
	index uint32,
	deleteCount uint32,
) (int, int, *OpID, error) {
	position := uint32(0)
	start := -1

	var (
		previous      *OpID
		previousValue OpID
	)

	for i, operation := range sequence {
		if position == index {
			start = i
			break
		}

		length := elementLength(operation)
		if position+length > index {
			// UTF-16 callers can address the middle of a surrogate pair.
			// Upstream Rust advances such a position to the boundary after
			// the character rather than rejecting the edit.
			position += length
			previousValue = operation.ID
			previous = &previousValue
			start = i + 1

			break
		}

		position += length
		previousValue = operation.ID
		previous = &previousValue
	}

	if start == -1 {
		if position != index {
			return 0, 0, nil, fmt.Errorf("text index %d is out of bounds", index)
		}

		start = len(sequence)
	}

	target := position + deleteCount

	end := start
	for end < len(sequence) && position < target {
		position += elementLength(sequence[end])
		end++
	}

	if position < target {
		return 0, 0, nil, fmt.Errorf("text deletion extends beyond the document")
	}

	return start, end, previous, nil
}

// elementLength returns the position an operation occupies in the unified
// rich-text index space: block markers count as a single position, while text
// characters count by their UTF-16 code-unit length.
func elementLength(operation Operation) uint32 {
	if operation.Action == ActionMakeMap {
		return 1
	}

	return uint32(utf16Length(operation))
}

func utf16Length(operation Operation) int {
	if operation.Value == nil || operation.Value.Type != ScalarString {
		return 0
	}

	length := 0

	for _, character := range operation.Value.String {
		if character > 0xffff {
			length += 2
		} else {
			length++
		}
	}

	return length
}

func decodeScalarWire(encoded []byte) (Scalar, error) {
	var wire scalarWire
	if err := json.Unmarshal(encoded, &wire); err != nil {
		return Scalar{}, fmt.Errorf("cannot decode scalar: %w", err)
	}

	value := Scalar{
		Bool:   wire.Bool,
		Uint:   wire.Uint,
		Int:    wire.Int,
		Float:  math.Float64frombits(wire.Float),
		String: wire.String,
	}
	switch wire.Type {
	case "null":
		value.Type = ScalarNull
	case "boolean":
		if wire.Bool {
			value.Type = ScalarTrue
		} else {
			value.Type = ScalarFalse
		}
	case "uint":
		value.Type = ScalarUint
	case "int":
		value.Type = ScalarInt
	case "float64":
		value.Type = ScalarFloat64
	case "string":
		value.Type = ScalarString
	case "bytes":
		value.Type = ScalarBytes

		bytes, err := hex.DecodeString(wire.Bytes)
		if err != nil {
			return Scalar{}, fmt.Errorf("cannot decode scalar bytes: %w", err)
		}

		value.Bytes = bytes
	case "counter":
		value.Type = ScalarCounter
	case "timestamp":
		value.Type = ScalarTimestamp
	default:
		return Scalar{}, fmt.Errorf("unknown scalar type %q", wire.Type)
	}

	return value, nil
}

func encodeScalarWire(value Scalar) ([]byte, error) {
	wire := scalarWire{
		Bool:   value.Bool,
		Uint:   value.Uint,
		Int:    value.Int,
		Float:  math.Float64bits(value.Float),
		String: value.String,
		Bytes:  hex.EncodeToString(value.Bytes),
	}
	switch value.Type {
	case ScalarNull:
		wire.Type = "null"
	case ScalarFalse, ScalarTrue:
		wire.Type = "boolean"
		wire.Bool = value.Type == ScalarTrue
	case ScalarUint:
		wire.Type = "uint"
	case ScalarInt:
		wire.Type = "int"
	case ScalarFloat64:
		wire.Type = "float64"
	case ScalarString:
		wire.Type = "string"
	case ScalarBytes:
		wire.Type = "bytes"
	case ScalarCounter:
		wire.Type = "counter"
	case ScalarTimestamp:
		wire.Type = "timestamp"
	default:
		return nil, fmt.Errorf("unsupported scalar type %d", value.Type)
	}

	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("cannot encode scalar: %w", err)
	}

	return encoded, nil
}

func scalarValuesEqual(left, right Scalar) bool {
	return left.Type == right.Type &&
		left.Bool == right.Bool &&
		left.Uint == right.Uint &&
		left.Int == right.Int &&
		math.Float64bits(left.Float) == math.Float64bits(right.Float) &&
		left.String == right.String &&
		bytes.Equal(left.Bytes, right.Bytes)
}

func randomActorID() (ActorID, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("cannot generate native actor ID: %w", err)
	}

	return NewActorID(value[:])
}

func encodeEmptyDocument() []byte {
	body := []byte{0, 0, 0, 0}
	hashInput := []byte{byte(ChunkDocument)}
	hashInput = appendULEB(hashInput, uint64(len(body)))
	hashInput = append(hashInput, body...)
	hash := sha256.Sum256(hashInput)

	raw := []byte{0x85, 0x6f, 0x4a, 0x83}
	raw = append(raw, hash[:4]...)
	raw = append(raw, byte(ChunkDocument))
	raw = appendULEB(raw, uint64(len(body)))

	return append(raw, body...)
}

func decodeCursor(data []byte) (OpID, byte, error) {
	r := &reader{data: data}

	version, err := r.byte()
	if err != nil || version != 1 {
		return OpID{}, 0, fmt.Errorf("invalid cursor version")
	}

	cursorType, err := r.byte()
	if err != nil || cursorType != 3 {
		return OpID{}, 0, fmt.Errorf("unsupported cursor type")
	}

	actorBytes, err := decodeLengthPrefixed(r)
	if err != nil {
		return OpID{}, 0, fmt.Errorf("cannot decode cursor actor: %w", err)
	}

	actor, err := NewActorID(actorBytes)
	if err != nil {
		return OpID{}, 0, err
	}

	counter, err := r.uleb()
	if err != nil {
		return OpID{}, 0, fmt.Errorf("cannot decode cursor counter: %w", err)
	}

	move, err := r.byte()
	if err != nil || (move != 1 && move != 2) || r.remaining() != 0 {
		return OpID{}, 0, fmt.Errorf("invalid cursor movement")
	}

	return OpID{Actor: actor, Counter: counter}, move, nil
}

func equalHashes(left, right [][32]byte) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}

func nativeHashes(heads [][32]byte) []ChangeHash {
	result := make([]ChangeHash, len(heads))
	for i, head := range heads {
		result[i] = ChangeHash(head)
	}

	return result
}

func syncMessageFlagBits(flags []byte) byte {
	var bits byte

	for _, flag := range flags {
		if flag&syncFlagMarker != 0 {
			bits |= flag &^ syncFlagMarker
		}
	}

	return bits
}
