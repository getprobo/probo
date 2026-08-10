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

package storage

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
)

type encodedColumn struct {
	specification uint32
	data          []byte
}

func EncodeChange(change *Change) ([]byte, error) {
	if len(change.Actor) == 0 {
		return nil, fmt.Errorf("change actor cannot be empty")
	}

	if change.Sequence == 0 || change.StartOp == 0 {
		return nil, fmt.Errorf("change sequence and start operation must be positive")
	}

	actors, actorIndexes, err := changeActorTable(change)
	if err != nil {
		return nil, err
	}

	columns, err := encodeOperationColumns(change, actorIndexes)
	if err != nil {
		return nil, err
	}

	var body []byte

	body = appendHashesNative(body, change.Dependencies)
	body = appendLengthPrefixedNative(body, change.Actor.Bytes())
	body = appendULEB(body, change.Sequence)
	body = appendULEB(body, change.StartOp)
	body = appendLEB(body, change.Time)
	body = appendLengthPrefixedNative(body, []byte(change.Message))

	body = appendULEB(body, uint64(len(actors)))
	for _, actor := range actors {
		body = appendLengthPrefixedNative(body, actor.Bytes())
	}

	body = appendColumns(body, columns)
	body = append(body, change.ExtraBytes...)

	hashInput := []byte{byte(ChunkChange)}
	hashInput = appendULEB(hashInput, uint64(len(body)))
	hashInput = append(hashInput, body...)
	hash := ChangeHash(sha256.Sum256(hashInput))
	change.Hash = new(hash)

	raw := []byte{0x85, 0x6f, 0x4a, 0x83}
	raw = append(raw, hash[:4]...)
	raw = append(raw, byte(ChunkChange))
	raw = appendULEB(raw, uint64(len(body)))
	raw = append(raw, body...)
	change.Raw = append([]byte(nil), raw...)

	return raw, nil
}

func changeActorTable(
	change *Change,
) ([]ActorID, map[ActorID]uint64, error) {
	actorSet := make(map[ActorID]struct{})
	add := func(actor ActorID) {
		if actor != "" && actor != change.Actor {
			actorSet[actor] = struct{}{}
		}
	}

	for i, operation := range change.Operations {
		expectedID := OpID{
			Actor:   change.Actor,
			Counter: change.StartOp + uint64(i),
		}
		if operation.ID != expectedID {
			return nil, nil, fmt.Errorf(
				"operation %d ID %v does not match implicit ID %v",
				i,
				operation.ID,
				expectedID,
			)
		}

		if !operation.Object.IsRoot {
			add(operation.Object.OpID.Actor)
		}

		if operation.Key.Element != nil {
			add(operation.Key.Element.Actor)
		}

		for _, predecessor := range operation.Predecessors {
			add(predecessor.Actor)
		}
	}

	actors := make([]ActorID, 0, len(actorSet))
	for actor := range actorSet {
		actors = append(actors, actor)
	}

	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Compare(actors[j]) < 0
	})

	indexes := map[ActorID]uint64{change.Actor: 0}
	for i, actor := range actors {
		indexes[actor] = uint64(i + 1)
	}

	return actors, indexes, nil
}

func encodeOperationColumns(
	change *Change,
	actorIndexes map[ActorID]uint64,
) ([]encodedColumn, error) {
	count := len(change.Operations)
	objActors := make([]optional[uint64], count)
	objCounters := make([]optional[uint64], count)
	keyActors := make([]optional[uint64], count)
	keyCounters := make([]optional[int64], count)
	keyStrings := make([]optional[string], count)
	inserts := make([]bool, count)
	actions := make([]optional[uint64], count)
	valueMetadata := make([]optional[uint64], count)
	predGroups := make([]optional[uint64], count)
	markExpands := make([]bool, count)
	markNames := make([]optional[string], count)

	var (
		valueData     []byte
		predActors    []optional[uint64]
		predCounters  []optional[int64]
		hasMarkExpand bool
	)

	for i, operation := range change.Operations {
		if !operation.Object.IsRoot {
			actorIndex, ok := actorIndexes[operation.Object.OpID.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d object actor is unknown", i)
			}

			objActors[i] = some(actorIndex)
			objCounters[i] = some(operation.Object.OpID.Counter)
		}

		switch {
		case operation.Key.Property != nil:
			keyStrings[i] = some(*operation.Key.Property)
		case operation.Key.IsHead:
			keyCounters[i] = some(int64(0))
		case operation.Key.Element != nil:
			actorIndex, ok := actorIndexes[operation.Key.Element.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d key actor is unknown", i)
			}

			keyActors[i] = some(actorIndex)
			keyCounters[i] = some(int64(operation.Key.Element.Counter))
		default:
			return nil, fmt.Errorf("operation %d has no key", i)
		}

		inserts[i] = operation.Insert
		actions[i] = some(uint64(operation.Action))

		meta, data, err := encodeScalar(operation.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot encode operation %d value: %w", i, err)
		}

		valueMetadata[i] = meta

		valueData = append(valueData, data...)

		predGroups[i] = some(uint64(len(operation.Predecessors)))
		for _, predecessor := range operation.Predecessors {
			actorIndex, ok := actorIndexes[predecessor.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d predecessor actor is unknown", i)
			}

			predActors = append(predActors, some(actorIndex))
			predCounters = append(predCounters, some(int64(predecessor.Counter)))
		}

		if operation.MarkExpand != nil {
			markExpands[i] = *operation.MarkExpand
			hasMarkExpand = true
		}

		if operation.MarkName != nil {
			markNames[i] = some(*operation.MarkName)
		}
	}

	var markExpandData []byte
	if hasMarkExpand {
		markExpandData = encodeBooleans(markExpands)
	}

	columns := []encodedColumn{
		{specification: 1, data: encodeRLE(objActors, appendULEB)},
		{specification: 2, data: encodeRLE(objCounters, appendULEB)},
		{specification: 17, data: encodeRLE(keyActors, appendULEB)},
		{specification: 19, data: encodeDelta(keyCounters)},
		{specification: 21, data: encodeStrings(keyStrings)},
		{specification: 52, data: encodeBooleans(inserts)},
		{specification: 66, data: encodeRLE(actions, appendULEB)},
		{specification: 86, data: encodeRLE(valueMetadata, appendULEB)},
		{specification: 87, data: valueData},
		{specification: 112, data: encodeRLE(predGroups, appendULEB)},
		{specification: 113, data: encodeRLE(predActors, appendULEB)},
		{specification: 115, data: encodeDelta(predCounters)},
		{specification: 148, data: markExpandData},
		{specification: 165, data: encodeStrings(markNames)},
	}

	filtered := columns[:0]
	for _, column := range columns {
		if len(column.data) > 0 {
			filtered = append(filtered, column)
		}
	}

	return filtered, nil
}

