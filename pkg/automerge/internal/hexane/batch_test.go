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
	"slices"
	"testing"

	"go.probo.inc/probo/pkg/automerge/internal/hexane"
)

func TestColumn_BatchSpliceRandomizedDifferential(t *testing.T) {
	t.Parallel()

	random := rand.New(rand.NewPCG(130, 140))

	for iteration := range 500 {
		model := make([]hexane.Value[int64], 1_024)
		for i := range model {
			model[i] = hexane.Some(int64(i % 7))
		}

		column := hexane.NewColumnFromValues(hexane.Int64Codec(), model...)

		splices := make([]hexane.ColumnSplice[int64], 0, 8)

		for index := 64; index < len(model); index += 127 {
			deleteCount := random.IntN(min(5, len(model)-index) + 1)

			inserted := make([]hexane.Value[int64], random.IntN(6))
			for i := range inserted {
				if random.IntN(5) == 0 {
					inserted[i] = hexane.Null[int64]()
				} else {
					inserted[i] = hexane.Some(int64(random.IntN(7)))
				}
			}

			splices = append(splices, hexane.ColumnSplice[int64]{
				Index:       index,
				DeleteCount: deleteCount,
				Inserted: hexane.NewColumnFromValues(
					hexane.Int64Codec(),
					inserted...,
				),
			})
		}

		for i := len(splices) - 1; i >= 0; i-- {
			splice := splices[i]
			model = spliceValues(
				model,
				splice.Index,
				splice.DeleteCount,
				splice.Inserted.Values(),
			)
		}

		random.Shuffle(len(splices), func(i, j int) {
			splices[i], splices[j] = splices[j], splices[i]
		})

		if err := column.BatchSplice(splices); err != nil {
			t.Fatalf("iteration %d: BatchSplice() error = %v", iteration, err)
		}

		if got := column.Values(); !slices.Equal(got, model) {
			t.Fatalf("iteration %d: Values() differ", iteration)
		}

		wantWire := canonicalRLE(
			model,
			func(data []byte, value int64) []byte {
				return testAppendLEB(data, value)
			},
		)

		gotWire, err := column.Bytes()
		if err != nil {
			t.Fatalf("iteration %d: Bytes() error = %v", iteration, err)
		}

		if !bytes.Equal(gotWire, wantWire) {
			t.Fatalf("iteration %d: Bytes() = %x, want %x", iteration, gotWire, wantWire)
		}
	}
}

func TestColumns_CachedTailAppendMatchesFreshEncoding(t *testing.T) {
	t.Parallel()

	values := []hexane.Value[int64]{
		hexane.Null[int64](),
		hexane.Some[int64](1),
		hexane.Some[int64](2),
	}

	column := hexane.NewColumnFromValues(hexane.Int64Codec(), values...)
	if _, err := column.Bytes(); err != nil {
		t.Fatal(err)
	}

	for _, value := range []hexane.Value[int64]{
		hexane.Some[int64](2),
		hexane.Some[int64](3),
		hexane.Null[int64](),
		hexane.Null[int64](),
		hexane.Some[int64](4),
	} {
		if err := column.BatchSplice([]hexane.ColumnSplice[int64]{{
			Index:    column.Len(),
			Inserted: hexane.NewColumnFromValues(hexane.Int64Codec(), value),
		}}); err != nil {
			t.Fatal(err)
		}

		values = append(values, value)

		got, err := column.Bytes()
		if err != nil {
			t.Fatal(err)
		}

		fresh, err := hexane.NewColumnFromValues(
			hexane.Int64Codec(),
			values...,
		).Bytes()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, fresh) {
			t.Fatalf("cached column bytes = %x, fresh bytes = %x", got, fresh)
		}
	}

	booleans := []bool{false, true, true}
	booleanColumn := hexane.NewBooleanColumnFromValues(booleans...)
	_ = booleanColumn.Bytes()

	for _, value := range []bool{true, false, false, true} {
		inserted := hexane.NewBooleanColumnFromValues(value)
		if err := booleanColumn.BatchSplice([]hexane.BooleanSplice{{
			Index:    booleanColumn.Len(),
			Inserted: inserted,
		}}); err != nil {
			t.Fatal(err)
		}

		booleans = append(booleans, value)

		fresh := hexane.NewBooleanColumnFromValues(booleans...)
		if !bytes.Equal(booleanColumn.Bytes(), fresh.Bytes()) {
			t.Fatal("cached boolean bytes differ from fresh encoding")
		}
	}

	deltas := []hexane.Value[int64]{hexane.Some[int64](3), hexane.Some[int64](8)}

	deltaColumn := hexane.NewDeltaColumnFromValues(deltas...)
	if _, err := deltaColumn.Bytes(); err != nil {
		t.Fatal(err)
	}

	for _, value := range []hexane.Value[int64]{
		hexane.Some[int64](13),
		hexane.Null[int64](),
		hexane.Some[int64](21),
	} {
		inserted := hexane.NewDeltaColumnFromValues(value)
		if err := deltaColumn.BatchSplice([]hexane.DeltaSplice{{
			Index:    deltaColumn.Len(),
			Inserted: inserted,
		}}); err != nil {
			t.Fatal(err)
		}

		deltas = append(deltas, value)

		got, err := deltaColumn.Bytes()
		if err != nil {
			t.Fatal(err)
		}

		fresh, err := hexane.NewDeltaColumnFromValues(deltas...).Bytes()
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, fresh) {
			t.Fatal("cached delta bytes differ from fresh encoding")
		}
	}
}

