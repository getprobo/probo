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
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge/internal/opset"
)

func writeRLE[T comparable](
	values []optional[T],
	writer rleWriter[T],
) []byte {
	for _, value := range values {
		writer.appendOptional(value)
	}

	return writer.finish()
}

func writeDelta(values []optional[int64]) []byte {
	writer := newDeltaWriter()

	for _, value := range values {
		if value.valid {
			writer.append(value.value)
		} else {
			writer.appendNull()
		}
	}

	return writer.finish()
}

func writeBooleans(values []bool) []byte {
	var writer booleanWriter

	for _, value := range values {
		writer.append(value)
	}

	return writer.finish()
}

func TestRLEWriter_MatchesSliceEncoder(t *testing.T) {
	t.Parallel()

	null := optional[uint64]{}
	value := func(v uint64) optional[uint64] { return some(v) }

	shapes := map[string][]optional[uint64]{
		"empty":                    {},
		"single null":              {null},
		"all null":                 {null, null, null},
		"single value":             {value(7)},
		"lone value then null":     {value(7), null},
		"null then lone value":     {null, null, value(7)},
		"run":                      {value(7), value(7), value(7)},
		"run then null":            {value(7), value(7), null},
		"run then lone value":      {value(7), value(7), value(8)},
		"run then run":             {value(7), value(7), value(8), value(8)},
		"literal":                  {value(1), value(2), value(3)},
		"literal then null":        {value(1), value(2), null, value(3)},
		"literal then run":         {value(1), value(2), value(3), value(3)},
		"literal run literal":      {value(1), value(2), value(2), value(3), value(4)},
		"lone value between runs":  {value(1), value(1), value(2), value(3), value(3)},
		"lone value between nulls": {null, value(1), null, value(1), null},
		"zero value run":           {value(0), value(0), null, value(0)},
	}

	for name, values := range shapes {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(
					t,
					encodeRLE(values, appendULEB),
					writeRLE(values, newRLEWriter(appendULEB)),
				)
			},
		)
	}
}

func TestRLEWriter_LongLiteralRun(t *testing.T) {
	t.Parallel()

	// A literal run's length is written into a one-byte slot once the run
	// ends. These lengths sit on either side of each point where the length
	// needs another byte and the run's values have to move up.
	for _, length := range []int{63, 64, 65, 66, 8191, 8192, 8193, 8194, 1 << 20} {
		values := make([]optional[uint64], 0, length+4)
		values = append(values, some(uint64(9)), some(uint64(9)))

		for i := range length {
			values = append(values, some(uint64(i%300)))
		}

		values = append(values, optional[uint64]{}, some(uint64(1)))

		require.Equal(
			t,
			encodeRLE(values, appendULEB),
			writeRLE(values, newRLEWriter(appendULEB)),
			"literal run of %d",
			length,
		)
	}
}

func TestColumnWriters_MatchSliceEncodersOnRandomColumns(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(53, 59))

	for range 2000 {
		length := random.IntN(80)

		unsigned := randomOptionalColumn(random, length, func() uint64 {
			return uint64(random.IntN(4))
		})
		require.Equal(
			t,
			encodeRLE(unsigned, appendULEB),
			writeRLE(unsigned, newRLEWriter(appendULEB)),
			"unsigned %v",
			unsigned,
		)

		strings := randomOptionalColumn(random, length, func() string {
			return string(rune('a' + random.IntN(3)))
		})
		require.Equal(
			t,
			encodeStrings(strings),
			writeRLE(strings, newStringWriter()),
			"strings %v",
			strings,
		)

		signed := randomOptionalColumn(random, length, func() int64 {
			return int64(random.IntN(2000)) - 1000
		})
		require.Equal(t, encodeDelta(signed), writeDelta(signed), "delta %v", signed)

		booleans := make([]bool, length)
		for i := range booleans {
			booleans[i] = random.IntN(3) == 0
		}

		require.Equal(
			t,
			encodeBooleans(booleans),
			writeBooleans(booleans),
			"booleans %v",
			booleans,
		)
	}
}

func TestAppendScalar_AppendsPayloadAfterExistingData(t *testing.T) {
	t.Parallel()

	values := []*opset.Scalar{
		nil,
		{Type: opset.ScalarNull},
		{Type: opset.ScalarTrue},
		{Type: opset.ScalarUint, Uint: 300},
		{Type: opset.ScalarInt, Int: -300},
		{Type: opset.ScalarCounter, Int: 12},
		{Type: opset.ScalarTimestamp, Int: 1 << 40},
		{Type: opset.ScalarFloat64, Float: 1.5},
		{Type: opset.ScalarString, String: "héllo"},
		{Type: opset.ScalarString},
		{Type: opset.ScalarBytes, Bytes: []byte{0, 1, 2}},
		{Type: 12, Raw: []byte{9, 9}},
	}

	prefix := []byte("prefix")
	data := append([]byte(nil), prefix...)

	for _, value := range values {
		wantMeta, wantData, err := encodeScalar(value)
		require.NoError(t, err)

		start := len(data)

		var meta optional[uint64]

		data, meta, err = appendScalar(data, value)
		require.NoError(t, err)

		assert.Equal(t, wantMeta, meta)
		assert.Equal(t, len(wantData), len(data)-start)
		assert.Equal(t, string(wantData), string(data[start:]))
	}

	assert.Equal(t, prefix, data[:len(prefix)])

	_, _, err := appendScalar(nil, &opset.Scalar{Type: 0x10})
	require.ErrorContains(t, err, "scalar type 16 exceeds 0x0f")

	_, _, err = appendScalar(nil, &opset.Scalar{Type: opset.ScalarString, String: "\xff"})
	require.ErrorContains(t, err, "string scalar is not valid UTF-8")
}
