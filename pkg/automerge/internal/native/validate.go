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
	"fmt"
	"math"
	"slices"
)

func decodeActorArray(r *reader, sorted bool) ([]ActorID, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, err
	}

	if count > maxDecodedItems {
		return nil, fmt.Errorf("actor count %d exceeds limit", count)
	}

	actors := make([]ActorID, 0, count)
	for i := uint64(0); i < count; i++ {
		value, err := decodeLengthPrefixed(r)
		if err != nil {
			return nil, fmt.Errorf("cannot decode actor %d: %w", i, err)
		}

		actor, err := NewActorID(value)
		if err != nil {
			return nil, fmt.Errorf("actor %d: %w", i, err)
		}

		if sorted && len(actors) > 0 && actors[len(actors)-1].Compare(actor) >= 0 {
			return nil, fmt.Errorf("actor IDs are not strictly sorted at index %d", i)
		}

		actors = append(actors, actor)
	}

	return actors, nil
}

func decodeLengthPrefixed(r *reader) ([]byte, error) {
	length, err := r.uleb()
	if err != nil {
		return nil, err
	}

	return r.bytes(length)
}

func decodeHashArray(r *reader, sorted bool) ([]ChangeHash, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, err
	}

	if count > maxDecodedItems {
		return nil, fmt.Errorf("hash count %d exceeds limit", count)
	}

	hashes := make([]ChangeHash, 0, count)
	for i := uint64(0); i < count; i++ {
		value, err := r.bytes(32)
		if err != nil {
			return nil, fmt.Errorf("cannot decode hash %d: %w", i, err)
		}

		hash := copyHash(value)
		if sorted && len(hashes) > 0 && bytes.Compare(hashes[len(hashes)-1][:], hash[:]) >= 0 {
			return nil, fmt.Errorf("hashes are not strictly sorted at index %d", i)
		}

		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func decodeRequiredDelta(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
) ([]optional[uint64], error) {
	data, err := requireColumn(columns, specification)
	if err != nil {
		return nil, err
	}

	values, err := decodeDeltaColumn(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireItems(name, values, expected, false); err != nil {
		return nil, err
	}

	return values, nil
}

func decodeOptionalDelta(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
) ([]optional[uint64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return make([]optional[uint64], expected), nil
	}

	values, err := decodeDeltaColumn(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireItems(name, values, expected, true); err != nil {
		return nil, err
	}

	return values, nil
}

func decodeOptionalSignedDelta(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
) ([]optional[int64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return make([]optional[int64], expected), nil
	}

	values, err := decodeSignedDeltaColumn(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireItems(name, values, expected, true); err != nil {
		return nil, err
	}

	return values, nil
}

func decodeRequiredULEB(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
) ([]optional[uint64], error) {
	data, err := requireColumn(columns, specification)
	if err != nil {
		return nil, err
	}

	values, err := decodeULEBColumn(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireItems(name, values, expected, false); err != nil {
		return nil, err
	}

	return values, nil
}

func decodeOptionalULEB(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
) ([]optional[uint64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return make([]optional[uint64], expected), nil
	}

	values, err := decodeULEBColumn(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireItems(name, values, expected, true); err != nil {
		return nil, err
	}

	return values, nil
}

func decodeOptionalStrings(
	columns map[uint32]column,
	specification uint32,
	name string,
	expected int,
) ([]optional[string], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		return make([]optional[string], expected), nil
	}

	values, err := decodeStringColumn(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s: %w", name, err)
	}

	if err := requireItems(name, values, expected, true); err != nil {
		return nil, err
	}

	return values, nil
}

func decodeOptionalBooleans(
	columns map[uint32]column,
	specification uint32,
	expected int,
) ([]optional[bool], error) {
	data := optionalColumn(columns, specification)

	values := make([]optional[bool], expected)
	if data == nil {
		return values, nil
	}

	decoded, err := decodeBooleanColumn(data, expected)
	if err != nil {
		return nil, err
	}

	for i, value := range decoded {
		values[i] = optional[bool]{value: value, valid: true}
	}

	return values, nil
}

func decodeOptionalScalars(
	columns map[uint32]column,
	metaSpecification uint32,
	rawSpecification uint32,
	expected int,
) ([]optional[Scalar], error) {
	meta := optionalColumn(columns, metaSpecification)

	raw := optionalColumn(columns, rawSpecification)
	if meta == nil {
		if raw != nil {
			return nil, fmt.Errorf("raw value column is missing metadata")
		}

		return make([]optional[Scalar], expected), nil
	}

	return decodeScalars(meta, raw, expected)
}

func sumGroups(groups []optional[uint64]) (int, error) {
	var total uint64

	for i, group := range groups {
		if !group.valid {
			return 0, fmt.Errorf("group %d is null", i)
		}

		if total > math.MaxUint64-group.value {
			return 0, fmt.Errorf("group sizes overflow uint64")
		}

		total += group.value
		if total > maxDecodedItems {
			return 0, fmt.Errorf("grouped column exceeds %d items", maxDecodedItems)
		}
	}

	return int(total), nil
}

func decodeGroupedOpIDs(
	columns map[uint32]column,
	actors []ActorID,
	groupSpec uint32,
	actorSpec uint32,
	counterSpec uint32,
	expected int,
	name string,
) ([][]OpID, error) {
	result := make([][]OpID, expected)

	groupData := optionalColumn(columns, groupSpec)
	if groupData == nil {
		return result, nil
	}

	groups, err := decodeULEBColumn(groupData)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s groups: %w", name, err)
	}

	if err := requireItems(name+" group", groups, expected, false); err != nil {
		return nil, err
	}

	count, err := sumGroups(groups)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return result, nil
	}

	actorData, err := requireColumn(columns, actorSpec)
	if err != nil {
		return nil, err
	}

	actorIndexes, err := decodeULEBColumn(actorData)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s actors: %w", name, err)
	}

	if err := requireItems(name+" actor", actorIndexes, count, false); err != nil {
		return nil, err
	}

	counters, err := decodeRequiredDelta(columns, counterSpec, name+" counter", count)
	if err != nil {
		return nil, err
	}

	offset := 0

	for i, group := range groups {
		result[i] = make([]OpID, int(group.value))
		for j := range result[i] {
			id, err := opIDFromIndexes(actorIndexes[offset+j], counters[offset+j], actors)
			if err != nil {
				return nil, fmt.Errorf("%s %d: %w", name, offset+j, err)
			}

			result[i][j] = id
		}

		offset += int(group.value)
	}

	return result, nil
}

func opIDFromIndexes(
	actorIndex optional[uint64],
	counter optional[uint64],
	actors []ActorID,
) (OpID, error) {
	if !actorIndex.valid || !counter.valid {
		return OpID{}, fmt.Errorf("actor or counter is null")
	}

	if actorIndex.value >= uint64(len(actors)) {
		return OpID{}, fmt.Errorf("actor index %d is out of bounds", actorIndex.value)
	}

	if counter.value == 0 {
		return OpID{}, fmt.Errorf("counter is zero")
	}

	return OpID{Actor: actors[actorIndex.value], Counter: counter.value}, nil
}

func objectIDFromIndexes(
	actorIndex optional[uint64],
	counter optional[uint64],
	actors []ActorID,
) (ObjectID, error) {
	if !actorIndex.valid && !counter.valid {
		return RootObject(), nil
	}

	id, err := opIDFromIndexes(actorIndex, counter, actors)
	if err != nil {
		return ObjectID{}, err
	}

	return ObjectID{OpID: id}, nil
}

func keyFromColumns(
	actorIndex optional[uint64],
	counter optional[uint64],
	property optional[string],
	actors []ActorID,
	insert bool,
) (Key, error) {
	if property.valid {
		if actorIndex.valid || counter.valid {
			return Key{}, fmt.Errorf("property key also has an element ID")
		}

		value := property.value

		return Key{Property: &value}, nil
	}

	if !actorIndex.valid && insert && (!counter.valid || counter.value == 0) {
		return Key{IsHead: true}, nil
	}

	id, err := opIDFromIndexes(actorIndex, counter, actors)
	if err != nil {
		return Key{}, err
	}

	return Key{Element: &id}, nil
}

func collectUnknown(columns map[uint32]column) []RawColumn {
	specifications := make([]uint32, 0, len(columns))
	for specification := range columns {
		specifications = append(specifications, specification)
	}

	slices.Sort(specifications)

	result := make([]RawColumn, 0, len(columns))
	for _, specification := range specifications {
		value := columns[specification]
		result = append(
			result,
			RawColumn{
				Specification: value.specification,
				Data:          append([]byte(nil), value.data...),
			},
		)
	}

	return result
}

func assignOperations(changes []Change, operations []Operation) error {
	byActor := make(map[ActorID][]int)
	for i := range changes {
		byActor[changes[i].Actor] = append(byActor[changes[i].Actor], i)
	}

	for actor, indexes := range byActor {
		slices.SortFunc(
			indexes,
			func(left, right int) int {
				switch {
				case changes[left].MaxOp < changes[right].MaxOp:
					return -1
				case changes[left].MaxOp > changes[right].MaxOp:
					return 1
				default:
					return 0
				}
			},
		)

		var previous uint64
		for _, index := range indexes {
			changes[index].StartOp = previous + 1
			previous = changes[index].MaxOp
		}

		byActor[actor] = indexes
	}

	for _, operation := range operations {
		indexes := byActor[operation.ID.Actor]

		index, found := slices.BinarySearchFunc(
			indexes,
			operation.ID.Counter,
			func(index int, counter uint64) int {
				switch {
				case changes[index].MaxOp < counter:
					return -1
				case changes[index].MaxOp > counter:
					return 1
				default:
					return 0
				}
			},
		)
		if !found {
			if index >= len(indexes) {
				return fmt.Errorf(
					"operation %s@%d has no containing change",
					operation.ID.Actor,
					operation.ID.Counter,
				)
			}

			// BinarySearchFunc returns the insertion point, which is the first
			// change whose maxOp is greater than this counter.
		}

		changeIndex := indexes[index]
		if operation.ID.Counter < changes[changeIndex].StartOp {
			return fmt.Errorf("operation counter precedes containing change start")
		}

		changes[changeIndex].Operations = append(changes[changeIndex].Operations, operation)
	}

	for i := range changes {
		slices.SortFunc(changes[i].Operations, func(left, right Operation) int {
			return left.ID.Compare(right.ID)
		})
	}

	return nil
}

func validateSnapshotGraph(changes []Change, heads []uint64) error {
	dependedOn := make([]bool, len(changes))
	for i, change := range changes {
		seen := make(map[uint64]struct{}, len(change.DependencyIndexes))
		for _, dependency := range change.DependencyIndexes {
			if dependency >= uint64(len(changes)) {
				return fmt.Errorf("change %d dependency %d is out of bounds", i, dependency)
			}

			if dependency == uint64(i) {
				return fmt.Errorf("change %d depends on itself", i)
			}

			if _, ok := seen[dependency]; ok {
				return fmt.Errorf("change %d repeats dependency %d", i, dependency)
			}

			seen[dependency] = struct{}{}
			dependedOn[dependency] = true
		}
	}

	if err := detectIndexCycle(changes); err != nil {
		return err
	}

	expectedHeads := make(map[uint64]struct{})

	for i, isDependedOn := range dependedOn {
		if !isDependedOn {
			expectedHeads[uint64(i)] = struct{}{}
		}
	}

	actualHeads := make(map[uint64]struct{}, len(heads))
	for _, head := range heads {
		if _, exists := actualHeads[head]; exists {
			return fmt.Errorf("head index %d is repeated", head)
		}

		actualHeads[head] = struct{}{}
	}

	if !mapsEqual(expectedHeads, actualHeads) {
		return fmt.Errorf("head indexes do not match graph frontier")
	}

	return validateActorSequences(changes)
}

func detectIndexCycle(changes []Change) error {
	state := make([]uint8, len(changes))

	var visit func(int) error

	visit = func(index int) error {
		switch state[index] {
		case 1:
			return fmt.Errorf("dependency cycle includes change %d", index)
		case 2:
			return nil
		}

		state[index] = 1
		for _, dependency := range changes[index].DependencyIndexes {
			if err := visit(int(dependency)); err != nil {
				return err
			}
		}

		state[index] = 2

		return nil
	}
	for i := range changes {
		if err := visit(i); err != nil {
			return err
		}
	}

	return nil
}

func mapsEqual[K comparable](left, right map[K]struct{}) bool {
	if len(left) != len(right) {
		return false
	}

	for key := range left {
		if _, ok := right[key]; !ok {
			return false
		}
	}

	return true
}

func validateActorSequences(changes []Change) error {
	byActor := make(map[ActorID][]Change)
	for _, change := range changes {
		byActor[change.Actor] = append(byActor[change.Actor], change)
	}

	for actor, actorChanges := range byActor {
		slices.SortFunc(actorChanges, func(left, right Change) int {
			switch {
			case left.Sequence < right.Sequence:
				return -1
			case left.Sequence > right.Sequence:
				return 1
			default:
				return 0
			}
		})

		var previousMax uint64

		for i, change := range actorChanges {
			expectedSequence := uint64(i + 1)
			if change.Sequence != expectedSequence {
				return fmt.Errorf(
					"actor %s has sequence %d, expected %d",
					actor,
					change.Sequence,
					expectedSequence,
				)
			}

			if change.MaxOp < change.StartOp {
				if change.MaxOp < previousMax {
					return fmt.Errorf(
						"actor %s sequence %d empty change maxOp %d precedes %d",
						actor,
						change.Sequence,
						change.MaxOp,
						previousMax,
					)
				}

				previousMax = change.MaxOp

				continue
			}

			if change.MaxOp <= previousMax {
				return fmt.Errorf(
					"actor %s sequence %d has non-increasing maxOp %d",
					actor,
					change.Sequence,
					change.MaxOp,
				)
			}

			previousMax = change.MaxOp
		}
	}

	return nil
}

func mergeActors(existing, additions []ActorID) []ActorID {
	set := make(map[ActorID]struct{}, len(existing)+len(additions))
	for _, actor := range existing {
		set[actor] = struct{}{}
	}

	for _, actor := range additions {
		set[actor] = struct{}{}
	}

	result := make([]ActorID, 0, len(set))
	for actor := range set {
		result = append(result, actor)
	}

	slices.SortFunc(result, func(left, right ActorID) int {
		return left.Compare(right)
	})

	return result
}

func validateDocument(document *Document) error {
	if len(document.Changes) == 0 {
		if len(document.Heads) != 0 {
			return fmt.Errorf("empty history has heads")
		}

		return nil
	}

	if len(document.ChunkTypes) > 0 && document.ChunkTypes[0] == ChunkDocument {
		return validateChangeChunksAfterSnapshot(document)
	}

	return validateChangeChunkGraph(document)
}

func validateChangeChunksAfterSnapshot(document *Document) error {
	known := make(map[ChangeHash]struct{})
	dependedOn := make(map[ChangeHash]struct{})

	for _, change := range document.Changes {
		if change.Hash != nil {
			if _, exists := known[*change.Hash]; exists {
				return fmt.Errorf("duplicate known change hash %s", change.Hash)
			}

			known[*change.Hash] = struct{}{}
		}
	}

	for _, change := range document.Changes {
		if len(change.DependencyIndexes) > 0 {
			continue
		}

		for _, dependency := range change.Dependencies {
			if _, ok := known[dependency]; !ok {
				return fmt.Errorf("change %s has missing dependency %s", change.Hash, dependency)
			}

			dependedOn[dependency] = struct{}{}
		}
	}

	document.Heads = document.Heads[:0]

	for hash := range known {
		if _, ok := dependedOn[hash]; !ok {
			document.Heads = append(document.Heads, hash)
		}
	}

	slices.SortFunc(document.Heads, func(left, right ChangeHash) int {
		return bytes.Compare(left[:], right[:])
	})

	return validateActorSequences(document.Changes)
}

func validateChangeChunkGraph(document *Document) error {
	changes := make(map[ChangeHash]Change, len(document.Changes))
	dependedOn := make(map[ChangeHash]struct{})

	for _, change := range document.Changes {
		if change.Hash == nil {
			return fmt.Errorf("change chunk has no hash")
		}

		if _, exists := changes[*change.Hash]; exists {
			return fmt.Errorf("duplicate change %s", change.Hash)
		}

		changes[*change.Hash] = change
	}

	for _, change := range document.Changes {
		for _, dependency := range change.Dependencies {
			if _, ok := changes[dependency]; !ok {
				return fmt.Errorf("change %s has missing dependency %s", change.Hash, dependency)
			}

			dependedOn[dependency] = struct{}{}
		}
	}

	document.Heads = document.Heads[:0]

	for hash := range changes {
		if _, ok := dependedOn[hash]; !ok {
			document.Heads = append(document.Heads, hash)
		}
	}

	slices.SortFunc(document.Heads, func(left, right ChangeHash) int {
		return bytes.Compare(left[:], right[:])
	})

	return validateActorSequences(document.Changes)
}