func TestColumn_BatchSpliceAdjacentLeafBoundariesAndCOW(t *testing.T) {
	t.Parallel()

	values := make([]hexane.Value[[]byte], 768)
	for i := range values {
		values[i] = hexane.Some([]byte{byte(i)})
	}

	column := hexane.NewColumnFromValues(hexane.BytesCodec(), values...)
	clone := column.Clone()

	splices := []hexane.ColumnSplice[[]byte]{
		{
			Index:       255,
			DeleteCount: 2,
			Inserted: hexane.NewColumnFromValues(
				hexane.BytesCodec(),
				hexane.Some([]byte("left")),
			),
		},
		{
			Index:       511,
			DeleteCount: 2,
			Inserted: hexane.NewColumnFromValues(
				hexane.BytesCodec(),
				hexane.Some([]byte("right")),
			),
		},
	}
	if err := column.BatchSplice(splices); err != nil {
		t.Fatalf("BatchSplice() error = %v", err)
	}

	if got := string(column.Get(255).Value); got != "left" {
		t.Errorf("Get(255) = %q, want left", got)
	}

	if got := clone.Get(255).Value; !bytes.Equal(got, []byte{255}) {
		t.Errorf("clone Get(255) = %v, want [255]", got)
	}

	if clone.Len() != len(values) || column.Len() != len(values)-2 {
		t.Errorf("lengths = (%d, %d)", clone.Len(), column.Len())
	}
}

func TestColumn_EncodedValuesAreCachedAndInvalidated(t *testing.T) {
	t.Parallel()

	appends := 0
	codec := hexane.CodecFuncs[uint64]{
		EqualFunc: func(left, right uint64) bool { return left == right },
		CloneFunc: func(value uint64) uint64 { return value },
		AppendFunc: func(data []byte, value uint64) ([]byte, error) {
			appends++
			return append(data, byte(value)), nil
		},
	}
	column := hexane.NewColumnFromValues(
		codec,
		hexane.Some(uint64(1)),
		hexane.Some(uint64(2)),
		hexane.Some(uint64(3)),
	)

	if appends != 3 {
		t.Fatalf("constructor appends = %d, want 3", appends)
	}

	_, _ = column.Bytes()
	_, _ = column.Bytes()

	if appends != 3 {
		t.Fatalf("cached Bytes() appends = %d, want 3", appends)
	}

	inserted := hexane.NewColumnFromValues(codec, hexane.Some(uint64(9)))
	if err := column.BatchSplice([]hexane.ColumnSplice[uint64]{{
		Index: 1, DeleteCount: 1, Inserted: inserted,
	}}); err != nil {
		t.Fatalf("BatchSplice() error = %v", err)
	}

	_, _ = column.Bytes()

	if appends != 4 {
		t.Errorf("post-splice appends = %d, want 4", appends)
	}
}

