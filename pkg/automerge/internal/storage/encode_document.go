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

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

// assembleChunk frames a chunk body: the magic bytes, the first four bytes of
// the digest covering the typed and length-prefixed body, then that body.
func assembleChunk(kind opset.ChunkType, body []byte) []byte {
	length := appendULEB(nil, uint64(len(body)))

	hasher := sha256.New()
	_, _ = hasher.Write([]byte{byte(kind)})
	_, _ = hasher.Write(length)
	_, _ = hasher.Write(body)
	digest := hasher.Sum(nil)

	chunk := make([]byte, 0, 4+4+1+len(length)+len(body))
	chunk = append(chunk, magic[:]...)
	chunk = append(chunk, digest[:4]...)
	chunk = append(chunk, byte(kind))
	chunk = append(chunk, length...)
	chunk = append(chunk, body...)

	return chunk
}

// EncodeDocument serializes a whole history as one compacted document chunk,
// the form Rust and JavaScript write from save().
//
// A document chunk is not a concatenation of changes. It stores the operation
// set once, in operation-set order rather than per change, and reduces each
// change to a row of metadata whose ancestry is expressed as indexes into that
// same table. Deletes disappear into the successor lists of the operations they
// removed, and every operation's predecessors are recovered from those lists on
// the way back in.
//
// order gives the operation-set sequence: object by object, and within an
// object the order a reader would see. The caller owns that sequence because it
// requires the sequence state only the engine maintains. Operations named by
// order must exist in the history, and deletes must be left out of it.
//
// compress DEFLATEs individual columns above a size threshold, which is what
// save() does and save_nocompress() does not. The compressed bytes are not
// byte-identical to the reference because the DEFLATE implementations differ, so
// byte identity only holds for histories small enough that no column crosses the
// threshold; compression is a size optimization, and every column round-trips
// because the decoder inflates any column whose specification carries the
// compressed bit.
func EncodeDocument(document *opset.Document, order []opset.OpID, compress bool) ([]byte, error) {
	return encodeDocument(document, order, nil, compress, true)
}

func EncodePreparedDocument(
	document *opset.Document,
	operations []opset.Operation,
	compress bool,
) ([]byte, error) {
	return encodeDocument(document, nil, operations, compress, true)
}

// EncodeTrustedPreparedDocument serializes state whose snapshot-domain
// invariants were checked when changes entered the engine. It avoids rescanning
// the complete operation set on every save.
func EncodeTrustedPreparedDocument(
	document *opset.Document,
	operations []opset.Operation,
	compress bool,
) ([]byte, error) {
	return encodeDocument(document, nil, operations, compress, false)
}

func encodeDocument(
	document *opset.Document,
	order []opset.OpID,
	preparedOperations []opset.Operation,
	compress bool,
	validate bool,
) ([]byte, error) {
	if validate {
		if err := validateSnapshotEncodeDomain(document); err != nil {
			return nil, err
		}
	}

	var changes []*opset.Change
	if preparedOperations != nil {
		changes = make([]*opset.Change, len(document.Changes))
		for i := range document.Changes {
			changes[i] = &document.Changes[i]
		}
	} else {
		var err error

		changes, err = documentChangeOrder(document)
		if err != nil {
			return nil, err
		}
	}

	operations := preparedOperations
	if operations == nil {
		var err error

		operations, err = documentOperations(changes, order)
		if err != nil {
			return nil, err
		}
	}

	actors := documentActorTable(changes, operations)
	actorIndexes := make(map[opset.ActorID]uint64, len(actors))

	for i, actor := range actors {
		actorIndexes[actor] = uint64(i)
	}

	heads, headIndexes, err := documentHeads(changes)
	if err != nil {
		return nil, err
	}

	changeColumns, err := encodeDocumentChangeColumns(changes, actorIndexes)
	if err != nil {
		return nil, err
	}

	operationColumns, err := encodeDocumentOperationColumns(operations, actorIndexes)
	if err != nil {
		return nil, err
	}

	changeColumns = compressColumns(
		sortColumns(
			append(changeColumns, retainedColumns(document, changeColumnSpecifications)...),
		),
		compress,
	)
	operationColumns = compressColumns(
		sortColumns(
			append(operationColumns, retainedColumns(document, operationColumnSpecifications)...),
		),
		compress,
	)

	var body []byte

	body = appendULEB(body, uint64(len(actors)))
	for _, actor := range actors {
		body = appendLengthPrefixedNative(body, actor.Bytes())
	}

	body = appendULEB(body, uint64(len(heads)))
	for _, head := range heads {
		body = append(body, head[:]...)
	}

	// Both column sets are described before either is written, so the metadata
	// has to be laid out ahead of the data it measures.
	body = appendColumnMetadata(body, changeColumns)
	body = appendColumnMetadata(body, operationColumns)
	body = appendColumnData(body, changeColumns)
	body = appendColumnData(body, operationColumns)

	for _, index := range headIndexes {
		body = appendULEB(body, index)
	}

	return assembleChunk(opset.ChunkDocument, body), nil
}

