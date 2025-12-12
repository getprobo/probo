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
