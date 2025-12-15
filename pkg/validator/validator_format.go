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
	"net/url"
	"regexp"

	"go.probo.inc/probo/pkg/gid"
)

var (
	uuidRegex   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
)

// URL validates that a string is a valid URL with http or https scheme.
func URL() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidURL, "value must be a string")
		}

		if str == "" {
			return nil
		}

		parsedURL, err := url.Parse(str)
		if err != nil {
			return newValidationError(ErrorCodeInvalidURL, "invalid URL format")
		}

		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return newValidationError(ErrorCodeInvalidURL, "URL must use http or https scheme")
		}

		if parsedURL.Host == "" {
			return newValidationError(ErrorCodeInvalidURL, "URL must have a host")
		}

		return nil
	}
}

// HTTPUrl validates that a string is a valid HTTP URL (not HTTPS).
func HTTPUrl() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidURL, "value must be a string")
		}

		if str == "" {
			return nil
		}

		parsedURL, err := url.Parse(str)
		if err != nil {
			return newValidationError(ErrorCodeInvalidURL, "invalid URL format")
		}

		if parsedURL.Scheme != "http" {
			return newValidationError(ErrorCodeInvalidURL, "URL must use http scheme")
		}

		if parsedURL.Host == "" {
			return newValidationError(ErrorCodeInvalidURL, "URL must have a host")
		}

		return nil
	}
}

// HTTPSUrl validates that a string is a valid HTTPS URL (not HTTP).
func HTTPSUrl() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidURL, "value must be a string")
		}

		if str == "" {
			return nil
		}

		parsedURL, err := url.Parse(str)
		if err != nil {
			return newValidationError(ErrorCodeInvalidURL, "invalid URL format")
		}

		if parsedURL.Scheme != "https" {
			return newValidationError(ErrorCodeInvalidURL, "URL must use https scheme")
		}

		if parsedURL.Host == "" {
			return newValidationError(ErrorCodeInvalidURL, "URL must have a host")
		}

		return nil
	}
}

// UUID validates that a string is a valid UUID.
func UUID() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		if !uuidRegex.MatchString(str) {
			return newValidationError(ErrorCodeInvalidFormat, "invalid UUID format")
		}

		return nil
	}
}

// GID validates that a string is a valid GID using gid.ParseGID.
// Optionally validates the entity type if provided.
//
// Example usage:
//   - GID() validates any GID format
//   - GID(100) validates GID with entity type 100
//   - GID(100, 200) validates GID with entity type 100 or 200
func GID(entityTypes ...uint16) ValidatorFunc {
	return func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		var gidValue gid.GID

		switch v := value.(type) {
		case gid.GID:
			gidValue = v
		case *gid.GID:
			if v == nil {
				return nil
			}
			gidValue = *v
		default:
			return newValidationError(ErrorCodeInvalidGID, "value must be a GID")
		}

		if len(entityTypes) > 0 {
			parsedEntityType := gidValue.EntityType()
			valid := false
			for _, expected := range entityTypes {
				if parsedEntityType == expected {
					valid = true
					break
				}
			}
			if !valid {
				return newValidationError(ErrorCodeInvalidGID, "GID has invalid entity type")
			}
		}

		return nil
	}
}

// Domain validates that a string is a valid domain name.
func Domain() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		if len(str) > 253 {
			return newValidationError(ErrorCodeInvalidFormat, "domain name too long (max 253 characters)")
		}

		if !domainRegex.MatchString(str) {
			return newValidationError(ErrorCodeInvalidFormat, "invalid domain name format")
		}

		return nil
	}
}
