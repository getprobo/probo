// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package security

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckHeader(t *testing.T) {
	t.Parallel()

	t.Run(
		"present header returns present true and value",
		func(t *testing.T) {
			t.Parallel()

			h := http.Header{}
			h.Set("X-Frame-Options", "DENY")

			result := checkHeader(h, "X-Frame-Options")

			assert.True(t, result.Present)
			assert.Equal(t, "DENY", result.Value)
		},
	)

	t.Run(
		"missing header returns present false",
		func(t *testing.T) {
			t.Parallel()

			h := http.Header{}

			result := checkHeader(h, "X-Frame-Options")

			assert.False(t, result.Present)
			assert.Equal(t, "", result.Value)
		},
	)

	t.Run(
		"empty header map returns present false",
		func(t *testing.T) {
			t.Parallel()

			result := checkHeader(http.Header{}, "Strict-Transport-Security")

			assert.False(t, result.Present)
			assert.Equal(t, "", result.Value)
		},
	)

	t.Run(
		"header lookup is case insensitive",
		func(t *testing.T) {
			t.Parallel()

			h := http.Header{}
			h.Set("content-security-policy", "default-src 'self'")

			result := checkHeader(h, "Content-Security-Policy")

			assert.True(t, result.Present)
			assert.Equal(t, "default-src 'self'", result.Value)
		},
	)
}

func TestHeadersFromResponse(t *testing.T) {
	t.Parallel()

	t.Run(
		"all security headers present",
		func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{
					"Strict-Transport-Security":    {"max-age=31536000; includeSubDomains"},
					"Content-Security-Policy":      {"default-src 'self'"},
					"X-Frame-Options":              {"DENY"},
					"X-Content-Type-Options":       {"nosniff"},
					"Referrer-Policy":              {"strict-origin-when-cross-origin"},
					"Permissions-Policy":           {"camera=(), microphone=()"},
					"Cross-Origin-Opener-Policy":   {"same-origin"},
					"Cross-Origin-Embedder-Policy": {"require-corp"},
					"Cross-Origin-Resource-Policy": {"same-origin"},
				},
			}

			result := headersFromResponse(resp)

			assert.True(t, result.HSTS.Present)
			assert.Equal(t, "max-age=31536000; includeSubDomains", result.HSTS.Value)
			assert.True(t, result.CSP.Present)
			assert.Equal(t, "default-src 'self'", result.CSP.Value)
			assert.True(t, result.XFrameOptions.Present)
			assert.Equal(t, "DENY", result.XFrameOptions.Value)
			assert.True(t, result.XContentTypeOptions.Present)
			assert.Equal(t, "nosniff", result.XContentTypeOptions.Value)
			assert.True(t, result.ReferrerPolicy.Present)
			assert.Equal(t, "strict-origin-when-cross-origin", result.ReferrerPolicy.Value)
			assert.True(t, result.PermissionsPolicy.Present)
			assert.Equal(t, "camera=(), microphone=()", result.PermissionsPolicy.Value)
			assert.True(t, result.CrossOriginOpenerPolicy.Present)
			assert.Equal(t, "same-origin", result.CrossOriginOpenerPolicy.Value)
			assert.True(t, result.CrossOriginEmbedderPolicy.Present)
			assert.Equal(t, "require-corp", result.CrossOriginEmbedderPolicy.Value)
			assert.True(t, result.CrossOriginResourcePolicy.Present)
			assert.Equal(t, "same-origin", result.CrossOriginResourcePolicy.Value)
		},
	)

	t.Run(
		"no security headers present",
		func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{},
			}

			result := headersFromResponse(resp)

			assert.False(t, result.HSTS.Present)
			assert.False(t, result.CSP.Present)
			assert.False(t, result.XFrameOptions.Present)
			assert.False(t, result.XContentTypeOptions.Present)
			assert.False(t, result.ReferrerPolicy.Present)
			assert.False(t, result.PermissionsPolicy.Present)
			assert.False(t, result.CrossOriginOpenerPolicy.Present)
			assert.False(t, result.CrossOriginEmbedderPolicy.Present)
			assert.False(t, result.CrossOriginResourcePolicy.Present)
			assert.False(t, result.RedirectsToHTTPS)
		},
	)

	t.Run(
		"partial headers present",
		func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{
					"Strict-Transport-Security": {"max-age=86400"},
					"X-Content-Type-Options":    {"nosniff"},
				},
			}

			result := headersFromResponse(resp)

			assert.True(t, result.HSTS.Present)
			assert.Equal(t, "max-age=86400", result.HSTS.Value)
			assert.False(t, result.CSP.Present)
			assert.False(t, result.XFrameOptions.Present)
			assert.True(t, result.XContentTypeOptions.Present)
			assert.Equal(t, "nosniff", result.XContentTypeOptions.Value)
			assert.False(t, result.ReferrerPolicy.Present)
			assert.False(t, result.PermissionsPolicy.Present)
			assert.False(t, result.CrossOriginOpenerPolicy.Present)
			assert.False(t, result.CrossOriginEmbedderPolicy.Present)
			assert.False(t, result.CrossOriginResourcePolicy.Present)
		},
	)

	t.Run(
		"does not set redirects to https",
		func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{
					"Strict-Transport-Security": {"max-age=31536000"},
				},
			}

			result := headersFromResponse(resp)

			assert.False(t, result.RedirectsToHTTPS)
		},
	)
}
