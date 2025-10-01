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

package pem

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const (
	BlockTypeCertificate     = "CERTIFICATE"
	BlockTypeECPrivateKey    = "EC PRIVATE KEY"
	BlockTypeRSAPrivateKey   = "RSA PRIVATE KEY"
	BlockTypePKCS8PrivateKey = "PRIVATE KEY"
)

func EncodeCertificate(der []byte) []byte {
	block := &pem.Block{
		Type:  BlockTypeCertificate,
		Bytes: der,
	}
	return pem.EncodeToMemory(block)
}

func EncodeCertificateChain(derCerts [][]byte) []byte {
	var chain []byte
	for _, der := range derCerts {
		chain = append(chain, EncodeCertificate(der)...)
	}
	return chain
}

func EncodePrivateKey(key crypto.Signer) ([]byte, error) {
	var keyDER []byte
	var keyType string

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		der, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal EC private key: %w", err)
		}
		keyDER = der
		keyType = BlockTypeECPrivateKey
	case *rsa.PrivateKey:
		keyDER = x509.MarshalPKCS1PrivateKey(k)
		keyType = BlockTypeRSAPrivateKey
	case ed25519.PrivateKey:
		der, err := x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal ED25519 private key: %w", err)
		}
		keyDER = der
		keyType = BlockTypePKCS8PrivateKey
	default:
		return nil, fmt.Errorf("unsupported key type: %T", key)
	}

	block := &pem.Block{
		Type:  keyType,
		Bytes: keyDER,
	}

	return pem.EncodeToMemory(block), nil
}

func DecodePrivateKey(pemData []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("cannot to decode PEM block")
	}

	switch block.Type {
	case BlockTypeECPrivateKey:
		return x509.ParseECPrivateKey(block.Bytes)
	case BlockTypeRSAPrivateKey:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case BlockTypePKCS8PrivateKey:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("cannot parse PKCS8 private key: %w", err)
		}
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("key is not a crypto.Signer")
		}
		return signer, nil
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}
