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

import "fmt"

// Min validates that a number is at least the specified minimum value.
func Min(min int) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		var num int
		switch v := actualValue.(type) {
		case int:
			num = v
		case int32:
			num = int(v)
		case int64:
			num = int(v)
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a number")
		}

		if num < min {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be at least %d", min),
			)
		}

		return nil
	}
}

// Max validates that a number does not exceed the specified maximum value.
func Max(max int) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		var num int
		switch v := actualValue.(type) {
		case int:
			num = v
		case int32:
			num = int(v)
		case int64:
			num = int(v)
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a number")
		}

		if num > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be at most %d", max),
			)
		}

		return nil
	}
}

// Range validates that a number is within the specified range (inclusive).
func Range(min, max int) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		var num int
		switch v := actualValue.(type) {
		case int:
			num = v
		case int32:
			num = int(v)
		case int64:
			num = int(v)
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a number")
		}

		if num < min || num > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be between %d and %d", min, max),
			)
		}

		return nil
	}
}

// MinFloat validates that a floating-point number is at least the specified minimum value.
func MinFloat(min float64) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		var num float64
		switch v := actualValue.(type) {
		case float32:
			num = float64(v)
		case float64:
			num = v
		case int:
			num = float64(v)
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a number")
		}

		if num < min {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be at least %g", min),
			)
		}

		return nil
	}
}

// MaxFloat validates that a floating-point number does not exceed the specified maximum value.
func MaxFloat(max float64) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		var num float64
		switch v := actualValue.(type) {
		case float32:
			num = float64(v)
		case float64:
			num = v
		case int:
			num = float64(v)
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a number")
		}

		if num > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be at most %g", max),
			)
		}

		return nil
	}
}

// RangeFloat validates that a floating-point number is within the specified range (inclusive).
func RangeFloat(min, max float64) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		var num float64
		switch v := actualValue.(type) {
		case float32:
			num = float64(v)
		case float64:
			num = v
		case int:
			num = float64(v)
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a number")
		}

		if num < min || num > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be between %g and %g", min, max),
			)
		}

		return nil
	}
}
