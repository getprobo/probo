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

	hasRequired := false
	hasOptional := false
	var actualValidators []ValidatorFunc

	for _, validator := range validators {
		testErr := validator(nil)
		if testErr == nil {
			optionalErr := validator("")
			if optionalErr == nil {
				hasOptional = true
				continue
			}
		}

		testReqErr := validator(nil)
		if testReqErr != nil && testReqErr.Code == ErrorCodeRequired {
			hasRequired = true
		}

		actualValidators = append(actualValidators, validator)
	}

	if hasRequired && hasOptional {
		panic("cannot use both Required() and Optional() on the same field")
	}

	if hasOptional {
		if value == nil {
			return
		}
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr && val.IsNil() {
			return
		}
	}

	for _, validator := range actualValidators {
		if err := validator(value); err != nil {
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

func (v *Validator) CheckNested(field string, fn func(v *Validator)) {
	nestedValidator := New()
	fn(nestedValidator)

	for _, err := range nestedValidator.errors {
		prefixedErr := &ValidationError{
			Field:   fmt.Sprintf("%s.%s", field, err.Field),
			Code:    err.Code,
			Message: err.Message,
			Value:   err.Value,
		}
		v.errors = append(v.errors, prefixedErr)
	}
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

func (v *Validator) Error() error {
	if len(v.errors) == 0 {
		return nil
	}
	return v.errors
}

type ValidatorFunc func(value any) *ValidationError
