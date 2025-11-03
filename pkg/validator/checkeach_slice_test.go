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

// CustomType simulates types like gid.GID
type CustomType string

func TestCheckEach_EmptyTypedSlice(t *testing.T) {
	v := New()

	// Simulate what happens with []gid.GID{} (empty slice of custom type)
	emptySlice := []CustomType{}

	v.CheckEach(emptySlice, "items", func(index int, item any) {
		// This callback should never be called for an empty slice
		t.Error("callback should not be called for empty slice")
	})

	if v.HasErrors() {
		t.Errorf("unexpected error for empty slice: %v", v.Error())
	}
}

func TestCheckEach_NonEmptyTypedSlice(t *testing.T) {
	v := New()

	// Simulate what happens with []gid.GID{"abc", "def"}
	slice := []CustomType{"abc", "def"}

	callCount := 0
	v.CheckEach(slice, "items", func(index int, item any) {
		callCount++
		// Verify the item is the correct type
		str, ok := item.(CustomType)
		if !ok {
			t.Errorf("expected CustomType, got %T", item)
		}
		if index == 0 && str != "abc" {
			t.Errorf("expected 'abc', got %s", str)
		}
		if index == 1 && str != "def" {
			t.Errorf("expected 'def', got %s", str)
		}
	})

	if callCount != 2 {
		t.Errorf("expected callback to be called 2 times, got %d", callCount)
	}

	if v.HasErrors() {
		t.Errorf("unexpected error: %v", v.Error())
	}
}

func TestCheckEach_NilTypedSlice(t *testing.T) {
	v := New()

	// Simulate what happens with var x []gid.GID (nil slice)
	var nilSlice []CustomType

	v.CheckEach(nilSlice, "items", func(index int, item any) {
		// This callback should never be called for a nil slice
		t.Error("callback should not be called for nil slice")
	})

	if v.HasErrors() {
		t.Errorf("unexpected error for nil slice: %v", v.Error())
	}
}

func TestCheckEach_PointerToNonEmptySlice(t *testing.T) {
	v := New()

	// Simulate what happens with *[]gid.GID (pointer to slice)
	slice := []CustomType{"abc", "def", "ghi"}
	ptrToSlice := &slice

	callCount := 0
	v.CheckEach(ptrToSlice, "items", func(index int, item any) {
		callCount++
		str, ok := item.(CustomType)
		if !ok {
			t.Errorf("expected CustomType, got %T", item)
		}
		expectedValues := []CustomType{"abc", "def", "ghi"}
		if str != expectedValues[index] {
			t.Errorf("at index %d: expected %s, got %s", index, expectedValues[index], str)
		}
	})

	if callCount != 3 {
		t.Errorf("expected callback to be called 3 times, got %d", callCount)
	}

	if v.HasErrors() {
		t.Errorf("unexpected error for pointer to slice: %v", v.Error())
	}
}

func TestCheckEach_PointerToEmptySlice(t *testing.T) {
	v := New()

	// Simulate what happens with *[]gid.GID{} (pointer to empty slice)
	slice := []CustomType{}
	ptrToSlice := &slice

	v.CheckEach(ptrToSlice, "items", func(index int, item any) {
		t.Error("callback should not be called for empty slice")
	})

	if v.HasErrors() {
		t.Errorf("unexpected error for pointer to empty slice: %v", v.Error())
	}
}

func TestCheckEach_NilPointerToSlice(t *testing.T) {
	v := New()

	// Simulate what happens with var x *[]gid.GID (nil pointer to slice)
	var nilPtrToSlice *[]CustomType

	v.CheckEach(nilPtrToSlice, "items", func(index int, item any) {
		t.Error("callback should not be called for nil pointer to slice")
	})

	if v.HasErrors() {
		t.Errorf("unexpected error for nil pointer to slice: %v", v.Error())
	}
}

func TestCheckEach_DoublePointerToSlice(t *testing.T) {
	v := New()

	// Simulate what happens with **[]gid.GID (double pointer to slice)
	slice := []CustomType{"x", "y"}
	ptrToSlice := &slice
	doublePtrToSlice := &ptrToSlice

	callCount := 0
	v.CheckEach(doublePtrToSlice, "items", func(index int, item any) {
		callCount++
		str, ok := item.(CustomType)
		if !ok {
			t.Errorf("expected CustomType, got %T", item)
		}
		expectedValues := []CustomType{"x", "y"}
		if str != expectedValues[index] {
			t.Errorf("at index %d: expected %s, got %s", index, expectedValues[index], str)
		}
	})

	if callCount != 2 {
		t.Errorf("expected callback to be called 2 times, got %d", callCount)
	}

	if v.HasErrors() {
		t.Errorf("unexpected error for double pointer to slice: %v", v.Error())
	}
}

func TestCheckEach_NonSliceValue(t *testing.T) {
	v := New()

	// Pass a non-slice value
	notASlice := "this is a string"

	v.CheckEach(notASlice, "items", func(index int, item any) {
		t.Error("callback should not be called for non-slice value")
	})

	if !v.HasErrors() {
		t.Error("expected error for non-slice value")
	}

	errors := v.Errors()
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Code != ErrorCodeInvalidFormat {
		t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, errors[0].Code)
	}
	if errors[0].Message != "expected a slice" {
		t.Errorf("expected message 'expected a slice', got '%s'", errors[0].Message)
	}
}
