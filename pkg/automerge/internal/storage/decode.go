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
	"maps"
	"math"
	"slices"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

var magic = [4]byte{0x85, 0x6f, 0x4a, 0x83}

type decodedChunk struct {
	kind    opset.ChunkType
	content []byte
	hash    opset.ChangeHash
	raw     []byte
}

type implicitOperationIDs struct {
	actorIndex uint64
	startOp    uint64
}

// Decode parses all chunks in data and validates the resulting dependency
// graph. It accepts document, change, and compressed change chunks.
func Decode(data []byte) (*opset.Document, error) {
	return decode(data, true, true)
}

// DecodeRescue performs strict structural and history validation while allowing
// the explicit rescue path to bypass mark ordering.
func DecodeRescue(data []byte) (*opset.Document, error) {
	return decode(data, true, false)
}

// DecodePartial parses chunks whose causal dependencies may already exist in
// another document. Column and operation ownership validation still happens
// while whole-history frontier validation is deferred to the caller.
func DecodePartial(data []byte) (*opset.Document, error) {
	return decode(data, false, false)
}

// DecodeIncremental parses the complete chunk prefix and ignores an incomplete
// or corrupt trailing fragment after at least one valid chunk.
func DecodeIncremental(data []byte) (*opset.Document, int, error) {
	r := &reader{data: data}
	budget := &decodeBudget{}
	consumed := 0

	for r.remaining() > 0 {
		start := r.offset

		chunk, err := decodeChunk(r, budget, false)
		if err != nil {
			r.offset = start
			break
		}

		if chunk.kind == opset.ChunkCompressedChange {
			releaseDecodedBytes(budget, uint64(cap(chunk.content)))
			chunk.content = nil
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

func decode(
	data []byte,
	validateHistory bool,
	validateMarks bool,
) (*opset.Document, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("automerge file is empty")
	}

	owned := append([]byte(nil), data...)
	r := &reader{data: owned}
	document := &opset.Document{OwnedData: owned}
	budget := &decodeBudget{}

	for r.remaining() > 0 {
		chunk, err := decodeChunk(r, budget, false)
		if err != nil {
			return nil, fmt.Errorf("cannot decode chunk %d: %w", len(document.ChunkTypes), err)
		}

		if err := chargeDecodedSliceGrowth[opset.ChunkType](budget, 1); err != nil {
			return nil, err
		}

		document.ChunkTypes = append(document.ChunkTypes, chunk.kind)

		switch chunk.kind {
		case opset.ChunkDocument:
			if len(document.ChunkTypes) != 1 {
				return nil, fmt.Errorf("only the first chunk may be a document chunk")
			}

			if err := decodeDocumentChunk(
				document,
				chunk.content,
				budget,
				validateMarks,
			); err != nil {
				return nil, fmt.Errorf("cannot decode document chunk: %w", err)
			}
		case opset.ChunkChange, opset.ChunkCompressedChange:
			// A trailing change extends the document beyond the snapshot roots.
			// The combined stream must use the explicit replay/rebuild path.
			document.Canonical = nil

			change, actors, unknown, err := decodeChangeChunk(
				chunk.content,
				chunk.hash,
				budget,
			)
			if err != nil {
				return nil, fmt.Errorf("cannot decode change chunk: %w", err)
			}

			change.Raw = chunk.raw

			if err := chargeDecodedSliceGrowth[opset.Change](budget, 1); err != nil {
				return nil, err
			}

			document.Changes = append(document.Changes, change)

			document.Actors, err = mergeActors(document.Actors, actors, budget)
			if err != nil {
				return nil, fmt.Errorf("cannot merge actor table: %w", err)
			}

			if err := chargeDecodedSliceGrowth[opset.RawColumn](
				budget,
				uint64(len(unknown)),
			); err != nil {
				return nil, err
			}

			document.UnknownColumns = append(document.UnknownColumns, unknown...)
		default:
			return nil, fmt.Errorf("unsupported chunk type %d", chunk.kind)
		}
	}

	if validateHistory {
		if err := validateDocument(document, budget); err != nil {
			return nil, fmt.Errorf("invalid dependency graph: %w", err)
		}
	}

	return document, nil
}

func decodeChunk(
	r *reader,
	budget *decodeBudget,
	ownContent bool,
) (decodedChunk, error) {
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

	kind := opset.ChunkType(rawType)
	if kind > opset.ChunkCompressedChange {
		return decodedChunk{}, fmt.Errorf("unknown chunk type %d", kind)
	}

	lengthStart := r.offset

	length, err := r.uleb()
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read chunk length: %w", err)
	}

	lengthBytes := r.data[lengthStart:r.offset]

	content, err := r.bytes(length)
	if err != nil {
		return decodedChunk{}, fmt.Errorf("cannot read chunk content: %w", err)
	}

	hashKind := kind
	hashContent := content

	if kind == opset.ChunkCompressedChange {
		hashKind = opset.ChunkChange

		hashContent, err = inflate(content, budget)
		if err != nil {
			return decodedChunk{}, fmt.Errorf("cannot inflate compressed change: %w", err)
		}

		lengthBytes = appendULEB(nil, uint64(len(hashContent)))
	}

	hasher := sha256.New()
	_, _ = hasher.Write([]byte{byte(hashKind)})
	_, _ = hasher.Write(lengthBytes)
	_, _ = hasher.Write(hashContent)

	digest := copyHash(hasher.Sum(nil))
	if !bytes.Equal(checksum, digest[:4]) {
		return decodedChunk{}, fmt.Errorf(
			"checksum mismatch: encoded %x, calculated %x",
			checksum,
			digest[:4],
		)
	}

	if kind != opset.ChunkCompressedChange && ownContent {
		if err := chargeDecodedBytes(budget, uint64(len(hashContent))); err != nil {
			return decodedChunk{}, fmt.Errorf("cannot retain chunk content: %w", err)
		}

		owned := make([]byte, len(hashContent))
		copy(owned, hashContent)
		hashContent = owned
	}

	chunk := decodedChunk{
		kind:    kind,
		content: hashContent,
		raw:     r.data[start:r.offset],
	}
	if kind == opset.ChunkChange || kind == opset.ChunkCompressedChange {
		chunk.hash = opset.ChangeHash(digest)
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

func decodeDocumentChunk(
	document *opset.Document,
	data []byte,
	budget *decodeBudget,
	validateMarks bool,
) error {
	r := &reader{data: data}

	actors, err := decodeActorArray(r, true, budget)
	if err != nil {
		return fmt.Errorf("cannot decode actors: %w", err)
	}

	heads, err := decodeHashArray(r, true, budget)
	if err != nil {
		return fmt.Errorf("cannot decode heads: %w", err)
	}

	changeMetadata, err := parseColumnMetadata(r, true, budget)
	if err != nil {
		return fmt.Errorf("cannot decode change column metadata: %w", err)
	}

	operationMetadata, err := parseColumnMetadata(r, true, budget)
	if err != nil {
		return fmt.Errorf("cannot decode operation column metadata: %w", err)
	}

	changeColumns, err := readColumns(r, changeMetadata, budget)
	if err != nil {
		return fmt.Errorf("cannot decode change columns: %w", err)
	}

	operationColumns, err := readColumns(r, operationMetadata, budget)
	if err != nil {
		return fmt.Errorf("cannot decode operation columns: %w", err)
	}

	canonicalChangeColumns := maps.Clone(changeColumns)
	canonicalOperationColumns := maps.Clone(operationColumns)

	changes, unknownChanges, err := decodeDocumentChanges(changeColumns, actors, budget)
	if err != nil {
		return fmt.Errorf("cannot decode changes: %w", err)
	}

	operations, unknownOperations, err := decodeOperations(
		operationColumns,
		actors,
		false,
		nil,
		budget,
	)
	if err != nil {
		return fmt.Errorf("cannot decode operations: %w", err)
	}

	operationOrder := make([]opset.OpID, len(operations))
	for i := range operations {
		operationOrder[i] = operations[i].ID
	}

	canonicalOperations := operations

	operations, err = restoreChangeOperations(operations, budget)
	if err != nil {
		return fmt.Errorf("cannot restore change operations: %w", err)
	}

	if validateMarks {
		if err := validateSnapshotMarkOrder(operations); err != nil {
			return err
		}
	}

	if err := assignOperations(changes, operations, budget); err != nil {
		return fmt.Errorf("cannot assign operations: %w", err)
	}

	if err := chargeDecoded[uint64](budget, uint64(len(heads))); err != nil {
		return err
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

		if changes[index].Hash != nil && *changes[index].Hash != heads[i] {
			return fmt.Errorf("head index %d is assigned conflicting hashes", index)
		}

		changes[index].Hash = &heads[i]
	}

	if r.remaining() != 0 {
		return fmt.Errorf("document chunk has %d trailing bytes", r.remaining())
	}

	if err := validateSnapshotGraph(changes, headIndexes, budget); err != nil {
		return err
	}

	if err := reconstructSnapshotChanges(changes, budget); err != nil {
		return err
	}

	snapshot, err := newDecodedSnapshotColumns(
		actors,
		heads,
		headIndexes,
		canonicalChangeColumns,
		canonicalOperationColumns,
		len(changes),
		len(canonicalOperations),
	)
	if err != nil {
		return err
	}

	document.Actors = actors
	document.Heads = heads
	document.Changes = changes
	document.OperationOrder = operationOrder
	document.ChangeColumns = snapshotRawColumns(
		canonicalChangeColumns,
		changeColumnSpecifications,
	)
	document.OperationColumns = snapshotRawColumns(
		canonicalOperationColumns,
		operationColumnSpecifications,
	)
	document.Canonical = &DecodedDocument{
		Snapshot:   snapshot,
		Operations: canonicalOperations,
	}

	unknownCount := len(unknownChanges) + len(unknownOperations)
	if err := chargeDecoded[opset.RawColumn](budget, uint64(unknownCount)); err != nil {
		return err
	}

	document.UnknownColumns = make([]opset.RawColumn, 0, unknownCount)
	document.UnknownColumns = append(document.UnknownColumns, unknownChanges...)
	document.UnknownColumns = append(document.UnknownColumns, unknownOperations...)

	return nil
}

func snapshotRawColumns(
	columns map[uint32]column,
	specifications []uint32,
) []opset.RawColumn {
	raw := make([]opset.RawColumn, 0, len(specifications))
	for _, specification := range specifications {
		column, ok := columns[specification]
		if !ok {
			continue
		}

		raw = append(raw, opset.RawColumn{
			Specification: specification,
			Data:          column.data,
		})
	}

	return raw
}

// restoreChangeOperations turns a snapshot's operation view back into the one a
// change carries. A snapshot records, for every surviving operation, the
// operations that superseded it, and it drops delete operations entirely because
// they survive only as those successor entries. A change instead names each
// operation's predecessors and spells its deletes out, so rebuilding a change
// means inverting the successor lists and recreating the deletes they imply.
//
// Successors are left in place: the engine materializes a loaded snapshot from
// them, and re-encoding a change only ever reads predecessors.
func restoreChangeOperations(
	operations []opset.Operation,
	budget *decodeBudget,
) ([]opset.Operation, error) {
	if err := chargeDecodedMap[opset.OpID, struct{}](
		budget,
		uint64(len(operations)),
	); err != nil {
		return nil, err
	}

	stored := make(map[opset.OpID]struct{}, len(operations))
	for _, operation := range operations {
		stored[operation.ID] = struct{}{}
	}

	var successorCount uint64
	for _, operation := range operations {
		if successorCount > math.MaxUint64-uint64(len(operation.Successors)) {
			return nil, fmt.Errorf("snapshot successor count overflows uint64")
		}

		successorCount += uint64(len(operation.Successors))
	}

	if successorCount > maxDecodedItems {
		return nil, fmt.Errorf("snapshot successors exceed %d items", maxDecodedItems)
	}

	if err := chargeDecodedMap[opset.OpID, []opset.OpID](budget, successorCount); err != nil {
		return nil, err
	}

	if err := chargeDecodedSliceGrowth[opset.OpID](budget, successorCount); err != nil {
		return nil, err
	}

	predecessors := make(map[opset.OpID][]opset.OpID)

	for _, operation := range operations {
		for _, successor := range operation.Successors {
			predecessors[successor] = append(predecessors[successor], operation.ID)
		}
	}

	for identifier := range predecessors {
		slices.SortFunc(
			predecessors[identifier],
			func(left, right opset.OpID) int {
				return left.Compare(right)
			},
		)
	}

	for i := range operations {
		operations[i].Predecessors = predecessors[operations[i].ID]
	}

	deleteCount := 0

	for identifier := range predecessors {
		if _, ok := stored[identifier]; ok {
			continue
		}

		deleteCount++
	}

	if deleteCount == 0 {
		return operations, nil
	}

	if err := chargeDecoded[opset.Operation](
		budget,
		uint64(len(operations)+deleteCount),
	); err != nil {
		return nil, err
	}

	if err := chargeDecoded[opset.OpID](budget, uint64(deleteCount)); err != nil {
		return nil, err
	}

	result := make([]opset.Operation, len(operations), len(operations)+deleteCount)
	copy(result, operations)

	for identifier, superseded := range predecessors {
		if _, ok := stored[identifier]; ok {
			continue
		}

		// Only a delete leaves no operation of its own behind, and every operation
		// it removed shares the object and key it targeted.
		source := operationByID(operations, superseded[0])
		if source == nil {
			return nil, fmt.Errorf(
				"operation %s@%d supersedes nothing that the snapshot retains",
				identifier.Actor,
				identifier.Counter,
			)
		}

		result = append(
			result,
			opset.Operation{
				ID:           identifier,
				Object:       source.Object,
				Key:          supersededKey(source),
				Action:       opset.ActionDelete,
				Predecessors: superseded,
			},
		)
	}

	slices.SortFunc(
		result[len(operations):],
		func(left, right opset.Operation) int {
			return left.ID.Compare(right.ID)
		},
	)

	return result, nil
}

// supersededKey names the location an operation occupies, which is what an
// operation superseding it addresses. A map operation is addressed by its
// property, while a sequence operation is addressed by the element identifier:
// an insertion creates the element it is named by, and any later operation on
// that element already carries it.
func supersededKey(operation *opset.Operation) opset.Key {
	if operation.Key.Property != nil || !operation.Insert {
		return operation.Key
	}

	element := operation.ID

	return opset.Key{Element: &element}
}

func operationByID(operations []opset.Operation, identifier opset.OpID) *opset.Operation {
	for i := range operations {
		if operations[i].ID == identifier {
			return &operations[i]
		}
	}

	return nil
}

// reconstructSnapshotChanges restores the change-chunk identity of every change
// in a document chunk. Snapshots store the frontier hashes only and reference
// ancestry by column index, so every non-head change decodes without a hash and
// without its original bytes. Re-encoding each change once its dependencies are
// known recovers both, which keeps the change graph whole: without it only the
// heads are addressable and any walk of their ancestry hits a missing change.
//
// Dependencies are rebuilt in the stored index order because the encoder writes
// dependency hashes in slice order, so that order is what the original hash was
// computed over.
func reconstructSnapshotChanges(
	changes []opset.Change,
	budget *decodeBudget,
) error {
	if err := chargeDecoded[bool](budget, uint64(len(changes))); err != nil {
		return err
	}

	var dependencyCount uint64
	for i := range changes {
		if dependencyCount > math.MaxUint64-uint64(len(changes[i].DependencyIndexes)) {
			return fmt.Errorf("snapshot dependency count overflows uint64")
		}

		dependencyCount += uint64(len(changes[i].DependencyIndexes))
	}

	if err := chargeDecoded[opset.ChangeHash](budget, dependencyCount); err != nil {
		return err
	}

	resolved := make([]bool, len(changes))
	remaining := len(changes)

	for remaining > 0 {
		progressed := false

		for i := range changes {
			change := &changes[i]

			if resolved[i] || !dependenciesResolved(change, resolved) {
				continue
			}

			change.Dependencies = make([]opset.ChangeHash, 0, len(change.DependencyIndexes))
			for _, index := range change.DependencyIndexes {
				change.Dependencies = append(change.Dependencies, *changes[index].Hash)
			}

			recorded := change.Hash
			change.Hash = nil

			if err := reserveChangeReconstruction(change, budget); err != nil {
				return fmt.Errorf("cannot reserve snapshot change %d: %w", i, err)
			}

			if _, err := EncodeChange(change); err != nil {
				return fmt.Errorf("cannot rebuild snapshot change %d: %w", i, err)
			}

			// A recorded hash only exists for frontier changes. Disagreeing with it
			// means the rebuilt bytes are not the ones the writer hashed, so the
			// whole graph would be keyed by identifiers no peer shares.
			if recorded != nil && *recorded != *change.Hash {
				return fmt.Errorf(
					"snapshot change %d rebuilds to hash %s but the document records %s",
					i,
					change.Hash,
					recorded,
				)
			}

			resolved[i] = true
			remaining--
			progressed = true
		}

		if !progressed {
			return fmt.Errorf("snapshot dependency graph cannot be ordered")
		}
	}

	return nil
}

func reserveChangeReconstruction(change *opset.Change, budget *decodeBudget) error {
	if err := chargeDecoded[[1024]byte](
		budget,
		uint64(len(change.Operations)+1),
	); err != nil {
		return err
	}

	if err := chargeDecoded[[128]byte](
		budget,
		uint64(len(change.Dependencies)),
	); err != nil {
		return err
	}

	if err := chargeDecodedPayload(budget, len(change.Actor), 2); err != nil {
		return err
	}

	if err := chargeDecodedPayload(budget, len(change.Message), 4); err != nil {
		return err
	}

	if err := chargeDecodedPayload(budget, len(change.ExtraBytes), 4); err != nil {
		return err
	}

	for i := range change.Operations {
		operation := &change.Operations[i]
		if err := chargeDecodedPayload(budget, len(operation.ID.Actor), 2); err != nil {
			return err
		}

		if !operation.Object.IsRoot {
			if err := chargeDecodedPayload(
				budget,
				len(operation.Object.OpID.Actor),
				2,
			); err != nil {
				return err
			}
		}

		if operation.Key.Property != nil {
			if err := chargeDecodedPayload(budget, len(*operation.Key.Property), 4); err != nil {
				return err
			}
		}

		if operation.Key.Element != nil {
			if err := chargeDecodedPayload(
				budget,
				len(operation.Key.Element.Actor),
				2,
			); err != nil {
				return err
			}
		}

		for _, predecessor := range operation.Predecessors {
			if err := chargeDecoded[[128]byte](budget, 1); err != nil {
				return err
			}

			if err := chargeDecodedPayload(budget, len(predecessor.Actor), 2); err != nil {
				return err
			}
		}

		if operation.Value != nil {
			var length int

			switch operation.Value.Type {
			case opset.ScalarString:
				length = len(operation.Value.String)
			case opset.ScalarBytes:
				length = len(operation.Value.Bytes)
			default:
				length = len(operation.Value.Raw)
			}

			if err := chargeDecodedPayload(budget, length, 4); err != nil {
				return err
			}
		}

		if operation.MarkName != nil {
			if err := chargeDecodedPayload(budget, len(*operation.MarkName), 4); err != nil {
				return err
			}
		}
	}

	return nil
}

func chargeDecodedPayload(budget *decodeBudget, length int, copies uint64) error {
	if uint64(length) > math.MaxUint64/copies {
		return fmt.Errorf("decoded payload size overflows uint64")
	}

	return chargeDecodedBytes(budget, uint64(length)*copies)
}

func dependenciesResolved(change *opset.Change, resolved []bool) bool {
	for _, index := range change.DependencyIndexes {
		if !resolved[index] {
			return false
		}
	}

	return true
}

func decodeDocumentChanges(
	columns map[uint32]column,
	actors []opset.ActorID,
	budget *decodeBudget,
) ([]opset.Change, []opset.RawColumn, error) {
	if len(columns) == 0 {
		return nil, nil, nil
	}

	actorData, err := requireColumn(columns, 1)
	if err != nil {
		return nil, nil, err
	}

	actorIndexes, err := decodeULEBColumnWithBudget(actorData, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode actor column: %w", err)
	}

	count := len(actorIndexes)
	if err := requireItems("actor", actorIndexes, count, false); err != nil {
		return nil, nil, err
	}

	sequence, err := decodeRequiredDelta(columns, 3, "sequence", count, budget)
	if err != nil {
		return nil, nil, err
	}

	maxOps, err := decodeRequiredDelta(columns, 19, "maxOp", count, budget)
	if err != nil {
		return nil, nil, err
	}

	for i := range count {
		if sequence[i].value > math.MaxUint32 {
			return nil, nil, fmt.Errorf(
				"change %d sequence %d exceeds snapshot uint32 domain",
				i,
				sequence[i].value,
			)
		}

		if maxOps[i].value > math.MaxUint32 {
			return nil, nil, fmt.Errorf(
				"change %d maxOp %d exceeds snapshot uint32 domain",
				i,
				maxOps[i].value,
			)
		}
	}

	times, err := decodeOptionalSignedDelta(columns, 35, "time", count, budget)
	if err != nil {
		return nil, nil, err
	}

	messages, err := decodeOptionalStrings(columns, 53, "message", count, budget)
	if err != nil {
		return nil, nil, err
	}

	groupData, err := requireColumn(columns, 64)
	if err != nil {
		return nil, nil, err
	}

	groups, err := decodeULEBColumnWithBudget(groupData, budget)
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

	var dependencies []optional[uint64]

	if dependencyCount > 0 {
		dependencyData, err := requireColumn(columns, 67)
		if err != nil {
			return nil, nil, err
		}

		dependencies, err = decodeDeltaColumnWithBudget(dependencyData, budget)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode dependencies: %w", err)
		}

		if err := requireItems("dependency", dependencies, dependencyCount, false); err != nil {
			return nil, nil, err
		}
	}

	extras, extraBytes, err := decodeSnapshotExtras(columns, count, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode change extras: %w", err)
	}

	if err := chargeDecoded[opset.Change](budget, uint64(count)); err != nil {
		return nil, nil, err
	}

	if err := chargeDecoded[uint64](budget, uint64(dependencyCount)); err != nil {
		return nil, nil, err
	}

	changes := make([]opset.Change, count)
	dependencyOffset := 0

	for i := range changes {
		actorIndex := actorIndexes[i].value
		if actorIndex >= uint64(len(actors)) {
			return nil, nil, fmt.Errorf("change %d actor index %d is out of bounds", i, actorIndex)
		}

		changes[i] = opset.Change{
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
			changes[i].Extra = &extras[i].value
			changes[i].ExtraBytes = extraBytes[i]
		}

		groupLength := int(groups[i].value)

		changes[i].DependencyIndexes = make([]uint64, groupLength)
		for j := range groupLength {
			changes[i].DependencyIndexes[j] = dependencies[dependencyOffset+j].value
		}

		dependencyOffset += groupLength
	}

	unknown, err := collectUnknown(columns, budget)
	if err != nil {
		return nil, nil, err
	}

	return changes, unknown, nil
}

func decodeChangeChunk(
	data []byte,
	hash opset.ChangeHash,
	budget *decodeBudget,
) (opset.Change, []opset.ActorID, []opset.RawColumn, error) {
	r := &reader{data: data}

	dependencies, err := decodeHashArray(r, false, budget)
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode dependencies: %w", err)
	}

	actorBytes, err := decodeLengthPrefixed(r)
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode actor: %w", err)
	}

	if err := chargeDecodedBytes(budget, uint64(len(actorBytes))); err != nil {
		return opset.Change{}, nil, nil, err
	}

	actor, err := opset.NewActorID(actorBytes)
	if err != nil {
		return opset.Change{}, nil, nil, err
	}

	sequence, err := r.uleb()
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode sequence: %w", err)
	}

	if sequence == 0 {
		return opset.Change{}, nil, nil, fmt.Errorf("sequence is zero")
	}

	startOp, err := r.uleb()
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode start op: %w", err)
	}

	if startOp == 0 {
		return opset.Change{}, nil, nil, fmt.Errorf("start op is zero")
	}

	timestamp, err := r.leb()
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode time: %w", err)
	}

	messageBytes, err := decodeLengthPrefixed(r)
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode message: %w", err)
	}

	if !utf8.Valid(messageBytes) {
		return opset.Change{}, nil, nil, fmt.Errorf("message is not valid UTF-8")
	}

	otherActors, err := decodeActorArray(r, true, budget)
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode other actors: %w", err)
	}

	if slices.Contains(otherActors, actor) {
		return opset.Change{}, nil, nil, fmt.Errorf("other actors contains the change actor")
	}

	metadata, err := parseColumnMetadata(r, false, budget)
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode operation metadata: %w", err)
	}

	columns, err := readColumns(r, metadata, budget)
	if err != nil {
		return opset.Change{}, nil, nil, fmt.Errorf("cannot decode operation columns: %w", err)
	}

	if err := chargeDecoded[opset.ActorID](budget, uint64(len(otherActors)+1)); err != nil {
		return opset.Change{}, nil, nil, err
	}

	actors := make([]opset.ActorID, 1, len(otherActors)+1)
	actors[0] = actor
	actors = append(actors, otherActors...)

	operations, unknown, err := decodeOperations(
		columns,
		actors,
		true,
		&implicitOperationIDs{actorIndex: 0, startOp: startOp},
		budget,
	)
	if err != nil {
		return opset.Change{}, nil, nil, err
	}

	maxOp := startOp - 1
	if len(operations) > 0 {
		if startOp > math.MaxUint64-uint64(len(operations))+1 {
			return opset.Change{}, nil, nil, fmt.Errorf("operation range overflows uint64")
		}

		maxOp = startOp + uint64(len(operations)) - 1
	}

	for i, operation := range operations {
		expected := startOp + uint64(i)
		if operation.ID.Actor != actor || operation.ID.Counter != expected {
			return opset.Change{}, nil, nil, fmt.Errorf(
				"operation %d has ID %s@%d, expected %s@%d",
				i,
				operation.ID.Actor,
				operation.ID.Counter,
				actor,
				expected,
			)
		}
	}

	if err := chargeDecodedBytes(budget, uint64(len(messageBytes))); err != nil {
		return opset.Change{}, nil, nil, err
	}

	if err := chargeDecoded[opset.ChangeHash](budget, 1); err != nil {
		return opset.Change{}, nil, nil, err
	}

	change := opset.Change{
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
		change.ExtraBytes = r.data[r.offset:]
	}

	return change, actors, unknown, nil
}

