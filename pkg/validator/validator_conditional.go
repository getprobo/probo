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

// When conditionally applies validators based on a boolean condition.
// If condition is true, all provided validators are run. If false, validation is skipped.
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

// RequiredIf conditionally requires a field based on a boolean condition.
// If condition is true, the field must have a value (using Required validation).
// If condition is false, the field is optional.
func RequiredIf(condition bool) ValidatorFunc {
	return func(value any) *ValidationError {
		if !condition {
			return nil
		}
		return Required()(value)
	}
}

// WhenSet conditionally applies validators based on whether a pointer is non-nil.
// If ptr is nil, validation is skipped. If ptr is non-nil, all provided validators are run.
func WhenSet(ptr any, validators ...ValidatorFunc) ValidatorFunc {
	return func(value any) *ValidationError {
		if ptr == nil {
			return nil
		}
		val := reflect.ValueOf(ptr)
		if val.Kind() == reflect.Ptr && val.IsNil() {
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
