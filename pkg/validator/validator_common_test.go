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

func TestOptionalByDefault(t *testing.T) {
	t.Run("nil value skips validation by default", func(t *testing.T) {
		v := New()
		v.Check(nil, "field", MinLen(5))

		if v.HasErrors() {
			t.Errorf("expected no errors for nil (optional by default), got: %v", v.Errors())
		}
	})

	t.Run("nil pointer skips validation by default", func(t *testing.T) {
		v := New()
		var str *string
		v.Check(str, "field", MinLen(5))

		if v.HasErrors() {
			t.Errorf("expected no errors for nil pointer (optional by default), got: %v", v.Errors())
		}
	})

	t.Run("valid value passes validation", func(t *testing.T) {
		v := New()
		str := "hello world"
		v.Check(&str, "field", MinLen(5))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("invalid value fails validation", func(t *testing.T) {
		v := New()
		str := "hi"
		v.Check(&str, "field", MinLen(5))

		if !v.HasErrors() {
			t.Error("expected validation error")
		}
	})

	t.Run("multiple validators", func(t *testing.T) {
		v := New()
		str := "hello"
		v.Check(&str, "field", MinLen(3), MaxLen(10))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("empty string is not nil and gets validated", func(t *testing.T) {
		v := New()
		str := ""
		v.Check(&str, "field", MinLen(5))

		if !v.HasErrors() {
			t.Error("expected validation error for empty string")
		}
	})

	t.Run("Required() validates nil values", func(t *testing.T) {
		v := New()
		var str *string
		v.Check(str, "field", Required())

		if !v.HasErrors() {
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
