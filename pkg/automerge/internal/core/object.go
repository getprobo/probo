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
	"encoding/json"
	"fmt"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"slices"
)

func (b *Engine) PutString(
	object uint32,
	key string,
	value string,
) error {
	if err := b.requireRoot(object); err != nil {
		return err
	}

	if existing, ok := b.state.visibleMapOperation(key, opset.ActionSet); ok &&
		existing.Value != nil &&
		existing.Value.Type == opset.ScalarString &&
		existing.Value.String == value {
		return nil
	}

	property := key

	operation := opset.Operation{
		ID:     b.nextOperationID(),
		Object: opset.RootObject(),
		Key:    opset.Key{Property: &property},
		Action: opset.ActionSet,
		Value:  &opset.Scalar{Type: opset.ScalarString, String: value},
	}
	for _, predecessor := range b.state.visibleMapOperations(key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	return b.addPending(operation)
}

func (b *Engine) GetString(
	object uint32,
	key string,
) (string, error) {
	if err := b.requireRoot(object); err != nil {
		return "", err
	}

	operation, ok := b.state.visibleMapOperation(key, opset.ActionSet)
	if !ok || operation.Value == nil || operation.Value.Type != opset.ScalarString {
		return "", fmt.Errorf("string property %q does not exist", key)
	}

	return operation.Value.String, nil
}

func (b *Engine) PutScalar(
	object uint32,
	key string,
	encoded []byte,
) error {
	objectID, err := b.mapObject(object)
	if err != nil {
		return err
	}

	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	property := key

	if existing, ok := b.state.visibleMapObjectValue(objectID, key); ok {
		existingValue, scalar := b.state.scalarValue(existing)
		if scalar && scalarValuesEqual(existingValue, value) {
			// Assigning the value the winning operation already holds changes
			// nothing, so an unconflicted key records no operation. A conflicted
			// key still has to collapse: the reference deletes the losing
			// siblings and keeps the winner rather than writing the value again.
			losing := make([]opset.OpID, 0)

			for _, operation := range b.state.visibleMapObjectOperations(objectID, key) {
				if operation.ID != existing.ID {
					losing = append(losing, operation.ID)
				}
			}

			if len(losing) == 0 {
				return nil
			}

			return b.addPending(opset.Operation{
				ID:           b.nextOperationID(),
				Object:       objectID,
				Key:          opset.Key{Property: &property},
				Action:       opset.ActionDelete,
				Predecessors: losing,
			},
			)
		}
	}

	operation := opset.Operation{
		ID:     b.nextOperationID(),
		Object: objectID,
		Key:    opset.Key{Property: &property},
		Action: opset.ActionSet,
		Value:  &value,
	}
	for _, predecessor := range b.state.visibleMapObjectOperations(objectID, key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	return b.addPending(operation)
}

func (b *Engine) GetScalar(
	object uint32,
	key string,
) ([]byte, error) {
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

func (b *Engine) GetScalarAtHeads(
	object uint32,
	key string,
	heads [][32]byte,
) ([]byte, error) {
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

func (b *Engine) GetAllScalars(
	object uint32,
	key string,
) ([]byte, error) {
	objectID, err := b.mapObject(object)
	if err != nil {
		return nil, err
	}

	var values []json.RawMessage

	for _, operation := range b.state.visibleMapObjectOperations(objectID, key) {
		if operation.Action == opset.ActionIncrement {
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

func (b *Engine) GetAllScalarsAt(
	object uint32,
	index uint64,
) ([]byte, error) {
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

func (b *Engine) PutObject(
	object uint32,
	key string,
	rawType string,
) (uint32, error) {
	objectID, err := b.mapObject(object)
	if err != nil {
		return 0, err
	}

	action, err := objectAction(rawType)
	if err != nil {
		return 0, err
	}

	property := key

	operation := opset.Operation{
		ID:     b.nextOperationID(),
		Object: objectID,
		Key:    opset.Key{Property: &property},
		Action: action,
	}
	for _, predecessor := range b.state.visibleMapObjectOperations(objectID, key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(opset.ObjectID{OpID: operation.ID}), nil
}

func (b *Engine) GetObject(
	object uint32,
	key string,
) (uint32, string, error) {
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

	return b.pushObject(opset.ObjectID{OpID: operation.ID}), rawType, nil
}

func (b *Engine) InsertScalar(
	object uint32,
	index uint64,
	encoded []byte,
) error {
	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	_, err = b.insertSequenceOperation(object, index, opset.ActionSet, &value)

	return err
}

func (b *Engine) PutScalarAt(
	object uint32,
	index uint64,
	encoded []byte,
) error {
	value, err := decodeScalarWire(encoded)
	if err != nil {
		return err
	}

	target, err := b.sequenceOperation(object, index)
	if err != nil {
		return err
	}

	objectID, err := b.object(object)
	if err != nil {
		return err
	}

	// Assigning the value the winning operation already holds changes nothing,
	// so an unconflicted element records no operation. A conflicted element
	// still has to collapse: the reference deletes the losing siblings and keeps
	// the winner rather than writing the same value again.
	if existingValue, scalar := b.state.scalarValue(target.Operation); scalar &&
		scalarValuesEqual(existingValue, value) {
		losing := make([]opset.OpID, 0)

		for _, operation := range b.state.visibleSequenceElementOperations(target.Element) {
			if operation.ID != target.Operation.ID {
				losing = append(losing, operation.ID)
			}
		}

		if len(losing) == 0 {
			return nil
		}

		return b.addPending(opset.Operation{
			ID:           b.nextOperationID(),
			Object:       objectID,
			Key:          opset.Key{Element: new(target.Element)},
			Action:       opset.ActionDelete,
			Predecessors: losing,
		},
		)
	}

	return b.addPending(opset.Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          opset.Key{Element: new(target.Element)},
		Action:       opset.ActionSet,
		Value:        &value,
		Predecessors: b.sequenceElementPredecessors(target.Element),
	},
	)
}

func (b *Engine) InsertObject(
	object uint32,
	index uint64,
	rawType string,
) (uint32, error) {
	action, err := objectAction(rawType)
	if err != nil {
		return 0, err
	}

	operation, err := b.insertSequenceOperation(
		object,
		index,
		action,
		nil,
	)
	if err != nil {
		return 0, err
	}

	return b.pushObject(opset.ObjectID{OpID: operation.ID}), nil
}

func (b *Engine) PutObjectAt(
	object uint32,
	index uint64,
	rawType string,
) (uint32, error) {
	action, err := objectAction(rawType)
	if err != nil {
		return 0, err
	}

	target, err := b.sequenceOperation(object, index)
	if err != nil {
		return 0, err
	}

	objectID, err := b.object(object)
	if err != nil {
		return 0, err
	}

	operation := opset.Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          opset.Key{Element: new(target.Element)},
		Action:       action,
		Predecessors: b.sequenceElementPredecessors(target.Element),
	}
	if err := b.addPending(operation); err != nil {
		return 0, err
	}

	return b.pushObject(opset.ObjectID{OpID: operation.ID}), nil
}

func (b *Engine) GetScalarAt(
	object uint32,
	index uint64,
) ([]byte, error) {
	operation, err := b.sequenceOperation(object, index)
	if err != nil {
		return nil, err
	}

	value, ok := b.state.scalarValue(operation.Operation)
	if !ok {
		return nil, fmt.Errorf("sequence value at index %d is not a scalar", index)
	}

	return encodeScalarWire(value)
}

func (b *Engine) GetObjectAt(
	object uint32,
	index uint64,
) (uint32, string, error) {
	operation, err := b.sequenceOperation(object, index)
	if err != nil {
		return 0, "", err
	}

	rawType, err := actionObjectType(operation.Operation.Action)
	if err != nil {
		return 0, "", err
	}

	return b.pushObject(opset.ObjectID{OpID: operation.Operation.ID}), rawType, nil
}

func (b *Engine) DeleteMap(
	object uint32,
	key string,
) error {
	objectID, err := b.mapObject(object)
	if err != nil {
		return err
	}

	property := key

	operation := opset.Operation{
		ID:     b.nextOperationID(),
		Object: objectID,
		Key:    opset.Key{Property: &property},
		Action: opset.ActionDelete,
	}
	for _, predecessor := range b.state.visibleMapObjectOperations(objectID, key) {
		operation.Predecessors = append(operation.Predecessors, predecessor.ID)
	}

	if len(operation.Predecessors) == 0 {
		return fmt.Errorf("map property %q does not exist", key)
	}

	return b.addPending(operation)
}

func (b *Engine) DeleteSequence(
	object uint32,
	index uint64,
) error {
	target, err := b.sequenceOperation(object, index)
	if err != nil {
		return err
	}

	objectID, err := b.object(object)
	if err != nil {
		return err
	}

	return b.addPending(opset.Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          opset.Key{Element: new(target.Element)},
		Action:       opset.ActionDelete,
		Predecessors: b.sequenceElementPredecessors(target.Element),
	},
	)
}

func (b *Engine) Increment(
	object uint32,
	key string,
	delta int64,
) error {
	objectID, err := b.mapObject(object)
	if err != nil {
		return err
	}

	visible := b.state.visibleMapObjectOperations(objectID, key)

	hasCounter := slices.ContainsFunc(visible, isCounterOperation)

	if !hasCounter {
		return fmt.Errorf("map property %q is not a counter", key)
	}

	property := key

	predecessors := make([]opset.OpID, 0, len(visible))
	for _, operation := range visible {
		predecessors = append(predecessors, operation.ID)
	}

	return b.addPending(opset.Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          opset.Key{Property: &property},
		Action:       opset.ActionIncrement,
		Value:        &opset.Scalar{Type: opset.ScalarInt, Int: delta},
		Predecessors: predecessors,
	},
	)
}

func (b *Engine) IncrementAt(
	object uint32,
	index uint64,
	delta int64,
) error {
	target, err := b.sequenceOperation(object, index)
	if err != nil {
		return err
	}

	visible := b.state.visibleSequenceElementOperations(target.Element)

	hasCounter := slices.ContainsFunc(visible, isCounterOperation)

	if !hasCounter {
		return fmt.Errorf("sequence value at index %d is not a counter", index)
	}

	objectID, err := b.object(object)
	if err != nil {
		return err
	}

	predecessors := make([]opset.OpID, 0, len(visible))
	for _, operation := range visible {
		predecessors = append(predecessors, operation.ID)
	}

	return b.addPending(opset.Operation{
		ID:           b.nextOperationID(),
		Object:       objectID,
		Key:          opset.Key{Element: new(target.Element)},
		Action:       opset.ActionIncrement,
		Value:        &opset.Scalar{Type: opset.ScalarInt, Int: delta},
		Predecessors: predecessors,
	},
	)
}

func (b *Engine) Keys(object uint32) ([]string, error) {
	objectID, err := b.mapObject(object)
	if err != nil {
		return nil, err
	}

	return b.state.mapKeys(objectID), nil
}

func (b *Engine) Length(object uint32) (uint64, error) {
	objectID, err := b.object(object)
	if err != nil {
		return 0, err
	}

	if objectID.IsRoot {
		return b.state.mapLength(objectID), nil
	}

	operation, ok := b.state.operation(objectID.OpID)
	if !ok {
		return 0, fmt.Errorf("object does not exist")
	}

	if operation.Action == opset.ActionMakeMap ||
		operation.Action == opset.ActionMakeTable {
		return b.state.mapLength(objectID), nil
	}

	if operation.Action != opset.ActionMakeList &&
		operation.Action != opset.ActionMakeText {
		return 0, fmt.Errorf("object does not have a length")
	}

	sequence := b.state.sequenceValues(objectID.OpID)

	if operation.Action == opset.ActionMakeText {
		total := uint64(0)
		for _, value := range sequence {
			total += sequenceValueUTF16Width(value)
		}

		return total, nil
	}

	return uint64(len(sequence)), nil
}
