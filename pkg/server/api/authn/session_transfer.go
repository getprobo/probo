// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package authn

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const sessionTransferTTL = 60 * time.Second

var (
	ErrInvalidSessionTransferToken = errors.New("invalid session transfer token")
	ErrSessionTransferTokenExpired = errors.New("session transfer token expired")
)

// SignSessionTransfer creates a signed, time-limited token containing a
// session ID and the intended redirect URL. The token format is
// base64(sessionID:continueURL:timestamp).signature.
func SignSessionTransfer(sessionID string, continueURL string, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("cannot sign session transfer token: secret is empty")
	}

	payload := sessionID + ":" + continueURL + ":" + strconv.FormatInt(time.Now().Unix(), 10)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig, nil
}

// SessionTransferClaims holds the verified claims from a session transfer
// token.
type SessionTransferClaims struct {
	SessionID   string
	ContinueURL string
}

// VerifySessionTransfer verifies a session transfer token and returns
// the session ID and continue URL if the token is valid and not expired.
func VerifySessionTransfer(token string, secret string) (SessionTransferClaims, error) {
	if secret == "" {
		return SessionTransferClaims{}, fmt.Errorf("cannot verify session transfer token: secret is empty")
	}

	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return SessionTransferClaims{}, ErrInvalidSessionTransferToken
	}

	encoded, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return SessionTransferClaims{}, ErrInvalidSessionTransferToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return SessionTransferClaims{}, ErrInvalidSessionTransferToken
	}

	// Payload format: sessionID:continueURL:timestamp
	// Use LastIndex to find the timestamp separator (timestamp is always last).
	idx := strings.LastIndex(string(payload), ":")
	if idx < 0 {
		return SessionTransferClaims{}, ErrInvalidSessionTransferToken
	}

	tsStr := string(payload[idx+1:])
	rest := string(payload[:idx])

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return SessionTransferClaims{}, ErrInvalidSessionTransferToken
	}

	if time.Since(time.Unix(ts, 0)) > sessionTransferTTL {
		return SessionTransferClaims{}, ErrSessionTransferTokenExpired
	}

	// Split rest into sessionID and continueURL.
	before, after, ok := strings.Cut(rest, ":")
	if !ok {
		return SessionTransferClaims{}, ErrInvalidSessionTransferToken
	}

	return SessionTransferClaims{
		SessionID:   before,
		ContinueURL: after,
	}, nil
}