func validateSnapshotEncodeDomain(document *opset.Document) error {
	for i := range document.Changes {
		change := &document.Changes[i]
		if change.Hash == nil {
			return fmt.Errorf("change %d has no hash", i)
		}

		if change.Sequence == 0 {
			return fmt.Errorf("change %d sequence is zero", i)
		}

		if change.Sequence > math.MaxUint32 {
			return fmt.Errorf(
				"change %d sequence %d exceeds snapshot uint32 domain",
				i,
				change.Sequence,
			)
		}

		if change.MaxOp > math.MaxUint32 {
			return fmt.Errorf(
				"change %d maxOp %d exceeds snapshot uint32 domain",
				i,
				change.MaxOp,
			)
		}

		for j, operation := range change.Operations {
			if operation.ID.Counter == 0 {
				return fmt.Errorf("operation %d:%d counter is zero", i, j)
			}

			if operation.ID.Counter > math.MaxUint32 {
				return fmt.Errorf(
					"operation %d:%d counter %d exceeds snapshot uint32 domain",
					i,
					j,
					operation.ID.Counter,
				)
			}

			if !operation.Object.IsRoot &&
				operation.Object.OpID.Counter > math.MaxUint32 {
				return fmt.Errorf(
					"operation %d:%d object counter %d exceeds snapshot uint32 domain",
					i,
					j,
					operation.Object.OpID.Counter,
				)
			}

			if operation.Key.Element != nil &&
				operation.Key.Element.Counter > math.MaxUint32 {
				return fmt.Errorf(
					"operation %d:%d key counter %d exceeds snapshot uint32 domain",
					i,
					j,
					operation.Key.Element.Counter,
				)
			}

			for k, predecessor := range operation.Predecessors {
				if predecessor.Counter > math.MaxUint32 {
					return fmt.Errorf(
						"operation %d:%d predecessor %d counter %d exceeds snapshot uint32 domain",
						i,
						j,
						k,
						predecessor.Counter,
					)
				}
			}
		}
	}

	return nil
}

// documentChangeOrder returns the changes in dependency order. A snapshot may
// legally store them in any order, but writing ancestors first keeps the index
// references pointing backwards, which is what every other implementation emits
// and what makes the result readable in one pass.
func documentChangeOrder(document *opset.Document) ([]*opset.Change, error) {
	byHash := make(map[opset.ChangeHash]*opset.Change, len(document.Changes))

	for i := range document.Changes {
		change := &document.Changes[i]
		byHash[*change.Hash] = change
	}

	ordered := make([]*opset.Change, 0, len(document.Changes))

	const (
		visiting uint8 = 1
		visited  uint8 = 2
	)

	states := make(map[opset.ChangeHash]uint8, len(document.Changes))

	type traversalFrame struct {
		change         *opset.Change
		nextDependency int
	}

	stack := make([]traversalFrame, 0, len(document.Changes))

	for i := range document.Changes {
		root := &document.Changes[i]
		if states[*root.Hash] == visited {
			continue
		}

		states[*root.Hash] = visiting
		stack = append(stack, traversalFrame{change: root})

		for len(stack) > 0 {
			frame := &stack[len(stack)-1]
			if frame.nextDependency >= len(frame.change.Dependencies) {
				ordered = append(ordered, frame.change)
				states[*frame.change.Hash] = visited
				stack = stack[:len(stack)-1]

				continue
			}

			dependency := frame.change.Dependencies[frame.nextDependency]
			frame.nextDependency++

			parent, ok := byHash[dependency]
			if !ok {
				return nil, fmt.Errorf(
					"change %s depends on %s which the history does not hold",
					frame.change.Hash,
					dependency,
				)
			}

			switch states[dependency] {
			case visiting:
				return nil, fmt.Errorf(
					"change dependency graph contains a cycle at %s",
					parent.Hash,
				)
			case visited:
				continue
			default:
				states[dependency] = visiting

				stack = append(stack, traversalFrame{change: parent})
			}
		}
	}

	return ordered, nil
}

