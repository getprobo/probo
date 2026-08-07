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

// Package native adapts pure-Go Automerge codecs and CRDT operations to Probo's
// backend contract.
package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
	"unicode/utf16"

	"github.com/MichaelMure/gotomerge/format"
	"github.com/MichaelMure/gotomerge/opset"
	"github.com/MichaelMure/gotomerge/types"
	automergeio "github.com/MichaelMure/gotomerge/utils/io"
)

type (
	Backend struct {
		operations  *opset.OpSet
		actor       types.ActorId
		sequence    uint64
		transaction *opset.Transaction
		history     []byte
		objects     []types.ObjectId
		syncStates  []*syncState
		closed      bool
	}

	syncState struct {
		LastSent [][32]byte `json:"lastSent,omitempty"`
	}

	syncMessage struct {
		Document []byte `json:"document"`
	}

	cursor struct {
		End     bool          `json:"end,omitempty"`
		Counter uint32        `json:"counter,omitempty"`
		Actor   types.ActorId `json:"actor,omitempty"`
	}

	plainSpan struct {
		Type  string         `json:"type"`
		Value string         `json:"value"`
		Marks map[string]any `json:"marks,omitempty"`
	}
)

var (
	ErrClosed              = errors.New("native Automerge backend is closed")
	ErrNoOperations        = errors.New("change contains no operations")
	ErrNegativeDeleteCount = errors.New("negative text delete count is not supported")
)

func New(ctx context.Context) (*Backend, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("cannot create native Automerge document: %w", err)
	}

	return &Backend{
		operations: opset.New(),
		sequence:   1,
		objects:    []types.ObjectId{types.RootObjectId()},
	}, nil
}

func Load(ctx context.Context, data []byte) (*Backend, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("cannot load native Automerge document: %w", err)
	}

	operations := opset.New()
	if err := applyStream(operations, data, true); err != nil {
		return nil, fmt.Errorf("cannot decode native Automerge document: %w", err)
	}

	return &Backend{
		operations: operations,
		sequence:   1,
		history:    bytes.Clone(data),
		objects:    []types.ObjectId{types.RootObjectId()},
	}, nil
}

func (b *Backend) Close(_ context.Context) error {
	b.closed = true
	b.operations = nil
	b.transaction = nil
	b.history = nil
	b.objects = nil
	b.syncStates = nil
	return nil
}

func (b *Backend) Save(ctx context.Context) ([]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}
	if len(b.history) > 0 {
		return bytes.Clone(b.history), nil
	}

	var output bytes.Buffer
	if err := b.operations.ExportDocument(&output); err != nil {
		return nil, fmt.Errorf("cannot encode empty native Automerge document: %w", err)
	}
	return output.Bytes(), nil
}

func (b *Backend) SetActor(ctx context.Context, actor []byte) error {
	if err := b.open(ctx); err != nil {
		return err
	}
	if len(actor) == 0 {
		return fmt.Errorf("actor ID cannot be empty")
	}
	if b.transaction != nil {
		return fmt.Errorf("cannot change actor with pending native operations")
	}

	b.actor = types.ActorId(bytes.Clone(actor))
	b.sequence = 1
	return nil
}

func (b *Backend) PutString(ctx context.Context, object uint32, key, value string) error {
	if err := b.ready(ctx); err != nil {
		return err
	}

	objectID, err := b.object(object)
	if err != nil {
		return fmt.Errorf("cannot resolve native map object: %w", err)
	}
	b.change().MapSet(objectID, key, value)
	return nil
}

func (b *Backend) PutText(ctx context.Context, object uint32, key string) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve native text parent: %w", err)
	}
	textID, _ := b.change().MakeText(objectID, key)
	return b.pushObject(textID)
}

func (b *Backend) GetText(ctx context.Context, object uint32, key string) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve native text parent: %w", err)
	}
	operation, ok := b.operations.MapGet(objectID, key)
	if !ok {
		return 0, fmt.Errorf("text does not exist")
	}
	if operation.Action.Kind != types.ActionMakeText {
		return 0, fmt.Errorf("value is not text")
	}

	return b.pushObject(types.ObjectId(operation.Id))
}

