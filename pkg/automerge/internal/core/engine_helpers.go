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

package core

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"go.probo.inc/probo/pkg/automerge/internal/encoding"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func (b *Engine) addPending(operation opset.Operation) error {
	if err := b.state.applyPending([]opset.Operation{operation}); err != nil {
		return err
	}

	b.pending = append(b.pending, operation)

	return nil
}

func (b *Engine) nextOperationID() opset.OpID {
	id := opset.OpID{
		Actor:   b.actor,
		Counter: b.nextOp,
	}
	b.nextOp++

	return id
}

func (b *Engine) requireRoot(handle uint32) error {

	object, err := b.object(handle)
	if err != nil {
		return err
	}

	if !object.IsRoot {
		return fmt.Errorf("object is not the root map")
	}

	return nil
}

func (b *Engine) object(handle uint32) (opset.ObjectID, error) {
	object, ok := b.objects[handle]
	if !ok {
		return opset.ObjectID{}, fmt.Errorf("invalid object handle %d", handle)
	}

	return object, nil
}

func (b *Engine) mapObject(handle uint32) (opset.ObjectID, error) {
	object, err := b.object(handle)
	if err != nil {
		return opset.ObjectID{}, err
	}

	if object.IsRoot {
		return object, nil
	}

	operation, ok := b.state.operations[object.OpID]
	if !ok ||
		(operation.Action != opset.ActionMakeMap &&
			operation.Action != opset.ActionMakeTable) {
		return opset.ObjectID{}, fmt.Errorf("object is not a map")
	}

	return object, nil
}

func (b *Engine) sequenceObject(handle uint32) (opset.ObjectID, error) {
	object, err := b.object(handle)
	if err != nil {
		return opset.ObjectID{}, err
	}

	if object.IsRoot {
		return opset.ObjectID{}, fmt.Errorf("root map is not a sequence")
	}

	operation, ok := b.state.operations[object.OpID]
	if !ok ||
		(operation.Action != opset.ActionMakeList &&
			operation.Action != opset.ActionMakeText) {
		return opset.ObjectID{}, fmt.Errorf("object is not a sequence")
	}

	return object, nil
}

func (b *Engine) textObject(handle uint32) (opset.ObjectID, error) {
	object, err := b.object(handle)
	if err != nil {
		return opset.ObjectID{}, err
	}

	if object.IsRoot {
		return opset.ObjectID{}, fmt.Errorf("root map is not text")
	}

	operation, ok := b.state.operations[object.OpID]
	if !ok || operation.Action != opset.ActionMakeText {
		return opset.ObjectID{}, fmt.Errorf("object is not text")
	}

	return object, nil
}

func (b *Engine) pushObject(object opset.ObjectID) uint32 {
	handle := b.nextHandle
	b.nextHandle++
	b.objects[handle] = object

	return handle
}

func (b *Engine) syncState(handle uint32) (*syncSessionState, error) {
	state, ok := b.syncStates[handle]
	if !ok {
		return nil, fmt.Errorf("invalid sync state %d", handle)
	}

	return state, nil
}

