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

import "strings"

// Required validates that a field has a value.
// For strings, it also checks that the value is not empty or just whitespace.
// For slices, it checks that the slice is not empty.
func Required() ValidatorFunc {
	return func(value any) *ValidationError {
		if value == nil {
			return newValidationError(ErrorCodeRequired, "field is required")
		}

		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case *string:
			if v == nil || strings.TrimSpace(*v) == "" {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case int, int8, int16, int32, int64:
		case *int:
			if v == nil {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case *int8:
			if v == nil {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case *int16:
			if v == nil {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case *int32:
			if v == nil {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case *int64:
			if v == nil {
				return newValidationError(ErrorCodeRequired, "field is required")
			}
		case []any:
			if len(v) == 0 {
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
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return newValidationError(ErrorCodeRequired, "field cannot be empty")
			}
		case *string:
			if v == nil || strings.TrimSpace(*v) == "" {
				return newValidationError(ErrorCodeRequired, "field cannot be empty")
			}
		case []any:
			if len(v) == 0 {
				return newValidationError(ErrorCodeRequired, "field cannot be empty")
			}
		}

		return nil
	}
}
