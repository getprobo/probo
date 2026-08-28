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
	"fmt"
	"math"
	"slices"
	"sort"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func decodeActorArray(
	r *reader,
	sorted bool,
	budget *decodeBudget,
) ([]opset.ActorID, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, err
	}

	if count > maxDecodedItems {
		return nil, fmt.Errorf("actor count %d exceeds limit", count)
	}

	if err := chargeDecoded[opset.ActorID](budget, count); err != nil {
		return nil, err
	}

	actors := make([]opset.ActorID, 0, count)
	for i := range count {
		value, err := decodeLengthPrefixed(r)
		if err != nil {
			return nil, fmt.Errorf("cannot decode actor %d: %w", i, err)
		}

		if err := chargeDecodedBytes(budget, uint64(len(value))); err != nil {
			return nil, err
		}

		actor, err := opset.NewActorID(value)
		if err != nil {
			return nil, fmt.Errorf("actor %d: %w", i, err)
		}

		if sorted &&
			len(actors) > 0 &&
			string(actors[len(actors)-1]) >= string(actor) {
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

func decodeHashArray(
	r *reader,
	sorted bool,
	budget *decodeBudget,
) ([]opset.ChangeHash, error) {
	count, err := r.uleb()
	if err != nil {
		return nil, err
	}

	if count > maxDecodedItems {
		return nil, fmt.Errorf("hash count %d exceeds limit", count)
	}

	if err := chargeDecoded[opset.ChangeHash](budget, count); err != nil {
		return nil, err
	}

	hashes := make([]opset.ChangeHash, 0, count)
	for i := range count {
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
	budget *decodeBudget,
) ([]optional[uint64], error) {
	data, err := requireColumn(columns, specification)
	if err != nil {
		return nil, err
	}

	values, err := decodeDeltaColumnWithBudget(data, budget)
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
	budget *decodeBudget,
) ([]optional[uint64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		if err := chargeDecoded[optional[uint64]](budget, uint64(expected)); err != nil {
			return nil, err
		}

		return make([]optional[uint64], expected), nil
	}

	values, err := decodeDeltaColumnWithBudget(data, budget)
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
	budget *decodeBudget,
) ([]optional[int64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		if err := chargeDecoded[optional[int64]](budget, uint64(expected)); err != nil {
			return nil, err
		}

		return make([]optional[int64], expected), nil
	}

	values, err := decodeSignedDeltaColumnWithBudget(data, budget)
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
	budget *decodeBudget,
) ([]optional[uint64], error) {
	data, err := requireColumn(columns, specification)
	if err != nil {
		return nil, err
	}

	values, err := decodeULEBColumnWithBudget(data, budget)
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
	budget *decodeBudget,
) ([]optional[uint64], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		if err := chargeDecoded[optional[uint64]](budget, uint64(expected)); err != nil {
			return nil, err
		}

		return make([]optional[uint64], expected), nil
	}

	values, err := decodeULEBColumnWithBudget(data, budget)
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
	budget *decodeBudget,
) ([]optional[string], error) {
	data := optionalColumn(columns, specification)
	if data == nil {
		if err := chargeDecoded[optional[string]](budget, uint64(expected)); err != nil {
			return nil, err
		}

		return make([]optional[string], expected), nil
	}

	values, err := decodeStringColumnWithBudget(data, budget)
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
	budget *decodeBudget,
) ([]optional[bool], error) {
	if err := chargeDecoded[optional[bool]](budget, uint64(expected)); err != nil {
		return nil, err
	}

	data := optionalColumn(columns, specification)

	values := make([]optional[bool], expected)
	if data == nil {
		return values, nil
	}

	decoded, err := decodeBooleanColumnWithBudget(data, expected, budget)
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
	budget *decodeBudget,
) ([]optional[opset.Scalar], error) {
	meta := optionalColumn(columns, metaSpecification)

	raw := optionalColumn(columns, rawSpecification)
	if meta == nil {
		if raw != nil {
			return nil, fmt.Errorf("raw value column is missing metadata")
		}

		if err := chargeDecoded[optional[opset.Scalar]](budget, uint64(expected)); err != nil {
			return nil, err
		}

		return make([]optional[opset.Scalar], expected), nil
	}

	values, _, err := decodeScalarsInternal(meta, raw, expected, budget, false)

	return values, err
}

func decodeSnapshotExtras(
	columns map[uint32]column,
	expected int,
	budget *decodeBudget,
) ([]optional[opset.Scalar], [][]byte, error) {
	meta := optionalColumn(columns, 86)

	raw := optionalColumn(columns, 87)
	if meta == nil {
		if raw != nil {
			return nil, nil, fmt.Errorf("raw value column is missing metadata")
		}

		if err := chargeDecoded[optional[opset.Scalar]](budget, uint64(expected)); err != nil {
			return nil, nil, err
		}

		if err := chargeDecoded[[]byte](budget, uint64(expected)); err != nil {
			return nil, nil, err
		}

		return make([]optional[opset.Scalar], expected), make([][]byte, expected), nil
	}

	return decodeScalarsWithRaw(meta, raw, expected, budget)
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
	actors []opset.ActorID,
	groupSpec uint32,
	actorSpec uint32,
	counterSpec uint32,
	expected int,
	name string,
	budget *decodeBudget,
) ([][]opset.OpID, error) {
	if err := chargeDecoded[[]opset.OpID](budget, uint64(expected)); err != nil {
		return nil, err
	}

	result := make([][]opset.OpID, expected)

	groupData := optionalColumn(columns, groupSpec)
	if groupData == nil {
		return result, nil
	}

	groups, err := decodeULEBColumnWithBudget(groupData, budget)
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

	actorIndexes, err := decodeULEBColumnWithBudget(actorData, budget)
	if err != nil {
		return nil, fmt.Errorf("cannot decode %s actors: %w", name, err)
	}

	if err := requireItems(name+" actor", actorIndexes, count, false); err != nil {
		return nil, err
	}

	counters, err := decodeRequiredDelta(
		columns,
		counterSpec,
		name+" counter",
		count,
		budget,
	)
	if err != nil {
		return nil, err
	}

	offset := 0

	for i, group := range groups {
		if err := chargeDecoded[opset.OpID](budget, group.value); err != nil {
			return nil, err
		}

		result[i] = make([]opset.OpID, int(group.value))
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
	actors []opset.ActorID,
) (opset.OpID, error) {
	if !actorIndex.valid || !counter.valid {
		return opset.OpID{}, fmt.Errorf("actor or counter is null")
	}

	if actorIndex.value >= uint64(len(actors)) {
		return opset.OpID{}, fmt.Errorf("actor index %d is out of bounds", actorIndex.value)
	}

	if counter.value == 0 {
		return opset.OpID{}, fmt.Errorf("counter is zero")
	}

	return opset.OpID{Actor: actors[actorIndex.value], Counter: counter.value}, nil
}

func objectIDFromIndexes(
	actorIndex optional[uint64],
	counter optional[uint64],
	actors []opset.ActorID,
) (opset.ObjectID, error) {
	if !actorIndex.valid && !counter.valid {
		return opset.RootObject(), nil
	}

	id, err := opIDFromIndexes(actorIndex, counter, actors)
	if err != nil {
		return opset.ObjectID{}, err
	}

	return opset.ObjectID{OpID: id}, nil
}

func keyFromColumns(
	actorIndex optional[uint64],
	counter optional[uint64],
	property optional[string],
	actors []opset.ActorID,
	insert bool,
) (opset.Key, error) {
	if property.valid {
		if actorIndex.valid || counter.valid {
			return opset.Key{}, fmt.Errorf("property key also has an element ID")
		}

		value := property.value

		return opset.Key{Property: &value}, nil
	}

	if !actorIndex.valid && insert && (!counter.valid || counter.value == 0) {
		return opset.Key{IsHead: true}, nil
	}

	id, err := opIDFromIndexes(actorIndex, counter, actors)
	if err != nil {
		return opset.Key{}, err
	}

	return opset.Key{Element: &id}, nil
}

func collectUnknown(
	columns map[uint32]column,
	budget *decodeBudget,
) ([]opset.RawColumn, error) {
	if err := chargeDecoded[uint32](budget, uint64(len(columns))); err != nil {
		return nil, err
	}

	if err := chargeDecoded[opset.RawColumn](budget, uint64(len(columns))); err != nil {
		return nil, err
	}

	specifications := make([]uint32, 0, len(columns))
	for specification := range columns {
		specifications = append(specifications, specification)
	}

	slices.Sort(specifications)

	result := make([]opset.RawColumn, 0, len(columns))
	for _, specification := range specifications {
		value := columns[specification]

		result = append(
			result, opset.RawColumn{
				Specification: value.specification,
				Data:          value.data,
			},
		)
	}

	return result, nil
}

func assignOperations(
	changes []opset.Change,
	operations []opset.Operation,
	budget *decodeBudget,
) error {
	const (
		visiting uint8 = 1
		visited  uint8 = 2
	)

	type traversalFrame struct {
		index          int
		nextDependency int
		dependencyMax  uint64
	}

	if err := chargeDecoded[uint8](budget, uint64(len(changes))); err != nil {
		return err
	}

	if err := chargeDecoded[traversalFrame](budget, uint64(len(changes))); err != nil {
		return err
	}

	states := make([]uint8, len(changes))
	stack := make([]traversalFrame, 0, len(changes))

	for root := range changes {
		if states[root] == visited {
			continue
		}

		states[root] = visiting
		stack = append(stack, traversalFrame{index: root})

		for len(stack) > 0 {
			frame := &stack[len(stack)-1]
			dependencies := changes[frame.index].DependencyIndexes

			if frame.nextDependency < len(dependencies) {
				dependency := dependencies[frame.nextDependency]
				frame.nextDependency++

				if dependency >= uint64(len(changes)) {
					return fmt.Errorf(
						"change %d dependency %d is out of bounds",
						frame.index,
						dependency,
					)
				}

				dependencyIndex := int(dependency)
				switch states[dependencyIndex] {
				case visiting:
					return fmt.Errorf(
						"snapshot dependency cycle includes change %d",
						dependency,
					)
				case visited:
					frame.dependencyMax = max(
						frame.dependencyMax,
						changes[dependencyIndex].MaxOp,
					)
				default:
					states[dependencyIndex] = visiting
					stack = append(stack, traversalFrame{index: dependencyIndex})
				}

				continue
			}

			if frame.dependencyMax == math.MaxUint64 {
				return fmt.Errorf("change %d start operation overflows uint64", frame.index)
			}

			changes[frame.index].StartOp = frame.dependencyMax + 1
			states[frame.index] = visited
			completedMax := changes[frame.index].MaxOp
			stack = stack[:len(stack)-1]

			if len(stack) > 0 {
				parent := &stack[len(stack)-1]
				parent.dependencyMax = max(parent.dependencyMax, completedMax)
			}
		}
	}

	if err := chargeDecodedMap[opset.ActorID, []int](budget, uint64(len(changes))); err != nil {
		return err
	}

	if err := chargeDecodedSliceGrowth[int](budget, uint64(len(changes))); err != nil {
		return err
	}

	byActor := make(map[opset.ActorID][]int)
	for i := range changes {
		if changes[i].MaxOp < changes[i].StartOp {
			continue
		}

		byActor[changes[i].Actor] = append(byActor[changes[i].Actor], i)
	}

	for actor, candidates := range byActor {
		slices.SortFunc(
			candidates,
			func(left, right int) int {
				switch {
				case changes[left].StartOp < changes[right].StartOp:
					return -1
				case changes[left].StartOp > changes[right].StartOp:
					return 1
				case changes[left].MaxOp < changes[right].MaxOp:
					return -1
				case changes[left].MaxOp > changes[right].MaxOp:
					return 1
				default:
					return left - right
				}
			},
		)

		for i := 1; i < len(candidates); i++ {
			previous := changes[candidates[i-1]]
			current := changes[candidates[i]]
			if current.StartOp <= previous.MaxOp {
				return fmt.Errorf(
					"actor %s has overlapping operation ranges %d..%d and %d..%d",
					actor,
					previous.StartOp,
					previous.MaxOp,
					current.StartOp,
					current.MaxOp,
				)
			}
		}
	}

	if err := chargeDecoded[int](budget, uint64(len(operations))); err != nil {
		return err
	}

	if err := chargeDecoded[int](budget, uint64(len(changes))); err != nil {
		return err
	}

	owners := make([]int, len(operations))
	counts := make([]int, len(changes))

	for i, operation := range operations {
		owner := -1
		candidates := byActor[operation.ID.Actor]
		position := sort.Search(
			len(candidates),
			func(index int) bool {
				return changes[candidates[index]].StartOp > operation.ID.Counter
			},
		)

		if position > 0 {
			candidate := candidates[position-1]
			if operation.ID.Counter <= changes[candidate].MaxOp {
				owner = candidate
			}
		}

		if owner < 0 {
			return fmt.Errorf(
				"operation %s@%d has no containing change",
				operation.ID.Actor,
				operation.ID.Counter,
			)
		}

		owners[i] = owner
		counts[owner]++
	}

	if err := chargeDecoded[opset.Operation](budget, uint64(len(operations))); err != nil {
		return err
	}

	for i := range changes {
		changes[i].Operations = make([]opset.Operation, 0, counts[i])
	}

	for i, operation := range operations {
		owner := owners[i]
		changes[owner].Operations = append(changes[owner].Operations, operation)
	}

	for i := range changes {
		slices.SortFunc(
			changes[i].Operations,
			func(left, right opset.Operation) int {
				return left.ID.Compare(right.ID)
			},
		)

		expected := uint64(0)
		if changes[i].MaxOp >= changes[i].StartOp {
			expected = changes[i].MaxOp - changes[i].StartOp + 1
		}

		if uint64(len(changes[i].Operations)) != expected {
			return fmt.Errorf(
				"change %d holds %d operations, expected %d for range %d..%d",
				i,
				len(changes[i].Operations),
				expected,
				changes[i].StartOp,
				changes[i].MaxOp,
			)
		}
	}

	return nil
}

func validateSnapshotMarkOrder(operations []opset.Operation) error {
	byID := make(map[opset.OpID]opset.Operation, len(operations))
	objects := make(map[opset.ObjectID]struct{})
	for _, operation := range operations {
		byID[operation.ID] = operation
		if operation.Action == opset.ActionMark {
			objects[operation.Object] = struct{}{}
		}
	}

	for object := range objects {
		children := make(map[opset.OpID][]opset.Operation)
		var head []opset.Operation
		for _, operation := range operations {
			if operation.Object != object || !operation.Insert {
				continue
			}
			if operation.Key.IsHead {
				head = append(head, operation)
			} else if operation.Key.Element != nil {
				children[*operation.Key.Element] = append(
					children[*operation.Key.Element],
					operation,
				)
			}
		}

		seen := make(map[opset.OpID]struct{})
		visited := make(map[opset.OpID]struct{})
		var visit func([]opset.Operation) error
		visit = func(rows []opset.Operation) error {
			slices.SortFunc(rows, func(left, right opset.Operation) int {
				return right.ID.Compare(left.ID)
			})
			for _, operation := range rows {
				if _, ok := visited[operation.ID]; ok {
					continue
				}
				visited[operation.ID] = struct{}{}
				if operation.Action == opset.ActionMark {
					if operation.MarkName != nil {
						seen[operation.ID] = struct{}{}
					} else if operation.ID.Counter > 0 {
						beginID := opset.OpID{
							Actor:   operation.ID.Actor,
							Counter: operation.ID.Counter - 1,
						}
						begin, ok := byID[beginID]
						if ok &&
							begin.Action == opset.ActionMark &&
							begin.MarkName != nil &&
							begin.Object == operation.Object {
							if _, ok := seen[beginID]; !ok {
								return fmt.Errorf(
									"invalid mark operation order: end %v precedes begin %v",
									operation.ID,
									beginID,
								)
							}
						}
					}
				}
				if err := visit(children[operation.ID]); err != nil {
					return err
				}
			}
			return nil
		}
		if err := visit(head); err != nil {
			return err
		}
	}

	return nil
}

func validateSnapshotGraph(
	changes []opset.Change,
	heads []uint64,
	budget *decodeBudget,
) error {
	if err := chargeDecoded[bool](budget, uint64(len(changes))); err != nil {
		return err
	}

	dependedOn := make([]bool, len(changes))
	for i, change := range changes {
		if err := chargeDecodedMap[uint64, struct{}](
			budget,
			uint64(len(change.DependencyIndexes)),
		); err != nil {
			return err
		}

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

	if err := detectIndexCycle(changes, budget); err != nil {
		return err
	}

	if err := chargeDecodedMap[uint64, struct{}](budget, uint64(len(changes))); err != nil {
		return err
	}

	expectedHeads := make(map[uint64]struct{})

	for i, isDependedOn := range dependedOn {
		if !isDependedOn {
			expectedHeads[uint64(i)] = struct{}{}
		}
	}

	if err := chargeDecodedMap[uint64, struct{}](budget, uint64(len(heads))); err != nil {
		return err
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

	return validateActorSequences(changes, budget)
}

func detectIndexCycle(changes []opset.Change, budget *decodeBudget) error {
	type traversalFrame struct {
		index          int
		nextDependency int
	}

	if err := chargeDecoded[uint8](budget, uint64(len(changes))); err != nil {
		return err
	}

	if err := chargeDecoded[traversalFrame](budget, uint64(len(changes))); err != nil {
		return err
	}

	state := make([]uint8, len(changes))
	stack := make([]traversalFrame, 0, len(changes))

	for root := range changes {
		if state[root] == 2 {
			continue
		}

		state[root] = 1
		stack = append(stack, traversalFrame{index: root})

		for len(stack) > 0 {
			frame := &stack[len(stack)-1]
			dependencies := changes[frame.index].DependencyIndexes

			if frame.nextDependency >= len(dependencies) {
				state[frame.index] = 2
				stack = stack[:len(stack)-1]

				continue
			}

			dependency := int(dependencies[frame.nextDependency])
			frame.nextDependency++

			switch state[dependency] {
			case 1:
				return fmt.Errorf("dependency cycle includes change %d", dependency)
			case 2:
				continue
			default:
				state[dependency] = 1
				stack = append(stack, traversalFrame{index: dependency})
			}
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

func validateActorSequences(
	changes []opset.Change,
	budget *decodeBudget,
) error {
	if err := chargeDecodedMap[opset.ActorID, []opset.Change](
		budget,
		uint64(len(changes)),
	); err != nil {
		return err
	}

	if err := chargeDecodedSliceGrowth[opset.Change](budget, uint64(len(changes))); err != nil {
		return err
	}

	byActor := make(map[opset.ActorID][]opset.Change)
	for _, change := range changes {
		byActor[change.Actor] = append(byActor[change.Actor], change)
	}

	for actor, actorChanges := range byActor {
		slices.SortFunc(
			actorChanges,
			func(left, right opset.Change) int {
				switch {
				case left.Sequence < right.Sequence:
					return -1
				case left.Sequence > right.Sequence:
					return 1
				default:
					return 0
				}
			},
		)

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

func mergeActors(
	existing []opset.ActorID,
	additions []opset.ActorID,
	budget *decodeBudget,
) ([]opset.ActorID, error) {
	count := uint64(len(existing) + len(additions))
	if err := chargeDecodedMap[opset.ActorID, struct{}](budget, count); err != nil {
		return nil, err
	}

	set := make(map[opset.ActorID]struct{}, len(existing)+len(additions))
	for _, actor := range existing {
		set[actor] = struct{}{}
	}

	for _, actor := range additions {
		set[actor] = struct{}{}
	}

	if err := chargeDecoded[opset.ActorID](budget, uint64(len(set))); err != nil {
		return nil, err
	}

	result := make([]opset.ActorID, 0, len(set))
	for actor := range set {
		result = append(result, actor)
	}

	slices.SortFunc(
		result,
		func(left, right opset.ActorID) int {
			switch {
			case string(left) < string(right):
				return -1
			case string(left) > string(right):
				return 1
			default:
				return 0
			}
		},
	)

	return result, nil
}

func validateDocument(document *opset.Document, budget *decodeBudget) error {
	if len(document.Changes) == 0 {
		if len(document.Heads) != 0 {
			return fmt.Errorf("empty history has heads")
		}

		return nil
	}

	if len(document.ChunkTypes) > 0 && document.ChunkTypes[0] == opset.ChunkDocument {
		return validateChangeChunksAfterSnapshot(document, budget)
	}

	return validateChangeChunkGraph(document, budget)
}

func validateChangeChunksAfterSnapshot(
	document *opset.Document,
	budget *decodeBudget,
) error {
	if err := chargeDecodedMap[opset.ChangeHash, struct{}](
		budget,
		uint64(len(document.Changes)),
	); err != nil {
		return err
	}

	if err := chargeDecodedMap[opset.ChangeHash, struct{}](
		budget,
		uint64(len(document.Changes)),
	); err != nil {
		return err
	}

	known := make(map[opset.ChangeHash]struct{})
	dependedOn := make(map[opset.ChangeHash]struct{})

	// A change may legitimately appear both inside the snapshot and as a trailing
	// change chunk, so repeats identify the same change rather than a conflict.
	for _, change := range document.Changes {
		if change.Hash != nil {
			known[*change.Hash] = struct{}{}
		}
	}

	for _, change := range document.Changes {
		for _, dependency := range change.Dependencies {
			if _, ok := known[dependency]; !ok {
				return fmt.Errorf("change %s has missing dependency %s", change.Hash, dependency)
			}

			dependedOn[dependency] = struct{}{}
		}
	}

	document.Heads = document.Heads[:0]
	if cap(document.Heads) < len(known) {
		if err := chargeDecoded[opset.ChangeHash](budget, uint64(len(known))); err != nil {
			return err
		}

		document.Heads = make([]opset.ChangeHash, 0, len(known))
	}

	for hash := range known {
		if _, ok := dependedOn[hash]; !ok {
			document.Heads = append(document.Heads, hash)
		}
	}

	slices.SortFunc(
		document.Heads,
		func(left, right opset.ChangeHash) int {
			return bytes.Compare(left[:], right[:])
		},
	)

	return validateActorSequences(document.Changes, budget)
}

func validateChangeChunkGraph(
	document *opset.Document,
	budget *decodeBudget,
) error {
	if err := chargeDecodedMap[opset.ChangeHash, opset.Change](
		budget,
		uint64(len(document.Changes)),
	); err != nil {
		return err
	}

	if err := chargeDecodedMap[opset.ChangeHash, struct{}](
		budget,
		uint64(len(document.Changes)),
	); err != nil {
		return err
	}

	changes := make(map[opset.ChangeHash]opset.Change, len(document.Changes))
	dependedOn := make(map[opset.ChangeHash]struct{})

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
	if cap(document.Heads) < len(changes) {
		if err := chargeDecoded[opset.ChangeHash](budget, uint64(len(changes))); err != nil {
			return err
		}

		document.Heads = make([]opset.ChangeHash, 0, len(changes))
	}

	for hash := range changes {
		if _, ok := dependedOn[hash]; !ok {
			document.Heads = append(document.Heads, hash)
		}
	}

	slices.SortFunc(
		document.Heads,
		func(left, right opset.ChangeHash) int {
			return bytes.Compare(left[:], right[:])
		},
	)

	return validateActorSequences(document.Changes, budget)
}
