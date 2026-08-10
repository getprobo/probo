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
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"slices"
	"unicode/utf8"
)

var magic = [4]byte{0x85, 0x6f, 0x4a, 0x83}

type decodedChunk struct {
	kind    ChunkType
	content []byte
	hash    *ChangeHash
	raw     []byte
}

type implicitOperationIDs struct {
	actorIndex uint64
	startOp    uint64
}

// Decode parses all chunks in data and validates the resulting dependency
// graph. It accepts document, change, and compressed change chunks.
func Decode(data []byte) (*Document, error) {
	return decode(data, true)
}

// DecodePartial parses chunks whose causal dependencies may already exist in
// another document. Column and operation ownership validation still happens
// while whole-history frontier validation is deferred to the caller.
func DecodePartial(data []byte) (*Document, error) {
	return decode(data, false)
}

// DecodeIncremental parses the complete chunk prefix and ignores an incomplete
// or corrupt trailing fragment after at least one valid chunk.
func DecodeIncremental(data []byte) (*Document, int, error) {
	r := &reader{data: data}
	consumed := 0

	for r.remaining() > 0 {
		start := r.offset
		if _, err := decodeChunk(r); err != nil {
			r.offset = start
			break
		}

		consumed = r.offset
	}

	if consumed == 0 {
		document, err := DecodePartial(data)

		return document, 0, err
	}

	document, err := DecodePartial(data[:consumed])

	return document, consumed, err
}

func decode(data []byte, validateHistory bool) (*Document, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("automerge file is empty")
	}

	r := &reader{data: data}
	document := &Document{}

	for r.remaining() > 0 {
		chunk, err := decodeChunk(r)
		if err != nil {
			return nil, fmt.Errorf("cannot decode chunk %d: %w", len(document.ChunkTypes), err)
		}

		document.ChunkTypes = append(document.ChunkTypes, chunk.kind)

		switch chunk.kind {
		case ChunkDocument:
			if len(document.ChunkTypes) != 1 {
				return nil, fmt.Errorf("only the first chunk may be a document chunk")
			}

			if err := decodeDocumentChunk(document, chunk.content); err != nil {
				return nil, fmt.Errorf("cannot decode document chunk: %w", err)
			}
		case ChunkChange, ChunkCompressedChange:
			change, actors, unknown, err := decodeChangeChunk(chunk.content, *chunk.hash)
			if err != nil {
				return nil, fmt.Errorf("cannot decode change chunk: %w", err)
			}

			change.Raw = append([]byte(nil), chunk.raw...)

			document.Changes = append(document.Changes, change)
			document.Actors = mergeActors(document.Actors, actors)
			document.UnknownColumns = append(document.UnknownColumns, unknown...)
		default:
			return nil, fmt.Errorf("unsupported chunk type %d", chunk.kind)
		}
	}

	if validateHistory {
		if err := validateDocument(document); err != nil {
			return nil, fmt.Errorf("invalid dependency graph: %w", err)
		}
	}

	return document, nil
}

func decodeChunk(r *reader) (decodedChunk, error) {
	start := r.offset

	header, err := r.bytes(4)
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read magic bytes: %w", err)
	}

	if !bytes.Equal(header, magic[:]) {
		return decodedChunk{}, fmt.Errorf("invalid magic bytes %x", header)
	}

	checksum, err := r.bytes(4)
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read checksum: %w", err)
	}

	rawType, err := r.byte()
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read chunk type: %w", err)
	}

	kind := ChunkType(rawType)
	if kind > ChunkCompressedChange {
		return decodedChunk{}, fmt.Errorf("unknown chunk type %d", kind)
	}

	lengthStart := r.offset

	length, err := r.uleb()
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read chunk length: %w", err)
	}

	lengthBytes := append([]byte(nil), r.data[lengthStart:r.offset]...)

	content, err := r.bytes(length)
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read chunk content: %w", err)
	}

	hashInput := make([]byte, 0, 1+len(lengthBytes)+len(content))
	hashKind := kind
	hashContent := content

	if kind == ChunkCompressedChange {
		hashKind = ChunkChange

		hashContent, err = inflate(content)
		if err != nil {
			return decodedChunk{}, fmt.Errorf("cannot inflate compressed change: %w", err)
		}

		lengthBytes = appendULEB(nil, uint64(len(hashContent)))
	}

	hashInput = append(hashInput, byte(hashKind))
	hashInput = append(hashInput, lengthBytes...)
	hashInput = append(hashInput, hashContent...)

	digest := sha256.Sum256(hashInput)
	if !bytes.Equal(checksum, digest[:4]) {
		return decodedChunk{}, fmt.Errorf(
			"checksum mismatch: encoded %x, calculated %x",
			checksum,
			digest[:4],
		)
	}

	chunk := decodedChunk{
		kind:    kind,
		content: hashContent,
		raw:     append([]byte(nil), r.data[start:r.offset]...),
	}
	if kind == ChunkChange || kind == ChunkCompressedChange {
		hash := ChangeHash(digest)
		chunk.hash = &hash
	}

	return chunk, nil
}

