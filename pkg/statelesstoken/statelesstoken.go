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

package statelesstoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type (
	// Config holds the configuration for tokens
	Config struct {
		// Secret is the secret key used for signing tokens
		Secret string

		// ExpirationTime is the duration after which a token expires
		ExpirationTime time.Duration
	}

	// Payload is a generic token payload that can hold any data
	Payload[T any] struct {
		ExpiresAt time.Time `json:"exp"`
		IssuedAt  time.Time `json:"iat"`
		Type      string    `json:"typ"`
		Data      T         `json:"data"`
	}

	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken struct {
		message string
	}

	// ErrExpiredToken is returned when a token has expired
	ErrExpiredToken struct {
		message string
	}
)

var (
	DefaultExpirationTime = 1 * time.Hour
)

// Error implementations
func (e ErrInvalidToken) Error() string {
	return e.message
}

func (e ErrExpiredToken) Error() string {
	return e.message
}

func NewToken[T any](secret string, tokenType string, expirationTime time.Duration, data T) (string, error) {
	now := time.Now()
	return NewDeterministicToken(secret, tokenType, now.Add(expirationTime), now, data)
}

func NewDeterministicToken[T any](secret string, tokenType string, expiresAt time.Time, issuedAt time.Time, data T) (string, error) {
	payload := Payload[T]{
		ExpiresAt: expiresAt,
		IssuedAt:  issuedAt,
		Type:      tokenType,
		Data:      data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("cannot marshal token payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	tokenString := encodedPayload + "." + signature

	return tokenString, nil
}

// DecodePayload decodes the token payload without verifying the signature.
// This is useful when you need to inspect the payload to determine which
// secret to use for full validation (e.g., extracting the provider from
// an OAuth2 state token to look up the correct connector).
func DecodePayload[T any](tokenString string) (*Payload[T], error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 2 {
		return nil, &ErrInvalidToken{message: "invalid token format"}
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode token payload: %w", err)
	}

	var payload Payload[T]
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("cannot unmarshal token payload: %w", err)
	}

	return &payload, nil
}

// ValidateToken validates a token and unmarshals the payload
// It returns an error if the token is invalid or expired
func ValidateToken[T any](secret string, tokenType string, tokenString string) (*Payload[T], error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 2 {
		return nil, &ErrInvalidToken{message: "invalid token format"}
	}

	encodedPayload := parts[0]
	providedSignature := parts[1]

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(encodedPayload))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(providedSignature), []byte(expectedSignature)) != 1 {
		return nil, &ErrInvalidToken{message: "invalid token signature"}
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("cannot decode token payload: %w", err)
	}

	var payload Payload[T]
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("cannot unmarshal token payload: %w", err)
	}

	if time.Now().After(payload.ExpiresAt) {
		return nil, &ErrExpiredToken{message: "token has expired"}
	}

	if payload.Type != tokenType {
		return nil, &ErrInvalidToken{message: "invalid token type"}
	}

	return &payload, nil
}
