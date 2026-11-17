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
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/testsupport/reference"
)

// collectChangeOperations reads the operations the index locates for a change.
func collectChangeOperations(index *ChangeOperationIndex, change int) []opset.Operation {
	operations := make([]opset.Operation, 0, index.Len(change))
	for _, operation := range index.Operations(change) {
		operations = append(operations, operation)
	}

	return operations
}

// assertEngineDecodeMatches requires a snapshot decoded for an engine to
// describe exactly the history Decode does, with each change's operations
// located among the rows instead of copied into the change.
func assertEngineDecodeMatches(t *testing.T, saved []byte) {
	t.Helper()

	full, err := Decode(saved)
	require.NoError(t, err)

	deferred, err := DecodeForEngine(saved)
	require.NoError(t, err)

	fullDecoded, ok := full.Canonical.(*DecodedDocument)
	require.True(t, ok)
	require.Nil(t, fullDecoded.ChangeOperations)

	decoded, ok := deferred.Canonical.(*DecodedDocument)
	require.True(t, ok)
	require.NotNil(t, decoded.ChangeOperations, "snapshot took the general path")

	assert.Equal(t, full.Actors, deferred.Actors)
	assert.Equal(t, full.Heads, deferred.Heads)
	assert.Equal(t, fullDecoded.Operations, decoded.Operations)
	require.Len(t, deferred.Changes, len(full.Changes))

	var deletes []opset.Operation

	for i := range full.Changes {
		want := full.Changes[i]
		got := deferred.Changes[i]

		assert.True(t, got.OperationsDeferred, "change %d", i)
		assert.Empty(t, got.Operations, "change %d", i)

		operations := collectChangeOperations(decoded.ChangeOperations, i)
		if len(want.Operations) == 0 {
			assert.Empty(t, operations, "change %d", i)
		} else {
			assert.Equal(t, want.Operations, operations, "change %d", i)
		}

		for operation := range decoded.ChangeOperations.ChangeDeletes(i) {
			deletes = append(deletes, operation)
		}

		got.Operations = want.Operations
		got.OperationsDeferred = false
		assert.Equal(t, want, got, "change %d", i)
	}

	slices.SortFunc(
		deletes,
		func(left, right opset.Operation) int {
			return left.ID.Compare(right.ID)
		},
	)

	if len(deletes) == 0 {
		assert.Empty(t, decoded.ChangeOperations.Deletes())
	} else {
		assert.Equal(t, deletes, decoded.ChangeOperations.Deletes())
	}
}

func referenceReplica(t *testing.T, shared []byte, actor byte) *reference.Engine {
	t.Helper()

	engine, err := reference.Load(shared)
	require.NoError(t, err)
	require.NoError(
		t,
		engine.SetActor(
			[]byte{
				actor, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
			},
		),
	)

	return engine
}

func TestDecodeForEngine_MatchesDecodeOnReferenceHistories(t *testing.T) {
	t.Parallel()

	for _, seed := range []uint64{1, 2, 3} {
		t.Run(
			fmt.Sprintf("seed %d", seed),
			func(t *testing.T) {
				t.Parallel()

				random := rand.New(rand.NewPCG(seed, 97))
				clock := int64(0)

				commit := func(engine *reference.Engine) {
					clock++

					_, err := engine.Commit("edit", time.Unix(clock, 0))
					require.NoError(t, err)
				}

				// The first actor sorts between the two that join later, so
				// actor order and join order disagree.
				first := newReference(t, 0x20)

				_, err := first.PutText(0, "body")
				require.NoError(t, err)
				require.NoError(t, first.PutScalar(0, "count", counterScalar()))
				require.NoError(t, first.PutString(0, "title", "start"))
				commit(first)

				shared, err := first.Save(true, true)
				require.NoError(t, err)

				assertEngineDecodeMatches(t, shared)

				replicas := []*reference.Engine{
					first,
					referenceReplica(t, shared, 0x10),
					referenceReplica(t, shared, 0x30),
				}

				edit := func(engine *reference.Engine) {
					body, _, err := engine.GetObject(0, "body")
					require.NoError(t, err)

					length, err := engine.Length(body)
					require.NoError(t, err)

					switch choice := random.IntN(7); {
					case choice < 2 || length < 4:
						require.NoError(
							t,
							engine.SpliceText(
								body,
								uint32(random.IntN(int(length)+1)),
								0,
								[]string{"a", "bc", "def ", "\n"}[random.IntN(4)],
							),
						)
					case choice == 2:
						start := random.IntN(int(length) - 2)
						require.NoError(
							t,
							engine.SpliceText(body, uint32(start), int32(1+random.IntN(2)), ""),
						)
					case choice == 3:
						start := random.IntN(int(length) - 2)
						require.NoError(
							t,
							engine.MarkText(
								body,
								uint32(start),
								uint32(start+2),
								"strong",
								[][]byte{boolScalar(), nullScalar()}[random.IntN(2)],
								"both",
							),
						)
					case choice == 4:
						require.NoError(t, engine.Increment(0, "count", int64(1+random.IntN(3))))
					case choice == 5:
						require.NoError(
							t,
							engine.PutString(0, "title", fmt.Sprintf("title %d", random.IntN(100))),
						)
					default:
						key := fmt.Sprintf("scratch-%d", random.IntN(3))

						require.NoError(t, engine.PutString(0, key, "temporary"))
						commit(engine)
						require.NoError(t, engine.DeleteMap(0, key))
					}

					commit(engine)
				}

				for round := range 6 {
					for _, replica := range replicas {
						for range 1 + random.IntN(4) {
							edit(replica)
						}
					}

					// Each round one replica catches up with another, so the
					// histories stay concurrent and keep distinct frontiers.
					source := replicas[(round+1)%len(replicas)]
					target := replicas[round%len(replicas)]

					saved, err := source.Save(true, true)
					require.NoError(t, err)

					_, err = target.Merge(saved)
					require.NoError(t, err)

					for _, replica := range replicas {
						saved, err := replica.Save(true, round%2 == 0)
						require.NoError(t, err)

						assertEngineDecodeMatches(t, saved)
					}
				}
			},
		)
	}
}

