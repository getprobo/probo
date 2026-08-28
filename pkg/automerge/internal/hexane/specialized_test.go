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

func TestDeltaColumn_EditingEncodingAndClone(t *testing.T) {
	t.Parallel()

	column := hexane.NewDeltaColumn()
	column.Insert(
		0,
		hexane.Some(int64(10)),
		hexane.Some(int64(12)),
		hexane.Null[int64](),
		hexane.Some(int64(9)),
	)

	data, err := column.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}

	// Absolute values become deltas [10, 2, null, -3].
	wantWire := []byte{0x7e, 0x0a, 0x02, 0x00, 0x01, 0x7f, 0x7d}
	if !bytes.Equal(data, wantWire) {
		t.Errorf("Bytes() = %x, want %x", data, wantWire)
	}

	cloned := column.Clone()
	column.Set(0, hexane.Some(int64(100)))
	cloned.Delete(1, 2)
	cloned.Insert(cloned.Len(), hexane.Some(int64(20)))

	if got := column.Get(0); got != hexane.Some(int64(100)) {
		t.Errorf("original Get(0) = %#v, want 100", got)
	}

	wantClone := []hexane.Value[int64]{
		hexane.Some(int64(10)),
		hexane.Some(int64(9)),
		hexane.Some(int64(20)),
	}
	if got := cloned.Values(); !reflect.DeepEqual(got, wantClone) {
		t.Errorf("clone Values() = %#v, want %#v", got, wantClone)
	}
}

func TestDeltaColumn_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(30, 40))

	model := make([]hexane.Value[int64], 1_024)
	for i := range model {
		if i%13 == 0 {
			model[i] = hexane.Null[int64]()
		} else {
			model[i] = hexane.Some(int64(i * 3))
		}
	}

	column := hexane.NewDeltaColumnFromValues(model...)

	for step := range 2_000 {
		index := random.IntN(len(model) + 1)

		deleteCount := 0
		if index < len(model) {
			deleteCount = random.IntN(len(model) - index + 1)
		}

		inserted := make([]hexane.Value[int64], random.IntN(5))
		for i := range inserted {
			if random.IntN(7) == 0 {
				inserted[i] = hexane.Null[int64]()
			} else {
				inserted[i] = hexane.Some(int64(random.IntN(1_000) - 500))
			}
		}

		column.Splice(index, deleteCount, inserted...)

		model = spliceValues(model, index, deleteCount, inserted)
		if got := column.Values(); !reflect.DeepEqual(got, model) {
			t.Fatalf("step %d: Values() = %#v, want %#v", step, got, model)
		}

		gotWire, err := column.Bytes()
		if err != nil {
			t.Fatalf("step %d: Bytes() error = %v", step, err)
		}

		wantWire := canonicalRLE(
			testAbsoluteToDeltas(model),
			func(data []byte, value int64) []byte {
				return testAppendLEB(data, value)
			},
		)
		if !bytes.Equal(gotWire, wantWire) {
			t.Fatalf("step %d: Bytes() = %x, want %x", step, gotWire, wantWire)
		}
	}
}

func TestPrefixColumn_PrefixEditingAndClone(t *testing.T) {
	t.Parallel()

	column := hexane.NewPrefixColumn()
	column.Insert(0, 2, 3, 5)
	column.Insert(0, 1)
	column.Set(2, 4)
	column.Splice(column.Len(), 0, 8)
	column.Delete(3, 1)

	want := []uint64{1, 2, 4, 8}
	if got := column.Values(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Values() = %v, want %v", got, want)
	}

	wantPrefixes := []uint64{0, 1, 3, 7, 15}
	for i, expected := range wantPrefixes {
		if got := column.Prefix(i); got != expected {
			t.Errorf("Prefix(%d) = %d, want %d", i, got, expected)
		}
	}

	cloned := column.Clone()
	column.Set(0, 100)
	cloned.Insert(0, 7)

	if got := cloned.Values(); !reflect.DeepEqual(got, []uint64{7, 1, 2, 4, 8}) {
		t.Errorf("clone Values() = %v", got)
	}

	if got := column.Values(); !reflect.DeepEqual(got, []uint64{100, 2, 4, 8}) {
		t.Errorf("original Values() = %v", got)
	}
}

