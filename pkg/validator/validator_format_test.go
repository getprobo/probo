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
	"strings"
	"testing"

	"getprobo.com/go/pkg/gid"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with plus", "test+tag@example.com", false},
		{"invalid email no @", "testexample.com", true},
		{"invalid email no domain", "test@", true},
		{"invalid email no TLD", "test@example", true},
		{"empty string", "", false},            // Empty is allowed, use Required() to enforce
		{"nil pointer", (*string)(nil), false}, // Skip validation
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Email()(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Email() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && err.Code != ErrorCodeInvalidEmail {
				t.Errorf("Expected error code %s, got %s", ErrorCodeInvalidEmail, err.Code)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantError bool
	}{
		{"valid http URL", "http://example.com", false},
		{"valid https URL", "https://example.com", false},
		{"valid URL with path", "https://example.com/path", false},
		{"invalid scheme", "ftp://example.com", true},
		{"no scheme", "example.com", true},
		{"no host", "https://", true},
		{"empty string", "", false}, // Empty is allowed
		{"nil pointer", (*string)(nil), false},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := URL()(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("URL() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && err.Code != ErrorCodeInvalidURL {
				t.Errorf("Expected error code %s, got %s", ErrorCodeInvalidURL, err.Code)
			}
		})
	}
}

