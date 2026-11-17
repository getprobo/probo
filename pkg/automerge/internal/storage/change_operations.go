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
	"cmp"
	"iter"
	"math"
	"slices"
	"sort"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

// changeOperationDelete marks a key whose operation is a delete, held apart
// from the snapshot's rows because a snapshot stores no row for it.
const changeOperationDelete = 1 << 31

type (
	// ChangeOperationIndex locates the operations each change of a decoded
	// snapshot wrote, so no change needs its own copy of them. A snapshot
	// stores every operation once, in document order; a change's operations
	// are the ones whose identifiers fall in its counter range.
	ChangeOperationIndex struct {
		rows    []opset.Operation
		deletes []opset.Operation
		// keys holds one entry per operation, grouped by actor and sorted by
		// counter within an actor. The counter is the high half of an entry
		// and the operation's place the low half: a row number, or a delete
		// number with changeOperationDelete set. The operations of one change
		// are consecutive entries.
		keys  []uint64
		spans []changeOperationSpan
	}

	changeOperationSpan struct {
		start uint32
		count uint32
	}

	// OperationRowLookup finds the row of an operation in the snapshot an
	// index was built from. It keeps the index's keys and not its rows, so
	// holding it does not keep a decoded operation table alive.
	OperationRowLookup struct {
		keys   []uint64
		ranges map[opset.ActorID][]operationCounterRange
	}

	// operationCounterRange is the counter range one change wrote and where
	// its keys start.
	operationCounterRange struct {
		start uint64
		end   uint64
		first uint32
	}

	// actorNumbers numbers the actors of one document. Consecutive lookups
	// nearly always name the same actor, so the last answer is remembered.
	actorNumbers struct {
		numbers    map[opset.ActorID]int
		lastActor  opset.ActorID
		lastNumber int
		hasLast    bool
	}
)

func newActorNumbers(actors []opset.ActorID) *actorNumbers {
	numbers := make(map[opset.ActorID]int, len(actors))
	for i, actor := range actors {
		numbers[actor] = i
	}

	return &actorNumbers{numbers: numbers}
}

func (n *actorNumbers) number(actor opset.ActorID) (int, bool) {
	if n.hasLast && actor == n.lastActor {
		return n.lastNumber, true
	}

	number, ok := n.numbers[actor]
	if ok {
		n.lastActor, n.lastNumber, n.hasLast = actor, number, true
	}

	return number, ok
}

// Deletes returns the delete operations the snapshot implies, ordered by
// identifier. Each carries the operations it removed as its predecessors.
func (x *ChangeOperationIndex) Deletes() []opset.Operation {
	return x.deletes
}

// ChangeDeletes yields the deletes the change at the given index wrote, in
// counter order.
func (x *ChangeOperationIndex) ChangeDeletes(change int) iter.Seq[opset.Operation] {
	span := x.spans[change]
	keys := x.keys[span.start : span.start+span.count]

	return func(yield func(opset.Operation) bool) {
		for _, key := range keys {
			place := uint32(key)
			if place&changeOperationDelete == 0 {
				continue
			}

			if !yield(x.deletes[place&^changeOperationDelete]) {
				return
			}
		}
	}
}

// RowLookup returns a lookup from operation identifier to row. Every change
// fills its counter range, so an operation is found by its actor's range and
// its offset in it, without one map entry per operation.
func (x *ChangeOperationIndex) RowLookup(changes []opset.Change) *OperationRowLookup {
	lookup := &OperationRowLookup{
		keys:   x.keys,
		ranges: make(map[opset.ActorID][]operationCounterRange),
	}

	for i := range changes {
		span := x.spans[i]
		if span.count == 0 {
			continue
		}

		actor := changes[i].Actor
		lookup.ranges[actor] = append(
			lookup.ranges[actor],
			operationCounterRange{
				start: changes[i].StartOp,
				end:   changes[i].MaxOp,
				first: span.start,
			},
		)
	}

	for _, ranges := range lookup.ranges {
		slices.SortFunc(
			ranges,
			func(left, right operationCounterRange) int {
				return cmp.Compare(left.start, right.start)
			},
		)
	}

	return lookup
}

// Row returns the row of the operation with the given identifier. It reports
// false for an identifier the snapshot holds no row for, which includes its
// deletes.
func (l *OperationRowLookup) Row(identifier opset.OpID) (int, bool) {
	ranges := l.ranges[identifier.Actor]

	position := sort.Search(
		len(ranges),
		func(i int) bool {
			return ranges[i].end >= identifier.Counter
		},
	)
	if position == len(ranges) || ranges[position].start > identifier.Counter {
		return 0, false
	}

	found := ranges[position]

	place := uint32(l.keys[uint64(found.first)+identifier.Counter-found.start])
	if place&changeOperationDelete != 0 {
		return 0, false
	}

	return int(place), true
}

