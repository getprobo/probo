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

package hexane_test

import (
	"bytes"
	"math/rand/v2"
	"reflect"
	"testing"

	"go.probo.inc/probo/pkg/automerge/internal/hexane"
)

func TestColumn_BoundaryEditsAndWireEncoding(t *testing.T) {
	t.Parallel()

	column := hexane.NewColumn(hexane.Uint64Codec())
	column.Insert(0, hexane.Some(uint64(2)), hexane.Some(uint64(2)))
	column.Insert(0, hexane.Some(uint64(1)))
	column.Insert(column.Len(), hexane.Null[uint64](), hexane.Null[uint64]())
	column.Set(3, hexane.Some(uint64(3)))
	column.Splice(2, 1, hexane.Some(uint64(2)), hexane.Some(uint64(2)))
	column.Delete(column.Len()-1, 1)

	wantValues := []hexane.Value[uint64]{
		hexane.Some(uint64(1)),
		hexane.Some(uint64(2)),
		hexane.Some(uint64(2)),
		hexane.Some(uint64(2)),
		hexane.Some(uint64(3)),
	}
	if got := column.Values(); !reflect.DeepEqual(got, wantValues) {
		t.Fatalf("Values() = %#v, want %#v", got, wantValues)
	}

	data, err := column.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}

	// One literal 1, a repeated run of three 2s, and one literal 3.
	wantWire := []byte{0x7f, 0x01, 0x03, 0x02, 0x7f, 0x03}
	if !bytes.Equal(data, wantWire) {
		t.Errorf("Bytes() = %x, want %x", data, wantWire)
	}

	var saved bytes.Buffer

	written, err := column.SaveTo(&saved)
	if err != nil {
		t.Fatalf("SaveTo() error = %v", err)
	}

	if written != int64(len(wantWire)) {
		t.Errorf("SaveTo() wrote %d, want %d", written, len(wantWire))
	}

	if !bytes.Equal(saved.Bytes(), wantWire) {
		t.Errorf("SaveTo() = %x, want %x", saved.Bytes(), wantWire)
	}
}

func TestColumn_NullsStringsAndOwnedBytes(t *testing.T) {
	t.Parallel()

	strings := hexane.NewColumn(hexane.StringCodec())
	strings.Insert(
		0,
		hexane.Some("a"),
		hexane.Null[string](),
		hexane.Null[string](),
		hexane.Some("bc"),
	)

	data, err := strings.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}

	wantWire := []byte{0x7f, 0x01, 'a', 0x00, 0x02, 0x7f, 0x02, 'b', 'c'}
	if !bytes.Equal(data, wantWire) {
		t.Errorf("Bytes() = %x, want %x", data, wantWire)
	}

	invalid := hexane.NewColumn(hexane.StringCodec())
	invalid.Insert(0, hexane.Some(string([]byte{0xff})))

	if _, err := invalid.Bytes(); err == nil {
		t.Error("invalid UTF-8 string encoded without error")
	}

	allNull := hexane.NewColumn(hexane.Uint64Codec())
	allNull.Insert(0, hexane.Null[uint64](), hexane.Null[uint64]())

	nullWire, err := allNull.Bytes()
	if err != nil {
		t.Fatalf("all-null Bytes() error = %v", err)
	}

	if nullWire != nil {
		t.Errorf("all-null Bytes() = %x, want nil", nullWire)
	}

	input := []byte{1, 2}
	byteColumn := hexane.NewColumn(hexane.BytesCodec())
	byteColumn.Insert(0, hexane.Some(input))
	input[0] = 9
	got := byteColumn.Get(0)
	got.Value[1] = 8

	if value := byteColumn.Get(0).Value; !bytes.Equal(value, []byte{1, 2}) {
		t.Errorf("owned value = %v, want [1 2]", value)
	}
}

func TestColumn_CloneIsolation(t *testing.T) {
	t.Parallel()

	column := hexane.NewColumn(hexane.BytesCodec())
	column.Insert(0, hexane.Some([]byte("a")), hexane.Some([]byte("b")))
	cloned := column.Clone()

	column.Set(0, hexane.Some([]byte("original")))

	clonedValue := cloned.Get(1)
	clonedValue.Value[0] = 'x'

	cloned.Insert(cloned.Len(), hexane.Some([]byte("clone")))

	if got := string(cloned.Get(0).Value); got != "a" {
		t.Errorf("clone first value = %q, want %q", got, "a")
	}

	if got := string(cloned.Get(1).Value); got != "b" {
		t.Errorf("clone second value = %q, want %q", got, "b")
	}

	if got := string(column.Get(0).Value); got != "original" {
		t.Errorf("original first value = %q, want %q", got, "original")
	}

	if column.Len() != 2 || cloned.Len() != 3 {
		t.Errorf("lengths = (%d, %d), want (2, 3)", column.Len(), cloned.Len())
	}
}

