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

package console_v1

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"strings"
)

// crispVerificationCodeLength bounds the human-typeable verification code. 12
// base32 characters carry ~60 bits, far beyond guessing given the code is only
// a proof-of-control challenge (not a secret) and is compared server side.
const crispVerificationCodeLength = 12

// computeCrispVerificationCode derives the deterministic ownership-verification
// code Probo shows for a (organization, Crisp website) pair. It is
// HMAC(tokenSecret, domain || orgID || websiteID) so only Probo can mint it and
// the value is bound to BOTH the organization and the website: a code minted for
// one organization cannot verify a website under another, and a code for one
// website cannot verify another. Nothing is stored: the same inputs always yield
// the same code, so the create-time check re-derives and compares it.
func computeCrispVerificationCode(secret, orgID, websiteID string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("probo/connector/crisp-verification:" + orgID + ":" + websiteID))

	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(mac.Sum(nil))

	return code[:crispVerificationCodeLength]
}

// verifyCrispVerificationCode reports whether provided matches the code Probo
// would show for (orgID, websiteID). Operators paste the code back through the
// Crisp dashboard, so the comparison tolerates surrounding whitespace and a
// lowercased value; it is otherwise a constant-time compare.
func verifyCrispVerificationCode(secret, orgID, websiteID, provided string) bool {
	expected := computeCrispVerificationCode(secret, orgID, websiteID)
	got := strings.ToUpper(strings.TrimSpace(provided))

	return subtle.ConstantTimeCompare([]byte(expected), []byte(got)) == 1
}
