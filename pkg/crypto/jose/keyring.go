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

package jose

import (
	"crypto/rsa"
	"fmt"
	"slices"
	"sync/atomic"
)

type (
	// SigningKey pairs an RSA private key with its key ID. Every key of a
	// KeyRing is published in its JWKS; only keys with Active set to true sign
	// new tokens. A retired key stays published so tokens a verifier already
	// cached keep verifying.
	SigningKey struct {
		PrivateKey *rsa.PrivateKey
		KID        string
		Active     bool
	}

	// KeyRing is a validated set of signing keys. It signs with the active
	// keys and publishes all of them, so a caller minting tokens needs to know
	// nothing about rotation.
	KeyRing struct {
		keys      []SigningKey
		activeIdx []int
		rrCounter atomic.Uint64
	}
)

// NewKeyRing validates keys and returns the ring that signs with them.
//
// Retired keys are validated alongside the active ones: they stay in the
// published set, so a verifier can resolve a token to any of them.
func NewKeyRing(keys []SigningKey) (*KeyRing, error) {
	activeIdx := make([]int, 0, len(keys))
	seenKIDs := make(map[string]struct{}, len(keys))

	for idx, key := range keys {
		if key.PrivateKey == nil {
			return nil, fmt.Errorf("cannot create key ring: signing key %q has no private key", key.KID)
		}

		if err := ValidateRSAPublicKey(&key.PrivateKey.PublicKey); err != nil {
			return nil, fmt.Errorf("cannot create key ring: signing key %q: %w", key.KID, err)
		}

		if key.KID == "" {
			return nil, fmt.Errorf("cannot create key ring: signing key at index %d has no key id", idx)
		}

		// A verifier resolves a token to the first published key carrying its
		// kid, so a second key under the same kid makes some minted tokens
		// unverifiable. Rotation introduces a new kid; it never reuses one.
		if _, ok := seenKIDs[key.KID]; ok {
			return nil, fmt.Errorf("cannot create key ring: signing key %q is duplicated", key.KID)
		}

		seenKIDs[key.KID] = struct{}{}

		if key.Active {
			activeIdx = append(activeIdx, idx)
		}
	}

	if len(activeIdx) == 0 {
		return nil, fmt.Errorf("cannot create key ring: at least one signing key must be active")
	}

	// Cloned so that a caller holding the original slice cannot swap in key
	// material that never passed the validation above.
	return &KeyRing{
		keys:      slices.Clone(keys),
		activeIdx: activeIdx,
	}, nil
}

// Sign signs claims as an RS256 JWT with one of the active keys, naming it in
// the "kid" header so a verifier can find it in the published set.
func (r *KeyRing) Sign(claims any) (string, error) {
	key := r.signingKey()

	return SignJWT(key.PrivateKey, key.KID, claims)
}

// JWKS returns the published key set. Retired keys are included so that a
// token signed before a rotation still verifies.
func (r *KeyRing) JWKS() *JWKS {
	jwks := &JWKS{
		Keys: make([]JWK, 0, len(r.keys)),
	}

	for _, key := range r.keys {
		jwks.Keys = append(
			jwks.Keys,
			RSAPublicKeyToJWK(&key.PrivateKey.PublicKey, key.KID),
		)
	}

	return jwks
}

// signingKey returns the next active signing key using round-robin.
func (r *KeyRing) signingKey() *SigningKey {
	n := r.rrCounter.Add(1)
	idx := r.activeIdx[n%uint64(len(r.activeIdx))]

	return &r.keys[idx]
}
