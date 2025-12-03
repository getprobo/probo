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

package securetoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.probo.inc/probo/pkg/bearertoken"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidSignature = errors.New("invalid signature")
)

func Get(req *http.Request, secret string) (string, error) {
	v := req.Header.Get("Authorization")
	if v == "" {
		return "", ErrTokenNotFound
	}

	token, err := bearertoken.Parse(v)
	if err != nil {
		return "", ErrInvalidToken
	}

	value, err := Verify(token, secret)
	if err != nil {
		return "", ErrInvalidToken
	}

	return value, nil
}

// Sign creates a signed value using HMAC-SHA256
func Sign(value, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(value))

	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return value + "." + signature, nil
}

// Verify checks if a signed value is valid
func Verify(signedValue, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}

	parts := strings.Split(signedValue, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid signed value format")
	}

	value := parts[0]

	expectedSignedValue, err := Sign(value, secret)
	if err != nil {
		return "", fmt.Errorf("cannot sign value: %w", err)
	}

	if signedValue != expectedSignedValue {
		return "", ErrInvalidSignature
	}

	return value, nil
}