func appendULEB(destination []byte, value uint64) []byte {
	for {
		current := byte(value & 0x7f)

		value >>= 7
		if value != 0 {
			current |= 0x80
		}

		destination = append(destination, current)
		if value == 0 {
			return destination
		}
	}
}

func decodeDocumentChunk(document *Document, data []byte) error {
	r := &reader{data: data}

	actors, err := decodeActorArray(r, true)
	if err != nil {
		return fmt.Errorf("cannot decode actors: %w", err)
	}

	heads, err := decodeHashArray(r, true)
	if err != nil {
		return fmt.Errorf("cannot decode heads: %w", err)
	}

	changeMetadata, err := parseColumnMetadata(r, true)
	if err != nil {
		return fmt.Errorf("cannot decode change column metadata: %w", err)
	}

	operationMetadata, err := parseColumnMetadata(r, true)
	if err != nil {
		return fmt.Errorf("cannot decode operation column metadata: %w", err)
	}

	changeColumns, err := readColumns(r, changeMetadata)
	if err != nil {
		return fmt.Errorf("cannot decode change columns: %w", err)
	}

	operationColumns, err := readColumns(r, operationMetadata)
	if err != nil {
		return fmt.Errorf("cannot decode operation columns: %w", err)
	}

	changes, unknownChanges, err := decodeDocumentChanges(changeColumns, actors)
	if err != nil {
		return fmt.Errorf("cannot decode changes: %w", err)
	}

	operations, unknownOperations, err := decodeOperations(operationColumns, actors, false, nil)
	if err != nil {
		return fmt.Errorf("cannot decode operations: %w", err)
	}

	if err := assignOperations(changes, operations); err != nil {
		return fmt.Errorf("cannot assign operations: %w", err)
	}

	headIndexes := make([]uint64, len(heads))
	for i := range heads {
		index, err := r.uleb()
		if err != nil {
			return fmt.Errorf("cannot decode head index %d: %w", i, err)
		}

		if index >= uint64(len(changes)) {
			return fmt.Errorf("head index %d is out of bounds", index)
		}

		headIndexes[i] = index

		hash := heads[i]
		if changes[index].Hash != nil && *changes[index].Hash != hash {
			return fmt.Errorf("head index %d is assigned conflicting hashes", index)
		}

		changes[index].Hash = &hash
	}

	if r.remaining() != 0 {
		return fmt.Errorf("document chunk has %d trailing bytes", r.remaining())
	}

	if err := validateSnapshotGraph(changes, headIndexes); err != nil {
		return err
	}

	document.Actors = actors
	document.Heads = heads
	document.Changes = changes

	document.UnknownColumns = append(unknownChanges, unknownOperations...)

	return nil
}

func decodeDocumentChanges(
	columns map[uint32]column,
	actors []ActorID,
) ([]Change, []RawColumn, error) {
	if len(columns) == 0 {
		return nil, nil, nil
	}

	actorData, err := requireColumn(columns, 1)
	if err != nil {
		return nil, nil, err
	}

	actorIndexes, err := decodeULEBColumn(actorData)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode actor column: %w", err)
	}

	count := len(actorIndexes)
	if err := requireItems("actor", actorIndexes, count, false); err != nil {
		return nil, nil, err
	}

	sequence, err := decodeRequiredDelta(columns, 3, "sequence", count)
	if err != nil {
		return nil, nil, err
	}

	maxOps, err := decodeRequiredDelta(columns, 19, "maxOp", count)
	if err != nil {
		return nil, nil, err
	}

	times, err := decodeOptionalSignedDelta(columns, 35, "time", count)
	if err != nil {
		return nil, nil, err
	}

	messages, err := decodeOptionalStrings(columns, 53, "message", count)
	if err != nil {
		return nil, nil, err
	}

	groupData, err := requireColumn(columns, 64)
	if err != nil {
		return nil, nil, err
	}

	groups, err := decodeULEBColumn(groupData)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode dependency groups: %w", err)
	}

	if err := requireItems("dependency group", groups, count, false); err != nil {
		return nil, nil, err
	}

	dependencyCount, err := sumGroups(groups)
	if err != nil {
		return nil, nil, err
	}

	dependencies := make([]optional[uint64], 0)

	if dependencyCount > 0 {
		dependencyData, err := requireColumn(columns, 67)
		if err != nil {
			return nil, nil, err
		}

		dependencies, err = decodeDeltaColumn(dependencyData)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode dependencies: %w", err)
		}

		if err := requireItems("dependency", dependencies, dependencyCount, false); err != nil {
			return nil, nil, err
		}
	}

	extras, err := decodeOptionalScalars(columns, 86, 87, count)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode change extras: %w", err)
	}

	changes := make([]Change, count)
	dependencyOffset := 0

	for i := range changes {
		actorIndex := actorIndexes[i].value
		if actorIndex >= uint64(len(actors)) {
			return nil, nil, fmt.Errorf("change %d actor index %d is out of bounds", i, actorIndex)
		}

		changes[i] = Change{
			Actor:    actors[actorIndex],
			Sequence: sequence[i].value,
			MaxOp:    maxOps[i].value,
		}
		if times[i].valid {
			changes[i].Time = times[i].value
		}

		if messages[i].valid {
			changes[i].Message = messages[i].value
		}

		if extras[i].valid {
			extra := extras[i].value
			changes[i].Extra = &extra
		}

		groupLength := int(groups[i].value)

		changes[i].DependencyIndexes = make([]uint64, groupLength)
		for j := range groupLength {
			changes[i].DependencyIndexes[j] = dependencies[dependencyOffset+j].value
		}

		dependencyOffset += groupLength
	}

	unknown := collectUnknown(columns)

	return changes, unknown, nil
}

