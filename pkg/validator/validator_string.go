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
	"fmt"
	"reflect"
	"strings"
)

// MinLen validates that a string has at least the specified minimum length.
func MinLen(minLength int) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if len(str) < minLength {
			return newValidationError(
				ErrorCodeTooShort,
				fmt.Sprintf("must be at least %d characters", minLength),
			)
		}

		return nil
	}
}

// MaxLen validates that a string does not exceed the specified maximum length.
func MaxLen(maxLength int) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if len(str) > maxLength {
			return newValidationError(
				ErrorCodeTooLong,
				fmt.Sprintf("must be at most %d characters", maxLength),
			)
		}

		return nil
	}
}

// OneOfSlice validates that a value is one of the allowed values in the slice.
// Accepts a slice of any type. Compares by value first, then by string representation.
func OneOfSlice[T any](allowed []T) ValidatorFunc {
	// Build allowed map with string keys for flexible comparison
	allowedMap := make(map[string]bool)
	allowedStrings := make([]string, 0, len(allowed))

	for _, v := range allowed {
		str := fmt.Sprint(v)
		allowedMap[str] = true
		allowedStrings = append(allowedStrings, str)
	}

	return func(value any) *ValidationError {
		// Handle nil values first
		if value == nil {
			return nil
		}

		// Dereference all pointer levels
		actualValue := value
		val := reflect.ValueOf(value)
		for val.Kind() == reflect.Pointer {
			if val.IsNil() {
				return nil
			}
			val = val.Elem()
			actualValue = val.Interface()
		}

		// First try exact match with DeepEqual
		for _, allowedVal := range allowed {
			if reflect.DeepEqual(actualValue, allowedVal) {
				return nil
			}
		}

		// Then try string comparison (for custom string types)
		valueStr := fmt.Sprint(actualValue)
		if allowedMap[valueStr] {
			return nil
		}

		return newValidationError(
			ErrorCodeInvalidEnum,
			fmt.Sprintf("must be one of: %s", strings.Join(allowedStrings, ", ")),
		)
	}
}