func (b *Backend) SpliceText(
	ctx context.Context,
	object uint32,
	index uint32,
	deleteCount int32,
	value string,
) error {
	if err := b.ready(ctx); err != nil {
		return err
	}
	if deleteCount < 0 {
		return ErrNegativeDeleteCount
	}

	objectID, err := b.object(object)
	if err != nil {
		return fmt.Errorf("cannot resolve native text object: %w", err)
	}
	transaction := b.change()
	current := textFromOperations(transaction.ListElements(objectID))
	position, err := runePosition(current, uint64(index))
	if err != nil {
		return fmt.Errorf("cannot resolve native text splice index: %w", err)
	}
	end, err := runePosition(current, uint64(index)+uint64(deleteCount))
	if err != nil {
		return fmt.Errorf("cannot resolve native text splice end: %w", err)
	}
	for range end - position {
		operation, ok := transaction.ListAt(objectID, int(position))
		if !ok {
			return fmt.Errorf("cannot find native text operation at %d", position)
		}
		transaction.ListDelete(objectID, operation.Id, operation.Id)
	}

	var predecessor types.Key = types.KeyOpId{}
	if position > 0 {
		operation, ok := transaction.ListAt(objectID, int(position-1))
		if !ok {
			return fmt.Errorf("cannot find native text predecessor at %d", position-1)
		}
		predecessor = types.KeyOpId(operation.Id)
	}
	for _, character := range value {
		identifier := transaction.ListInsert(objectID, predecessor, string(character))
		predecessor = types.KeyOpId(identifier)
	}

	return nil
}

func (b *Backend) Text(ctx context.Context, object uint32) (string, error) {
	if err := b.ready(ctx); err != nil {
		return "", err
	}

	objectID, err := b.object(object)
	if err != nil {
		return "", fmt.Errorf("cannot resolve native text object: %w", err)
	}
	if b.transaction != nil {
		return textFromOperations(b.transaction.ListElements(objectID)), nil
	}
	if kind, ok := b.operations.ObjType(objectID); !ok || kind != types.ActionMakeText {
		return "", fmt.Errorf("value is not text")
	}

	return b.operations.Text(objectID), nil
}

func (b *Backend) TextSpans(ctx context.Context, object uint32) ([]byte, error) {
	text, err := b.Text(ctx, object)
	if err != nil {
		return nil, fmt.Errorf("cannot read native text spans: %w", err)
	}

	data, err := json.Marshal([]plainSpan{{Type: "text", Value: text}})
	if err != nil {
		return nil, fmt.Errorf("cannot encode native text spans: %w", err)
	}
	return data, nil
}

func (b *Backend) TextCursor(ctx context.Context, object, index uint32) ([]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve native cursor text object: %w", err)
	}
	operations := b.textOperations(objectID)
	current := textFromOperations(operations)
	position, err := runePosition(current, uint64(index))
	if err != nil {
		return nil, fmt.Errorf("cannot resolve native cursor index: %w", err)
	}

	nativeCursor := cursor{End: position == uint64(len(operations))}
	if !nativeCursor.End {
		operation := operations[position]
		nativeCursor.Counter = operation.Id.Counter
		nativeCursor.Actor = bytes.Clone(b.operations.Actor(operation.Id.ActorIdx))
	}
	data, err := json.Marshal(nativeCursor)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native cursor: %w", err)
	}
	return data, nil
}

func (b *Backend) TextCursorPosition(
	ctx context.Context,
	object uint32,
	data []byte,
) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, fmt.Errorf("cannot resolve native cursor text object: %w", err)
	}
	var nativeCursor cursor
	if err := json.Unmarshal(data, &nativeCursor); err != nil {
		return 0, fmt.Errorf("cannot decode native cursor: %w", err)
	}
	operations := b.textOperations(objectID)
	position := len(operations)
	if !nativeCursor.End {
		position = -1
		for index, operation := range operations {
			if operation.Id.Counter == nativeCursor.Counter &&
				bytes.Equal(b.operations.Actor(operation.Id.ActorIdx), nativeCursor.Actor) {
				position = index
				break
			}
		}
		if position < 0 {
			return 0, fmt.Errorf("native cursor target is not visible")
		}
	}
	index, err := utf16Position(textFromOperations(operations), uint64(position))
	if err != nil {
		return 0, fmt.Errorf("cannot convert native cursor position: %w", err)
	}
	if index > math.MaxUint32 {
		return 0, fmt.Errorf("native cursor position %d exceeds uint32", index)
	}
	return uint32(index), nil
}

