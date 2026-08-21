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
	"context"
	"fmt"
	"testing"

	"go.probo.inc/probo/pkg/automerge"
)

func FuzzLoad(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("invalid"))
	f.Add([]byte{0x85, 0x6f, 0x4a, 0x83})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024*1024 {
			t.Skip()
		}

		document, err := automerge.Load(context.Background(), data, actor(255))
		if err != nil {
			return
		}

		_ = document.Close(context.Background())
	})
}

func FuzzCoreOperations(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Add([]byte("automerge fuzz operations"))
	f.Add([]byte{255, 0, 255, 1, 254, 2, 253, 3})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}

		ctx := context.Background()

		document, err := automerge.New(ctx, actor(254))
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = document.Close(context.Background()) }()

		root := document.Root()

		values, err := root.CreateObject(ctx, "values", automerge.ObjectTypeMap)
		if err != nil {
			t.Fatal(err)
		}

		list, err := root.CreateObject(ctx, "list", automerge.ObjectTypeList)
		if err != nil {
			t.Fatal(err)
		}

		mapModel := make(map[string]int64)

		var listModel []int64

		for index, operation := range data {
			key := fmt.Sprintf("key-%d", operation%8)
			value := int64(int8(operation))

			switch operation % 5 {
			case 0:
				mapModel[key] = value
				err = values.PutScalar(
					ctx,
					key,
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  value,
					},
				)
			case 1:
				if _, ok := mapModel[key]; ok {
					delete(mapModel, key)
					err = values.DeleteKey(ctx, key)
				}
			case 2:
				position := 0
				if len(listModel) > 0 {
					position = int(operation) % (len(listModel) + 1)
				}

				listModel = append(listModel, 0)
				copy(listModel[position+1:], listModel[position:])
				listModel[position] = value
				err = list.InsertScalar(
					ctx,
					uint64(position),
					automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  value,
					},
				)
			case 3:
				if len(listModel) > 0 {
					position := int(operation) % len(listModel)
					listModel[position] = value
					err = list.PutScalarAt(
						ctx,
						uint64(position),
						automerge.Scalar{
							Type: automerge.ScalarTypeInt,
							Int:  value,
						},
					)
				}
			case 4:
				if len(listModel) > 0 {
					position := int(operation) % len(listModel)
					listModel = append(
						listModel[:position],
						listModel[position+1:]...,
					)
					err = list.DeleteIndex(ctx, uint64(position))
				}
			}

			if err != nil {
				t.Fatalf("operation %d failed: %v", index, err)
			}
		}

		if _, err := document.Commit(ctx, "fuzz operations", commitTime); err != nil {
			t.Fatal(err)
		}

		saved, err := document.Save(ctx)
		if err != nil {
			t.Fatal(err)
		}

		loaded, err := automerge.Load(ctx, saved, actor(253))
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = loaded.Close(context.Background()) }()

		loadedValues, err := loaded.Root().Object(ctx, "values")
		if err != nil {
			t.Fatal(err)
		}

		for key, expected := range mapModel {
			value, err := loadedValues.Scalar(ctx, key)
			if err != nil {
				t.Fatal(err)
			}

			if value.Int != expected {
				t.Fatalf("map value %q is %d, expected %d", key, value.Int, expected)
			}
		}

		loadedList, err := loaded.Root().Object(ctx, "list")
		if err != nil {
			t.Fatal(err)
		}

		length, err := loadedList.Len(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if length != uint64(len(listModel)) {
			t.Fatalf("list length is %d, expected %d", length, len(listModel))
		}

		for index, expected := range listModel {
			value, err := loadedList.ScalarAt(ctx, uint64(index))
			if err != nil {
				t.Fatal(err)
			}

			if value.Int != expected {
				t.Fatalf(
					"list value %d is %d, expected %d",
					index,
					value.Int,
					expected,
				)
			}
		}
	})
}