func TestPrefixColumn_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(50, 60))

	model := make([]uint64, 1_024)
	for i := range model {
		model[i] = uint64(i % 10)
	}

	column := hexane.NewPrefixColumnFromValues(model...)

	for step := range 2_000 {
		index := random.IntN(len(model) + 1)

		deleteCount := 0
		if index < len(model) {
			deleteCount = random.IntN(len(model) - index + 1)
		}

		inserted := make([]uint64, random.IntN(5))
		for i := range inserted {
			inserted[i] = uint64(random.IntN(10))
		}

		column.Splice(index, deleteCount, inserted...)

		model = spliceSlice(model, index, deleteCount, inserted)
		if got := column.Values(); !reflect.DeepEqual(got, model) {
			t.Fatalf("step %d: Values() = %v, want %v", step, got, model)
		}

		wire, err := column.Bytes()
		if err != nil {
			t.Fatalf("step %d: Bytes() error = %v", step, err)
		}

		wantWire := canonicalRLE(
			testPresentValues(model),
			func(data []byte, value uint64) []byte {
				return testAppendULEB(data, value)
			},
		)
		if !bytes.Equal(wire, wantWire) {
			t.Fatalf("step %d: Bytes() = %x, want %x", step, wire, wantWire)
		}

		prefixIndex := random.IntN(len(model) + 1)

		var expected uint64
		for _, value := range model[:prefixIndex] {
			expected += value
		}

		if got := column.Prefix(prefixIndex); got != expected {
			t.Fatalf(
				"step %d: Prefix(%d) = %d, want %d",
				step,
				prefixIndex,
				got,
				expected,
			)
		}
	}
}

func TestBooleanColumn_EditingEncodingAndClone(t *testing.T) {
	t.Parallel()

	column := hexane.NewBooleanColumn()
	column.Insert(0, true, true, false)
	column.Insert(column.Len(), true)

	wantWire := []byte{0x00, 0x02, 0x01, 0x01}
	if got := column.Bytes(); !bytes.Equal(got, wantWire) {
		t.Errorf("Bytes() = %x, want %x", got, wantWire)
	}

	cloned := column.Clone()
	column.Set(0, false)
	cloned.Splice(1, 2, false, false, true)

	if got := column.Values(); !reflect.DeepEqual(got, []bool{false, true, false, true}) {
		t.Errorf("original Values() = %v", got)
	}

	if got := cloned.Values(); !reflect.DeepEqual(got, []bool{true, false, false, true, true}) {
		t.Errorf("clone Values() = %v", got)
	}
}

func TestBooleanColumn_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(70, 80))

	model := make([]bool, 1_024)
	for i := range model {
		model[i] = i%3 == 0
	}

	column := hexane.NewBooleanColumnFromValues(model...)

	for step := range 2_000 {
		index := random.IntN(len(model) + 1)

		deleteCount := 0
		if index < len(model) {
			deleteCount = random.IntN(len(model) - index + 1)
		}

		inserted := make([]bool, random.IntN(6))
		for i := range inserted {
			inserted[i] = random.IntN(2) == 1
		}

		column.Splice(index, deleteCount, inserted...)

		model = spliceSlice(model, index, deleteCount, inserted)
		if got := column.Values(); !reflect.DeepEqual(got, model) {
			t.Fatalf("step %d: Values() = %v, want %v", step, got, model)
		}

		if got, want := column.Bytes(), canonicalBooleans(model); !bytes.Equal(got, want) {
			t.Fatalf("step %d: Bytes() = %x, want %x", step, got, want)
		}
	}
}

func TestColumns_ConstructFromLogicalValues(t *testing.T) {
	t.Parallel()

	generic := hexane.NewColumnFromValues(
		hexane.Uint64Codec(),
		hexane.Some(uint64(1)),
		hexane.Null[uint64](),
		hexane.Some(uint64(2)),
	)
	if generic.Len() != 3 || generic.Get(2) != hexane.Some(uint64(2)) {
		t.Errorf("generic constructor produced %#v", generic.Values())
	}

	deltaValues := []hexane.Value[int64]{
		hexane.Some(int64(4)),
		hexane.Null[int64](),
		hexane.Some(int64(9)),
	}

	delta := hexane.NewDeltaColumnFromValues(deltaValues...)
	if got := delta.Values(); !reflect.DeepEqual(got, deltaValues) {
		t.Errorf("delta constructor = %#v, want %#v", got, deltaValues)
	}

	prefix := hexane.NewPrefixColumnFromValues(2, 3, 5)
	if prefix.Prefix(3) != 10 {
		t.Errorf("prefix constructor total = %d, want 10", prefix.Prefix(3))
	}

	boolean := hexane.NewBooleanColumnFromValues(true, false, true)
	if got := boolean.Values(); !reflect.DeepEqual(got, []bool{true, false, true}) {
		t.Errorf("boolean constructor = %v", got)
	}

	rawInput := []byte{1, 2, 3}
	raw := hexane.NewRawColumnFromBytes(rawInput)
	rawInput[0] = 9

	if got := raw.Bytes(); !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Errorf("raw constructor = %v", got)
	}
}

