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

package deviceagent

import (
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
)

const (
	// Leaf SPKI SHA-256 pins for the current ACM certificates.
	// A key rotation fails Probo API calls until these are updated.
	usConsoleLeafSPKI = "bc0eedef1d599ad6a77372303b8cc80b3e849c5142a94e8eaec21e369b07b370"
	euConsoleLeafSPKI = "8df5e1fd2cf07827d75ad186510f5bb203b438617640ecbfc0543ef81c029dad"
)

var (
	proboCloudLeafSPKIPins = mustSPKIPins(usConsoleLeafSPKI, euConsoleLeafSPKI)

	ErrProboCloudPinMismatch = errors.New("TLS certificate is not a pinned Probo key")
)

func proboCloudTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:            tls.VersionTLS12,
		VerifyPeerCertificate: verifyProboCloudSPKI,
	}
}

func verifyProboCloudSPKI(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
	return verifySPKIPins(verifiedChains, proboCloudLeafSPKIPins)
}

func verifySPKIPins(verifiedChains [][]*x509.Certificate, pins [][32]byte) error {
	for _, chain := range verifiedChains {
		for _, cert := range chain {
			if spkiPinned(cert, pins) {
				return nil
			}
		}
	}

	return ErrProboCloudPinMismatch
}

func spkiPinned(cert *x509.Certificate, pins [][32]byte) bool {
	if cert == nil {
		return false
	}

	sum := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	matched := false

	for _, pin := range pins {
		if subtle.ConstantTimeCompare(sum[:], pin[:]) == 1 {
			matched = true
		}
	}

	return matched
}

func mustSPKIPins(hexes ...string) [][32]byte {
	pins := make([][32]byte, len(hexes))

	for i, raw := range hexes {
		decoded, err := hex.DecodeString(raw)
		if err != nil || len(decoded) != sha256.Size {
			panic("invalid Probo SPKI pin")
		}

		copy(pins[i][:], decoded)
	}

	return pins
}