func encodeScalar(value *Scalar) (optional[uint64], []byte, error) {
	if value == nil {
		return some(uint64(ScalarNull)), nil, nil
	}

	var data []byte

	switch value.Type {
	case ScalarNull:
	case ScalarFalse, ScalarTrue:
	case ScalarUint:
		data = appendULEB(data, value.Uint)
	case ScalarInt, ScalarCounter, ScalarTimestamp:
		data = appendLEB(data, value.Int)
	case ScalarFloat64:
		var encoded [8]byte
		binary.LittleEndian.PutUint64(encoded[:], math.Float64bits(value.Float))
		data = encoded[:]
	case ScalarString:
		data = []byte(value.String)
	case ScalarBytes:
		data = append([]byte(nil), value.Bytes...)
	default:
		data = append([]byte(nil), value.Raw...)
	}

	meta := uint64(len(data))<<4 | uint64(value.Type)

	return some(meta), data, nil
}

func appendColumns(data []byte, columns []encodedColumn) []byte {
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].specification < columns[j].specification
	})

	data = appendULEB(data, uint64(len(columns)))
	for _, column := range columns {
		data = appendULEB(data, uint64(column.specification))
		data = appendULEB(data, uint64(len(column.data)))
	}

	for _, column := range columns {
		data = append(data, column.data...)
	}

	return data
}

func encodeRLE[T comparable](
	values []optional[T],
	appendValue func([]byte, T) []byte,
) []byte {
	if allNull(values) {
		return nil
	}

	// Change hashes cover these encoded bytes. Use the reference encoder's
	// canonical run grouping so another implementation does not re-encode an
	// otherwise valid change under a different hash.
	var data []byte

	for index := 0; index < len(values); {
		if !values[index].valid {
			end := index + 1
			for end < len(values) && !values[end].valid {
				end++
			}

			data = appendLEB(data, 0)
			data = appendULEB(data, uint64(end-index))
			index = end

			continue
		}

		if index+1 < len(values) &&
			values[index+1].valid &&
			values[index+1].value == values[index].value {
			end := index + 2
			for end < len(values) &&
				values[end].valid &&
				values[end].value == values[index].value {
				end++
			}

			data = appendLEB(data, int64(end-index))
			data = appendValue(data, values[index].value)
			index = end

			continue
		}

		end := index + 1
		for end < len(values) && values[end].valid {
			if end+1 < len(values) &&
				values[end+1].valid &&
				values[end+1].value == values[end].value {
				break
			}

			end++
		}

		data = appendLEB(data, -int64(end-index))
		for _, value := range values[index:end] {
			data = appendValue(data, value.value)
		}

		index = end
	}

	return data
}

func encodeDelta(values []optional[int64]) []byte {
	var (
		previous int64
		deltas   = make([]optional[int64], len(values))
	)
	for i, value := range values {
		if !value.valid {
			continue
		}

		deltas[i] = some(value.value - previous)
		previous = value.value
	}

	return encodeRLE(deltas, appendLEB)
}

func encodeStrings(values []optional[string]) []byte {
	return encodeRLE(
		values,
		func(data []byte, value string) []byte {
			return appendLengthPrefixedNative(data, []byte(value))
		},
	)
}

func encodeBooleans(values []bool) []byte {
	if len(values) == 0 {
		return nil
	}

	var (
		data    []byte
		current bool
		count   uint64
	)
	for _, value := range values {
		if value == current {
			count++
			continue
		}

		data = appendULEB(data, count)
		current = value
		count = 1
	}

	return appendULEB(data, count)
}

func appendLEB(data []byte, value int64) []byte {
	for {
		current := byte(value & 0x7f)
		value >>= 7

		done := (value == 0 && current&0x40 == 0) ||
			(value == -1 && current&0x40 != 0)
		if !done {
			current |= 0x80
		}

		data = append(data, current)
		if done {
			return data
		}
	}
}

func appendHashesNative(data []byte, hashes []ChangeHash) []byte {
	data = appendULEB(data, uint64(len(hashes)))
	for _, hash := range hashes {
		data = append(data, hash[:]...)
	}

	return data
}

func appendLengthPrefixedNative(data, value []byte) []byte {
	data = appendULEB(data, uint64(len(value)))
	return append(data, value...)
}

func some[T any](value T) optional[T] {
	return optional[T]{value: value, valid: true}
}

func allNull[T any](values []optional[T]) bool {
	for _, value := range values {
		if value.valid {
			return false
		}
	}

	return true
}
