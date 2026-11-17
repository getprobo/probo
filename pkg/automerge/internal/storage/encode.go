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
	"iter"
	"math"
	"slices"
	"sort"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

type encodedColumn struct {
	specification uint32
	data          []byte
}

// EncodeChange encodes a change chunk from the operations the change holds,
// and records the resulting hash and bytes on the change.
func EncodeChange(change *opset.Change) ([]byte, error) {
	return encodeChange(change, slices.All(change.Operations))
}

// encodeChange encodes a change chunk whose operations come from a sequence,
// which is read twice. It lets a caller that stores operations elsewhere
// encode a change without first assembling them into one slice.
func encodeChange(
	change *opset.Change,
	operations iter.Seq2[int, opset.Operation],
) ([]byte, error) {
	if change.Actor.Len() == 0 {
		return nil, fmt.Errorf("change actor cannot be empty")
	}

	if change.Sequence == 0 || change.StartOp == 0 {
		return nil, fmt.Errorf("change sequence and start operation must be positive")
	}

	actors, actorIndexes, err := changeActorTable(change, operations)
	if err != nil {
		return nil, err
	}

	columns, err := encodeOperationColumns(operations, actorIndexes)
	if err != nil {
		return nil, err
	}

	// The body is sized up front so that appending the columns to it does not
	// reallocate it along the way.
	size := len(change.Dependencies)*len(opset.ChangeHash{}) +
		change.Actor.Len() +
		len(change.Message) +
		len(change.ExtraBytes) +
		16*(len(actors)+len(columns)+4)

	for _, actor := range actors {
		size += actor.Len()
	}

	for _, column := range columns {
		size += len(column.data)
	}

	body := make([]byte, 0, size)

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

	// A change's hash covers its chunk type, body length and body.
	header := appendULEB([]byte{byte(opset.ChunkChange)}, uint64(len(body)))

	digest := sha256.New()
	digest.Write(header)
	digest.Write(body)

	var hash opset.ChangeHash

	digest.Sum(hash[:0])
	change.Hash = new(hash)

	raw := make([]byte, 0, len(magic)+4+len(header)+len(body))
	raw = append(raw, magic[:]...)
	raw = append(raw, hash[:4]...)
	raw = append(raw, header...)
	raw = append(raw, body...)
	change.Raw = slices.Clone(raw)

	return raw, nil
}

