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
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestBuildHexaneDocumentChangeColumns_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(0x484558414e45))
	actors := testHexaneActors()

	for iteration := range 200 {
		count := 1 + rng.Intn(40)
		actorIndexes := randomHexaneActorIndexes(rng, actors)
		changes := randomHexaneChanges(rng, actors, count)

		want, err := encodeDocumentChangeColumns(changes, actorIndexes)
		require.NoError(t, err)

		columnSet, err := newHexaneChangeColumns(changes, actorIndexes)
		require.NoError(t, err)
		got, err := columnSet.Encoded()
		require.NoError(t, err)
		assert.Equal(t, want, got, "iteration %d", iteration)
	}
}

func TestBuildHexaneDocumentOperationColumns_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(0x434f4c554d4e))
	actors := testHexaneActors()

	for iteration := range 200 {
		count := 1 + rng.Intn(80)
		actorIndexes := randomHexaneActorIndexes(rng, actors)
		operations := randomHexaneOperations(rng, actors, count)

		want, err := encodeDocumentOperationColumns(operations, actorIndexes)
		require.NoError(t, err)

		columnSet, err := newHexaneOperationColumns(operations, actorIndexes)
		require.NoError(t, err)
		got, err := columnSet.Encoded()
		require.NoError(t, err)
		assert.Equal(t, want, got, "iteration %d", iteration)
	}
}

