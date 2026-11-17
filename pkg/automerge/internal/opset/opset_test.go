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

package opset_test

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func TestActorID_ZeroValue(t *testing.T) {
	t.Parallel()

	var zero opset.ActorID

	assert.True(t, zero.IsZero())
	assert.Equal(t, 0, zero.Len())
	assert.Empty(t, zero.Raw())
	assert.Empty(t, zero.Bytes())
	assert.Empty(t, zero.String())
	assert.Equal(t, zero, opset.MakeActorID(""))

	_, err := opset.NewActorID(nil)
	require.Error(t, err)
}

func TestActorID_Identity(t *testing.T) {
	t.Parallel()

	raw := []byte{0x00, 0xff, 0x10, 0x7f}

	first, err := opset.NewActorID(raw)
	require.NoError(t, err)

	// The identifier must not alias the caller's buffer.
	second := opset.MakeActorID(string(raw))
	raw[0] = 0x01

	assert.Equal(t, first, second)
	assert.True(t, first == second)
	assert.False(t, first.IsZero())
	assert.Equal(t, 4, first.Len())
	assert.Equal(t, "\x00\xff\x10\x7f", first.Raw())
	assert.Equal(t, []byte{0x00, 0xff, 0x10, 0x7f}, first.Bytes())
	assert.Equal(t, "00ff107f", first.String())
	assert.NotEqual(t, first, opset.MakeActorID("\x00\xff\x10"))

	// An operation identifier is a map key across the engine; two identifiers
	// built separately from the same bytes must be the same key.
	index := map[opset.OpID]int{{Actor: first, Counter: 7}: 1}
	assert.Equal(t, 1, index[opset.OpID{Actor: second, Counter: 7}])
	assert.Equal(
		t,
		"{00ff107f 7}",
		fmt.Sprintf("%v", opset.OpID{Actor: first, Counter: 7}),
	)
}

func TestActorID_Compare(t *testing.T) {
	t.Parallel()

	raws := []string{
		"",
		"\x00",
		"\x00\x00",
		"\x01",
		"\x7f\xff",
		"\x80",
		"\xff",
		"\xff\x00",
		"ab",
		"abc",
		"b",
	}

	for _, left := range raws {
		for _, right := range raws {
			assert.Equal(
				t,
				bytes.Compare([]byte(left), []byte(right)),
				opset.MakeActorID(left).Compare(opset.MakeActorID(right)),
				"%q against %q",
				left,
				right,
			)
		}
	}
}

func TestOpID_Compare(t *testing.T) {
	t.Parallel()

	// Actors are interned in an order unrelated to their byte order, so the
	// comparison must come from the actor bytes and never from the handles.
	high := opset.MakeActorID("\xfe-compare-high")
	low := opset.MakeActorID("\x01-compare-low")

	identifiers := []opset.OpID{
		{Actor: high, Counter: 2},
		{Actor: low, Counter: 3},
		{Actor: high, Counter: 1},
		{Actor: low, Counter: 2},
		{Actor: low, Counter: 1},
	}

	slices.SortFunc(identifiers, opset.OpID.Compare)

	assert.Equal(
		t,
		[]opset.OpID{
			{Actor: low, Counter: 1},
			{Actor: high, Counter: 1},
			{Actor: low, Counter: 2},
			{Actor: high, Counter: 2},
			{Actor: low, Counter: 3},
		},
		identifiers,
	)
}

func TestActorID_Concurrency(t *testing.T) {
	t.Parallel()

	base := opset.MakeActorID("\x01\x02\x03\x04\x05\x06\x07\x08")

	assert.Equal(t, base, base.WithConcurrency(0))

	for _, level := range []uint64{1, 127, 128, 1 << 40} {
		derived := base.WithConcurrency(level)
		assert.NotEqual(t, base, derived)

		recovered, recoveredLevel := derived.Concurrency()
		assert.Equal(t, base, recovered)
		assert.Equal(t, level, recoveredLevel)
	}

	recovered, level := base.Concurrency()
	assert.Equal(t, base, recovered)
	assert.Zero(t, level)
}
