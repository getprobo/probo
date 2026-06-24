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

// Package bearertoken parses Bearer tokens and builds Bearer WWW-Authenticate
// challenges according to RFC 6750.
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

// IsAttempt reports whether credentials use the Bearer scheme.
func IsAttempt(credentials string) bool {
	if len(credentials) < len(scheme) {
		return false
	}

	return strings.EqualFold(credentials[:len(scheme)], scheme)
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