// documentOperations collects the operations to store, in the given order, with
// each one's successors derived from the predecessors recorded across the whole
// history. Deletes are dropped: they exist in the result only as the successor
// entries they contribute.
func documentOperations(
	changes []*opset.Change,
	order []opset.OpID,
) ([]opset.Operation, error) {
	sources := make(map[opset.OpID]*opset.Operation)

	for _, change := range changes {
		for i := range change.Operations {
			operation := &change.Operations[i]
			if _, ok := sources[operation.ID]; ok {
				return nil, fmt.Errorf(
					"operation %s@%d occurs twice",
					operation.ID.Actor,
					operation.ID.Counter,
				)
			}

			sources[operation.ID] = operation
		}
	}

	successors := make(map[opset.OpID][]opset.OpID)

	for _, change := range changes {
		for _, operation := range change.Operations {
			for _, predecessor := range operation.Predecessors {
				successors[predecessor] = append(successors[predecessor], operation.ID)
			}
		}
	}

	for identifier := range successors {
		slices.SortFunc(
			successors[identifier],
			func(left, right opset.OpID) int {
				return left.Compare(right)
			},
		)
	}

	operations := make([]opset.Operation, 0, len(order))

	for _, identifier := range order {
		source, ok := sources[identifier]
		if !ok {
			return nil, fmt.Errorf(
				"operation %s@%d is ordered but absent from the history",
				identifier.Actor,
				identifier.Counter,
			)
		}

		if source.Action == opset.ActionDelete {
			return nil, fmt.Errorf(
				"operation %s@%d is a delete and cannot be stored",
				identifier.Actor,
				identifier.Counter,
			)
		}

		stored := *source
		stored.Predecessors = nil
		stored.Successors = successors[identifier]

		operations = append(operations, stored)
	}

	return operations, nil
}

func documentActorTable(changes []*opset.Change, operations []opset.Operation) []opset.ActorID {
	seen := make(map[opset.ActorID]struct{})

	add := func(actor opset.ActorID) {
		if actor != "" {
			seen[actor] = struct{}{}
		}
	}

	for _, change := range changes {
		add(change.Actor)
	}

	for _, operation := range operations {
		add(operation.ID.Actor)

		if !operation.Object.IsRoot {
			add(operation.Object.OpID.Actor)
		}

		if operation.Key.Element != nil {
			add(operation.Key.Element.Actor)
		}

		for _, successor := range operation.Successors {
			add(successor.Actor)
		}
	}

	actors := make([]opset.ActorID, 0, len(seen))
	for actor := range seen {
		actors = append(actors, actor)
	}

	slices.SortFunc(
		actors,
		func(left, right opset.ActorID) int {
			return left.Compare(right)
		},
	)

	return actors
}

// documentHeads returns the frontier and the index of each head, which is how a
// snapshot names its heads.
func documentHeads(changes []*opset.Change) ([]opset.ChangeHash, []uint64, error) {
	indexes := make(map[opset.ChangeHash]uint64, len(changes))
	dependedOn := make(map[opset.ChangeHash]struct{}, len(changes))

	for i, change := range changes {
		indexes[*change.Hash] = uint64(i)

		for _, dependency := range change.Dependencies {
			dependedOn[dependency] = struct{}{}
		}
	}

	heads := make([]opset.ChangeHash, 0)

	for _, change := range changes {
		if _, ok := dependedOn[*change.Hash]; !ok {
			heads = append(heads, *change.Hash)
		}
	}

	slices.SortFunc(
		heads,
		func(left, right opset.ChangeHash) int {
			return bytes.Compare(left[:], right[:])
		},
	)

	headIndexes := make([]uint64, len(heads))
	for i, head := range heads {
		headIndexes[i] = indexes[head]
	}

	return heads, headIndexes, nil
}

