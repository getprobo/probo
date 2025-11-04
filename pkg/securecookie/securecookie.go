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

package securecookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidCookie    = errors.New("invalid cookie")
	ErrCookieNotFound   = errors.New("cookie not found")
	ErrInvalidSignature = errors.New("invalid signature")
)

// Config holds the configuration for secure cookies
type Config struct {
	// Name is the name of the cookie
	Name string

	// Secret is the secret key used for signing cookies
	Secret string

	// Domain is the cookie domain
	Domain string

	// Path is the cookie path
	Path string

	// MaxAge is the maximum age of the cookie in seconds
	MaxAge int

	// Secure indicates if the cookie should only be sent over HTTPS
	Secure bool

	// HTTPOnly indicates if the cookie should be inaccessible to JavaScript
	HTTPOnly bool

	// SameSite defines the SameSite attribute of the cookie
	SameSite http.SameSite
}

// DefaultConfig returns a default secure cookie configuration
func DefaultConfig(name, secret string, secure bool) Config {
	sameSite := http.SameSiteNoneMode // None mode required for SAML (cross-site POST from IdP)
	if !secure {
		sameSite = http.SameSiteLaxMode
	}

	return Config{
		Name:     name,
		Secret:   secret,
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		Secure:   secure,
		HTTPOnly: true,
		SameSite: sameSite,
	}
}

// Set creates and sets a secure cookie with the given value
func Set(w http.ResponseWriter, config Config, value string) error {
	signedValue, err := Sign(value, config.Secret)
	if err != nil {
		return fmt.Errorf("cannot sign cookie value: %w", err)
	}

	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    signedValue,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   config.MaxAge,
		Secure:   config.Secure,
		HttpOnly: config.HTTPOnly,
		SameSite: config.SameSite,
	}

	http.SetCookie(w, cookie)
	return nil
}

// Get retrieves and verifies a secure cookie
func Get(r *http.Request, config Config) (string, error) {
	cookie, err := r.Cookie(config.Name)
	if err != nil {
		return "", ErrCookieNotFound
	}

	value, err := Verify(cookie.Value, config.Secret)
	if err != nil {
		return "", ErrInvalidCookie
	}

	return value, nil
}

// Clear removes a cookie by setting its expiration in the past
func Clear(w http.ResponseWriter, config Config) {
	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    "",
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		Secure:   config.Secure,
		HttpOnly: config.HTTPOnly,
		SameSite: config.SameSite,
	}
	http.SetCookie(w, cookie)
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
