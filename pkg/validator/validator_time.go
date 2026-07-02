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
