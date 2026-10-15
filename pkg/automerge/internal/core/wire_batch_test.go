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

package core

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
	"go.probo.inc/probo/pkg/automerge/internal/storage"
)

func TestOperationSpliceRuns_MatchesFlatReplacement(t *testing.T) {
	t.Parallel()

	current := make([]opset.Operation, 200)
	next := make([]opset.Operation, 0, 216)

	for i := range current {
		current[i].ID = opset.OpID{
			Actor:   opset.ActorID("existing"),
			Counter: uint64(i + 1),
		}
		if i%13 == 0 {
			next = append(next, opset.Operation{ID: opset.OpID{
				Actor:   opset.ActorID("incoming"),
				Counter: uint64(1_000 + i),
			}})
		}

		operation := current[i]
		if i%17 == 0 {
			operation.Successors = []opset.OpID{{
				Actor:   opset.ActorID("incoming"),
				Counter: uint64(2_000 + i),
			}}
		}

		next = append(next, operation)
	}

	runs, ok := operationSpliceRuns(current, next)
	require.True(t, ok)
	require.NotEmpty(t, runs)

	actual := append([]opset.Operation(nil), current...)
	for _, run := range runs {
		actual = slices.Replace(
			actual,
			run.Index,
			run.Index+run.DeleteCount,
			run.Operations...,
		)
	}

	assert.Equal(t, next, actual)
	assertMinimalSpliceRuns(t, current, runs)
}

func TestOperationSpliceRuns_CurrentOperationAbsent(t *testing.T) {
	t.Parallel()

	current := []opset.Operation{
		{ID: opset.OpID{Actor: "existing", Counter: 1}},
		{ID: opset.OpID{Actor: "existing", Counter: 2}},
		{ID: opset.OpID{Actor: "existing", Counter: 3}},
	}
	next := []opset.Operation{
		current[0],
		current[2],
	}

	runs, ok := operationSpliceRuns(current, next)
	assert.False(t, ok)
	assert.Nil(t, runs)
}

func assertMinimalSpliceRuns(
	t *testing.T,
	current []opset.Operation,
	runs []storage.SnapshotOperationSplice,
) {
	t.Helper()

	for _, run := range runs {
		require.NotEmpty(t, run.Operations)
		require.GreaterOrEqual(t, run.Index, 0)
		require.LessOrEqual(t, run.Index+run.DeleteCount, len(current))
		if run.Index >= len(current) {
			continue
		}

		assert.False(
			t,
			wireOperationsEqual(current[run.Index], run.Operations[0]),
			"run at %d starts on an unchanged operation",
			run.Index,
		)
		if run.DeleteCount == 0 {
			continue
		}

		end := run.Index + run.DeleteCount - 1
		assert.False(
			t,
			wireOperationsEqual(
				current[end],
				run.Operations[len(run.Operations)-1],
			),
			"run at %d ends on an unchanged operation",
			run.Index,
		)
	}
}
