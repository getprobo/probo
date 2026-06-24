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
	"reflect"
	"strings"

	"go.probo.inc/probo/pkg/mail"
)

// Required validates that a field has a value.
// For strings, it also checks that the value is not empty or just whitespace.
// For slices, it checks that the slice is not empty.
func Required() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return newValidationError(ErrorCodeRequired, "field is required")
		}

		switch v := actualValue.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		default:
			rv := reflect.ValueOf(actualValue)
			if rv.Kind() == reflect.Slice && rv.Len() == 0 {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		}

		return nil
	}
}

// NoDuplicates validates that a slice contains no duplicate elements.
func NoDuplicates() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		rv := reflect.ValueOf(actualValue)
		if rv.Kind() != reflect.Slice {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a slice")
		}

		if !rv.Type().Elem().Comparable() {
			return newValidationError(ErrorCodeInvalidFormat, "slice elements must be comparable")
		}

		seen := make(map[any]struct{}, rv.Len())
		for i := range rv.Len() {
			elem := rv.Index(i).Interface()
			if _, ok := seen[elem]; ok {
				return newValidationError(ErrorCodeInvalidFormat, "must not contain duplicates")
			}

			seen[elem] = struct{}{}
		}

		return nil
	}
}

// NotEmpty validates that a field is not empty.
// Similar to Required, but can be used independently.
func NotEmpty() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		switch v := actualValue.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return newValidationError(ErrorCodeRequired, "field cannot be empty")
			}
		case mail.Addr:
			if v == mail.Nil {
				return newValidationError(ErrorCodeRequired, "field cannot be empty")
			}
		default:
			rv := reflect.ValueOf(actualValue)
			if rv.Kind() == reflect.Slice && rv.Len() == 0 {
				return newValidationError(ErrorCodeRequired, "field cannot be empty")
			}
		}

		return nil
	}
}
