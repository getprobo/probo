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
	"unicode/utf8"
	"unsafe"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

type encodedColumn struct {
	specification uint32
	data          []byte
}

func EncodeChange(change *opset.Change) ([]byte, error) {
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

	// Hash the framed body without copying it into a second buffer.
	length := appendULEB(nil, uint64(len(body)))
	hasher := sha256.New()
	_, _ = hasher.Write([]byte{byte(opset.ChunkChange)})
	_, _ = hasher.Write(length)
	_, _ = hasher.Write(body)

	var hash opset.ChangeHash

	_ = hasher.Sum(hash[:0])
	change.Hash = new(hash)

	raw := []byte{0x85, 0x6f, 0x4a, 0x83}
	raw = append(raw, hash[:4]...)
	raw = append(raw, byte(opset.ChunkChange))
	raw = append(raw, length...)
	raw = append(raw, body...)
	change.Raw = append([]byte(nil), raw...)

	return raw, nil
}

func changeActorTable(
	change *opset.Change,
) ([]opset.ActorID, map[opset.ActorID]uint64, error) {
	actorSet := make(map[opset.ActorID]struct{})
	add := func(actor opset.ActorID) {
		if actor != "" && actor != change.Actor {
			actorSet[actor] = struct{}{}
		}
	}

	for i, operation := range change.Operations {
		expectedID := opset.OpID{
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

	actors := make([]opset.ActorID, 0, len(actorSet))
	for actor := range actorSet {
		actors = append(actors, actor)
	}

	sort.Slice(
		actors,
		func(i, j int) bool {
			return actors[i].Compare(actors[j]) < 0
		},
	)

	indexes := map[opset.ActorID]uint64{change.Actor: 0}
	for i, actor := range actors {
		indexes[actor] = uint64(i + 1)
	}

	return actors, indexes, nil
}

func encodeOperationColumns(
	change *opset.Change,
	actorIndexes map[opset.ActorID]uint64,
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

		// An all-false expand column carries no information and is left out, so a
		// mark that expands in neither direction encodes without one.
		if operation.MarkExpand != nil && *operation.MarkExpand {
			markExpands[i] = true
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

func encodeScalar(value *opset.Scalar) (optional[uint64], []byte, error) {
	if value == nil {
		return some(uint64(opset.ScalarNull)), nil, nil
	}

	if value.Type > 0x0f {
		return optional[uint64]{}, nil, fmt.Errorf("scalar type %d exceeds 0x0f", value.Type)
	}

	if value.Type == opset.ScalarString && !utf8.ValidString(value.String) {
		return optional[uint64]{}, nil, fmt.Errorf("string scalar is not valid UTF-8")
	}

	var data []byte

	switch value.Type {
	case opset.ScalarNull:
	case opset.ScalarFalse, opset.ScalarTrue:
	case opset.ScalarUint:
		data = appendULEB(data, value.Uint)
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		data = appendLEB(data, value.Int)
	case opset.ScalarFloat64:
		var encoded [8]byte
		binary.LittleEndian.PutUint64(encoded[:], math.Float64bits(value.Float))
		data = encoded[:]
	case opset.ScalarString:
		data = []byte(value.String)
	case opset.ScalarBytes:
		data = append([]byte(nil), value.Bytes...)
	default:
		data = append([]byte(nil), value.Raw...)
	}

	meta := uint64(len(data))<<4 | uint64(value.Type)

	return some(meta), data, nil
}

func appendColumns(data []byte, columns []encodedColumn) []byte {
	sort.Slice(
		columns,
		func(i, j int) bool {
			return columns[i].specification < columns[j].specification
		},
	)

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

func appendHashesNative(data []byte, hashes []opset.ChangeHash) []byte {
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

// reconstructionScratch bounds the bytes EncodeChange allocates while rebuilding
// one snapshot change, on top of the change value already in hand.
//
// The encoded columns, the body, and both framed copies can be live together,
// and append grows a buffer by at most twice. Six times the encoded length
// covers those buffers. Column inputs that are not part of the encoded bytes
// (the per-operation slices and predecessor runs) are added separately.
func reconstructionScratch(change *opset.Change) (uint64, error) {
	encoded, err := encodedBodySize(change)
	if err != nil {
		return 0, err
	}

	if encoded > math.MaxUint64/6 {
		return 0, fmt.Errorf("change encoding size overflows uint64")
	}

	operations := len(change.Operations)

	perOperation := 6*uint64(unsafe.Sizeof(optional[uint64]{})) +
		uint64(unsafe.Sizeof(optional[int64]{})) +
		2*uint64(unsafe.Sizeof(optional[string]{})) +
		2*uint64(unsafe.Sizeof(bool(false)))
	if perOperation == 0 || uint64(operations) > math.MaxUint64/perOperation {
		return 0, fmt.Errorf("change operation count overflows scratch size")
	}

	fixed := uint64(operations) * perOperation

	var predecessors uint64
	for i := range change.Operations {
		if predecessors > math.MaxUint64-uint64(len(change.Operations[i].Predecessors)) {
			return 0, fmt.Errorf("change predecessor count overflows uint64")
		}

		predecessors += uint64(len(change.Operations[i].Predecessors))
	}

	predWidth := uint64(unsafe.Sizeof(optional[uint64]{})) +
		uint64(unsafe.Sizeof(optional[int64]{}))
	if predecessors > math.MaxUint64/2/predWidth {
		return 0, fmt.Errorf("change predecessor scratch overflows uint64")
	}

	// encodedBodySize builds predecessor and delta columns while measuring.
	// Those slices can still be on the heap when EncodeChange allocates.
	sizing := uint64(operations)*uint64(unsafe.Sizeof(optional[int64]{})) +
		predecessors*predWidth

	scratch := fixed + predecessors*predWidth*2 + encoded*6 + sizing
	if scratch < fixed {
		return 0, fmt.Errorf("change reconstruction scratch overflows uint64")
	}

	return scratch, nil
}

func encodedBodySize(change *opset.Change) (uint64, error) {
	actors, indexes, err := changeActorTable(change)
	if err != nil {
		return 0, err
	}

	var total uint64

	add := func(extra int) error {
		if extra < 0 {
			return fmt.Errorf("change encoding size overflows uint64")
		}

		return addUint(&total, uint64(extra))
	}

	if err := add(ulebSize(uint64(len(change.Dependencies))) + 32*len(change.Dependencies)); err != nil {
		return 0, err
	}

	if err := add(ulebSize(uint64(len(change.Actor))) + len(change.Actor)); err != nil {
		return 0, err
	}

	if err := add(ulebSize(change.Sequence)); err != nil {
		return 0, err
	}

	if err := add(ulebSize(change.StartOp)); err != nil {
		return 0, err
	}

	if err := add(lebSize(change.Time)); err != nil {
		return 0, err
	}

	if err := add(ulebSize(uint64(len(change.Message))) + len(change.Message)); err != nil {
		return 0, err
	}

	if err := add(ulebSize(uint64(len(actors)))); err != nil {
		return 0, err
	}

	for _, actor := range actors {
		if err := add(ulebSize(uint64(len(actor))) + len(actor)); err != nil {
			return 0, err
		}
	}

	columns, err := encodedColumnSizes(change, indexes)
	if err != nil {
		return 0, err
	}

	if err := add(ulebSize(uint64(len(columns)))); err != nil {
		return 0, err
	}

	for _, column := range columns {
		if err := add(ulebSize(uint64(column.specification))); err != nil {
			return 0, err
		}

		if err := add(ulebSize(column.size)); err != nil {
			return 0, err
		}
	}

	for _, column := range columns {
		if err := addUint(&total, column.size); err != nil {
			return 0, err
		}
	}

	if err := addUint(&total, uint64(len(change.ExtraBytes))); err != nil {
		return 0, err
	}

	return total, nil
}

func addUint(total *uint64, extra uint64) error {
	if extra > math.MaxUint64-*total {
		return fmt.Errorf("change encoding size overflows uint64")
	}

	*total += extra

	return nil
}

type sizedColumn struct {
	specification uint32
	size          uint64
}

func encodedColumnSizes(
	change *opset.Change,
	indexes map[opset.ActorID]uint64,
) ([]sizedColumn, error) {
	n := len(change.Operations)
	columns := make([]sizedColumn, 0, 14)
	add := func(specification uint32, size uint64) {
		if size == 0 {
			return
		}

		columns = append(columns, sizedColumn{
			specification: specification,
			size:          size,
		})
	}

	objActors, err := rleSize(
		n,
		func(i int) optional[uint64] {
			operation := &change.Operations[i]
			if operation.Object.IsRoot {
				return optional[uint64]{}
			}

			return some(indexes[operation.Object.OpID.Actor])
		},
		ulebSize,
	)
	if err != nil {
		return nil, err
	}

	add(1, objActors)

	objCounters, err := rleSize(
		n,
		func(i int) optional[uint64] {
			operation := &change.Operations[i]
			if operation.Object.IsRoot {
				return optional[uint64]{}
			}

			return some(operation.Object.OpID.Counter)
		},
		ulebSize,
	)
	if err != nil {
		return nil, err
	}

	add(2, objCounters)

	keyActors, err := rleSize(
		n,
		func(i int) optional[uint64] {
			operation := &change.Operations[i]
			if operation.Key.Element == nil {
				return optional[uint64]{}
			}

			return some(indexes[operation.Key.Element.Actor])
		},
		ulebSize,
	)
	if err != nil {
		return nil, err
	}

	add(17, keyActors)

	keyCounterSize, err := deltaRLESize(
		n,
		func(i int) optional[int64] {
			operation := &change.Operations[i]
			switch {
			case operation.Key.Property != nil:
				return optional[int64]{}
			case operation.Key.IsHead:
				return some(int64(0))
			case operation.Key.Element != nil:
				return some(int64(operation.Key.Element.Counter))
			default:
				return optional[int64]{}
			}
		},
	)
	if err != nil {
		return nil, err
	}

	add(19, keyCounterSize)

	keyStrings, err := rleSize(
		n,
		func(i int) optional[string] {
			property := change.Operations[i].Key.Property
			if property == nil {
				return optional[string]{}
			}

			return some(*property)
		},
		func(value string) int {
			return ulebSize(uint64(len(value))) + len(value)
		},
	)
	if err != nil {
		return nil, err
	}

	add(21, keyStrings)

	inserts, err := booleanColumnSize(
		n,
		func(i int) bool {
			return change.Operations[i].Insert
		},
	)
	if err != nil {
		return nil, err
	}

	add(52, inserts)

	actions, err := rleSize(
		n,
		func(i int) optional[uint64] {
			return some(uint64(change.Operations[i].Action))
		},
		ulebSize,
	)
	if err != nil {
		return nil, err
	}

	add(66, actions)

	metadata, valueLen, err := valueColumnSize(change)
	if err != nil {
		return nil, err
	}

	add(86, metadata)
	add(87, valueLen)

	groups, err := rleSize(
		n,
		func(i int) optional[uint64] {
			return some(uint64(len(change.Operations[i].Predecessors)))
		},
		ulebSize,
	)
	if err != nil {
		return nil, err
	}

	add(112, groups)

	predActors, predCounters, err := predecessorColumnSizes(change, indexes)
	if err != nil {
		return nil, err
	}

	add(113, predActors)
	add(115, predCounters)

	var hasExpand bool

	for i := range change.Operations {
		if change.Operations[i].MarkExpand != nil && *change.Operations[i].MarkExpand {
			hasExpand = true
			break
		}
	}

	if hasExpand {
		expands, err := booleanColumnSize(
			n,
			func(i int) bool {
				expand := change.Operations[i].MarkExpand

				return expand != nil && *expand
			},
		)
		if err != nil {
			return nil, err
		}

		add(148, expands)
	}

	names, err := rleSize(
		n,
		func(i int) optional[string] {
			name := change.Operations[i].MarkName
			if name == nil {
				return optional[string]{}
			}

			return some(*name)
		},
		func(value string) int {
			return ulebSize(uint64(len(value))) + len(value)
		},
	)
	if err != nil {
		return nil, err
	}

	add(165, names)

	return columns, nil
}

func valueColumnSize(change *opset.Change) (uint64, uint64, error) {
	n := len(change.Operations)

	metadata, err := rleSize(
		n,
		func(i int) optional[uint64] {
			meta, _, metaErr := scalarMeta(change.Operations[i].Value)
			if metaErr != nil {
				return optional[uint64]{}
			}

			return some(meta)
		},
		ulebSize,
	)
	if err != nil {
		return 0, 0, err
	}

	var valueLen uint64

	for i := range change.Operations {
		_, dataLen, err := scalarMeta(change.Operations[i].Value)
		if err != nil {
			return 0, 0, err
		}

		if valueLen > math.MaxUint64-dataLen {
			return 0, 0, fmt.Errorf("change value column overflows uint64")
		}

		valueLen += dataLen
	}

	return metadata, valueLen, nil
}

func scalarMeta(value *opset.Scalar) (uint64, uint64, error) {
	if value == nil {
		return uint64(opset.ScalarNull), 0, nil
	}

	if value.Type > 0x0f {
		return 0, 0, fmt.Errorf("scalar type %d exceeds 0x0f", value.Type)
	}

	if value.Type == opset.ScalarString && !utf8.ValidString(value.String) {
		return 0, 0, fmt.Errorf("string scalar is not valid UTF-8")
	}

	var dataLen uint64

	switch value.Type {
	case opset.ScalarNull, opset.ScalarFalse, opset.ScalarTrue:
	case opset.ScalarUint:
		dataLen = uint64(ulebSize(value.Uint))
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		dataLen = uint64(lebSize(value.Int))
	case opset.ScalarFloat64:
		dataLen = 8
	case opset.ScalarString:
		dataLen = uint64(len(value.String))
	case opset.ScalarBytes:
		dataLen = uint64(len(value.Bytes))
	default:
		dataLen = uint64(len(value.Raw))
	}

	if dataLen > math.MaxUint64>>4 {
		return 0, 0, fmt.Errorf("scalar length overflows metadata")
	}

	return dataLen<<4 | uint64(value.Type), dataLen, nil
}

func predecessorColumnSizes(
	change *opset.Change,
	indexes map[opset.ActorID]uint64,
) (uint64, uint64, error) {
	var count int
	for i := range change.Operations {
		count += len(change.Operations[i].Predecessors)
	}

	if count == 0 {
		return 0, 0, nil
	}

	actors := make([]optional[uint64], count)
	counters := make([]optional[int64], count)
	index := 0

	for i := range change.Operations {
		for _, predecessor := range change.Operations[i].Predecessors {
			actors[index] = some(indexes[predecessor.Actor])
			counters[index] = some(int64(predecessor.Counter))
			index++
		}
	}

	actorSize, err := rleSize(
		count,
		func(i int) optional[uint64] {
			return actors[i]
		},
		ulebSize,
	)
	if err != nil {
		return 0, 0, err
	}

	var previous int64

	deltas := make([]optional[int64], count)

	for i, counter := range counters {
		if !counter.valid {
			continue
		}

		deltas[i] = some(counter.value - previous)
		previous = counter.value
	}

	counterSize, err := rleSize(
		count,
		func(i int) optional[int64] {
			return deltas[i]
		},
		lebSize,
	)
	if err != nil {
		return 0, 0, err
	}

	return actorSize, counterSize, nil
}

func deltaRLESize(n int, at func(int) optional[int64]) (uint64, error) {
	if n == 0 {
		return 0, nil
	}

	deltas := make([]optional[int64], n)

	var previous int64

	for i := range n {
		value := at(i)
		if !value.valid {
			continue
		}

		deltas[i] = some(value.value - previous)
		previous = value.value
	}

	return rleSize(
		n,
		func(i int) optional[int64] {
			return deltas[i]
		},
		lebSize,
	)
}

func rleSize[T comparable](
	n int,
	at func(int) optional[T],
	valueSize func(T) int,
) (uint64, error) {
	if n == 0 || allNullFunc(n, at) {
		return 0, nil
	}

	var total uint64

	add := func(extra int) error {
		if extra < 0 || uint64(extra) > math.MaxUint64-total {
			return fmt.Errorf("run length column overflows uint64")
		}

		total += uint64(extra)

		return nil
	}

	for index := 0; index < n; {
		current := at(index)
		if !current.valid {
			end := index + 1
			for end < n && !at(end).valid {
				end++
			}

			if err := add(lebSize(0) + ulebSize(uint64(end-index))); err != nil {
				return 0, err
			}

			index = end

			continue
		}

		if index+1 < n {
			next := at(index + 1)
			if next.valid && next.value == current.value {
				end := index + 2
				for end < n {
					step := at(end)
					if !step.valid || step.value != current.value {
						break
					}

					end++
				}

				if err := add(lebSize(int64(end-index)) + valueSize(current.value)); err != nil {
					return 0, err
				}

				index = end

				continue
			}
		}

		end := index + 1
		for end < n && at(end).valid {
			if end+1 < n {
				left := at(end)

				right := at(end + 1)
				if left.valid && right.valid && left.value == right.value {
					break
				}
			}

			end++
		}

		if err := add(lebSize(-int64(end - index))); err != nil {
			return 0, err
		}

		for i := index; i < end; i++ {
			if err := add(valueSize(at(i).value)); err != nil {
				return 0, err
			}
		}

		index = end
	}

	return total, nil
}

func allNullFunc[T any](n int, at func(int) optional[T]) bool {
	for i := range n {
		if at(i).valid {
			return false
		}
	}

	return true
}

func booleanColumnSize(n int, at func(int) bool) (uint64, error) {
	if n == 0 {
		return 0, nil
	}

	var (
		total   uint64
		current bool
		count   uint64
	)

	add := func(extra int) error {
		if extra < 0 || uint64(extra) > math.MaxUint64-total {
			return fmt.Errorf("boolean column overflows uint64")
		}

		total += uint64(extra)

		return nil
	}

	for i := range n {
		value := at(i)
		if value == current {
			count++
			continue
		}

		if err := add(ulebSize(count)); err != nil {
			return 0, err
		}

		current = value
		count = 1
	}

	if err := add(ulebSize(count)); err != nil {
		return 0, err
	}

	return total, nil
}

func lebSize(value int64) int {
	size := 0

	for {
		current := byte(value & 0x7f)
		value >>= 7
		size++

		done := (value == 0 && current&0x40 == 0) ||
			(value == -1 && current&0x40 != 0)
		if done {
			return size
		}
	}
}