func (b *Engine) rootTextObjects() map[string]opset.Operation {
	objects := make(map[string]opset.Operation)

	for _, operation := range b.state.operations {
		if operation.Object.IsRoot &&
			operation.Key.Property != nil &&
			operation.Action == opset.ActionMakeText &&
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

func (b *Engine) insertSequenceOperation(
	handle uint32,
	index uint64,
	action opset.Action,
	value *opset.Scalar,
) (opset.Operation, error) {

	object, err := b.sequenceObject(handle)
	if err != nil {
		return opset.Operation{}, err
	}

	sequence := b.state.sequenceValues(object.OpID)

	element, ok := b.resolveSequenceIndex(object, sequence, index)
	if !ok || element > uint64(len(sequence)) {
		return opset.Operation{}, fmt.Errorf(
			"sequence index %d is out of bounds for length %d",
			index,
			len(sequence),
		)
	}

	key := opset.Key{IsHead: element == 0}
	if element > 0 {
		key.Element = new(sequence[element-1].Element)
	}

	key = b.state.insertAnchorKey(object.OpID, key)

	operation := opset.Operation{
		ID:     b.nextOperationID(),
		Object: object,
		Key:    key,
		Insert: true,
		Action: action,
		Value:  value,
	}
	if err := b.addPending(operation); err != nil {
		return opset.Operation{}, err
	}

	return operation, nil
}

// sequenceElementPredecessors returns the IDs of every visible operation at the
// list element, in ascending order. A put, delete, or increment must reference
// all of them so that concurrent conflicting values are overwritten identically
// to upstream Rust.
func (b *Engine) sequenceElementPredecessors(element opset.OpID) []opset.OpID {
	visible := b.state.visibleSequenceElementOperations(element)
	predecessors := make([]opset.OpID, 0, len(visible))

	for _, operation := range visible {
		predecessors = append(predecessors, operation.ID)
	}

	return predecessors
}

func (b *Engine) sequenceOperation(
	handle uint32,
	index uint64,
) (sequenceValue, error) {

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
func (b *Engine) resolveSequenceIndex(
	object opset.ObjectID,
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

func (b *Engine) isTextObject(object opset.ObjectID) bool {
	if object.IsRoot {
		return false
	}

	operation, ok := b.state.operations[object.OpID]

	return ok && operation.Action == opset.ActionMakeText
}

func sequenceValueUTF16Width(value sequenceValue) uint64 {
	operation := value.Operation
	if operation.Value != nil && operation.Value.Type == opset.ScalarString {
		return uint64(utf16Width(operation.Value.String))
	}

	return 1
}

func objectAction(rawType string) (opset.Action, error) {
	switch rawType {
	case "map":
		return opset.ActionMakeMap, nil
	case "list":
		return opset.ActionMakeList, nil
	case "text":
		return opset.ActionMakeText, nil
	case "table":
		return opset.ActionMakeTable, nil
	default:
		return 0, fmt.Errorf("unknown object type %q", rawType)
	}
}

func actionObjectType(action opset.Action) (string, error) {
	switch action {
	case opset.ActionMakeMap:
		return "map", nil
	case opset.ActionMakeList:
		return "list", nil
	case opset.ActionMakeText:
		return "text", nil
	case opset.ActionMakeTable:
		return "table", nil
	default:
		return "", fmt.Errorf("operation is not an object")
	}
}

func (b *Engine) textMarkKey(
	object opset.ObjectID,
	index uint32,
) (opset.Key, error) {
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
		return opset.Key{}, err
	}

	if previous == nil {
		return opset.Key{IsHead: true}, nil
	}

	return opset.Key{Element: new(*previous)}, nil
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
	sequence []opset.Operation,
	index uint32,
) (*opset.Operation, *opset.OpID, error) {
	var (
		position uint32
		previous *opset.OpID
	)

	for i := range sequence {
		operation := &sequence[i]
		if position == index {
			return operation, previous, nil
		}

		length := uint32(utf16Length(*operation))
		if operation.Action == opset.ActionMakeMap {
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

// sequenceRange resolves a UTF-16 index and delete count against the visible
// sequence using the precomputed cumulative offsets, so a splice locates its
// position by binary search rather than walking the whole sequence. offsets has
// one entry per element plus a trailing total, where offsets[i] is the width
// before element i.
func sequenceRange(
	sequence []opset.Operation,
	offsets []uint32,
	index uint32,
	deleteCount uint32,
) (int, int, *opset.OpID, error) {
	total := offsets[len(offsets)-1]
	if index > total {
		return 0, 0, nil, fmt.Errorf("text index %d is out of bounds", index)
	}

	// Find the element whose starting offset is the last one at or before index.
	// When that offset equals index the insertion sits on the boundary before
	// the element; when it is smaller the index fell inside the element (a UTF-16
	// caller addressing the middle of a surrogate pair), so advance to the
	// boundary after it, matching the reference.
	boundary := sort.Search(len(offsets), func(i int) bool { return offsets[i] > index })
	start := boundary - 1
	if offsets[start] < index {
		start++
	}

	var previous *opset.OpID
	if start > 0 {
		previousValue := sequence[start-1].ID
		previous = &previousValue
	}

	// A deletion that runs past the end of the sequence is clamped to the
	// remaining elements rather than rejected, matching the reference, whose
	// splice stops once there are no more elements to delete. end is the first
	// element boundary at or past the deletion target.
	target := offsets[start] + deleteCount
	end := start + sort.Search(
		len(sequence)-start+1,
		func(i int) bool {
			return offsets[start+i] >= target
		},
	)
	if end > len(sequence) {
		end = len(sequence)
	}

	return start, end, previous, nil
}

// elementLength returns the position an operation occupies in the unified
// rich-text index space: block markers count as a single position, while text
// characters count by their UTF-16 code-unit length.
func elementLength(operation opset.Operation) uint32 {
	if operation.Action == opset.ActionMakeMap {
		return 1
	}

	return uint32(utf16Length(operation))
}

func utf16Length(operation opset.Operation) int {
	if operation.Value == nil || operation.Value.Type != opset.ScalarString {
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

func decodeScalarWire(encoded []byte) (opset.Scalar, error) {
	var wire scalarWire
	if err := json.Unmarshal(encoded, &wire); err != nil {
		return opset.Scalar{}, fmt.Errorf("cannot decode scalar: %w", err)
	}

	value := opset.Scalar{
		Bool:   wire.Bool,
		Uint:   wire.Uint,
		Int:    wire.Int,
		Float:  math.Float64frombits(wire.Float),
		String: wire.String,
	}
	switch wire.Type {
	case "null":
		value.Type = opset.ScalarNull
	case "boolean":
		if wire.Bool {
			value.Type = opset.ScalarTrue
		} else {
			value.Type = opset.ScalarFalse
		}
	case "uint":
		value.Type = opset.ScalarUint
	case "int":
		value.Type = opset.ScalarInt
	case "float64":
		value.Type = opset.ScalarFloat64
	case "string":
		value.Type = opset.ScalarString
	case "bytes":
		value.Type = opset.ScalarBytes

		bytes, err := hex.DecodeString(wire.Bytes)
		if err != nil {
			return opset.Scalar{}, fmt.Errorf("cannot decode scalar bytes: %w", err)
		}

		value.Bytes = bytes
	case "counter":
		value.Type = opset.ScalarCounter
	case "timestamp":
		value.Type = opset.ScalarTimestamp
	default:
		return opset.Scalar{}, fmt.Errorf("unknown scalar type %q", wire.Type)
	}

	return value, nil
}

func encodeScalarWire(value opset.Scalar) ([]byte, error) {
	wire := scalarWire{
		Bool:   value.Bool,
		Uint:   value.Uint,
		Int:    value.Int,
		Float:  math.Float64bits(value.Float),
		String: value.String,
		Bytes:  hex.EncodeToString(value.Bytes),
	}
	switch value.Type {
	case opset.ScalarNull:
		wire.Type = "null"
	case opset.ScalarFalse, opset.ScalarTrue:
		wire.Type = "boolean"
		wire.Bool = value.Type == opset.ScalarTrue
	case opset.ScalarUint:
		wire.Type = "uint"
	case opset.ScalarInt:
		wire.Type = "int"
	case opset.ScalarFloat64:
		wire.Type = "float64"
	case opset.ScalarString:
		wire.Type = "string"
	case opset.ScalarBytes:
		wire.Type = "bytes"
	case opset.ScalarCounter:
		wire.Type = "counter"
	case opset.ScalarTimestamp:
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

func scalarValuesEqual(left, right opset.Scalar) bool {
	return left.Type == right.Type &&
		left.Bool == right.Bool &&
		left.Uint == right.Uint &&
		left.Int == right.Int &&
		math.Float64bits(left.Float) == math.Float64bits(right.Float) &&
		left.String == right.String &&
		bytes.Equal(left.Bytes, right.Bytes)
}

func randomActorID() (opset.ActorID, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("cannot generate native actor ID: %w", err)
	}

	return opset.NewActorID(value[:])
}

func encodeEmptyDocument() []byte {
	body := []byte{0, 0, 0, 0}
	hashInput := []byte{byte(opset.ChunkDocument)}
	hashInput = encoding.AppendULEB(hashInput, uint64(len(body)))
	hashInput = append(hashInput, body...)
	hash := sha256.Sum256(hashInput)

	raw := []byte{0x85, 0x6f, 0x4a, 0x83}
	raw = append(raw, hash[:4]...)
	raw = append(raw, byte(opset.ChunkDocument))
	raw = encoding.AppendULEB(raw, uint64(len(body)))

	return append(raw, body...)
}

func decodeCursor(data []byte) (opset.OpID, byte, error) {
	r := encoding.NewReader(data)

	version, err := r.Byte()
	if err != nil || version != 1 {
		return opset.OpID{}, 0, fmt.Errorf("invalid cursor version")
	}

	cursorType, err := r.Byte()
	if err != nil || cursorType != 3 {
		return opset.OpID{}, 0, fmt.Errorf("unsupported cursor type")
	}

	actorBytes, err := encoding.DecodeLengthPrefixed(r)
	if err != nil {
		return opset.OpID{}, 0, fmt.Errorf("cannot decode cursor actor: %w", err)
	}

	actor, err := opset.NewActorID(actorBytes)
	if err != nil {
		return opset.OpID{}, 0, err
	}

	counter, err := r.ULEB()
	if err != nil {
		return opset.OpID{}, 0, fmt.Errorf("cannot decode cursor counter: %w", err)
	}

	move, err := r.Byte()
	if err != nil || (move != 1 && move != 2) || r.Remaining() != 0 {
		return opset.OpID{}, 0, fmt.Errorf("invalid cursor movement")
	}

	return opset.OpID{Actor: actor, Counter: counter}, move, nil
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

func nativeHashes(heads [][32]byte) []opset.ChangeHash {
	result := make([]opset.ChangeHash, len(heads))
	for i, head := range heads {
		result[i] = opset.ChangeHash(head)
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
