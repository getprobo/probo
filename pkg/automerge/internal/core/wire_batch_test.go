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
}
