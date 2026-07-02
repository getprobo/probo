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

package securetoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
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

	if subtle.ConstantTimeCompare([]byte(signedValue), []byte(expectedSignedValue)) != 1 {
		return "", ErrInvalidSignature
	}

	return value, nil
}
