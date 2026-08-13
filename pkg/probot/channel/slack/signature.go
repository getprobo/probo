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

package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// VerifySignature implements Slack Events API request verification.
func VerifySignature(signingSecret, timestamp, signature string, body []byte) error {
	return VerifyAnySignature(timestamp, signature, body, signingSecret)
}

// VerifyAnySignature accepts a request signed by any of the given secrets so
// legacy Slack apps and Slackbot can share one interactive endpoint.
func VerifyAnySignature(
	timestamp, signature string,
	body []byte,
	signingSecrets ...string,
) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	if absUnixDelta(time.Now().Unix(), ts) > 5*60 {
		return fmt.Errorf("request timestamp too old")
	}

	matched := false
	seen := make(map[string]struct{}, len(signingSecrets))
	for _, signingSecret := range signingSecrets {
		if signingSecret == "" {
			continue
		}
		if _, ok := seen[signingSecret]; ok {
			continue
		}
		seen[signingSecret] = struct{}{}

		if signatureMatches(signingSecret, timestamp, signature, body) {
			matched = true
			break
		}
	}

	if !matched {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

func signatureMatches(signingSecret, timestamp, signature string, body []byte) bool {
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(baseString))

	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

func absUnixDelta(a, b int64) int64 {
	if a < b {
		return b - a
	}

	return a - b
}