func changeActorTable(
	change *opset.Change,
	operations iter.Seq2[int, opset.Operation],
) ([]opset.ActorID, map[opset.ActorID]uint64, error) {
	actorSet := make(map[opset.ActorID]struct{})

	// Nearly every reference names the actor the previous one did, so the
	// set is only consulted when the actor changes.
	previous := change.Actor
	add := func(actor opset.ActorID) {
		if actor == previous {
			return
		}

		previous = actor

		if !actor.IsZero() && actor != change.Actor {
			actorSet[actor] = struct{}{}
		}
	}

	for i, operation := range operations {
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

// changeColumnWriter encodes the operation columns of a change chunk. Each
// column is written as its operations arrive, so encoding costs the encoded
// size and not one table entry per operation and column.
type changeColumnWriter struct {
	actorIndexes map[opset.ActorID]uint64

	// Consecutive references nearly always name the same actor, so its index
	// is remembered instead of looked up again.
	cachedActor opset.ActorID
	cachedIndex uint64
	cached      bool

	objActors     rleWriter[uint64]
	objCounters   rleWriter[uint64]
	keyActors     rleWriter[uint64]
	keyCounters   deltaWriter
	keyStrings    rleWriter[string]
	inserts       booleanWriter
	actions       rleWriter[uint64]
	valueMetadata rleWriter[uint64]
	valueData     []byte
	predGroups    rleWriter[uint64]
	predActors    rleWriter[uint64]
	predCounters  deltaWriter
	markExpands   booleanWriter
	markNames     rleWriter[string]
	hasMarkExpand bool
}

func newChangeColumnWriter(
	actorIndexes map[opset.ActorID]uint64,
) *changeColumnWriter {
	return &changeColumnWriter{
		actorIndexes:  actorIndexes,
		objActors:     newRLEWriter(appendULEB),
		objCounters:   newRLEWriter(appendULEB),
		keyActors:     newRLEWriter(appendULEB),
		keyCounters:   newDeltaWriter(),
		keyStrings:    newStringWriter(),
		actions:       newRLEWriter(appendULEB),
		valueMetadata: newRLEWriter(appendULEB),
		predGroups:    newRLEWriter(appendULEB),
		predActors:    newRLEWriter(appendULEB),
		predCounters:  newDeltaWriter(),
		markNames:     newStringWriter(),
	}
}

func (w *changeColumnWriter) actorIndex(actor opset.ActorID) (uint64, bool) {
	if w.cached && actor == w.cachedActor {
		return w.cachedIndex, true
	}

	index, ok := w.actorIndexes[actor]
	if ok {
		w.cachedActor, w.cachedIndex, w.cached = actor, index, true
	}

	return index, ok
}

// append writes the operation at position i of the change.
func (w *changeColumnWriter) append(i int, operation *opset.Operation) error {
	if operation.Object.IsRoot {
		w.objActors.appendNull()
		w.objCounters.appendNull()
	} else {
		index, ok := w.actorIndex(operation.Object.OpID.Actor)
		if !ok {
			return fmt.Errorf("operation %d object actor is unknown", i)
		}

		w.objActors.append(index)
		w.objCounters.append(operation.Object.OpID.Counter)
	}

	switch {
	case operation.Key.Property != nil:
		w.keyActors.appendNull()
		w.keyCounters.appendNull()
		w.keyStrings.append(*operation.Key.Property)
	case operation.Key.IsHead:
		w.keyActors.appendNull()
		w.keyCounters.append(0)
		w.keyStrings.appendNull()
	case operation.Key.Element != nil:
		index, ok := w.actorIndex(operation.Key.Element.Actor)
		if !ok {
			return fmt.Errorf("operation %d key actor is unknown", i)
		}

		w.keyActors.append(index)
		w.keyCounters.append(int64(operation.Key.Element.Counter))
		w.keyStrings.appendNull()
	default:
		return fmt.Errorf("operation %d has no key", i)
	}

	w.inserts.append(operation.Insert)
	w.actions.append(uint64(operation.Action))

	valueData, meta, err := appendScalar(w.valueData, operation.Value)
	if err != nil {
		return fmt.Errorf("cannot encode operation %d value: %w", i, err)
	}

	w.valueData = valueData
	w.valueMetadata.appendOptional(meta)

	w.predGroups.append(uint64(len(operation.Predecessors)))

	for _, predecessor := range operation.Predecessors {
		index, ok := w.actorIndex(predecessor.Actor)
		if !ok {
			return fmt.Errorf("operation %d predecessor actor is unknown", i)
		}

		w.predActors.append(index)
		w.predCounters.append(int64(predecessor.Counter))
	}

	// An all-false expand column carries no information and is left out, so a
	// mark that expands in neither direction encodes without one.
	expands := operation.MarkExpand != nil && *operation.MarkExpand
	w.hasMarkExpand = w.hasMarkExpand || expands
	w.markExpands.append(expands)

	if operation.MarkName != nil {
		w.markNames.append(*operation.MarkName)
	} else {
		w.markNames.appendNull()
	}

	return nil
}

// columns returns the non-empty encoded columns. The writer must not be used
// afterwards.
func (w *changeColumnWriter) columns() []encodedColumn {
	var markExpandData []byte
	if w.hasMarkExpand {
		markExpandData = w.markExpands.finish()
	}

	columns := []encodedColumn{
		{specification: 1, data: w.objActors.finish()},
		{specification: 2, data: w.objCounters.finish()},
		{specification: 17, data: w.keyActors.finish()},
		{specification: 19, data: w.keyCounters.finish()},
		{specification: 21, data: w.keyStrings.finish()},
		{specification: 52, data: w.inserts.finish()},
		{specification: 66, data: w.actions.finish()},
		{specification: 86, data: w.valueMetadata.finish()},
		{specification: 87, data: w.valueData},
		{specification: 112, data: w.predGroups.finish()},
		{specification: 113, data: w.predActors.finish()},
		{specification: 115, data: w.predCounters.finish()},
		{specification: 148, data: markExpandData},
		{specification: 165, data: w.markNames.finish()},
	}

	filtered := columns[:0]
	for _, column := range columns {
		if len(column.data) > 0 {
			filtered = append(filtered, column)
		}
	}

	return filtered
}

func encodeOperationColumns(
	operations iter.Seq2[int, opset.Operation],
	actorIndexes map[opset.ActorID]uint64,
) ([]encodedColumn, error) {
	writer := newChangeColumnWriter(actorIndexes)

	for i, operation := range operations {
		if err := writer.append(i, &operation); err != nil {
			return nil, err
		}
	}

	return writer.columns(), nil
}

func encodeScalar(value *opset.Scalar) (optional[uint64], []byte, error) {
	data, meta, err := appendScalar(nil, value)

	return meta, data, err
}

// appendScalar appends a value's payload to data and returns its metadata,
// which packs the payload length with the scalar type.
func appendScalar(
	data []byte,
	value *opset.Scalar,
) ([]byte, optional[uint64], error) {
	if value == nil {
		return data, some(uint64(opset.ScalarNull)), nil
	}

	if value.Type > 0x0f {
		return data, optional[uint64]{}, fmt.Errorf("scalar type %d exceeds 0x0f", value.Type)
	}

	if value.Type == opset.ScalarString && !utf8.ValidString(value.String) {
		return data, optional[uint64]{}, fmt.Errorf("string scalar is not valid UTF-8")
	}

	start := len(data)

	switch value.Type {
	case opset.ScalarNull:
	case opset.ScalarFalse, opset.ScalarTrue:
	case opset.ScalarUint:
		data = appendULEB(data, value.Uint)
	case opset.ScalarInt, opset.ScalarCounter, opset.ScalarTimestamp:
		data = appendLEB(data, value.Int)
	case opset.ScalarFloat64:
		data = binary.LittleEndian.AppendUint64(data, math.Float64bits(value.Float))
	case opset.ScalarString:
		data = append(data, value.String...)
	case opset.ScalarBytes:
		data = append(data, value.Bytes...)
	default:
		data = append(data, value.Raw...)
	}

	meta := uint64(len(data)-start)<<4 | uint64(value.Type)

	return data, some(meta), nil
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
