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

package automerge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestChangeHistory_CausalOrderAndActors(t *testing.T) {
	t.Parallel()

	actorA, err := automerge.NewActorID()
	require.NoError(t, err)
	actorB, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorA)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	text, err := document.CreateText("body")
	require.NoError(t, err)

	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	require.NoError(t, text.Splice(0, 0, "hello"))
	_, err = document.Commit("a1", base)
	require.NoError(t, err)

	fork, err := document.Fork(actorB)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fork.Close() })

	forkText, err := fork.Text("body")
	require.NoError(t, err)
	require.NoError(t, forkText.Splice(5, 0, " world"))
	_, err = fork.Commit("b1", base.Add(time.Minute))
	require.NoError(t, err)

	_, err = document.Merge(fork)
	require.NoError(t, err)

	history, err := document.ChangeHistory()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(history), 2)

	assert.Equal(t, actorA[:], history[0].Actor)
	assert.Equal(t, base, history[0].Time)

	foundB := false
	for _, entry := range history[1:] {
		if string(entry.Actor) == string(actorB[:]) {
			foundB = true
			assert.Equal(t, base.Add(time.Minute), entry.Time)
			break
		}
	}
	assert.True(t, foundB, "expected actor B change in history")
}
