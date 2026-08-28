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

func TestLazyLoad_LoadSaveAndForkDoNotHydrateQueries(t *testing.T) {
	data := lazyLoadFixture(t)
	ResetRuntimeMetrics()

	engine, err := LoadEngine(data)
	require.NoError(t, err)
	saved, err := engine.Save(true, false)
	require.NoError(t, err)
	fork, err := engine.Fork([]byte("lazy-fork"))
	require.NoError(t, err)

	assert.Equal(t, data, saved)
	assert.NotNil(t, fork)
	metrics := ReadRuntimeMetrics()
	assert.Zero(t, metrics.SemanticChangeRows)
	assert.Zero(t, metrics.SemanticOperationRows)
}

func lazyLoadFixture(t *testing.T) []byte {
	t.Helper()

	engine, err := NewEngine()
	require.NoError(t, err)
	require.NoError(t, engine.SetActor([]byte("lazy-load")))
	value, err := encodeScalarWire(
		opset.Scalar{Type: opset.ScalarString, String: "value"},
	)
	require.NoError(t, err)
	require.NoError(t, engine.PutScalar(0, "value", value))
	text, err := engine.PutText(0, "body")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "body"))
	_, err = engine.Commit("fixture", time.Time{})
	require.NoError(t, err)
	data, err := engine.Save(true, false)
	require.NoError(t, err)

	return data
}
