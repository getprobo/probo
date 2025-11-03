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

import "reflect"

// When conditionally applies validators based on a condition.
// If the condition is true, the specified validators are applied.
// If the condition is false, all validators are skipped.
func When(condition bool, validators ...ValidatorFunc) ValidatorFunc {
	return func(value any) *ValidationError {
		if !condition {
			return nil
		}
		for _, validator := range validators {
			if err := validator(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// RequiredIf validates that a field is required when the condition is true.
func RequiredIf(condition bool) ValidatorFunc {
	return When(condition, Required())
}

// EqualTo validates that a value equals another value using deep equality.
func EqualTo(other any) ValidatorFunc {
	return func(value any) *ValidationError {
		if !reflect.DeepEqual(value, other) {
			return newValidationError(ErrorCodeInvalidFormat, "values must match")
		}
		return nil
	}
}

// NotEqualTo validates that a value does not equal another value using deep equality.
func NotEqualTo(other any) ValidatorFunc {
	return func(value any) *ValidationError {
		if reflect.DeepEqual(value, other) {
			return newValidationError(ErrorCodeInvalidFormat, "values must not match")
		}
		return nil
	}
}
