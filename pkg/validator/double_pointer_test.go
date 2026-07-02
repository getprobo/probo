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

		if v.Error() != nil {
			t.Errorf("expected no errors, got: %v", v.Error())
		}
	})

	t.Run("invalid double pointer string - empty", func(t *testing.T) {
		v := validator.New()
		str := ""
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.Required(), validator.NotEmpty())

		if v.Error() == nil {
			t.Error("expected errors for empty string")
		}
	})

	t.Run("invalid double pointer string - too long", func(t *testing.T) {
		v := validator.New()
		str := "this is a very long string that exceeds the maximum length"
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.Required(), validator.MaxLen(10))

		if v.Error() == nil {
			t.Error("expected errors for string exceeding max length")
		}
	})

	t.Run("optional double pointer - nil outer pointer", func(t *testing.T) {
		v := validator.New()

		var doublePtr **string = nil

		v.Check(doublePtr, "name", validator.NotEmpty(), validator.MaxLen(1000))

		if v.Error() != nil {
			t.Errorf("expected no errors for nil optional field, got: %v", v.Error())
		}
	})

	t.Run("optional double pointer - nil inner pointer", func(t *testing.T) {
		v := validator.New()

		var ptr *string = nil

		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.NotEmpty(), validator.MaxLen(1000))

		if v.Error() != nil {
			t.Errorf("expected no errors for nil optional field, got: %v", v.Error())
		}
	})

	t.Run("optional double pointer - valid value", func(t *testing.T) {
		v := validator.New()
		str := "hello"
		ptr := &str
		doublePtr := &ptr

		v.Check(doublePtr, "name", validator.NotEmpty(), validator.MaxLen(1000))

		if v.Error() != nil {
			t.Errorf("expected no errors, got: %v", v.Error())
		}
	})
}
