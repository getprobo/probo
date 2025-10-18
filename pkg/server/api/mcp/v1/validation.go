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

package v1

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidationError represents a validation error with field information
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidateRequired checks if a required field is present
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(field, "is required")
	}
	return nil
}

// ValidateURL checks if a URL is valid
func ValidateURL(field, urlStr string) error {
	if urlStr == "" {
		return nil // Optional field
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return NewValidationError(field, fmt.Sprintf("invalid URL: %v", err))
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return NewValidationError(field, "URL must have a scheme and host")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return NewValidationError(field, "URL scheme must be http or https")
	}

	return nil
}

// ValidateRange checks if an integer is within the specified range
func ValidateRange(field string, value, min, max int) error {
	if value < min || value > max {
		return NewValidationError(field, fmt.Sprintf("must be between %d and %d", min, max))
	}
	return nil
}

// ValidateEnum checks if a value is in the allowed set
func ValidateEnum(field, value string, allowedValues []string) error {
	if value == "" {
		return nil // Optional field
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return NewValidationError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", ")))
}

// ValidateMaxLength checks if a string exceeds the maximum length
func ValidateMaxLength(field, value string, maxLength int) error {
	if len(value) > maxLength {
		return NewValidationError(field, fmt.Sprintf("exceeds maximum length of %d characters", maxLength))
	}
	return nil
}

// ValidationErrors collects multiple validation errors
type ValidationErrors []error

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}
