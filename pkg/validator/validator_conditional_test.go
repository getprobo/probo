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

func TestWhen(t *testing.T) {
	t.Run("condition true - validation runs", func(t *testing.T) {
		str := "abc"
		err := When(true, MinLen(5))(&str)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("condition false - validation skipped", func(t *testing.T) {
		str := "abc"
		err := When(false, MinLen(5))(&str)
		if err != nil {
			t.Errorf("expected no error when condition is false, got: %v", err)
		}
	})

	t.Run("multiple validators when true", func(t *testing.T) {
		str := "abc"
		err := When(true, MinLen(5), AlphaNumeric())(&str)
		if err == nil {
			t.Error("expected validation error for MinLen")
		}
	})

	t.Run("multiple validators when false", func(t *testing.T) {
		str := "abc"
		err := When(false, MinLen(5), AlphaNumeric())(&str)
		if err != nil {
			t.Errorf("expected no error when condition is false, got: %v", err)
		}
	})
}

func TestRequiredIf(t *testing.T) {
	t.Run("condition true - field required", func(t *testing.T) {
		str := ""
		err := RequiredIf(true)(&str)
		if err == nil {
			t.Error("expected validation error")
		}
		if err.Code != ErrorCodeRequired {
			t.Errorf("expected error code %s, got %s", ErrorCodeRequired, err.Code)
		}
	})

	t.Run("condition false - field not required", func(t *testing.T) {
		str := ""
		err := RequiredIf(false)(&str)
		if err != nil {
			t.Errorf("expected no error when condition is false, got: %v", err)
		}
	})

	t.Run("condition true - field has value", func(t *testing.T) {
		str := "hello"
		err := RequiredIf(true)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("condition false - field has value", func(t *testing.T) {
		str := "hello"
		err := RequiredIf(false)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

func TestEqualTo(t *testing.T) {
	t.Run("equal strings", func(t *testing.T) {
		str1 := "password"
		str2 := "password"
		err := EqualTo(&str2)(&str1)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("different strings", func(t *testing.T) {
		str1 := "password"
		str2 := "different"
		err := EqualTo(&str2)(&str1)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("equal integers", func(t *testing.T) {
		num1 := 42
		num2 := 42
		err := EqualTo(&num2)(&num1)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("different integers", func(t *testing.T) {
		num1 := 42
		num2 := 43
		err := EqualTo(&num2)(&num1)
		if err == nil {
			t.Error("expected validation error")
		}
	})
}

func TestNotEqualTo(t *testing.T) {
	t.Run("different strings", func(t *testing.T) {
		str1 := "password"
		str2 := "different"
		err := NotEqualTo(&str2)(&str1)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("equal strings", func(t *testing.T) {
		str1 := "password"
		str2 := "password"
		err := NotEqualTo(&str2)(&str1)
		if err == nil {
			t.Error("expected validation error")
		}
	})
}
