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
