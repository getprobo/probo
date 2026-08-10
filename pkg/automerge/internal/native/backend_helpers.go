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
)

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

	key = b.state.insertAnchorKey(object.OpID, key)

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

	// A mark boundary past the end of the text is rejected, matching the
	// reference. The reference applies the begin boundary before failing on the
	// out-of-range end, leaving a begin operation with no matching end; span
	// computation extends such an unmatched begin to the end of the text.
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

	// A deletion that runs past the end of the sequence is clamped to the
	// remaining elements rather than rejected, matching the reference, whose
	// splice stops once there are no more elements to delete.
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
