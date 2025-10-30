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

package baseurl

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			input:   "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			input:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid https URL with port",
			input:   "https://example.com:8443",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "relative URL",
			input:   "/path/to/resource",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			input:   "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "no host",
			input:   "http://",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			input:   "ht!tp://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Parse() returned nil without error")
			}
		})
	}
}

func TestBaseURL_Accessors(t *testing.T) {
	b := MustParse("https://example.com:8443")

	if got := b.String(); got != "https://example.com:8443" {
		t.Errorf("String() = %v, want %v", got, "https://example.com:8443")
	}

	if got := b.Scheme(); got != "https" {
		t.Errorf("Scheme() = %v, want %v", got, "https")
	}

	if got := b.Host(); got != "example.com:8443" {
		t.Errorf("Host() = %v, want %v", got, "example.com:8443")
	}

	if got := b.Hostname(); got != "example.com" {
		t.Errorf("Hostname() = %v, want %v", got, "example.com")
	}

	if got := b.Port(); got != "8443" {
		t.Errorf("Port() = %v, want %v", got, "8443")
	}
}

func TestBaseURL_WithPath(t *testing.T) {
	b := MustParse("https://example.com")

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "path with leading slash",
			path: "/auth/login",
			want: "https://example.com/auth/login",
		},
		{
			name: "path without leading slash",
			path: "auth/login",
			want: "https://example.com/auth/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.WithPath(tt.path).String()
			if err != nil {
				t.Errorf("WithPath().String() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("WithPath().String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseURL_WithQuery(t *testing.T) {
	b := MustParse("https://example.com")

	got, err := b.WithPath("/search").
		WithQuery("q", "test").
		WithQuery("limit", "10").
		String()

	if err != nil {
		t.Fatalf("WithPath().WithQuery().String() error = %v", err)
	}

	// Parse the result to check query parameters
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("Failed to parse result URL: %v", err)
	}

	if parsed.Query().Get("q") != "test" {
		t.Errorf("Query param 'q' = %v, want %v", parsed.Query().Get("q"), "test")
	}

	if parsed.Query().Get("limit") != "10" {
		t.Errorf("Query param 'limit' = %v, want %v", parsed.Query().Get("limit"), "10")
	}
}

func TestBaseURL_WithQueryValues(t *testing.T) {
	b := MustParse("https://example.com")

	values := url.Values{}
	values.Add("foo", "bar")
	values.Add("baz", "qux")

	got, err := b.WithPath("/test").WithQueryValues(values).String()
	if err != nil {
		t.Fatalf("WithPath().WithQueryValues().String() error = %v", err)
	}

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("Failed to parse result URL: %v", err)
	}

	if parsed.Query().Get("foo") != "bar" {
		t.Errorf("Query param 'foo' = %v, want %v", parsed.Query().Get("foo"), "bar")
	}

	if parsed.Query().Get("baz") != "qux" {
		t.Errorf("Query param 'baz' = %v, want %v", parsed.Query().Get("baz"), "qux")
	}
}

func TestBaseURL_JSON(t *testing.T) {
	original := MustParse("https://example.com:8443")

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal
	var restored BaseURL
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if restored.String() != original.String() {
		t.Errorf("After JSON round-trip: got %v, want %v", restored.String(), original.String())
	}
}

func TestBaseURL_NilSafety(t *testing.T) {
	var b *BaseURL

	if got := b.String(); got != "" {
		t.Errorf("nil.String() = %v, want empty string", got)
	}

	if got := b.Scheme(); got != "" {
		t.Errorf("nil.Scheme() = %v, want empty string", got)
	}

	if got := b.Host(); got != "" {
		t.Errorf("nil.Host() = %v, want empty string", got)
	}

	if got := b.Hostname(); got != "" {
		t.Errorf("nil.Hostname() = %v, want empty string", got)
	}

	if got := b.Port(); got != "" {
		t.Errorf("nil.Port() = %v, want empty string", got)
	}

	builder := b.WithPath("/test")
	if _, err := builder.String(); err == nil {
		t.Error("nil.WithPath().String() expected error, got nil")
	}
}
