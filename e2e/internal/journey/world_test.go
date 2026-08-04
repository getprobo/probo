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

package journey

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorld_StepRecordsReadableDiagnostics(t *testing.T) {
	t.Parallel()

	world := &World{
		t:         t,
		startedAt: time.Now(),
	}

	world.Step(
		"Alice",
		"creates a document",
		func() error {
			return nil
		},
	)
	world.Step(
		"Bob",
		"approves the document",
		func() error {
			return nil
		},
	)

	require.Len(t, world.steps, 2)
	assert.Equal(t, 1, world.steps[0].Number)
	assert.Equal(t, "Alice", world.steps[0].Actor)
	assert.Equal(t, "creates a document", world.steps[0].Name)
	assert.Empty(t, world.steps[0].Failure)
	assert.Positive(t, world.steps[0].Duration)
	assert.Equal(t, 2, world.steps[1].Number)
	assert.Equal(t, "Bob approves the document", describeStep(
		world.steps[1].Actor,
		world.steps[1].Name,
	))
}