// Len returns the number of operations the change at the given index wrote.
func (x *ChangeOperationIndex) Len(change int) int {
	return int(x.spans[change].count)
}

// Operations yields the operations of the change at the given index in
// counter order, which is the order a change chunk stores them in.
func (x *ChangeOperationIndex) Operations(
	change int,
) iter.Seq2[int, opset.Operation] {
	span := x.spans[change]
	keys := x.keys[span.start : span.start+span.count]

	return func(yield func(int, opset.Operation) bool) {
		for i, key := range keys {
			place := uint32(key)

			var operation opset.Operation
			if place&changeOperationDelete != 0 {
				operation = x.deletes[place&^changeOperationDelete]
			} else {
				operation = x.rows[place]
			}

			if !yield(i, operation) {
				return
			}
		}
	}
}

// indexChangeOperations restores the change view of a snapshot's operations
// in place. It sets every row's predecessors, derives the deletes the rows
// imply, sets every change's first operation counter, and returns the index
// locating each change's operations.
//
// It handles the snapshot every writer produces: each change's operations
// fill its counter range exactly. It reports false for anything else, and for
// a budget it cannot meet, having charged nothing; the caller then decodes
// the snapshot the long way, which either accepts it or names what is wrong.
func indexChangeOperations(
	changes []opset.Change,
	actors []opset.ActorID,
	rows []opset.Operation,
	budget *decodeBudget,
) (*ChangeOperationIndex, bool) {
	chargedFrom := budgetUsed(budget)

	index, ok := buildChangeOperationIndex(changes, actors, rows, budget)
	if !ok {
		releaseDecodedBytes(budget, budgetUsed(budget)-chargedFrom)

		return nil, false
	}

	return index, true
}

func buildChangeOperationIndex(
	changes []opset.Change,
	actors []opset.ActorID,
	rows []opset.Operation,
	budget *decodeBudget,
) (*ChangeOperationIndex, bool) {
	if len(rows) >= changeOperationDelete || len(changes) > math.MaxInt32 {
		return nil, false
	}

	if chargeDecodedMap[opset.ActorID, int](budget, uint64(len(actors))) != nil {
		return nil, false
	}

	var (
		numbers = newActorNumbers(actors)
		index   = &ChangeOperationIndex{rows: rows}
	)

	keysFrom := budgetUsed(budget)

	segments, ok := index.buildKeys(numbers, len(actors), budget)
	if !ok {
		return nil, false
	}

	rowKeys := budgetUsed(budget) - keysFrom

	if !index.restorePredecessors(segments, numbers, budget) {
		return nil, false
	}

	if len(index.deletes) > 0 {
		// The deletes are operations of their changes too, so the keys are
		// built again with them in place of the keys that held rows alone.
		segments, ok = index.buildKeys(numbers, len(actors), budget)
		if !ok {
			return nil, false
		}

		releaseDecodedBytes(budget, rowKeys)
	}

	if !index.assignChanges(changes, segments, numbers, budget) {
		return nil, false
	}

	return index, true
}

// buildKeys writes one key per row and delete into its actor's segment and
// sorts each segment by counter. It returns the bounds of each actor's
// segment: segments[a] to segments[a+1] for actor number a. It reports false
// when two operations of one actor share a counter.
func (x *ChangeOperationIndex) buildKeys(
	numbers *actorNumbers,
	actorCount int,
	budget *decodeBudget,
) ([]int, bool) {
	total := uint64(len(x.rows)) + uint64(len(x.deletes))

	if total > math.MaxUint32 ||
		chargeDecoded[int](budget, 2*uint64(actorCount)+1) != nil ||
		chargeDecoded[uint64](budget, total) != nil {
		return nil, false
	}

	segments := make([]int, actorCount+1)

	for _, operations := range [][]opset.Operation{x.rows, x.deletes} {
		for i := range operations {
			identifier := operations[i].ID
			if identifier.Counter > math.MaxUint32 {
				return nil, false
			}

			actor, ok := numbers.number(identifier.Actor)
			if !ok {
				return nil, false
			}

			segments[actor+1]++
		}
	}

	for actor := range actorCount {
		segments[actor+1] += segments[actor]
	}

	var (
		keys   = make([]uint64, total)
		filled = make([]int, actorCount)
	)

	for kind, operations := range [][]opset.Operation{x.rows, x.deletes} {
		flag := uint64(kind) * changeOperationDelete

		for i := range operations {
			identifier := operations[i].ID
			actor, _ := numbers.number(identifier.Actor)

			keys[segments[actor]+filled[actor]] = identifier.Counter<<32 | flag | uint64(i)
			filled[actor]++
		}
	}

	for actor := range actorCount {
		segment := keys[segments[actor]:segments[actor+1]]
		if !slices.IsSorted(segment) {
			slices.Sort(segment)
		}

		for i := 1; i < len(segment); i++ {
			if segment[i]>>32 == segment[i-1]>>32 {
				return nil, false
			}
		}
	}

	x.keys = keys

	return segments, true
}

