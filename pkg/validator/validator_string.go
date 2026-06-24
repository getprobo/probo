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

// ContainsSubstring validates that a string contains the specified substring.
func ContainsSubstring(substr string) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if !strings.Contains(str, substr) {
			return newValidationError(
				ErrorCodeInvalidFormat,
				fmt.Sprintf("must contain %q", substr),
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