func decodeChangeChunk(
	data []byte,
	hash ChangeHash,
) (Change, []ActorID, []RawColumn, error) {
	r := &reader{data: data}

	dependencies, err := decodeHashArray(r, false)
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode dependencies: %w", err)
	}

	actorBytes, err := decodeLengthPrefixed(r)
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode actor: %w", err)
	}

	actor, err := NewActorID(actorBytes)
	if err != nil {
		return Change{}, nil, nil, err
	}

	sequence, err := r.uleb()
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode sequence: %w", err)
	}

	if sequence == 0 {
		return Change{}, nil, nil, fmt.Errorf("sequence is zero")
	}

	startOp, err := r.uleb()
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode start op: %w", err)
	}

	if startOp == 0 {
		return Change{}, nil, nil, fmt.Errorf("start op is zero")
	}

	timestamp, err := r.leb()
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode time: %w", err)
	}

	messageBytes, err := decodeLengthPrefixed(r)
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode message: %w", err)
	}

	if !utf8.Valid(messageBytes) {
		return Change{}, nil, nil, fmt.Errorf("message is not valid UTF-8")
	}

	otherActors, err := decodeActorArray(r, true)
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode other actors: %w", err)
	}

	if slices.Contains(otherActors, actor) {
		return Change{}, nil, nil, fmt.Errorf("other actors contains the change actor")
	}

	metadata, err := parseColumnMetadata(r, false)
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode operation metadata: %w", err)
	}

	columns, err := readColumns(r, metadata)
	if err != nil {
		return Change{}, nil, nil, fmt.Errorf("cannot decode operation columns: %w", err)
	}

	actors := append([]ActorID{actor}, otherActors...)

	operations, unknown, err := decodeOperations(
		columns,
		actors,
		true,
		&implicitOperationIDs{actorIndex: 0, startOp: startOp},
	)
	if err != nil {
		return Change{}, nil, nil, err
	}

	maxOp := startOp - 1
	if len(operations) > 0 {
		if startOp > math.MaxUint64-uint64(len(operations))+1 {
			return Change{}, nil, nil, fmt.Errorf("operation range overflows uint64")
		}

		maxOp = startOp + uint64(len(operations)) - 1
	}

	for i, operation := range operations {
		expected := startOp + uint64(i)
		if operation.ID.Actor != actor || operation.ID.Counter != expected {
			return Change{}, nil, nil, fmt.Errorf(
				"operation %d has ID %s@%d, expected %s@%d",
				i,
				operation.ID.Actor,
				operation.ID.Counter,
				actor,
				expected,
			)
		}
	}

	change := Change{
		Hash:         new(hash),
		Actor:        actor,
		Sequence:     sequence,
		StartOp:      startOp,
		MaxOp:        maxOp,
		Time:         timestamp,
		Message:      string(messageBytes),
		Dependencies: dependencies,
		Operations:   operations,
	}
	if r.remaining() > 0 {
		change.ExtraBytes = append([]byte(nil), r.data[r.offset:]...)
	}

	return change, actors, unknown, nil
}