func TestHTTPUrl(t *testing.T) {
	t.Run("valid http URL", func(t *testing.T) {
		str := "http://example.com"
		err := HTTPUrl()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid http URL with path", func(t *testing.T) {
		str := "http://example.com/path/to/resource"
		err := HTTPUrl()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid http URL with query", func(t *testing.T) {
		str := "http://example.com?foo=bar"
		err := HTTPUrl()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - https scheme", func(t *testing.T) {
		str := "https://example.com"
		err := HTTPUrl()(&str)
		if err == nil {
			t.Error("expected validation error for https")
		}
		if err.Message != "URL must use http scheme" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - no scheme", func(t *testing.T) {
		str := "example.com"
		err := HTTPUrl()(&str)
		if err == nil {
			t.Error("expected validation error for missing scheme")
		}
	})

	t.Run("invalid - no host", func(t *testing.T) {
		str := "http://"
		err := HTTPUrl()(&str)
		if err == nil {
			t.Error("expected validation error for missing host")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := HTTPUrl()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := HTTPUrl()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestHTTPSUrl(t *testing.T) {
	t.Run("valid https URL", func(t *testing.T) {
		str := "https://example.com"
		err := HTTPSUrl()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid https URL with path", func(t *testing.T) {
		str := "https://example.com/path/to/resource"
		err := HTTPSUrl()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid https URL with query", func(t *testing.T) {
		str := "https://api.example.com/v1/users?page=1"
		err := HTTPSUrl()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - http scheme", func(t *testing.T) {
		str := "http://example.com"
		err := HTTPSUrl()(&str)
		if err == nil {
			t.Error("expected validation error for http")
		}
		if err.Message != "URL must use https scheme" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - ftp scheme", func(t *testing.T) {
		str := "ftp://example.com"
		err := HTTPSUrl()(&str)
		if err == nil {
			t.Error("expected validation error for ftp")
		}
	})

	t.Run("invalid - no scheme", func(t *testing.T) {
		str := "example.com"
		err := HTTPSUrl()(&str)
		if err == nil {
			t.Error("expected validation error for missing scheme")
		}
	})

	t.Run("invalid - no host", func(t *testing.T) {
		str := "https://"
		err := HTTPSUrl()(&str)
		if err == nil {
			t.Error("expected validation error for missing host")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := HTTPSUrl()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := HTTPSUrl()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestDomain(t *testing.T) {
	t.Run("valid domain", func(t *testing.T) {
		str := "example.com"
		err := Domain()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid subdomain", func(t *testing.T) {
		str := "api.example.com"
		err := Domain()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid nested subdomain", func(t *testing.T) {
		str := "api.v1.example.com"
		err := Domain()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid domain with hyphens", func(t *testing.T) {
		str := "my-api.example-site.com"
		err := Domain()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("single word domain", func(t *testing.T) {
		str := "localhost"
		err := Domain()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - starts with hyphen", func(t *testing.T) {
		str := "-example.com"
		err := Domain()(&str)
		if err == nil {
			t.Error("expected validation error for domain starting with hyphen")
		}
	})

	t.Run("invalid - ends with hyphen", func(t *testing.T) {
		str := "example-.com"
		err := Domain()(&str)
		if err == nil {
			t.Error("expected validation error for domain ending with hyphen")
		}
	})

	t.Run("invalid - contains underscore", func(t *testing.T) {
		str := "example_site.com"
		err := Domain()(&str)
		if err == nil {
			t.Error("expected validation error for underscore")
		}
	})

	t.Run("invalid - contains spaces", func(t *testing.T) {
		str := "example site.com"
		err := Domain()(&str)
		if err == nil {
			t.Error("expected validation error for spaces")
		}
	})

	t.Run("invalid - empty label", func(t *testing.T) {
		str := "example..com"
		err := Domain()(&str)
		if err == nil {
			t.Error("expected validation error for empty label")
		}
	})

	t.Run("invalid - too long", func(t *testing.T) {
		str := strings.Repeat("a", 254)
		err := Domain()(&str)
		if err == nil {
			t.Error("expected validation error for domain too long")
		}
		if err.Message != "domain name too long (max 253 characters)" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := Domain()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := Domain()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestGID(t *testing.T) {
	// Create a valid GID for testing
	tenantID := gid.TenantID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	validGID := gid.New(tenantID, 100)
	validGIDString := validGID.String()

	anotherGID := gid.New(tenantID, 200)
	anotherGIDString := anotherGID.String()

	t.Run("valid GID string - no entity type validation", func(t *testing.T) {
		err := GID()(validGIDString)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid GID type - no entity type validation", func(t *testing.T) {
		err := GID()(validGID)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid GID string - with matching entity type", func(t *testing.T) {
		err := GID(100)(validGIDString)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid GID type - with matching entity type", func(t *testing.T) {
		err := GID(100)(validGID)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid GID string - with multiple entity types", func(t *testing.T) {
		err := GID(100, 200, 300)(validGIDString)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - wrong entity type", func(t *testing.T) {
		err := GID(200)(validGIDString)
		if err == nil {
			t.Error("expected validation error for wrong entity type")
		}
		if err.Code != ErrorCodeInvalidGID {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidGID, err.Code)
		}
		if err.Message != "GID has invalid entity type" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - wrong entity type with multiple options", func(t *testing.T) {
		err := GID(200, 300)(validGIDString)
		if err == nil {
			t.Error("expected validation error for wrong entity type")
		}
	})

	t.Run("valid - entity type matches one of multiple options", func(t *testing.T) {
		err := GID(99, 100, 101)(validGIDString)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid GID format", func(t *testing.T) {
		err := GID()("not-a-valid-gid")
		if err == nil {
			t.Error("expected validation error for invalid GID format")
		}
		if err.Code != ErrorCodeInvalidGID {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidGID, err.Code)
		}
		if err.Message != "invalid GID format" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - too short", func(t *testing.T) {
		err := GID()("abc123")
		if err == nil {
			t.Error("expected validation error for short string")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		err := GID()("")
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := GID()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})

	t.Run("non-string non-GID type", func(t *testing.T) {
		err := GID()(123)
		if err == nil {
			t.Error("expected validation error for non-string type")
		}
		if err.Message != "value must be a string or GID" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("string pointer with valid GID", func(t *testing.T) {
		str := validGIDString
		err := GID()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("string pointer with entity type validation", func(t *testing.T) {
		str := validGIDString
		err := GID(100)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}
