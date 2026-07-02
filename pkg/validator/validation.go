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
	"reflect"
)

type Validator struct {
	errors ValidationErrors
}

func New() *Validator {
	return &Validator{
		errors: ValidationErrors{},
	}
}

func (v *Validator) Check(value any, field string, validators ...ValidatorFunc) {
	if len(validators) == 0 {
		return
	}

	// Dereference pointer values to get the actual value for validation
	actualValue := value
	if value != nil {
		val := reflect.ValueOf(value)
		// Dereference all pointer levels
		for val.Kind() == reflect.Pointer && !val.IsNil() {
			val = val.Elem()
			actualValue = val.Interface()
		}

		// If we ended up with a nil pointer at any level, set actualValue to nil
		if val.Kind() == reflect.Pointer && val.IsNil() {
			actualValue = nil
		}
	}

	for _, validator := range validators {
		if err := validator(actualValue); err != nil {
			v.errors = append(v.errors, &ValidationError{
				Field:   field,
				Code:    err.Code,
				Message: err.Message,
				Value:   value,
			})
		}
	}
}

func (v *Validator) CheckEach(items any, field string, fn func(index int, item any)) {
	if items == nil {
		return
	}

	if slice, ok := items.([]any); ok {
		for i, item := range slice {
			fn(i, item)
		}

		return
	}

	val := reflect.ValueOf(items)
	// Dereference pointer levels to get to the actual slice
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}

		val = val.Elem()
	}

	if val.Kind() != reflect.Slice {
		v.errors = append(v.errors, &ValidationError{
			Field:   field,
			Code:    ErrorCodeInvalidFormat,
			Message: "expected a slice",
			Value:   items,
		})

		return
	}

	for i := 0; i < val.Len(); i++ {
		fn(i, val.Index(i).Interface())
	}
}

func (v *Validator) Error() error {
	if len(v.errors) == 0 {
		return nil
	}

	return v.errors
}

type ValidatorFunc func(value any) *ValidationError

// dereferenceValue recursively dereferences all pointer levels.
// Returns the final dereferenced value and a boolean indicating if any pointer in the chain was nil.
func dereferenceValue(value any) (any, bool) {
	if value == nil {
		return nil, true
	}

	val := reflect.ValueOf(value)
	// Dereference all pointer levels
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, true
		}

		val = val.Elem()
	}

	return val.Interface(), false
}