func (b *Backend) Commit(
	ctx context.Context,
	_ string,
	_ time.Time,
) ([32]byte, error) {
	if err := b.ready(ctx); err != nil {
		return [32]byte{}, err
	}
	if b.transaction == nil || !b.transaction.HasOps() {
		return [32]byte{}, ErrNoOperations
	}

	var change bytes.Buffer
	if err := b.transaction.Commit(&change); err != nil {
		return [32]byte{}, fmt.Errorf("cannot commit native Automerge change: %w", err)
	}
	b.transaction = nil
	b.sequence++
	b.history = append(b.history, change.Bytes()...)

	heads := copyHeads(b.operations.Heads())
	if len(heads) != 1 {
		return [32]byte{}, fmt.Errorf("native commit produced %d heads", len(heads))
	}
	return heads[0], nil
}

func (b *Backend) Heads(ctx context.Context) ([][32]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}
	return copyHeads(b.operations.Heads()), nil
}

func (b *Backend) Merge(ctx context.Context, other []byte) ([][32]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}
	if b.transaction != nil {
		return nil, fmt.Errorf("cannot merge with pending native operations")
	}

	additions, err := mergeStream(b.operations, other)
	if err != nil {
		return nil, fmt.Errorf("cannot merge native document: %w", err)
	}
	b.history = append(b.history, additions...)
	return copyHeads(b.operations.Heads()), nil
}

func (b *Backend) NewSyncState(ctx context.Context) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}
	if len(b.syncStates) == math.MaxUint32 {
		return 0, fmt.Errorf("too many native sync states")
	}

	handle := uint32(len(b.syncStates))
	b.syncStates = append(b.syncStates, &syncState{})
	return handle, nil
}

func (b *Backend) CloseSyncState(ctx context.Context, handle uint32) error {
	if err := b.ready(ctx); err != nil {
		return err
	}
	if _, err := b.state(handle); err != nil {
		return err
	}
	b.syncStates[handle] = nil
	return nil
}

func (b *Backend) GenerateSyncMessage(
	ctx context.Context,
	handle uint32,
) ([]byte, bool, error) {
	if err := b.ready(ctx); err != nil {
		return nil, false, err
	}

	state, err := b.state(handle)
	if err != nil {
		return nil, false, err
	}
	heads := copyHeads(b.operations.Heads())
	if equalHeads(heads, state.LastSent) {
		return nil, false, nil
	}
	document, err := b.Save(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("cannot save native sync document: %w", err)
	}
	message, err := json.Marshal(syncMessage{Document: document})
	if err != nil {
		return nil, false, fmt.Errorf("cannot encode native sync message: %w", err)
	}
	state.LastSent = heads
	return message, true, nil
}

func (b *Backend) ReceiveSyncMessage(ctx context.Context, handle uint32, data []byte) error {
	if err := b.ready(ctx); err != nil {
		return err
	}
	if _, err := b.state(handle); err != nil {
		return err
	}

	var message syncMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("cannot decode native sync message: %w", err)
	}
	_, err := b.Merge(ctx, message.Document)
	return err
}

func (b *Backend) SaveSyncState(ctx context.Context, handle uint32) ([]byte, error) {
	if err := b.ready(ctx); err != nil {
		return nil, err
	}

	state, err := b.state(handle)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("cannot encode native sync state: %w", err)
	}
	return data, nil
}

func (b *Backend) LoadSyncState(ctx context.Context, data []byte) (uint32, error) {
	if err := b.ready(ctx); err != nil {
		return 0, err
	}

	var state syncState
	if err := json.Unmarshal(data, &state); err != nil {
		return 0, fmt.Errorf("cannot decode native sync state: %w", err)
	}
	if len(b.syncStates) == math.MaxUint32 {
		return 0, fmt.Errorf("too many native sync states")
	}
	handle := uint32(len(b.syncStates))
	b.syncStates = append(b.syncStates, &state)
	return handle, nil
}

func (b *Backend) change() *opset.Transaction {
	if b.transaction == nil {
		b.transaction = b.operations.Begin(b.actor, b.sequence)
	}
	return b.transaction
}

func (b *Backend) ready(ctx context.Context) error {
	if err := b.open(ctx); err != nil {
		return err
	}
	if len(b.actor) == 0 {
		return fmt.Errorf("native Automerge actor is not initialized")
	}
	return nil
}

func (b *Backend) open(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("native Automerge operation canceled: %w", err)
	}
	if b.closed {
		return ErrClosed
	}
	return nil
}

func (b *Backend) object(handle uint32) (types.ObjectId, error) {
	if int(handle) >= len(b.objects) {
		return types.ObjectId{}, fmt.Errorf("invalid object handle %d", handle)
	}
	return b.objects[handle], nil
}