func decodeOperations(
	columns map[uint32]column,
	actors []ActorID,
	changeChunk bool,
	implicitIDs *implicitOperationIDs,
) ([]Operation, []RawColumn, error) {
	if len(columns) == 0 {
		return nil, nil, nil
	}

	actionData, err := requireColumn(columns, 66)
	if err != nil {
		return nil, nil, err
	}

	actions, err := decodeULEBColumn(actionData)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode actions: %w", err)
	}

	count := len(actions)
	if err := requireItems("action", actions, count, false); err != nil {
		return nil, nil, err
	}

	idActors := make([]optional[uint64], count)
	if idActorData := optionalColumn(columns, 33); idActorData != nil {
		idActors, err = decodeULEBColumn(idActorData)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode operation actors: %w", err)
		}

		if err := requireItems("operation actor", idActors, count, false); err != nil {
			return nil, nil, err
		}
	} else if implicitIDs != nil {
		for i := range idActors {
			idActors[i] = optional[uint64]{value: implicitIDs.actorIndex, valid: true}
		}
	} else {
		return nil, nil, fmt.Errorf("required column 33 is missing")
	}

	idCounters := make([]optional[uint64], count)
	if idCounterData := optionalColumn(columns, 35); idCounterData != nil {
		idCounters, err = decodeDeltaColumn(idCounterData)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode operation counters: %w", err)
		}

		if err := requireItems("operation counter", idCounters, count, false); err != nil {
			return nil, nil, err
		}
	} else if implicitIDs != nil {
		for i := range idCounters {
			if implicitIDs.startOp > math.MaxUint64-uint64(i) {
				return nil, nil, fmt.Errorf("implicit operation counter overflows uint64")
			}

			idCounters[i] = optional[uint64]{
				value: implicitIDs.startOp + uint64(i),
				valid: true,
			}
		}
	} else {
		return nil, nil, fmt.Errorf("required column 35 is missing")
	}

	objectActors, err := decodeOptionalULEB(columns, 1, "object actor", count)
	if err != nil {
		return nil, nil, err
	}

	objectCounters, err := decodeOptionalULEB(columns, 2, "object counter", count)
	if err != nil {
		return nil, nil, err
	}

	keyActors, err := decodeOptionalULEB(columns, 17, "key actor", count)
	if err != nil {
		return nil, nil, err
	}

	keyCounters, err := decodeOptionalDelta(columns, 19, "key counter", count)
	if err != nil {
		return nil, nil, err
	}

	keyStrings, err := decodeOptionalStrings(columns, 21, "key string", count)
	if err != nil {
		return nil, nil, err
	}

	inserts := make([]bool, count)
	if insertData := optionalColumn(columns, 52); insertData != nil {
		inserts, err = decodeBooleanColumn(insertData, count)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode insert column: %w", err)
		}
	}

	values, err := decodeOptionalScalars(columns, 86, 87, count)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode values: %w", err)
	}

	var related [][]OpID
	if changeChunk {
		related, err = decodeGroupedOpIDs(columns, actors, 112, 113, 115, count, "predecessor")
	} else {
		related, err = decodeGroupedOpIDs(columns, actors, 128, 129, 131, count, "successor")
	}

	if err != nil {
		return nil, nil, err
	}

	markExpand, err := decodeOptionalBooleans(columns, 148, count)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode mark expand: %w", err)
	}

	markNames, err := decodeOptionalStrings(columns, 165, "mark name", count)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode mark name: %w", err)
	}

	operations := make([]Operation, count)
	for i := range operations {
		id, err := opIDFromIndexes(idActors[i], idCounters[i], actors)
		if err != nil {
			return nil, nil, fmt.Errorf("operation %d ID: %w", i, err)
		}

		object, err := objectIDFromIndexes(objectActors[i], objectCounters[i], actors)
		if err != nil {
			return nil, nil, fmt.Errorf("operation %d object: %w", i, err)
		}

		key, err := keyFromColumns(
			keyActors[i],
			keyCounters[i],
			keyStrings[i],
			actors,
			inserts[i],
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"operation %d key (insert=%t, action=%d): %w",
				i,
				inserts[i],
				actions[i].value,
				err,
			)
		}

		operations[i] = Operation{
			ID:     id,
			Object: object,
			Key:    key,
			Insert: inserts[i],
			Action: Action(actions[i].value),
		}
		if values[i].valid {
			value := values[i].value
			operations[i].Value = &value
		}

		if changeChunk {
			operations[i].Predecessors = related[i]
		} else {
			operations[i].Successors = related[i]
			if operations[i].Action == ActionDelete {
				return nil, nil, fmt.Errorf("document operation %d explicitly encodes a delete", i)
			}
		}

		if markExpand[i].valid {
			value := markExpand[i].value
			operations[i].MarkExpand = &value
		}

		if markNames[i].valid {
			value := markNames[i].value
			operations[i].MarkName = &value
		}
	}

	return operations, collectUnknown(columns), nil
}
