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

// Package bearertoken parses Bearer tokens according to RFC 6750.
//
// The grammar is defined as:
//
//	b64token    = 1*( ALPHA / DIGIT / "-" / "." / "_" / "~" / "+" / "/" ) *"="
//	credentials = "Bearer" 1*SP b64token
package bearertoken

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidCredentials is returned when the credentials string is malformed.
	ErrInvalidCredentials = errors.New("invalid bearer credentials")

	// ErrMissingToken is returned when the token part is empty.
	ErrMissingToken = errors.New("missing bearer token")

	// ErrInvalidToken is returned when the token contains invalid characters.
	ErrInvalidToken = errors.New("invalid bearer token")
)

const (
	scheme = "Bearer"
)

// Parse extracts the b64token from a Bearer credentials string.
// The input must follow the format: "Bearer" 1*SP b64token
func Parse(credentials string) (string, error) {
	if len(credentials) <= len(scheme) {
		return "", ErrInvalidCredentials
	}

	if !strings.EqualFold(credentials[:len(scheme)], scheme) {
		return "", ErrInvalidCredentials
	}

	rest := credentials[len(scheme):]
	if len(rest) == 0 || rest[0] != ' ' {
		return "", ErrInvalidCredentials
	}

	// Skip all spaces (1*SP)
	token := strings.TrimLeft(rest, " ")
	if token == "" {
		return "", ErrMissingToken
	}

	if !isValidToken(token) {
		return "", ErrInvalidToken
	}

	return token, nil
}

// isValidToken checks if the given string is a valid b64token.
// A valid b64token consists of 1 or more characters from the set
// [A-Za-z0-9-._~+/] followed by zero or more '=' characters.
func isValidToken(token string) bool {
	if len(token) == 0 {
		return false
	}

	// Find where the padding starts (if any)
	paddingStart := strings.IndexByte(token, '=')
	if paddingStart == -1 {
		paddingStart = len(token)
	}

	// Must have at least one non-padding character
	if paddingStart == 0 {
		return false
	}

	// Validate the base part (before padding)
	for i := 0; i < paddingStart; i++ {
		if !isB64Char(token[i]) {
			return false
		}
	}

	// Validate padding (only '=' allowed after first '=')
	for i := paddingStart; i < len(token); i++ {
		if token[i] != '=' {
			return false
		}
	}

	return true
}

// isB64Char returns true if c is a valid b64token character (excluding padding).
// Valid characters: ALPHA / DIGIT / "-" / "." / "_" / "~" / "+" / "/"
func isB64Char(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '.' ||
		c == '_' ||
		c == '~' ||
		c == '+' ||
		c == '/'
}
