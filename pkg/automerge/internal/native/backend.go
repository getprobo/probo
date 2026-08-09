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

func (b *Backend) Close(context.Context) error {
	return nil
}

func (b *Backend) Save(ctx context.Context) ([]byte, error) {
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
		data = append(data, change...)
	}

	b.saveCursor = len(b.appended)

	return data, nil
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

	return uint64(len(b.state.sequenceValues(objectID.OpID))), nil
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

	sequence := b.state.sequence(object.OpID)

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

	inserted := make([]Operation, 0, len(value))
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

		inserted = append(inserted, operation)
		previous = new(operation.ID)
	}

	var updated []Operation
	if start == len(sequence) && end == len(sequence) {
		updated = append(sequence, inserted...)
	} else {
		updated = make(
			[]Operation,
			0,
			len(sequence)-(end-start)+len(inserted),
		)
		updated = append(updated, sequence[:start]...)
		updated = append(updated, inserted...)
		updated = append(updated, sequence[end:]...)
	}

	b.state.setSequenceCache(object.OpID, updated)

	return nil
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

	for _, operation := range b.state.sequence(object.OpID) {
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
	if index > uint64(len(sequence)) {
		return Operation{}, fmt.Errorf(
			"sequence index %d is out of bounds for length %d",
			index,
			len(sequence),
		)
	}

	key := Key{IsHead: index == 0}
	if index > 0 {
		key.Element = new(sequence[index-1].Element)
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
	if index >= uint64(len(sequence)) {
		return sequenceValue{}, fmt.Errorf(
			"sequence index %d is out of bounds for length %d",
			index,
			len(sequence),
		)
	}

	return sequence[index], nil
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
	sequence := b.state.sequence(object.OpID)

	_, _, previous, err := sequenceRange(sequence, index, 0)
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

		length := uint32(utf16Length(operation))
		if position+length > index {
			return 0, 0, nil, fmt.Errorf("text index splits a Unicode character")
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

	target := index + deleteCount

	end := start
	for end < len(sequence) && position < target {
		position += uint32(utf16Length(sequence[end]))
		if position > target {
			return 0, 0, nil, fmt.Errorf("text deletion splits a Unicode character")
		}

		end++
	}

	if position != target {
		return 0, 0, nil, fmt.Errorf("text deletion extends beyond the document")
	}

	return start, end, previous, nil
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
