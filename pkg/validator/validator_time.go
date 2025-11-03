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
func After(t time.Time) ValidatorFunc {
	return func(value any) *ValidationError {
		var timeVal time.Time
		switch v := value.(type) {
		case time.Time:
			timeVal = v
		case *time.Time:
			if v == nil {
				return nil
			}
			timeVal = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.After(t) {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be after %s", t.Format(time.RFC3339)),
			)
		}

		return nil
	}
}

// Before validates that a time is before the specified reference time.
func Before(t time.Time) ValidatorFunc {
	return func(value any) *ValidationError {
		var timeVal time.Time
		switch v := value.(type) {
		case time.Time:
			timeVal = v
		case *time.Time:
			if v == nil {
				return nil
			}
			timeVal = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.Before(t) {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must be before %s", t.Format(time.RFC3339)),
			)
		}

		return nil
	}
}

// FutureDate validates that a time is in the future.
func FutureDate() ValidatorFunc {
	return func(value any) *ValidationError {
		var timeVal time.Time
		switch v := value.(type) {
		case time.Time:
			timeVal = v
		case *time.Time:
			if v == nil {
				return nil
			}
			timeVal = *v
		default:
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
		var timeVal time.Time
		switch v := value.(type) {
		case time.Time:
			timeVal = v
		case *time.Time:
			if v == nil {
				return nil
			}
			timeVal = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a time.Time")
		}

		if !timeVal.Before(time.Now()) {
			return newValidationError(ErrorCodeOutOfRange, "must be a past date")
		}

		return nil
	}
}