func decodeOperations(
	columns map[uint32]column,
	actors []opset.ActorID,
	changeChunk bool,
	implicitIDs *implicitOperationIDs,
	budget *decodeBudget,
) ([]opset.Operation, []opset.RawColumn, error) {
	if len(columns) == 0 {
		return nil, nil, nil
	}

	actionData, err := requireColumn(columns, 66)
	if err != nil {
		return nil, nil, err
	}

	actions, err := decodeULEBColumnWithBudget(actionData, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode actions: %w", err)
	}

	count := len(actions)
	if err := requireItems("action", actions, count, false); err != nil {
		return nil, nil, err
	}

	var idActors []optional[uint64]
	if idActorData := optionalColumn(columns, 33); idActorData != nil {
		idActors, err = decodeULEBColumnWithBudget(idActorData, budget)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode operation actors: %w", err)
		}

		if err := requireItems("operation actor", idActors, count, false); err != nil {
			return nil, nil, err
		}
	} else if implicitIDs != nil {
		if err := chargeDecoded[optional[uint64]](budget, uint64(count)); err != nil {
			return nil, nil, err
		}

		idActors = make([]optional[uint64], count)
		for i := range idActors {
			idActors[i] = optional[uint64]{value: implicitIDs.actorIndex, valid: true}
		}
	} else {
		return nil, nil, fmt.Errorf("required column 33 is missing")
	}

	var idCounters []optional[uint64]
	if idCounterData := optionalColumn(columns, 35); idCounterData != nil {
		idCounters, err = decodeDeltaColumnWithBudget(idCounterData, budget)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode operation counters: %w", err)
		}

		if err := requireItems("operation counter", idCounters, count, false); err != nil {
			return nil, nil, err
		}
	} else if implicitIDs != nil {
		if err := chargeDecoded[optional[uint64]](budget, uint64(count)); err != nil {
			return nil, nil, err
		}

		idCounters = make([]optional[uint64], count)
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

	objectActors, err := decodeOptionalULEB(columns, 1, "object actor", count, budget)
	if err != nil {
		return nil, nil, err
	}

	objectCounters, err := decodeOptionalULEB(columns, 2, "object counter", count, budget)
	if err != nil {
		return nil, nil, err
	}

	keyActors, err := decodeOptionalULEB(columns, 17, "key actor", count, budget)
	if err != nil {
		return nil, nil, err
	}

	keyCounters, err := decodeOptionalDelta(columns, 19, "key counter", count, budget)
	if err != nil {
		return nil, nil, err
	}

	keyStrings, err := decodeOptionalStrings(columns, 21, "key string", count, budget)
	if err != nil {
		return nil, nil, err
	}

	var inserts []bool
	if insertData := optionalColumn(columns, 52); insertData != nil {
		inserts, err = decodeBooleanColumnWithBudget(insertData, count, budget)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot decode insert column: %w", err)
		}
	} else {
		if err := chargeDecoded[bool](budget, uint64(count)); err != nil {
			return nil, nil, err
		}

		inserts = make([]bool, count)
	}

	values, err := decodeOptionalScalars(columns, 86, 87, count, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode values: %w", err)
	}

	var related [][]opset.OpID
	if changeChunk {
		related, err = decodeGroupedOpIDs(
			columns,
			actors,
			112,
			113,
			115,
			count,
			"predecessor",
			budget,
		)
	} else {
		related, err = decodeGroupedOpIDs(
			columns,
			actors,
			128,
			129,
			131,
			count,
			"successor",
			budget,
		)
	}

	if err != nil {
		return nil, nil, err
	}

	markExpand, err := decodeOptionalBooleans(columns, 148, count, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode mark expand: %w", err)
	}

	markNames, err := decodeOptionalStrings(columns, 165, "mark name", count, budget)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decode mark name: %w", err)
	}

	if err := chargeDecoded[opset.Operation](budget, uint64(count)); err != nil {
		return nil, nil, err
	}

	type keyPayload struct {
		property string
		element  opset.OpID
	}

	if err := chargeDecoded[keyPayload](budget, uint64(count)); err != nil {
		return nil, nil, err
	}

	operations := make([]opset.Operation, count)
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

		operations[i] = opset.Operation{
			ID:     id,
			Object: object,
			Key:    key,
			Insert: inserts[i],
			Action: opset.Action(actions[i].value),
		}
		if values[i].valid {
			operations[i].Value = &values[i].value
		}

		if changeChunk {
			operations[i].Predecessors = related[i]
		} else {
			operations[i].Successors = related[i]
			if operations[i].Action == opset.ActionDelete {
				return nil, nil, fmt.Errorf("document operation %d explicitly encodes a delete", i)
			}
		}

		// Expand is a property of a mark, and its column is dense because booleans
		// cannot be null. A document chunk shares one column across every change,
		// so keeping the flag on ordinary operations would make a change that never
		// carried an expand column re-encode with one and hash differently.
		if markExpand[i].valid && operations[i].Action == opset.ActionMark {
			operations[i].MarkExpand = &markExpand[i].value
		}

		if markNames[i].valid {
			operations[i].MarkName = &markNames[i].value
		}

		if !changeChunk {
			if err := validateSnapshotOperationCounters(i, operations[i]); err != nil {
				return nil, nil, err
			}
		}
	}

	unknown, err := collectUnknown(columns, budget)
	if err != nil {
		return nil, nil, err
	}

	return operations, unknown, nil
}

func validateSnapshotOperationCounters(index int, operation opset.Operation) error {
	check := func(name string, counter uint64) error {
		if counter > math.MaxUint32 {
			return fmt.Errorf(
				"document operation %d %s counter %d exceeds uint32",
				index,
				name,
				counter,
			)
		}

		return nil
	}

	if err := check("ID", operation.ID.Counter); err != nil {
		return err
	}

	if !operation.Object.IsRoot {
		if err := check("object", operation.Object.OpID.Counter); err != nil {
			return err
		}
	}

	if operation.Key.Element != nil {
		if err := check("key", operation.Key.Element.Counter); err != nil {
			return err
		}
	}

	for _, successor := range operation.Successors {
		if err := check("successor", successor.Counter); err != nil {
			return err
		}
	}

	return nil
}
