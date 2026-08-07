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
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
	"unicode/utf16"
)

type Backend struct {
	state         *State
	actor         ActorID
	base          []byte
	appended      [][]byte
	pending       []Operation
	objects       map[uint32]ObjectID
	nextHandle    uint32
	syncStates    map[uint32]*nativeSyncState
	nextSyncState uint32
}

type nativeSyncState struct {
	RemoteHeads [][32]byte `json:"remoteHeads"`
	Need        [][32]byte `json:"need"`
	NeedsAck    bool       `json:"needsAck"`
}

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
		base:          base,
		objects:       map[uint32]ObjectID{0: RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*nativeSyncState),
		nextSyncState: 1,
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
		base:          append([]byte(nil), data...),
		objects:       map[uint32]ObjectID{0: RootObject()},
		nextHandle:    1,
		syncStates:    make(map[uint32]*nativeSyncState),
		nextSyncState: 1,
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
	return data, nil
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
	object, err := b.object(handle)
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

	for _, character := range []rune(value) {
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

func (b *Backend) Text(ctx context.Context, handle uint32) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	object, err := b.object(handle)
	if err != nil {
		return "", err
	}
	for property, operation := range b.rootTextObjects() {
		if operation.ID == object.OpID {
			return b.state.Text(property)
		}
	}
	return "", fmt.Errorf("text object does not exist")
}

func (b *Backend) TextSpans(
	ctx context.Context,
	handle uint32,
) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	object, err := b.object(handle)
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	object, err := b.object(handle)
	if err != nil {
		return nil, err
	}
	sequence := b.state.sequence(object.OpID)
	position := uint32(0)
	for _, operation := range sequence {
		if position == index {
			data := []byte{1, 3}
			data = appendLengthPrefixedNative(data, operation.ID.Actor.Bytes())
			data = appendULEB(data, operation.ID.Counter)
			data = append(data, 2)
			return data, nil
		}
		position += uint32(utf16Length(operation))
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
	object, err := b.object(handle)
	if err != nil {
		return 0, err
	}
	target, _, err := decodeCursor(cursor)
	if err != nil {
		return 0, err
	}

	position := uint32(0)
	for _, operation := range b.state.sequenceAll(object.OpID) {
		if operation.ID == target {
			return position, nil
		}
		if !b.state.isSuperseded(operation.ID) {
			position += uint32(utf16Length(operation))
		}
	}
	return 0, fmt.Errorf("text cursor target does not exist")
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
	dependencies := b.state.Heads()
	change := &Change{
		Actor:        b.actor,
		Sequence:     b.state.sequenceForActor(b.actor) + 1,
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
		b.base = append([]byte(nil), data...)
		b.appended = nil
		return b.Heads(ctx)
	}
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash == nil {
			continue
		}
		if _, exists := b.state.changes[*change.Hash]; exists {
			continue
		}
		raw, err := EncodeChange(change)
		if err != nil {
			return nil, fmt.Errorf("cannot encode merged native change: %w", err)
		}
		if err := b.state.ApplyChange(change); err != nil {
			return nil, fmt.Errorf("cannot apply merged native change: %w", err)
		}
		b.appended = append(b.appended, raw)
	}
	return b.Heads(ctx)
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
	if !state.NeedsAck && equalHashes(heads, state.RemoteHeads) {
		return nil, false, nil
	}

	message := SyncMessage{
		Version: SyncMessageVersion2,
		Heads:   heads,
		Need:    append([][32]byte(nil), state.Need...),
	}
	if !state.NeedsAck && len(state.Need) == 0 {
		document, err := b.Save(ctx)
		if err != nil {
			return nil, false, err
		}
		message.Changes = [][]byte{document}
	}
	state.NeedsAck = false
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
	for _, change := range message.Changes {
		if _, err := b.Merge(ctx, change); err != nil {
			return fmt.Errorf("cannot merge native sync payload: %w", err)
		}
	}
	state.RemoteHeads = append([][32]byte(nil), message.Heads...)
	state.Need = state.Need[:0]
	for _, head := range message.Heads {
		if _, ok := b.state.changes[ChangeHash(head)]; !ok {
			state.Need = append(state.Need, head)
		}
	}
	state.NeedsAck = len(message.Changes) > 0
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
	return OpID{
		Actor:   b.actor,
		Counter: b.state.maxOpGlobal() + 1,
	}
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
			objects[*operation.Key.Property] = operation
		}
	}
	return objects
}

func sequenceRange(
	sequence []Operation,
	index uint32,
	deleteCount uint32,
) (int, int, *OpID, error) {
	position := uint32(0)
	start := -1
	var previous *OpID
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
		previous = new(operation.ID)
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
	return len(utf16.Encode([]rune(operation.Value.String)))
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