func (b *Backend) pushObject(object types.ObjectId) (uint32, error) {
	if len(b.objects) == math.MaxUint32 {
		return 0, fmt.Errorf("too many object handles")
	}
	handle := uint32(len(b.objects))
	b.objects = append(b.objects, object)
	return handle, nil
}

func (b *Backend) textOperations(object types.ObjectId) []opset.Op {
	if b.transaction != nil {
		return b.transaction.ListElements(object)
	}
	return b.operations.ListElements(object)
}

func (b *Backend) state(handle uint32) (*syncState, error) {
	if int(handle) >= len(b.syncStates) || b.syncStates[handle] == nil {
		return nil, fmt.Errorf("invalid sync state handle %d", handle)
	}
	return b.syncStates[handle], nil
}

func applyStream(operations *opset.OpSet, data []byte, documents bool) error {
	reader := automergeio.NewSubReader(data)
	for !reader.Empty() {
		chunk, skip, err := format.ReadChunk(reader)
		if err != nil {
			return fmt.Errorf("cannot read Automerge chunk: %w", err)
		}
		switch value := chunk.(type) {
		case *format.DocumentChunk:
			if !documents {
				break
			}
			if err := operations.ApplyDocument(value); err != nil {
				return fmt.Errorf("cannot apply Automerge document chunk: %w", err)
			}
		case *format.ChangeChunk:
			if _, exists := operations.AppliedHashes()[value.Hash]; exists {
				break
			}
			if err := operations.ApplyChange(value); err != nil {
				return fmt.Errorf("cannot apply Automerge change chunk: %w", err)
			}
		}
		if err := reader.Skip(skip); err != nil {
			return fmt.Errorf("cannot advance Automerge chunk reader: %w", err)
		}
	}
	return nil
}

func mergeStream(operations *opset.OpSet, data []byte) ([]byte, error) {
	reader := automergeio.NewSubReader(data)
	var additions []byte
	for !reader.Empty() {
		start := reader.Consumed()
		chunk, skip, err := format.ReadChunk(reader)
		if err != nil {
			return nil, fmt.Errorf("cannot read merge chunk: %w", err)
		}
		if change, ok := chunk.(*format.ChangeChunk); ok {
			if _, exists := operations.AppliedHashes()[change.Hash]; !exists {
				if err := operations.ApplyChange(change); err != nil {
					return nil, fmt.Errorf("cannot apply merge change: %w", err)
				}
				additions = append(additions, data[start:start+skip]...)
			}
		}
		if err := reader.Skip(skip); err != nil {
			return nil, fmt.Errorf("cannot advance merge chunk reader: %w", err)
		}
	}
	return additions, nil
}

func textFromOperations(operations []opset.Op) string {
	var output bytes.Buffer
	for _, operation := range operations {
		if value, ok := operation.Action.Value.(string); ok {
			output.WriteString(value)
		}
	}
	return output.String()
}

func runePosition(value string, utf16Index uint64) (uint64, error) {
	var units uint64
	var runes uint64
	for _, character := range value {
		if units == utf16Index {
			return runes, nil
		}
		width := uint64(1)
		if utf16.RuneLen(character) == 2 {
			width = 2
		}
		if utf16Index < units+width {
			return 0, fmt.Errorf("UTF-16 index %d splits a surrogate pair", utf16Index)
		}
		units += width
		runes++
	}
	if units == utf16Index {
		return runes, nil
	}
	return 0, fmt.Errorf("UTF-16 index %d exceeds text length %d", utf16Index, units)
}

func utf16Position(value string, runeIndex uint64) (uint64, error) {
	var units uint64
	var runes uint64
	for _, character := range value {
		if runes == runeIndex {
			return units, nil
		}
		if utf16.RuneLen(character) == 2 {
			units += 2
		} else {
			units++
		}
		runes++
	}
	if runes == runeIndex {
		return units, nil
	}
	return 0, fmt.Errorf("rune index %d exceeds text length %d", runeIndex, runes)
}

func copyHeads(source []types.ChangeHash) [][32]byte {
	heads := make([][32]byte, len(source))
	for index := range source {
		heads[index] = [32]byte(source[index])
	}
	sort.Slice(
		heads,
		func(left, right int) bool {
			return bytes.Compare(heads[left][:], heads[right][:]) < 0
		},
	)
	return heads
}

func equalHeads(left, right [][32]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