func encodeDocumentChangeColumns(
	changes []*opset.Change,
	actorIndexes map[opset.ActorID]uint64,
) ([]encodedColumn, error) {
	count := len(changes)

	indexes := make(map[opset.ChangeHash]uint64, count)
	for i, change := range changes {
		indexes[*change.Hash] = uint64(i)
	}

	var (
		actors         = make([]optional[uint64], count)
		sequences      = make([]optional[int64], count)
		maxOps         = make([]optional[int64], count)
		times          = make([]optional[int64], count)
		messages       = make([]optional[string], count)
		dependencySize = make([]optional[uint64], count)
		dependencies   []optional[int64]
		extraMetadata  = make([]optional[uint64], count)
		extraData      []byte
	)

	for i, change := range changes {
		actorIndex, ok := actorIndexes[change.Actor]
		if !ok {
			return nil, fmt.Errorf("change %d actor is not in the actor table", i)
		}

		actors[i] = some(actorIndex)
		sequences[i] = some(int64(change.Sequence))
		maxOps[i] = some(int64(change.MaxOp))
		times[i] = some(change.Time)

		if change.Message != "" {
			messages[i] = some(change.Message)
		}

		dependencySize[i] = some(uint64(len(change.Dependencies)))

		for _, dependency := range change.Dependencies {
			index, ok := indexes[dependency]
			if !ok {
				return nil, fmt.Errorf("change %d depends on an absent change", i)
			}

			dependencies = append(dependencies, some(int64(index)))
		}

		metadata, data, err := encodeScalar(changeExtra(change))
		if err != nil {
			return nil, fmt.Errorf("cannot encode change %d extra: %w", i, err)
		}

		extraMetadata[i] = metadata

		extraData = append(extraData, data...)
	}

	columns := []encodedColumn{
		{specification: 1, data: encodeRLE(actors, appendULEB)},
		{specification: 3, data: encodeDelta(sequences)},
		{specification: 19, data: encodeDelta(maxOps)},
		{specification: 35, data: encodeDelta(times)},
		{specification: 53, data: encodeStrings(messages)},
		{specification: 64, data: encodeRLE(dependencySize, appendULEB)},
		{specification: 67, data: encodeDelta(dependencies)},
		{specification: 86, data: encodeRLE(extraMetadata, appendULEB)},
		{specification: 87, data: extraData},
	}

	return withData(columns), nil
}

// changeExtra reports the change's extra payload as the scalar a snapshot
// stores. A change chunk keeps the payload as trailing bytes, so the two forms
// have to be reconciled in whichever direction carries the value.
func changeExtra(change *opset.Change) *opset.Scalar {
	if change.Extra != nil {
		return change.Extra
	}

	if len(change.ExtraBytes) > 0 {
		return &opset.Scalar{Type: opset.ScalarBytes, Bytes: change.ExtraBytes}
	}

	// The payload is a byte string even when a change carries none, so an absent
	// one is empty rather than null.
	return &opset.Scalar{Type: opset.ScalarBytes}
}

func encodeDocumentOperationColumns(
	operations []opset.Operation,
	actorIndexes map[opset.ActorID]uint64,
) ([]encodedColumn, error) {
	count := len(operations)

	var (
		idActors         = make([]optional[uint64], count)
		idCounters       = make([]optional[int64], count)
		objectActors     = make([]optional[uint64], count)
		objectCounters   = make([]optional[uint64], count)
		keyActors        = make([]optional[uint64], count)
		keyCounters      = make([]optional[int64], count)
		keyStrings       = make([]optional[string], count)
		inserts          = make([]bool, count)
		actions          = make([]optional[uint64], count)
		valueMetadata    = make([]optional[uint64], count)
		valueData        []byte
		successorSize    = make([]optional[uint64], count)
		successorActors  []optional[uint64]
		successorCounter []optional[int64]
		markExpands      = make([]bool, count)
		hasMarkExpand    bool
		markNames        = make([]optional[string], count)
	)

	for i, operation := range operations {
		index, ok := actorIndexes[operation.ID.Actor]
		if !ok {
			return nil, fmt.Errorf("operation %d actor is not in the actor table", i)
		}

		idActors[i] = some(index)
		idCounters[i] = some(int64(operation.ID.Counter))

		if !operation.Object.IsRoot {
			index, ok := actorIndexes[operation.Object.OpID.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d object actor is unknown", i)
			}

			objectActors[i] = some(index)
			objectCounters[i] = some(operation.Object.OpID.Counter)
		}

		switch {
		case operation.Key.Property != nil:
			keyStrings[i] = some(*operation.Key.Property)
		case operation.Key.IsHead:
			keyCounters[i] = some(int64(0))
		case operation.Key.Element != nil:
			index, ok := actorIndexes[operation.Key.Element.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d key actor is unknown", i)
			}

			keyActors[i] = some(index)
			keyCounters[i] = some(int64(operation.Key.Element.Counter))
		default:
			return nil, fmt.Errorf("operation %d has no key", i)
		}

		inserts[i] = operation.Insert
		actions[i] = some(uint64(operation.Action))

		metadata, data, err := encodeScalar(operation.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot encode operation %d value: %w", i, err)
		}

		valueMetadata[i] = metadata

		valueData = append(valueData, data...)

		successorSize[i] = some(uint64(len(operation.Successors)))

		for _, successor := range operation.Successors {
			index, ok := actorIndexes[successor.Actor]
			if !ok {
				return nil, fmt.Errorf("operation %d successor actor is unknown", i)
			}

			successorActors = append(successorActors, some(index))
			successorCounter = append(successorCounter, some(int64(successor.Counter)))
		}

		// Expand only means anything on a mark, and an all-false column is left
		// out, so the flag is written only where it is actually set.
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
		{specification: 1, data: encodeRLE(objectActors, appendULEB)},
		{specification: 2, data: encodeRLE(objectCounters, appendULEB)},
		{specification: 17, data: encodeRLE(keyActors, appendULEB)},
		{specification: 19, data: encodeDelta(keyCounters)},
		{specification: 21, data: encodeStrings(keyStrings)},
		{specification: 33, data: encodeRLE(idActors, appendULEB)},
		{specification: 35, data: encodeDelta(idCounters)},
		{specification: 52, data: encodeBooleans(inserts)},
		{specification: 66, data: encodeRLE(actions, appendULEB)},
		{specification: 86, data: encodeRLE(valueMetadata, appendULEB)},
		{specification: 87, data: valueData},
		{specification: 128, data: encodeRLE(successorSize, appendULEB)},
		{specification: 129, data: encodeRLE(successorActors, appendULEB)},
		{specification: 131, data: encodeDelta(successorCounter)},
		{specification: 148, data: markExpandData},
		{specification: 165, data: encodeStrings(markNames)},
	}

	return withData(columns), nil
}