func BenchmarkHexaneOperationColumns_LocalizedEdit(b *testing.B) {
	for _, test := range []struct {
		name string
		rows int
	}{
		{name: "1KRows", rows: 1_000},
		{name: "100KRows", rows: 100_000},
	} {
		b.Run(test.name, func(b *testing.B) {
			rng := rand.New(rand.NewSource(42))
			actors := testHexaneActors()
			indexes := randomHexaneActorIndexes(rng, actors)
			base, err := newHexaneOperationColumns(
				randomHexaneOperations(rng, actors, test.rows),
				indexes,
			)
			require.NoError(b, err)
			inserted, err := newHexaneOperationColumns(
				randomHexaneOperations(rng, actors, 1),
				indexes,
			)
			require.NoError(b, err)

			index := test.rows / 2

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				clone := base.Clone()
				if err := clone.Splice(index, 1, inserted); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func testHexaneActors() []opset.ActorID {
	return []opset.ActorID{
		opset.ActorID("\x00"),
		opset.ActorID("actor-a"),
		opset.ActorID("\xff\x10"),
		opset.ActorID("longer-actor-index"),
		opset.ActorID("\x01\x02\x03"),
	}
}

func randomHexaneActorIndexes(
	rng *rand.Rand,
	actors []opset.ActorID,
) map[opset.ActorID]uint64 {
	permutation := rng.Perm(len(actors))

	indexes := make(map[opset.ActorID]uint64, len(actors))
	for i, actor := range actors {
		indexes[actor] = uint64(permutation[i])
	}

	return indexes
}

func randomHexaneChanges(
	rng *rand.Rand,
	actors []opset.ActorID,
	count int,
) []*opset.Change {
	changes := make([]*opset.Change, count)

	hashes := make([]opset.ChangeHash, count)
	for i := range hashes {
		_, _ = rng.Read(hashes[i][:])
	}

	for i := range changes {
		dependencyCount := 0
		if i > 0 {
			dependencyCount = rng.Intn(min(i, 5) + 1)
		}

		dependencies := make([]opset.ChangeHash, dependencyCount)
		for j := range dependencies {
			dependencies[j] = hashes[rng.Intn(i)]
		}

		change := &opset.Change{
			Hash:         new(hashes[i]),
			Actor:        actors[rng.Intn(len(actors))],
			Sequence:     uint64(1 + rng.Intn(10_000)),
			MaxOp:        uint64(rng.Intn(20_000)),
			Time:         int64(rng.Intn(2_000_000)) - 1_000_000,
			Dependencies: dependencies,
		}
		if i%3 != 0 {
			change.Message = randomHexaneString(rng)
		}

		switch i % 4 {
		case 0:
			change.Extra = randomHexaneScalar(rng, i)
		case 1:
			change.ExtraBytes = randomHexaneBytes(rng, 24)
		case 2:
			change.Extra = &opset.Scalar{Type: opset.ScalarNull}
		}

		changes[i] = change
	}

	return changes
}

func randomHexaneOperations(
	rng *rand.Rand,
	actors []opset.ActorID,
	count int,
) []opset.Operation {
	operations := make([]opset.Operation, count)
	for i := range operations {
		idActor := actors[rng.Intn(len(actors))]
		operation := opset.Operation{
			ID: opset.OpID{
				Actor:   idActor,
				Counter: uint64(1 + rng.Intn(50_000)),
			},
			Insert: i%3 == 0,
			Action: opset.Action(rng.Intn(8)),
			Value:  randomHexaneScalar(rng, i),
		}

		if i%4 == 0 {
			operation.Object = opset.RootObject()
		} else {
			operation.Object = opset.ObjectID{
				OpID: opset.OpID{
					Actor:   actors[rng.Intn(len(actors))],
					Counter: uint64(1 + rng.Intn(50_000)),
				},
			}
		}

		switch i % 3 {
		case 0:
			operation.Key.Property = new(randomHexaneString(rng))
		case 1:
			operation.Key.IsHead = true
		case 2:
			operation.Key.Element = &opset.OpID{
				Actor:   actors[rng.Intn(len(actors))],
				Counter: uint64(1 + rng.Intn(50_000)),
			}
		}

		successorCount := rng.Intn(5)

		operation.Successors = make([]opset.OpID, successorCount)
		for j := range operation.Successors {
			operation.Successors[j] = opset.OpID{
				Actor:   actors[rng.Intn(len(actors))],
				Counter: uint64(1 + rng.Intn(50_000)),
			}
		}

		if i%5 == 0 {
			operation.Action = opset.ActionMark
			operation.MarkName = new(randomHexaneString(rng))
			operation.MarkExpand = new(i%10 == 0)
		} else if i%7 == 0 {
			operation.MarkExpand = new(false)
		}

		operations[i] = operation
	}

	return operations
}

func randomHexaneScalar(rng *rand.Rand, index int) *opset.Scalar {
	switch index % 12 {
	case 0:
		return nil
	case 1:
		return &opset.Scalar{Type: opset.ScalarNull}
	case 2:
		return &opset.Scalar{Type: opset.ScalarFalse}
	case 3:
		return &opset.Scalar{Type: opset.ScalarTrue}
	case 4:
		return &opset.Scalar{
			Type: opset.ScalarUint,
			Uint: rng.Uint64(),
		}
	case 5:
		return &opset.Scalar{
			Type: opset.ScalarInt,
			Int:  int64(rng.Uint64()),
		}
	case 6:
		return &opset.Scalar{
			Type:  opset.ScalarFloat64,
			Float: math.Float64frombits(rng.Uint64()),
		}
	case 7:
		return &opset.Scalar{
			Type:   opset.ScalarString,
			String: randomHexaneString(rng),
		}
	case 8:
		return &opset.Scalar{
			Type:  opset.ScalarBytes,
			Bytes: randomHexaneBytes(rng, 32),
		}
	case 9:
		return &opset.Scalar{
			Type: opset.ScalarCounter,
			Int:  int64(rng.Uint64()),
		}
	case 10:
		return &opset.Scalar{
			Type: opset.ScalarTimestamp,
			Int:  int64(rng.Uint64()),
		}
	default:
		return &opset.Scalar{
			Type: opset.ScalarType(15),
			Raw:  randomHexaneBytes(rng, 20),
		}
	}
}

func randomHexaneString(rng *rand.Rand) string {
	values := []string{
		"",
		"property",
		"éclair",
		"文書",
		"emoji-\U0001f9ea",
		"a\x00b",
	}

	return values[rng.Intn(len(values))]
}

func randomHexaneBytes(rng *rand.Rand, maximum int) []byte {
	data := make([]byte, rng.Intn(maximum+1))
	_, _ = rng.Read(data)

	return data
}