func TestDecodeForEngine_KeepsOperationsWhenChangesFollowTheSnapshot(t *testing.T) {
	t.Parallel()

	engine := newReference(t, 0x20)

	require.NoError(t, engine.PutString(0, "title", "one"))

	_, err := engine.Commit("one", time.Unix(1, 0))
	require.NoError(t, err)

	saved, err := engine.Save(true, true)
	require.NoError(t, err)

	require.NoError(t, engine.PutString(0, "title", "two"))

	_, err = engine.Commit("two", time.Unix(2, 0))
	require.NoError(t, err)

	trailing, err := engine.SaveIncremental()
	require.NoError(t, err)
	require.NotEmpty(t, trailing)

	data := append(slices.Clone(saved), trailing...)

	want, err := Decode(data)
	require.NoError(t, err)

	got, err := DecodeForEngine(data)
	require.NoError(t, err)

	// A change chunk after the snapshot sends the stream down the replay
	// path, which reads every change's operations.
	require.Len(t, got.Changes, 2)
	assert.Nil(t, got.Canonical)

	for i := range got.Changes {
		assert.False(t, got.Changes[i].OperationsDeferred, "change %d", i)
		assert.NotEmpty(t, got.Changes[i].Operations, "change %d", i)
	}

	assert.Equal(t, want.Changes, got.Changes)
}

// snapshotTables builds the decoded tables of a linear history: one change per
// entry of sizes, written by the actors in turn. Every operation sets one of a
// few root keys and supersedes the operations visible there, and about one in
// four is a delete, which the snapshot keeps only as a successor entry.
func snapshotTables(
	random *rand.Rand,
	sizes []int,
) ([]opset.Change, []opset.ActorID, []opset.Operation) {
	actors := []opset.ActorID{
		opset.MakeActorID("\x01actor"),
		opset.MakeActorID("\x02actor"),
		opset.MakeActorID("\x03actor"),
	}

	var (
		changes   = make([]opset.Change, 0, len(sizes))
		rows      []opset.Operation
		rowByID   = make(map[opset.OpID]int)
		visible   = make(map[string][]opset.OpID)
		sequences = make(map[opset.ActorID]uint64)
		counter   uint64
	)

	for i, size := range sizes {
		actor := actors[i%len(actors)]
		sequences[actor]++

		change := opset.Change{
			Actor:    actor,
			Sequence: sequences[actor],
			MaxOp:    counter + uint64(size),
		}
		if i > 0 {
			change.DependencyIndexes = []uint64{uint64(i - 1)}
		}

		for range size {
			counter++

			var (
				identifier = opset.OpID{Actor: actor, Counter: counter}
				property   = fmt.Sprintf("key-%d", random.IntN(4))
			)

			for _, superseded := range visible[property] {
				row := rowByID[superseded]
				rows[row].Successors = append(rows[row].Successors, identifier)
			}

			if len(visible[property]) > 0 && random.IntN(4) == 0 {
				delete(visible, property)

				continue
			}

			visible[property] = []opset.OpID{identifier}
			rowByID[identifier] = len(rows)
			rows = append(
				rows,
				opset.Operation{
					ID:     identifier,
					Object: opset.RootObject(),
					Key:    opset.Key{Property: &property},
					Action: opset.ActionSet,
					Value:  &opset.Scalar{Type: opset.ScalarUint, Uint: counter},
				},
			)
		}

		changes = append(changes, change)
	}

	// A snapshot stores rows in document order, which is unrelated to the
	// order the changes wrote them in.
	random.Shuffle(len(rows), func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })

	return changes, actors, rows
}

func cloneSnapshotTables(
	changes []opset.Change,
	rows []opset.Operation,
) ([]opset.Change, []opset.Operation) {
	clonedRows := slices.Clone(rows)
	for i := range clonedRows {
		clonedRows[i].Successors = slices.Clone(rows[i].Successors)
	}

	return slices.Clone(changes), clonedRows
}

