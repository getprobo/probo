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
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

// reconstructionChain builds a linear history of single-operation changes and
// lays it out in the given order, so position[i] is the row holding the i-th
// change of the chain.
func reconstructionChain(t *testing.T, position []int) []opset.Change {
	t.Helper()

	actor, err := opset.NewActorID([]byte{7})
	require.NoError(t, err)

	changes := make([]opset.Change, len(position))

	for link, row := range position {
		property := fmt.Sprintf("key-%d", link)
		counter := uint64(link + 1)

		changes[row] = opset.Change{
			Actor:    actor,
			Sequence: counter,
			StartOp:  counter,
			MaxOp:    counter,
			Operations: []opset.Operation{{
				ID:     opset.OpID{Actor: actor, Counter: counter},
				Object: opset.RootObject(),
				Key:    opset.Key{Property: &property},
				Action: opset.ActionSet,
				Value:  &opset.Scalar{Type: opset.ScalarString, String: property},
			}},
		}
		if link > 0 {
			changes[row].DependencyIndexes = []uint64{uint64(position[link-1])}
		}
	}

	return changes
}

func TestReconstructSnapshotChanges(t *testing.T) {
	t.Parallel()

	t.Run(
		"rebuilds the same hashes whatever order the rows are stored in",
		func(t *testing.T) {
			t.Parallel()

			forward := reconstructionChain(t, []int{0, 1, 2, 3})
			require.NoError(t, reconstructSnapshotChanges(forward, nil, &decodeBudget{}))

			for _, position := range [][]int{{3, 2, 1, 0}, {2, 0, 3, 1}} {
				shuffled := reconstructionChain(t, position)
				require.NoError(t, reconstructSnapshotChanges(shuffled, nil, &decodeBudget{}))

				for link, row := range position {
					require.NotNil(t, shuffled[row].Hash)
					assert.Equal(t, *forward[link].Hash, *shuffled[row].Hash)
					assert.Equal(t, forward[link].Raw, shuffled[row].Raw)
				}
			}
		},
	)

	t.Run(
		"rejects a recorded hash the rebuilt change does not match",
		func(t *testing.T) {
			t.Parallel()

			changes := reconstructionChain(t, []int{0, 1})
			changes[1].Hash = &opset.ChangeHash{1}

			err := reconstructSnapshotChanges(changes, nil, &decodeBudget{})
			require.Error(t, err)
			assert.ErrorContains(t, err, "rebuilds to hash")
		},
	)

	t.Run(
		"rejects a dependency cycle",
		func(t *testing.T) {
			t.Parallel()

			changes := reconstructionChain(t, []int{0, 1})
			changes[0].DependencyIndexes = []uint64{1}

			err := reconstructSnapshotChanges(changes, nil, &decodeBudget{})
			require.Error(t, err)
			assert.ErrorContains(t, err, "cannot be ordered")
		},
	)

	t.Run(
		"rejects an out of range dependency",
		func(t *testing.T) {
			t.Parallel()

			changes := reconstructionChain(t, []int{0, 1})
			changes[1].DependencyIndexes = []uint64{9}

			err := reconstructSnapshotChanges(changes, nil, &decodeBudget{})
			require.Error(t, err)
			assert.ErrorContains(t, err, "out of bounds")
		},
	)

	t.Run(
		"keeps only the rebuilt changes charged",
		func(t *testing.T) {
			t.Parallel()

			changes := reconstructionChain(t, []int{0, 1, 2, 3})
			budget := &decodeBudget{}
			require.NoError(t, reconstructSnapshotChanges(changes, nil, budget))

			const dependencies = 3

			want := uint64(dependencies) * uint64(unsafe.Sizeof(opset.ChangeHash{}))
			want += uint64(4*len(changes)+1+dependencies) * uint64(unsafe.Sizeof(int(0)))

			for i := range changes {
				want += uint64(len(changes[i].Raw))
			}

			// The encoder reservation is several kilobytes per change; none
			// of it may outlive the change it was taken for.
			assert.Equal(t, want, budget.used)
		},
	)
}

func TestValidateSnapshotMarkOrder_RootOnly(t *testing.T) {
	t.Parallel()

	actor, err := opset.NewActorID([]byte{1})
	require.NoError(t, err)

	t.Run(
		"leaves sequence objects to the caller",
		func(t *testing.T) {
			t.Parallel()

			require.Error(t, validateSnapshotMarkOrder(markEndBeforeBegin(actor), false))
			require.NoError(t, validateSnapshotMarkOrder(markEndBeforeBegin(actor), true))
		},
	)

	t.Run(
		"still checks marks recorded against the root",
		func(t *testing.T) {
			t.Parallel()

			operations := markEndBeforeBegin(actor)
			for i := range operations {
				operations[i].Object = opset.RootObject()
			}

			err := validateSnapshotMarkOrder(operations, true)
			require.Error(t, err)
			assert.ErrorContains(t, err, "invalid mark operation order")
		},
	)
}
