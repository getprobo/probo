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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestEngineRollback_PreservesSurvivingHandles(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)

	stable, err := engine.PutObject(0, "stable", "map")
	require.NoError(t, err)
	_, err = engine.Commit("stable", time.Time{})
	require.NoError(t, err)

	transient, err := engine.PutObject(0, "transient", "map")
	require.NoError(t, err)

	cancelled, err := engine.Rollback()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cancelled)

	value, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarString, String: "yes"})
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(stable, "after", value))

	_, err = engine.PutObject(transient, "invalid", "map")
	require.Error(t, err)

	replacement, err := engine.PutObject(0, "replacement", "map")
	require.NoError(t, err)
	assert.Greater(t, replacement, transient)
}

func TestEngineRollback_PreservesIsolationStates(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)

	value, err := encodeScalarWire(opset.Scalar{Type: opset.ScalarString, String: "seed"})
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", value))
	hash, err := engine.Commit("seed", time.Time{})
	require.NoError(t, err)

	require.NoError(t, engine.Isolate([][32]byte{hash}))
	pinned := engine.state
	full := engine.fullState

	pending, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "pending"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", pending))

	cancelled, err := engine.Rollback()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cancelled)
	assert.Same(t, pinned, engine.state)
	assert.Same(t, full, engine.fullState)
	assert.True(t, engine.isolationActive)
}