func TestIndexChangeOperations_MatchesGeneralPath(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(61, 67))

	for range 200 {
		sizes := make([]int, 1+random.IntN(12))
		for i := range sizes {
			// An empty change is legal and owns no counter.
			sizes[i] = random.IntN(6)
		}

		changes, actors, rows := snapshotTables(random, sizes)

		wantChanges, wantRows := cloneSnapshotTables(changes, rows)

		restored, err := restoreChangeOperations(wantRows, &decodeBudget{})
		require.NoError(t, err)
		require.NoError(t, assignOperations(wantChanges, restored, &decodeBudget{}))

		gotChanges, gotRows := cloneSnapshotTables(changes, rows)
		budget := &decodeBudget{}

		index, ok := indexChangeOperations(gotChanges, actors, gotRows, budget)
		require.True(t, ok)
		require.NotZero(t, budget.used)

		assert.Equal(t, wantRows, gotRows)

		for i := range wantChanges {
			assert.Equal(t, wantChanges[i].StartOp, gotChanges[i].StartOp, "change %d", i)
			assert.Equal(t, len(wantChanges[i].Operations), index.Len(i), "change %d", i)

			if len(wantChanges[i].Operations) > 0 {
				assert.Equal(
					t,
					wantChanges[i].Operations,
					collectChangeOperations(index, i),
					"change %d",
					i,
				)
			}

			var wantDeletes, gotDeletes []opset.Operation

			for _, operation := range wantChanges[i].Operations {
				if operation.Action == opset.ActionDelete {
					wantDeletes = append(wantDeletes, operation)
				}
			}

			for operation := range index.ChangeDeletes(i) {
				gotDeletes = append(gotDeletes, operation)
			}

			assert.Equal(t, wantDeletes, gotDeletes, "change %d", i)
		}
	}
}

func TestIndexChangeOperations_LeavesIrregularSnapshotsToTheGeneralPath(t *testing.T) {
	t.Parallel()

	build := func() ([]opset.Change, []opset.ActorID, []opset.Operation) {
		return snapshotTables(rand.New(rand.NewPCG(71, 73)), []int{4, 3, 5, 2})
	}

	for _, testCase := range []struct {
		name    string
		corrupt func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation)
		budget  uint64
	}{
		{
			name: "two rows share an identifier",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				return changes, append(rows, rows[0])
			},
		},
		{
			name: "a change is missing one of its operations",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				return changes, rows[1:]
			},
		},
		{
			name: "an operation belongs to no change",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				extra := rows[0]
				extra.ID.Counter = 1000
				extra.Successors = nil

				return changes, append(rows, extra)
			},
		},
		{
			name: "an operation names an actor outside the table",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				rows[0].ID.Actor = opset.MakeActorID("\x09stranger")

				return changes, rows
			},
		},
		{
			name: "a counter exceeds the snapshot domain",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				rows[0].ID.Counter = math.MaxUint32 + 1

				return changes, rows
			},
		},
		{
			name: "a change's maximum counter cannot be followed",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				changes[len(changes)-1].MaxOp = math.MaxUint64

				return changes, rows
			},
		},
		{
			name: "a dependency lies outside the table",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				changes[1].DependencyIndexes = []uint64{99}

				return changes, rows
			},
		},
		{
			name: "the budget is nearly spent",
			corrupt: func(changes []opset.Change, rows []opset.Operation) ([]opset.Change, []opset.Operation) {
				return changes, rows
			},
			budget: maxDecodedMemoryBytes - 64,
		},
	} {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				changes, actors, rows := build()
				changes, rows = testCase.corrupt(changes, rows)
				budget := &decodeBudget{used: testCase.budget}

				index, ok := indexChangeOperations(changes, actors, rows, budget)

				assert.False(t, ok)
				assert.Nil(t, index)
				// Whatever the attempt charged is handed back, so the general
				// path starts from the budget it would have had anyway.
				assert.Equal(t, testCase.budget, budget.used)
			},
		)
	}
}

func TestIndexChangeOperations_ReleasesWorkingMemory(t *testing.T) {
	t.Parallel()

	changes, actors, rows := snapshotTables(rand.New(rand.NewPCG(79, 83)), []int{40, 40, 40})
	budget := &decodeBudget{}

	index, ok := indexChangeOperations(changes, actors, rows, budget)
	require.True(t, ok)
	require.NotEmpty(t, index.Deletes())

	var successors uint64
	for i := range rows {
		successors += uint64(len(rows[i].Successors))
	}

	// What stays charged is what the index and the rows keep: one key per
	// operation, one identifier per successor entry, the deletes, and the
	// per-change and per-actor tables. The tables used to invert the
	// successor lists, and the first set of keys, are released.
	operations := uint64(len(rows) + len(index.Deletes()))
	retained := operations*8 +
		successors*16 +
		uint64(len(index.Deletes()))*(152+16) +
		uint64(len(changes))*(8+4)

	assert.Greater(t, budget.used, retained)
	assert.Less(t, budget.used, retained+4096)
}