func TestColumn_InvalidBoundsPanic(t *testing.T) {
	t.Parallel()

	column := hexane.NewColumn(hexane.Uint64Codec())
	column.Insert(0, hexane.Some(uint64(1)))

	cases := map[string]func(){
		"negative get":    func() { column.Get(-1) },
		"past-end get":    func() { column.Get(1) },
		"negative insert": func() { column.Insert(-1, hexane.Some(uint64(2))) },
		"past insert":     func() { column.Insert(2, hexane.Some(uint64(2))) },
		"negative delete": func() { column.Delete(0, -1) },
		"long delete":     func() { column.Delete(0, 2) },
	}

	for name, operation := range cases {
		name := name
		operation := operation

		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				defer func() {
					if recover() == nil {
						t.Error("operation did not panic")
					}
				}()

				operation()
			},
		)
	}
}

func TestColumn_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(10, 20))

	model := make([]hexane.Value[int64], 1_024)
	for i := range model {
		if i%11 == 0 {
			model[i] = hexane.Null[int64]()
		} else {
			model[i] = hexane.Some(int64(i%9 - 4))
		}
	}

	column := hexane.NewColumnFromValues(hexane.Int64Codec(), model...)

	for step := range 5_000 {
		index := random.IntN(len(model) + 1)

		deleteCount := 0
		if len(model) > index {
			deleteCount = random.IntN(len(model) - index + 1)
		}

		inserted := make([]hexane.Value[int64], random.IntN(6))
		for i := range inserted {
			if random.IntN(5) == 0 {
				inserted[i] = hexane.Null[int64]()
			} else {
				inserted[i] = hexane.Some(int64(random.IntN(9) - 4))
			}
		}

		column.Splice(index, deleteCount, inserted...)
		model = spliceValues(model, index, deleteCount, inserted)
		assertColumnState(t, step, column, model)
	}
}

func assertColumnState(
	t *testing.T,
	step int,
	column *hexane.Column[int64],
	model []hexane.Value[int64],
) {
	t.Helper()

	if column.Len() != len(model) {
		t.Fatalf("step %d: Len() = %d, want %d", step, column.Len(), len(model))
	}

	if got := column.Values(); !reflect.DeepEqual(got, model) {
		t.Fatalf("step %d: Values() = %#v, want %#v", step, got, model)
	}

	for i, want := range model {
		if got := column.Get(i); !reflect.DeepEqual(got, want) {
			t.Fatalf("step %d index %d: Get() = %#v, want %#v", step, i, got, want)
		}
	}

	gotWire, err := column.Bytes()
	if err != nil {
		t.Fatalf("step %d: Bytes() error = %v", step, err)
	}

	wantWire := canonicalRLE(
		model,
		func(data []byte, value int64) []byte {
			return testAppendLEB(data, value)
		},
	)
	if !bytes.Equal(gotWire, wantWire) {
		t.Fatalf("step %d: Bytes() = %x, want %x", step, gotWire, wantWire)
	}
}

func spliceValues[T any](
	model []hexane.Value[T],
	index int,
	deleteCount int,
	inserted []hexane.Value[T],
) []hexane.Value[T] {
	next := make([]hexane.Value[T], 0, len(model)-deleteCount+len(inserted))
	next = append(next, model[:index]...)
	next = append(next, inserted...)
	next = append(next, model[index+deleteCount:]...)

	return next
}

func canonicalRLE[T comparable](
	values []hexane.Value[T],
	appendValue func([]byte, T) []byte,
) []byte {
	hasPresent := false
	for _, value := range values {
		hasPresent = hasPresent || value.Valid
	}

	if !hasPresent {
		return nil
	}

	var data []byte

	for index := 0; index < len(values); {
		if !values[index].Valid {
			end := index + 1
			for end < len(values) && !values[end].Valid {
				end++
			}

			data = testAppendLEB(data, 0)
			data = testAppendULEB(data, uint64(end-index))
			index = end

			continue
		}

		if index+1 < len(values) &&
			values[index+1].Valid &&
			values[index+1].Value == values[index].Value {
			end := index + 2
			for end < len(values) &&
				values[end].Valid &&
				values[end].Value == values[index].Value {
				end++
			}

			data = testAppendLEB(data, int64(end-index))
			data = appendValue(data, values[index].Value)
			index = end

			continue
		}

		end := index + 1
		for end < len(values) && values[end].Valid {
			if end+1 < len(values) &&
				values[end+1].Valid &&
				values[end+1].Value == values[end].Value {
				break
			}

			end++
		}

		data = testAppendLEB(data, -int64(end-index))
		for _, value := range values[index:end] {
			data = appendValue(data, value.Value)
		}

		index = end
	}

	return data
}

func testAppendULEB(data []byte, value uint64) []byte {
	for {
		current := byte(value & 0x7f)

		value >>= 7
		if value != 0 {
			current |= 0x80
		}

		data = append(data, current)
		if value == 0 {
			return data
		}
	}
}

func testAppendLEB(data []byte, value int64) []byte {
	for {
		current := byte(value & 0x7f)
		value >>= 7

		done := (value == 0 && current&0x40 == 0) ||
			(value == -1 && current&0x40 != 0)
		if !done {
			current |= 0x80
		}

		data = append(data, current)
		if done {
			return data
		}
	}
}