// row returns the row holding the operation with the given identifier.
func (x *ChangeOperationIndex) row(
	segments []int,
	numbers *actorNumbers,
	identifier opset.OpID,
) (int32, bool) {
	actor, ok := numbers.number(identifier.Actor)
	if !ok {
		return 0, false
	}

	segment := x.keys[segments[actor]:segments[actor+1]]

	position, found := sort.Find(
		len(segment),
		func(i int) int {
			switch counter := segment[i] >> 32; {
			case identifier.Counter < counter:
				return -1
			case identifier.Counter > counter:
				return 1
			default:
				return 0
			}
		},
	)
	if !found {
		return 0, false
	}

	return int32(uint32(segment[position])), true
}

// restorePredecessors inverts the rows' successor lists. A snapshot records,
// for every surviving operation, the operations that superseded it, and drops
// delete operations because they survive only as those successor entries. A
// change instead names each operation's predecessors and spells its deletes
// out. Successors are left in place: the engine reads a loaded snapshot from
// them, and encoding a change only reads predecessors.
func (x *ChangeOperationIndex) restorePredecessors(
	segments []int,
	numbers *actorNumbers,
	budget *decodeBudget,
) bool {
	rows := x.rows

	var successorCount uint64
	for i := range rows {
		successorCount += uint64(len(rows[i].Successors))
	}

	if successorCount == 0 {
		for i := range rows {
			rows[i].Predecessors = nil
		}

		return true
	}

	if successorCount > maxDecodedItems ||
		chargeDecoded[opset.OpID](budget, successorCount) != nil {
		return false
	}

	// The tables below are working memory, released once the lists are cut.
	scratchFrom := budgetUsed(budget)

	if chargeDecoded[int32](budget, 2*uint64(len(rows))+1) != nil ||
		chargeDecoded[int32](budget, successorCount) != nil ||
		chargeDecodedMap[opset.OpID, int32](budget, successorCount) != nil ||
		chargeDecodedSliceGrowth[int32](budget, successorCount) != nil {
		return false
	}

	scratch := budgetUsed(budget) - scratchFrom

	var (
		// offsets[r]..offsets[r+1] bounds the predecessors of row r.
		offsets     = make([]int32, len(rows)+1)
		deleteIndex = make(map[opset.OpID]int32)
		deleted     [][]int32
	)

	for i := range rows {
		for _, successor := range rows[i].Successors {
			if target, ok := x.row(segments, numbers, successor); ok {
				offsets[target+1]++

				continue
			}

			number, ok := deleteIndex[successor]
			if !ok {
				number = int32(len(deleted))
				deleteIndex[successor] = number
				deleted = append(deleted, nil)
			}

			deleted[number] = append(deleted[number], int32(i))
		}
	}

	for i := range rows {
		offsets[i+1] += offsets[i]
	}

	var (
		predecessorRows = make([]int32, offsets[len(rows)])
		filled          = make([]int32, len(rows))
	)

	for i := range rows {
		for _, successor := range rows[i].Successors {
			if target, ok := x.row(segments, numbers, successor); ok {
				predecessorRows[offsets[target]+filled[target]] = int32(i)
				filled[target]++
			}
		}
	}

	// Every predecessor list is cut from one table, capped so that appending
	// to one can never write into the next.
	identifiers := make([]opset.OpID, 0, successorCount)

	list := func(predecessors []int32) []opset.OpID {
		if len(predecessors) > 1 {
			slices.SortFunc(
				predecessors,
				func(left, right int32) int {
					return rows[left].ID.Compare(rows[right].ID)
				},
			)
		}

		start := len(identifiers)
		for _, row := range predecessors {
			identifiers = append(identifiers, rows[row].ID)
		}

		return identifiers[start:len(identifiers):len(identifiers)]
	}

	for i := range rows {
		rows[i].Predecessors = nil

		if predecessors := predecessorRows[offsets[i]:offsets[i+1]]; len(predecessors) > 0 {
			rows[i].Predecessors = list(predecessors)
		}
	}

	if len(deleted) == 0 {
		releaseDecodedBytes(budget, scratch)

		return true
	}

	// Each delete may allocate the element key it addresses.
	if len(deleted) >= changeOperationDelete ||
		chargeDecoded[opset.Operation](budget, uint64(len(deleted))) != nil ||
		chargeDecoded[opset.OpID](budget, uint64(len(deleted))) != nil {
		return false
	}

	deletes := make([]opset.Operation, len(deleted))

	for identifier, number := range deleteIndex {
		predecessors := list(deleted[number])

		// Only a delete leaves no operation of its own behind, and every
		// operation it removed shares the object and key it targeted.
		source := &rows[deleted[number][0]]

		deletes[number] = opset.Operation{
			ID:           identifier,
			Object:       source.Object,
			Key:          supersededKey(source),
			Action:       opset.ActionDelete,
			Predecessors: predecessors,
		}
	}

	// The map numbered the deletes in the order the rows named them, and
	// yields them in no order at all.
	slices.SortFunc(
		deletes,
		func(left, right opset.Operation) int {
			return left.ID.Compare(right.ID)
		},
	)

	x.deletes = deletes

	releaseDecodedBytes(budget, scratch)

	return true
}

