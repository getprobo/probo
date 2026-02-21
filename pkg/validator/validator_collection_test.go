// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package validator

import (
	"testing"
)

func TestMinItems(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		items := []string{"a", "b", "c"}
		err := MinItems(2)(&items)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("exact minimum", func(t *testing.T) {
		items := []int{1, 2}
		err := MinItems(2)(&items)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("too few items", func(t *testing.T) {
		items := []string{"a"}
		err := MinItems(2)(&items)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		var items *[]string
		err := MinItems(2)(items)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})

	t.Run("non-slice value", func(t *testing.T) {
		value := "not a slice"
		err := MinItems(2)(&value)
		if err == nil || err.Code != ErrorCodeInvalidFormat {
			t.Error("expected invalid format error")
		}
	})
}

func TestMaxItems(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		items := []string{"a", "b"}
		err := MaxItems(5)(&items)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("exact maximum", func(t *testing.T) {
		items := []int{1, 2, 3}
		err := MaxItems(3)(&items)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("too many items", func(t *testing.T) {
		items := []string{"a", "b", "c", "d"}
		err := MaxItems(2)(&items)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeOutOfRange {
			t.Errorf("expected error code %s, got %s", ErrorCodeOutOfRange, err.Code)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		var items *[]string
		err := MaxItems(2)(items)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestUniqueItems(t *testing.T) {
	t.Run("unique items", func(t *testing.T) {
		items := []string{"a", "b", "c"}
		err := UniqueItems()(&items)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duplicate items", func(t *testing.T) {
		items := []string{"a", "b", "a"}
		err := UniqueItems()(&items)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeInvalidFormat {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
		}
	})

	t.Run("unique integers", func(t *testing.T) {
		items := []int{1, 2, 3}
		err := UniqueItems()(&items)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duplicate integers", func(t *testing.T) {
		items := []int{1, 2, 1}
		err := UniqueItems()(&items)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		var items *[]string
		err := UniqueItems()(items)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		items := []string{}
		err := UniqueItems()(&items)
		if err != nil {
			t.Errorf("expected no error for empty slice, got: %v", err)
		}
	})

	t.Run("non-comparable type - slice of slices", func(t *testing.T) {
		items := [][]int{{1, 2}, {3, 4}}
		err := UniqueItems()(&items)
		if err == nil {
			t.Fatal("expected validation error for non-comparable type")
		} else {
			if err.Code != ErrorCodeInvalidFormat {
				t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
			}
			if err.Message != "cannot validate uniqueness for non-comparable types" {
				t.Errorf("unexpected error message: %s", err.Message)
			}
		}
	})

	t.Run("non-comparable type - slice of maps", func(t *testing.T) {
		items := []map[string]int{{"a": 1}, {"b": 2}}
		err := UniqueItems()(&items)
		if err == nil {
			t.Fatal("expected validation error for non-comparable type")
		} else if err.Code != ErrorCodeInvalidFormat {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
		}
	})

	t.Run("non-comparable type - struct with slice field", func(t *testing.T) {
		type NonComparable struct {
			Items []int
		}
		items := []NonComparable{{Items: []int{1, 2}}, {Items: []int{3, 4}}}
		err := UniqueItems()(&items)
		if err == nil {
			t.Fatal("expected validation error for non-comparable type")
		} else if err.Code != ErrorCodeInvalidFormat {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
		}
	})

	t.Run("comparable struct with unique values", func(t *testing.T) {
		type ComparableStruct struct {
			ID   int
			Name string
		}
		items := []ComparableStruct{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
		err := UniqueItems()(&items)
		if err != nil {
			t.Errorf("expected no error for comparable structs, got: %v", err)
		}
	})

	t.Run("comparable struct with duplicate values", func(t *testing.T) {
		type ComparableStruct struct {
			ID   int
			Name string
		}
		items := []ComparableStruct{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 1, Name: "a"}}
		err := UniqueItems()(&items)
		if err == nil {
			t.Fatal("expected validation error for duplicate comparable structs")
		} else if err.Code != ErrorCodeInvalidFormat {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
		}
	})
}
