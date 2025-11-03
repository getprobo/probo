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
	"time"
)

// After validates that a time is after the specified reference time.
// The reference time can be either time.Time or *time.Time.
func After(t any) ValidatorFunc {
	return func(value any) *ValidationError {
		// Extract the reference time
		refValue, refIsNil := dereferenceValue(t)
		if refIsNil {
			return nil // No reference time to compare against
		}

		refTime, ok := refValue.(time.Time)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "reference time must be time.Time")
		}

		// Extract the value being validated
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		timeVal, ok := actualValue.(time.Time)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.After(refTime) {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be after %s", refTime.Format(time.RFC3339)),
			)
		}

		return nil
	}
}

// Before validates that a time is before the specified reference time.
// The reference time can be either time.Time or *time.Time.
func Before(t any) ValidatorFunc {
	return func(value any) *ValidationError {
		// Extract the reference time
		refValue, refIsNil := dereferenceValue(t)
		if refIsNil {
			return nil // No reference time to compare against
		}

		refTime, ok := refValue.(time.Time)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "reference time must be time.Time")
		}

		// Extract the value being validated
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		timeVal, ok := actualValue.(time.Time)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.Before(refTime) {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be before %s", refTime.Format(time.RFC3339)),
			)
		}

		return nil
	}
}

// FutureDate validates that a time is in the future.
func FutureDate() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		timeVal, ok := actualValue.(time.Time)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.After(time.Now()) {
			return newValidationError(ErrorCodeOutOfRange, "must be a future date")
		}

		return nil
	}
}

// PastDate validates that a time is in the past.
func PastDate() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		timeVal, ok := actualValue.(time.Time)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.Before(time.Now()) {
			return newValidationError(ErrorCodeOutOfRange, "must be a past date")
		}

		return nil
	}
}

// MinDuration validates that a duration is at least the specified minimum value.
func MinDuration(min time.Duration) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		duration, ok := actualValue.(time.Duration)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Duration")
		}

		if duration < min {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be at least %s", min),
			)
		}

		return nil
	}
}

// MaxDuration validates that a duration does not exceed the specified maximum value.
func MaxDuration(max time.Duration) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		duration, ok := actualValue.(time.Duration)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Duration")
		}

		if duration > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be at most %s", max),
			)
		}

		return nil
	}
}

// RangeDuration validates that a duration is within the specified range (inclusive).
func RangeDuration(min, max time.Duration) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		duration, ok := actualValue.(time.Duration)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Duration")
		}

		if duration < min || duration > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be between %s and %s", min, max),
			)
		}

		return nil
	}
}
