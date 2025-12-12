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

package statelesstoken

import (
	"crypto/hmac"
	"crypto/sha256"
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

	if providedSignature != expectedSignature {
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
