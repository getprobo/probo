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
	"fmt"
	"net/url"
	"strings"
)

// BaseURL represents a validated base URL for the application.
// It provides convenient methods for building URLs with paths and query parameters.
type BaseURL struct {
	raw    string
	parsed *url.URL
}

// Parse creates a new BaseURL from a string, validating that it's a valid absolute URL.
func Parse(rawURL string) (*BaseURL, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if !parsed.IsAbs() {
		return nil, fmt.Errorf("base URL must be absolute (include scheme)")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("base URL scheme must be http or https, got: %s", parsed.Scheme)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("base URL must include a host")
	}

	return &BaseURL{
		raw:    rawURL,
		parsed: parsed,
	}, nil
}

// MustParse creates a new BaseURL from a string, panicking if it's invalid.
// This should only be used in tests or with known-valid URLs.
func MustParse(rawURL string) *BaseURL {
	b, err := Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return b
}

// String returns the base URL as a string.
func (b *BaseURL) String() string {
	if b == nil {
		return ""
	}
	return b.raw
}

// Scheme returns the URL scheme (http or https).
func (b *BaseURL) Scheme() string {
	if b == nil || b.parsed == nil {
		return ""
	}
	return b.parsed.Scheme
}

// Host returns the host:port portion of the URL.
func (b *BaseURL) Host() string {
	if b == nil || b.parsed == nil {
		return ""
	}
	return b.parsed.Host
}

// Hostname returns just the hostname without the port.
func (b *BaseURL) Hostname() string {
	if b == nil || b.parsed == nil {
		return ""
	}
	return b.parsed.Hostname()
}

// Port returns the port portion of the URL, or empty string if not specified.
func (b *BaseURL) Port() string {
	if b == nil || b.parsed == nil {
		return ""
	}
	return b.parsed.Port()
}

// URLBuilder provides a fluent interface for building URLs.
type URLBuilder struct {
	base  *BaseURL
	path  string
	query url.Values
	err   error
}

// WithPath returns a URLBuilder with the specified path.
// The path will be properly joined with the base URL.
func (b *BaseURL) WithPath(path string) *URLBuilder {
	if b == nil {
		return &URLBuilder{err: fmt.Errorf("base URL is nil")}
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Ensure path does not end with /
	path = strings.TrimSuffix(path, "/")

	return &URLBuilder{
		base:  b,
		path:  path,
		query: make(url.Values),
	}
}

// AppendPath returns a URLBuilder with the specified path.
// The path will be properly joined with the base URL's path.
func (b *BaseURL) AppendPath(path string) *URLBuilder {
	if b == nil {
		return &URLBuilder{err: fmt.Errorf("base URL is nil")}
	}

	basePath := strings.TrimSuffix(b.parsed.Path, "/")

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	path = basePath + path

	// Ensure path does not end with /
	path = strings.TrimSuffix(path, "/")

	return &URLBuilder{
		base:  b,
		path:  path,
		query: make(url.Values),
	}
}

// WithQuery adds a query parameter to the URL.
func (ub *URLBuilder) WithQuery(key, value string) *URLBuilder {
	if ub.err != nil {
		return ub
	}
	ub.query.Add(key, value)
	return ub
}

// WithQueryValues sets multiple query parameters at once.
func (ub *URLBuilder) WithQueryValues(values url.Values) *URLBuilder {
	if ub.err != nil {
		return ub
	}
	for key, vals := range values {
		for _, val := range vals {
			ub.query.Add(key, val)
		}
	}
	return ub
}

// String builds and returns the final URL string.
func (ub *URLBuilder) String() (string, error) {
	if ub.err != nil {
		return "", ub.err
	}

	u := &url.URL{
		Scheme:   ub.base.Scheme(),
		Host:     ub.base.Host(),
		Path:     ub.path,
		RawQuery: ub.query.Encode(),
	}

	return u.String(), nil
}

// MustString builds and returns the final URL string, panicking on error.
// This should only be used when you're certain the URL is valid.
func (ub *URLBuilder) MustString() string {
	s, err := ub.String()
	if err != nil {
		panic(err)
	}
	return s
}

// UnmarshalJSON implements json.Unmarshaler for BaseURL.
func (b *BaseURL) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := Parse(s)
	if err != nil {
		return err
	}

	*b = *parsed
	return nil
}

// MarshalJSON implements json.Marshaler for BaseURL.
func (b *BaseURL) MarshalJSON() ([]byte, error) {
	if b == nil {
		return json.Marshal("")
	}
	return json.Marshal(b.raw)
}

// UnmarshalText implements encoding.TextUnmarshaler for BaseURL.
func (b *BaseURL) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*b = *parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler for BaseURL.
func (b *BaseURL) MarshalText() ([]byte, error) {
	if b == nil {
		return []byte(""), nil
	}
	return []byte(b.raw), nil
}

func (b *URLBuilder) URL() url.URL {
	return url.URL{
		Scheme:   b.base.Scheme(),
		Host:     b.base.Host(),
		Path:     b.path,
		RawQuery: b.query.Encode(),
	}
}
