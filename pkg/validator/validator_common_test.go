// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package validator

import (
	"testing"
)

func TestOptionalByDefault(t *testing.T) {
	t.Run("nil value skips validation by default", func(t *testing.T) {
		v := New()
		v.Check(nil, "field", MinLen(5))

		if v.Error() != nil {
			t.Errorf("expected no errors for nil (optional by default), got: %v", v.Error())
		}
	})

	t.Run("nil pointer skips validation by default", func(t *testing.T) {
		v := New()

		var str *string
		v.Check(str, "field", MinLen(5))

		if v.Error() != nil {
			t.Errorf("expected no errors for nil pointer (optional by default), got: %v", v.Error())
		}
	})

	t.Run("valid value passes validation", func(t *testing.T) {
		v := New()
		str := "hello world"
		v.Check(&str, "field", MinLen(5))

		if v.Error() != nil {
			t.Errorf("expected no errors, got: %v", v.Error())
		}
	})

	t.Run("invalid value fails validation", func(t *testing.T) {
		v := New()
		str := "hi"
		v.Check(&str, "field", MinLen(5))

		if v.Error() == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("multiple validators", func(t *testing.T) {
		v := New()
		str := "hello"
		v.Check(&str, "field", MinLen(3), MaxLen(10))

		if v.Error() != nil {
			t.Errorf("expected no errors, got: %v", v.Error())
		}
	})

	t.Run("empty string is not nil and gets validated", func(t *testing.T) {
		v := New()
		str := ""
		v.Check(&str, "field", MinLen(5))

		if v.Error() == nil {
			t.Error("expected validation error for empty string")
		}
	})

	t.Run("Required() validates nil values", func(t *testing.T) {
		v := New()

		var str *string
		v.Check(str, "field", Required())

		if v.Error() == nil {
			t.Error("expected validation error for nil with Required()")
		}
	})
}

func TestRequired(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		str := "hello"

		err := Required()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""

		err := Required()(&str)
		if err == nil {
			t.Fatal("expected validation error")
		} else if err.Code != ErrorCodeRequired {
			t.Errorf("expected error code %s, got %s", ErrorCodeRequired, err.Code)
		}
	})

	t.Run("whitespace string", func(t *testing.T) {
		str := "   "

		err := Required()(&str)
		if err == nil {
			t.Error("expected validation error for whitespace")
		}
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var str *string

		err := Required()(str)
		if err == nil {
			t.Error("expected validation error for nil pointer")
		}
	})

	t.Run("valid string pointer", func(t *testing.T) {
		str := "hello"

		err := Required()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("nil interface", func(t *testing.T) {
		err := Required()(nil)
		if err == nil {
			t.Error("expected validation error for nil")
		}
	})

	t.Run("zero int", func(t *testing.T) {
		num := 0

		err := Required()(&num)
		if err != nil {
			t.Errorf("expected no error for zero int, got: %v", err)
		}
	})

	t.Run("positive int", func(t *testing.T) {
		num := 42

		err := Required()(&num)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("nil int pointer", func(t *testing.T) {
		var num *int

		err := Required()(num)
		if err == nil {
			t.Error("expected validation error for nil int pointer")
		}
	})

	t.Run("valid int pointer", func(t *testing.T) {
		num := 42

		err := Required()(&num)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := []any{}

		err := Required()(slice)
		if err == nil {
			t.Error("expected validation error for empty slice")
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		slice := []any{1, 2, 3}

		err := Required()(slice)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("empty string slice", func(t *testing.T) {
		slice := []string{}

		err := Required()(slice)
		if err == nil {
			t.Fatal("expected validation error for empty []string slice")
		} else if err.Code != ErrorCodeRequired {
			t.Errorf("expected error code %s, got %s", ErrorCodeRequired, err.Code)
		}
	})

	t.Run("non-empty string slice", func(t *testing.T) {
		slice := []string{"a", "b", "c"}

		err := Required()(slice)
		if err != nil {
			t.Errorf("expected no error for non-empty []string, got: %v", err)
		}
	})

	t.Run("empty int slice", func(t *testing.T) {
		slice := []int{}

		err := Required()(slice)
		if err == nil {
			t.Fatal("expected validation error for empty []int slice")
		} else if err.Code != ErrorCodeRequired {
			t.Errorf("expected error code %s, got %s", ErrorCodeRequired, err.Code)
		}
	})

	t.Run("non-empty int slice", func(t *testing.T) {
		slice := []int{1, 2, 3}

		err := Required()(slice)
		if err != nil {
			t.Errorf("expected no error for non-empty []int, got: %v", err)
		}
	})

	t.Run("empty custom type slice", func(t *testing.T) {
		type CustomType struct {
			ID int
		}

		slice := []CustomType{}

		err := Required()(slice)
		if err == nil {
			t.Fatal("expected validation error for empty custom type slice")
		} else if err.Code != ErrorCodeRequired {
			t.Errorf("expected error code %s, got %s", ErrorCodeRequired, err.Code)
		}
	})

	t.Run("non-empty custom type slice", func(t *testing.T) {
		type CustomType struct {
			ID int
		}

		slice := []CustomType{{ID: 1}, {ID: 2}}

		err := Required()(slice)
		if err != nil {
			t.Errorf("expected no error for non-empty custom type slice, got: %v", err)
		}
	})

	t.Run("empty pointer slice", func(t *testing.T) {
		slice := []*string{}

		err := Required()(slice)
		if err == nil {
			t.Error("expected validation error for empty []*string slice")
		}
	})

	t.Run("non-empty pointer slice", func(t *testing.T) {
		str1, str2 := "a", "b"
		slice := []*string{&str1, &str2}

		err := Required()(slice)
		if err != nil {
			t.Errorf("expected no error for non-empty []*string, got: %v", err)
		}
	})
}

func TestNoDuplicates(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var slice []string

		err := NoDuplicates()(slice)
		if err != nil {
			t.Errorf("expected no error for nil slice, got: %v", err)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := []string{}

		err := NoDuplicates()(slice)
		if err != nil {
			t.Errorf("expected no error for empty slice, got: %v", err)
		}
	})

	t.Run("unique strings", func(t *testing.T) {
		slice := []string{"a", "b", "c"}

		err := NoDuplicates()(slice)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duplicate strings", func(t *testing.T) {
		slice := []string{"a", "b", "a"}

		err := NoDuplicates()(slice)
		if err == nil {
			t.Fatal("expected validation error for duplicates")
		} else if err.Code != ErrorCodeInvalidFormat {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
		}
	})

	t.Run("unique ints", func(t *testing.T) {
		slice := []int{1, 2, 3}

		err := NoDuplicates()(slice)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("duplicate ints", func(t *testing.T) {
		slice := []int{1, 2, 1}

		err := NoDuplicates()(slice)
		if err == nil {
			t.Fatal("expected validation error for duplicates")
		}
	})

	t.Run("non-comparable elements", func(t *testing.T) {
		slice := []map[string]string{{"a": "b"}}

		err := NoDuplicates()(slice)
		if err == nil {
			t.Fatal("expected validation error for non-comparable elements")
		} else if err.Code != ErrorCodeInvalidFormat {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidFormat, err.Code)
		}
	})
}
