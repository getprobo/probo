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
)

// MinItems validates that a slice or array has at least the specified minimum number of items.
func MinItems(min int) ValidatorFunc {
	return func(value any) *ValidationError {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a slice or array")
		}

		if v.Len() < min {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must contain at least %d items", min),
			)
		}

		return nil
	}
}

// MaxItems validates that a slice or array does not exceed the specified maximum number of items.
func MaxItems(max int) ValidatorFunc {
	return func(value any) *ValidationError {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a slice or array")
		}

		if v.Len() > max {
			return newValidationError(
				ErrorCodeOutOfRange,
				fmt.Sprintf("must contain at most %d items", max),
			)
		}

		return nil
	}
}

// UniqueItems validates that all items in a slice or array are unique.
func UniqueItems() ValidatorFunc {
	return func(value any) *ValidationError {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a slice or array")
		}

		if v.Len() == 0 {
			return nil
		}

		elemType := v.Type().Elem()
		if !elemType.Comparable() {
			return newValidationError(ErrorCodeInvalidFormat, "cannot validate uniqueness for non-comparable types")
		}

		seen := make(map[any]bool)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			if seen[item] {
				return newValidationError(ErrorCodeInvalidFormat, "items must be unique")
			}
			seen[item] = true
		}

		return nil
	}
}