var (
	changeColumnSpecifications    = []uint32{1, 3, 19, 35, 53, 64, 67, 86, 87}
	operationColumnSpecifications = []uint32{
		1, 2, 17, 19, 21, 33, 35, 52, 66, 86, 87, 128, 129, 131, 148, 165,
	}
)

// retainedColumns returns the columns a previous reader did not understand but
// kept, so writing a history back does not quietly drop what a newer version of
// the format put there. Each retained column is matched to the table it came
// from by its specification.
func retainedColumns(document *opset.Document, known []uint32) []encodedColumn {
	retained := make([]encodedColumn, 0)
	seen := make(map[uint32]struct{})

	for _, column := range document.UnknownColumns {
		normalized := column.Specification &^ 8
		if slices.Contains(known, normalized) || len(column.Data) == 0 {
			continue
		}

		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}

		retained = append(
			retained,
			encodedColumn{
				specification: normalized,
				data:          append([]byte(nil), column.Data...),
			},
		)
	}

	return retained
}

// sortColumns puts columns in the strictly ascending order a reader requires,
// which both the metadata and the data must follow. Ordering is by the
// normalized specification so a column keeps its place whether or not it carries
// the compressed bit.
func sortColumns(columns []encodedColumn) []encodedColumn {
	slices.SortFunc(
		columns,
		func(left, right encodedColumn) int {
			leftSpec := left.specification &^ compressedColumnBit
			rightSpec := right.specification &^ compressedColumnBit

			switch {
			case leftSpec < rightSpec:
				return -1
			case leftSpec > rightSpec:
				return 1
			default:
				return 0
			}
		},
	)

	return columns
}

// compressedColumnBit marks a column whose data is DEFLATE-compressed. A reader
// strips it to recover the logical specification and inflates the data.
const compressedColumnBit = 8

// columnDeflateMinSize is the smallest column worth compressing, matching the
// change-chunk threshold. Below it, compression tends to grow the data.
const columnDeflateMinSize = 250

// compressColumns DEFLATEs each column whose data is large enough to benefit,
// marking it with the compressed bit. A column that is already compressed (a
// retained unknown column) or that does not shrink is left untouched, so the
// result is never larger than the input.
func compressColumns(columns []encodedColumn, compress bool) []encodedColumn {
	if !compress {
		return columns
	}

	for i := range columns {
		column := &columns[i]

		if column.specification&compressedColumnBit != 0 ||
			len(column.data) < columnDeflateMinSize {
			continue
		}

		deflated, err := deflate(column.data)
		if err != nil || len(deflated) >= len(column.data) {
			continue
		}

		column.specification |= compressedColumnBit
		column.data = deflated
	}

	return columns
}

func withData(columns []encodedColumn) []encodedColumn {
	filtered := make([]encodedColumn, 0, len(columns))

	for _, column := range columns {
		if len(column.data) > 0 {
			filtered = append(filtered, column)
		}
	}

	return filtered
}

func appendColumnMetadata(data []byte, columns []encodedColumn) []byte {
	data = appendULEB(data, uint64(len(columns)))
	for _, column := range columns {
		data = appendULEB(data, uint64(column.specification))
		data = appendULEB(data, uint64(len(column.data)))
	}

	return data
}

func appendColumnData(data []byte, columns []encodedColumn) []byte {
	for _, column := range columns {
		data = append(data, column.data...)
	}

	return data
}
