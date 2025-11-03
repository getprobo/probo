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
	"regexp"
	"strings"
)

var (
	alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	slugRegex         = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// MinLen validates that a string has at least the specified minimum length.
func MinLen(minLength int) ValidatorFunc {
	return func(value any) *ValidationError {
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
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
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
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

// Pattern validates that a string matches the specified regular expression pattern.
func Pattern(pattern string, message string) ValidatorFunc {
	regex := regexp.MustCompile(pattern)

	return func(value any) *ValidationError {
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if !regex.MatchString(str) {
			if message == "" {
				message = fmt.Sprintf("must match pattern: %s", pattern)
			}
			return newValidationError(ErrorCodeInvalidFormat, message)
		}

		return nil
	}
}

// AlphaNumeric validates that a string contains only letters and numbers.
func AlphaNumeric() ValidatorFunc {
	return func(value any) *ValidationError {
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		if !alphaNumericRegex.MatchString(str) {
			return newValidationError(ErrorCodeInvalidFormat, "must contain only letters and numbers")
		}

		return nil
	}
}

// NoSpaces validates that a string does not contain any spaces.
func NoSpaces() ValidatorFunc {
	return func(value any) *ValidationError {
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		if strings.Contains(str, " ") {
			return newValidationError(ErrorCodeInvalidFormat, "must not contain spaces")
		}

		return nil
	}
}

// Slug validates that a string is a valid URL slug (lowercase letters, numbers, and hyphens).
func Slug() ValidatorFunc {
	return func(value any) *ValidationError {
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		if !slugRegex.MatchString(str) {
			return newValidationError(ErrorCodeInvalidFormat, "must be a valid slug (lowercase letters, numbers, and hyphens)")
		}

		return nil
	}
}

// OneOf validates that a string is one of the allowed values.
func OneOf(allowed ...string) ValidatorFunc {
	allowedMap := make(map[string]bool)
	for _, v := range allowed {
		allowedMap[v] = true
	}

	return func(value any) *ValidationError {
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case *string:
			if v == nil {
				return nil
			}
			str = *v
		default:
			return newValidationError(ErrorCodeInvalidEnum, "value must be a string")
		}

		if !allowedMap[str] {
			return newValidationError(
				ErrorCodeInvalidEnum,
				fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
			)
		}

		return nil
	}
}
