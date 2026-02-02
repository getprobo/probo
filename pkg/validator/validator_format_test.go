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

	"go.probo.inc/probo/pkg/gid"
)

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
			t.Fatal("expected validation error for https")
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
			t.Fatal("expected validation error for http")
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
			t.Fatal("expected validation error for domain too long")
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

	t.Run("valid GID type - no entity type validation", func(t *testing.T) {
		err := GID()(validGID)
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

	t.Run("valid GID type - with multiple entity types", func(t *testing.T) {
		err := GID(100, 200, 300)(validGID)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - wrong entity type", func(t *testing.T) {
		err := GID(200)(validGID)
		if err == nil {
			t.Fatal("expected validation error for wrong entity type")
		}
		if err.Code != ErrorCodeInvalidGID {
			t.Errorf("expected error code %s, got %s", ErrorCodeInvalidGID, err.Code)
		}
		if err.Message != "GID has invalid entity type" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - wrong entity type with multiple options", func(t *testing.T) {
		err := GID(200, 300)(validGID)
		if err == nil {
			t.Error("expected validation error for wrong entity type")
		}
	})

	t.Run("valid - entity type matches one of multiple options", func(t *testing.T) {
		err := GID(99, 100, 101)(validGID)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("nil GID pointer", func(t *testing.T) {
		var gidPtr *gid.GID
		err := GID()(gidPtr)
		if err != nil {
			t.Errorf("expected no error for nil GID pointer, got: %v", err)
		}
	})

	t.Run("valid GID pointer", func(t *testing.T) {
		err := GID()(&validGID)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid GID pointer with entity type validation", func(t *testing.T) {
		err := GID(100)(&validGID)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("non-GID type", func(t *testing.T) {
		err := GID()(123)
		if err == nil {
			t.Fatal("expected validation error for non-GID type")
		}
		if err.Message != "value must be a GID" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("string type not supported", func(t *testing.T) {
		err := GID()("some-string")
		if err == nil {
			t.Fatal("expected validation error for string type")
		}
		if err.Message != "value must be a GID" {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})
}