func TestColumns_BatchSpliceFailuresAreTransactional(t *testing.T) {
	t.Parallel()

	column := hexane.NewColumnFromValues(
		hexane.Uint64Codec(),
		hexane.Some(uint64(1)),
		hexane.Some(uint64(2)),
	)
	beforeColumn := column.Values()

	err := column.BatchSplice([]hexane.ColumnSplice[uint64]{
		{Index: 0, DeleteCount: 2},
		{Index: 1, DeleteCount: 1},
	})
	if err == nil || !slices.Equal(column.Values(), beforeColumn) {
		t.Error("generic column overlap was not transactional")
	}

	delta := hexane.NewDeltaColumnFromValues(
		hexane.Some(int64(10)),
		hexane.Some(int64(20)),
	)
	beforeDelta := delta.Values()

	err = delta.BatchSplice([]hexane.DeltaSplice{{Index: 3}})
	if err == nil || !slices.Equal(delta.Values(), beforeDelta) {
		t.Error("delta column bounds failure was not transactional")
	}

	prefix := hexane.NewPrefixColumnFromValues(1, 2)
	beforePrefix := prefix.Values()

	err = prefix.BatchSplice([]hexane.PrefixSplice{{Index: -1}})
	if err == nil || !slices.Equal(prefix.Values(), beforePrefix) {
		t.Error("prefix column bounds failure was not transactional")
	}

	boolean := hexane.NewBooleanColumnFromValues(false, true)
	beforeBoolean := boolean.Values()

	err = boolean.BatchSplice([]hexane.BooleanSplice{{Index: 1, DeleteCount: 2}})
	if err == nil || !slices.Equal(boolean.Values(), beforeBoolean) {
		t.Error("boolean column bounds failure was not transactional")
	}

	raw := hexane.NewRawColumnFromBytes([]byte{1, 2})
	beforeRaw := raw.Bytes()

	err = raw.BatchSplice([]hexane.RawSplice{{Index: 0, DeleteCount: 3}})
	if err == nil || !bytes.Equal(raw.Bytes(), beforeRaw) {
		t.Error("raw column bounds failure was not transactional")
	}
}

func TestSpecializedColumns_BatchSpliceMatchesFlatReplacement(t *testing.T) {
	t.Parallel()

	delta := hexane.NewDeltaColumnFromValues(
		hexane.Some(int64(10)),
		hexane.Null[int64](),
		hexane.Some(int64(20)),
		hexane.Some(int64(30)),
		hexane.Some(int64(40)),
	)

	err := delta.BatchSplice([]hexane.DeltaSplice{
		{
			Index: 3, DeleteCount: 1,
			Inserted: hexane.NewDeltaColumnFromValues(hexane.Some(int64(31))),
		},
		{
			Index: 1, DeleteCount: 1,
			Inserted: hexane.NewDeltaColumnFromValues(
				hexane.Some(int64(15)),
				hexane.Some(int64(16)),
			),
		},
	})
	if err != nil {
		t.Fatalf("delta BatchSplice() error = %v", err)
	}

	wantDelta := []hexane.Value[int64]{
		hexane.Some(int64(10)),
		hexane.Some(int64(15)),
		hexane.Some(int64(16)),
		hexane.Some(int64(20)),
		hexane.Some(int64(31)),
		hexane.Some(int64(40)),
	}
	if got := delta.Values(); !slices.Equal(got, wantDelta) {
		t.Errorf("delta Values() = %#v, want %#v", got, wantDelta)
	}

	raw := hexane.NewRawColumnFromBytes(make([]byte, 8_192))

	err = raw.BatchSplice([]hexane.RawSplice{
		{
			Index: 4_095, DeleteCount: 2,
			Inserted: hexane.NewRawColumnFromBytes([]byte{1, 2, 3}),
		},
		{
			Index: 4_097, DeleteCount: 2,
			Inserted: hexane.NewRawColumnFromBytes([]byte{4}),
		},
	})
	if err != nil {
		t.Fatalf("raw BatchSplice() error = %v", err)
	}

	if got := raw.Slice(4_095, 4_099); !bytes.Equal(got, []byte{1, 2, 3, 4}) {
		t.Errorf("raw boundary bytes = %v, want [1 2 3 4]", got)
	}
}
