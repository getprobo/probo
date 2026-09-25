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

package probodconfig

import (
	"crypto"
	"crypto/rsa"
	"encoding"
	"encoding/json"
	"fmt"

	"go.probo.inc/probo/pkg/crypto/pem"
)

type (
	// PrivateKey is a PEM-encoded private key decoded when the configuration is
	// read, so that malformed key material is rejected at load time instead of
	// at the point of use.
	//
	// The zero value holds no key: a field whose key is optional reports that
	// with IsZero.
	PrivateKey struct {
		pem    string
		signer crypto.Signer
	}

	// RSAPrivateKey is a PrivateKey restricted to RSA. The signing key rings
	// and the SAML service provider sign with RSA only, so a key of another
	// algorithm is rejected when the configuration is read.
	RSAPrivateKey struct {
		pem string
		key *rsa.PrivateKey
	}
)

var (
	_ encoding.TextUnmarshaler = (*PrivateKey)(nil)
	_ encoding.TextMarshaler   = (*PrivateKey)(nil)
	_ json.Unmarshaler         = (*PrivateKey)(nil)
	_ json.Marshaler           = (*PrivateKey)(nil)

	_ encoding.TextUnmarshaler = (*RSAPrivateKey)(nil)
	_ encoding.TextMarshaler   = (*RSAPrivateKey)(nil)
	_ json.Unmarshaler         = (*RSAPrivateKey)(nil)
	_ json.Marshaler           = (*RSAPrivateKey)(nil)
)

// ParsePrivateKey decodes a PEM-encoded private key. An empty input returns the
// zero value, so a caller that requires a key checks IsZero.
func ParsePrivateKey(pemData string) (PrivateKey, error) {
	if pemData == "" {
		return PrivateKey{}, nil
	}

	signer, err := pem.DecodePrivateKey([]byte(pemData))
	if err != nil {
		return PrivateKey{}, fmt.Errorf("cannot decode private key: %w", err)
	}

	return PrivateKey{pem: pemData, signer: signer}, nil
}

// IsZero reports whether no key was configured.
func (k PrivateKey) IsZero() bool {
	return k.signer == nil
}

// PEM returns the key as it was configured. The encoding is preserved verbatim
// because a re-encoded key is different bytes, and callers derive secrets from
// this material.
func (k PrivateKey) PEM() string {
	return k.pem
}

// Signer returns the decoded key, or nil when none was configured.
func (k PrivateKey) Signer() crypto.Signer {
	return k.signer
}

func (k *PrivateKey) UnmarshalText(text []byte) error {
	privateKey, err := ParsePrivateKey(string(text))
	if err != nil {
		return err
	}

	*k = privateKey

	return nil
}

func (k *PrivateKey) UnmarshalJSON(data []byte) error {
	var pemData string
	if err := json.Unmarshal(data, &pemData); err != nil {
		return fmt.Errorf("cannot unmarshal private key: %w", err)
	}

	return k.UnmarshalText([]byte(pemData))
}

func (k PrivateKey) MarshalText() ([]byte, error) {
	return []byte(k.pem), nil
}

func (k PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.pem)
}

// ParseRSAPrivateKey decodes a PEM-encoded RSA private key. An empty input
// returns the zero value, so a caller that requires a key checks IsZero.
func ParseRSAPrivateKey(pemData string) (RSAPrivateKey, error) {
	privateKey, err := ParsePrivateKey(pemData)
	if err != nil {
		return RSAPrivateKey{}, err
	}

	if privateKey.IsZero() {
		return RSAPrivateKey{}, nil
	}

	rsaKey, ok := privateKey.Signer().(*rsa.PrivateKey)
	if !ok {
		return RSAPrivateKey{}, fmt.Errorf("private key is not an RSA key, got %T", privateKey.Signer())
	}

	return RSAPrivateKey{pem: privateKey.PEM(), key: rsaKey}, nil
}

// IsZero reports whether no key was configured.
func (k RSAPrivateKey) IsZero() bool {
	return k.key == nil
}

// PEM returns the key as it was configured. The encoding is preserved verbatim
// because a re-encoded key is different bytes, and callers derive secrets from
// this material.
func (k RSAPrivateKey) PEM() string {
	return k.pem
}

// PrivateKey returns the decoded key, or nil when none was configured.
func (k RSAPrivateKey) PrivateKey() *rsa.PrivateKey {
	return k.key
}

func (k *RSAPrivateKey) UnmarshalText(text []byte) error {
	privateKey, err := ParseRSAPrivateKey(string(text))
	if err != nil {
		return err
	}

	*k = privateKey

	return nil
}

func (k *RSAPrivateKey) UnmarshalJSON(data []byte) error {
	var pemData string
	if err := json.Unmarshal(data, &pemData); err != nil {
		return fmt.Errorf("cannot unmarshal RSA private key: %w", err)
	}

	return k.UnmarshalText([]byte(pemData))
}

func (k RSAPrivateKey) MarshalText() ([]byte, error) {
	return []byte(k.pem), nil
}

func (k RSAPrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.pem)
}