// assignChanges sets every change's first operation counter and locates its
// operations among the keys. It reports false unless each change's operations
// fill its counter range exactly and every operation belongs to a change.
func (x *ChangeOperationIndex) assignChanges(
	changes []opset.Change,
	segments []int,
	numbers *actorNumbers,
	budget *decodeBudget,
) bool {
	dependencyMax, err := changeDependencyMaximums(changes, budget)
	if err != nil {
		return false
	}

	if chargeDecoded[changeOperationSpan](budget, uint64(len(changes))) != nil ||
		chargeDecoded[int32](budget, uint64(len(changes))) != nil ||
		chargeDecoded[int](budget, uint64(len(segments))) != nil {
		return false
	}

	x.spans = make([]changeOperationSpan, len(changes))

	// StartOp is the first counter the change actually wrote. Counters are
	// document-wide, so a change can depend on a frontier whose maximum is
	// below its own first counter; the counters between belong to others.
	for i := range changes {
		change := &changes[i]
		if change.MaxOp == math.MaxUint64 {
			return false
		}

		change.StartOp = change.MaxOp + 1

		if change.MaxOp <= dependencyMax[i] {
			continue
		}

		actor, ok := numbers.number(change.Actor)
		if !ok {
			return false
		}

		segment := x.keys[segments[actor]:segments[actor+1]]
		position := sort.Search(
			len(segment),
			func(cursor int) bool {
				return segment[cursor]>>32 > dependencyMax[i]
			},
		)

		if position < len(segment) && segment[position]>>32 <= change.MaxOp {
			change.StartOp = segment[position] >> 32
		}
	}

	// Group the changes by actor and order each group by counter range, then
	// hand every change the next keys of its actor.
	var (
		order = make([]int32, 0, len(changes))
		next  = slices.Clone(segments[:len(segments)-1])
	)

	for i := range changes {
		if changes[i].MaxOp >= changes[i].StartOp {
			order = append(order, int32(i))
		}
	}

	actorOf := func(change int32) int {
		actor, _ := numbers.number(changes[change].Actor)

		return actor
	}

	slices.SortFunc(
		order,
		func(left, right int32) int {
			switch leftActor, rightActor := actorOf(left), actorOf(right); {
			case leftActor != rightActor:
				return leftActor - rightActor
			case changes[left].StartOp < changes[right].StartOp:
				return -1
			case changes[left].StartOp > changes[right].StartOp:
				return 1
			default:
				return int(left - right)
			}
		},
	)

	for _, number := range order {
		var (
			change = &changes[number]
			actor  = actorOf(number)
			start  = next[actor]
			count  = change.MaxOp - change.StartOp + 1
		)

		if count > uint64(segments[actor+1]-start) {
			return false
		}

		// The keys of an actor hold distinct, ascending counters, so a run of
		// them that starts and ends on the range's bounds is the whole range.
		end := start + int(count)
		if x.keys[start]>>32 != change.StartOp || x.keys[end-1]>>32 != change.MaxOp {
			return false
		}

		x.spans[number] = changeOperationSpan{start: uint32(start), count: uint32(count)}
		next[actor] = end
	}

	for actor := range next {
		if next[actor] != segments[actor+1] {
			return false
		}
	}

	return true
}
