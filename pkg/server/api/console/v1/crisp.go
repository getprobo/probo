// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
