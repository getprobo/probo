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

package validator_test

import (
	"testing"

	"go.probo.inc/probo/pkg/validator"
)

func TestDoublePointerValidation(t *testing.T) {
	t.Run("valid double pointer string", func(t *testing.T) {
		v := validator.New()
		str := "hello"
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.Required(), validator.NotEmpty(), validator.MaxLen(1000))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Error())
		}
	})

	t.Run("invalid double pointer string - empty", func(t *testing.T) {
		v := validator.New()
		str := ""
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.Required(), validator.NotEmpty())

		if !v.HasErrors() {
			t.Error("expected errors for empty string")
		}
	})

	t.Run("invalid double pointer string - too long", func(t *testing.T) {
		v := validator.New()
		str := "this is a very long string that exceeds the maximum length"
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.Required(), validator.MaxLen(10))

		if !v.HasErrors() {
			t.Error("expected errors for string exceeding max length")
		}
	})

	t.Run("optional double pointer - nil outer pointer", func(t *testing.T) {
		v := validator.New()
		var doublePtr **string = nil

		v.Check(doublePtr, "name", validator.NotEmpty(), validator.MaxLen(1000))

		if v.HasErrors() {
			t.Errorf("expected no errors for nil optional field, got: %v", v.Error())
		}
	})

	t.Run("optional double pointer - nil inner pointer", func(t *testing.T) {
		v := validator.New()
		var ptr *string = nil
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.NotEmpty(), validator.MaxLen(1000))

		if v.HasErrors() {
			t.Errorf("expected no errors for nil optional field, got: %v", v.Error())
		}
	})

	t.Run("optional double pointer - valid value", func(t *testing.T) {
		v := validator.New()
		str := "hello"
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.NotEmpty(), validator.MaxLen(1000))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Error())
		}
	})
}