func TestRawColumn_ChunkBoundariesEditingAndClone(t *testing.T) {
	t.Parallel()

	initial := make([]byte, 8_300)
	for i := range initial {
		initial[i] = byte(i)
	}

	column := hexane.NewRawColumn()
	column.Insert(0, initial...)
	column.Splice(4_095, 3, []byte{9, 8, 7, 6})
	column.Set(0, 5)
	column.Delete(column.Len()-2, 2)
	column.Insert(column.Len(), 1, 2)

	model := bytes.Clone(initial)
	model = spliceSlice(model, 4_095, 3, []byte{9, 8, 7, 6})
	model[0] = 5
	model = model[:len(model)-2]
	model = append(model, 1, 2)

	if got := column.Bytes(); !bytes.Equal(got, model) {
		t.Fatalf("Bytes() differs across chunk-boundary edits")
	}

	if got := column.Slice(4_093, 4_102); !bytes.Equal(got, model[4_093:4_102]) {
		t.Errorf("Slice() = %v, want %v", got, model[4_093:4_102])
	}

	cloned := column.Clone()
	column.Set(4_096, 0xff)
	cloned.Insert(0, 0xee)

	if cloned.Get(4_097) == 0xff {
		t.Error("original mutation leaked into clone")
	}

	if column.Get(0) == 0xee {
		t.Error("clone mutation leaked into original")
	}

	var saved bytes.Buffer

	written, err := cloned.SaveTo(&saved)
	if err != nil {
		t.Fatalf("SaveTo() error = %v", err)
	}

	if written != int64(cloned.Len()) || !bytes.Equal(saved.Bytes(), cloned.Bytes()) {
		t.Errorf("SaveTo() wrote inconsistent raw bytes")
	}
}

func TestRawColumn_RandomizedDifferential(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(90, 100))

	model := make([]byte, 10_000)
	for i := range model {
		model[i] = byte(i)
	}

	column := hexane.NewRawColumnFromBytes(model)

	for step := range 3_000 {
		index := random.IntN(len(model) + 1)

		deleteCount := 0
		if index < len(model) {
			deleteCount = random.IntN(min(20, len(model)-index) + 1)
		}

		inserted := make([]byte, random.IntN(20))
		for i := range inserted {
			inserted[i] = byte(random.Uint32())
		}

		column.Splice(index, deleteCount, inserted)

		model = spliceSlice(model, index, deleteCount, inserted)
		if column.Len() != len(model) {
			t.Fatalf("step %d: Len() = %d, want %d", step, column.Len(), len(model))
		}

		if got := column.Bytes(); !bytes.Equal(got, model) {
			t.Fatalf("step %d: Bytes() differs", step)
		}

		if len(model) > 0 {
			check := random.IntN(len(model))
			if got := column.Get(check); got != model[check] {
				t.Fatalf("step %d: Get(%d) = %d, want %d", step, check, got, model[check])
			}
		}
	}
}

func spliceSlice[T any](model []T, index, deleteCount int, inserted []T) []T {
	next := make([]T, 0, len(model)-deleteCount+len(inserted))
	next = append(next, model[:index]...)
	next = append(next, inserted...)
	next = append(next, model[index+deleteCount:]...)

	return next
}

func testAbsoluteToDeltas(values []hexane.Value[int64]) []hexane.Value[int64] {
	deltas := make([]hexane.Value[int64], len(values))

	var previous int64

	for i, value := range values {
		if value.Valid {
			deltas[i] = hexane.Some(value.Value - previous)
			previous = value.Value
		}
	}

	return deltas
}

func testPresentValues[T any](values []T) []hexane.Value[T] {
	present := make([]hexane.Value[T], len(values))
	for i, value := range values {
		present[i] = hexane.Some(value)
	}

	return present
}

func canonicalBooleans(values []bool) []byte {
	if len(values) == 0 {
		return nil
	}

	var (
		data    []byte
		current bool
		count   uint64
	)
	for _, value := range values {
		if value == current {
			count++
		} else {
			data = testAppendULEB(data, count)
			current = value
			count = 1
		}
	}

	return testAppendULEB(data, count)
}
