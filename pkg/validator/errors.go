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
	"strings"
)

type ErrorCode string

const (
	ErrorCodeRequired      ErrorCode = "REQUIRED"
	ErrorCodeInvalidFormat ErrorCode = "INVALID_FORMAT"
	ErrorCodeOutOfRange    ErrorCode = "OUT_OF_RANGE"
	ErrorCodeTooShort      ErrorCode = "TOO_SHORT"
	ErrorCodeTooLong       ErrorCode = "TOO_LONG"
	ErrorCodeInvalidEmail  ErrorCode = "INVALID_EMAIL"
	ErrorCodeInvalidURL    ErrorCode = "INVALID_URL"
	ErrorCodeInvalidEnum   ErrorCode = "INVALID_ENUM"
	ErrorCodeInvalidGID    ErrorCode = "INVALID_GID"
	ErrorCodeUnsafeContent ErrorCode = "UNSAFE_CONTENT"
	ErrorCodeCustom        ErrorCode = "CUSTOM"
)

type ValidationError struct {
	Field   string
	Code    ErrorCode
	Message string
	Value   any
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Message)
}

type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}

func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

func (ve ValidationErrors) Fields() []string {
	fields := make([]string, 0, len(ve))
	for _, err := range ve {
		fields = append(fields, err.Field)
	}

	return fields
}

func (ve ValidationErrors) ByField(field string) ValidationErrors {
	var errors ValidationErrors

	for _, err := range ve {
		if err.Field == field {
			errors = append(errors, err)
		}
	}

	return errors
}

func (ve ValidationErrors) ByCode(code ErrorCode) ValidationErrors {
	var errors ValidationErrors

	for _, err := range ve {
		if err.Code == code {
			errors = append(errors, err)
		}
	}

	return errors
}

func (ve ValidationErrors) First() *ValidationError {
	if len(ve) == 0 {
		return nil
	}

	return ve[0]
}

func newValidationError(code ErrorCode, message string) *ValidationError {
	return &ValidationError{
		Code:    code,
		Message: message,
	}
}
