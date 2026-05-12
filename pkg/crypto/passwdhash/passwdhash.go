// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package passwdhash

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type (
	Profile struct {
		iterations uint32
		saltLength uint
		keyLength  uint
		pepper     []byte
	}
)

const (
	versionByte   = 0x01 // Version identifier
	algorithmByte = 0x01 // Algorithm identifier (0x01 for PBKDF2-SHA256)

	minIterations = 600000
)

func NewProfile(pepper []byte, iterations uint32) (*Profile, error) {
	if len(pepper) < 32 {
		return nil, fmt.Errorf("pepper must be at least 32 bytes")
	}

	if iterations < minIterations {
		return nil, fmt.Errorf("iterations below minimum security threshold")
	}

	return &Profile{
		iterations: iterations,
		saltLength: 32,
		keyLength:  32,
		pepper:     pepper,
	}, nil
}

func (hp Profile) applyPepper(input []byte) []byte {
	mac := hmac.New(sha256.New, hp.pepper)
	mac.Write(input)
	return mac.Sum(nil)
}

func (hp Profile) HashPassword(password []byte) ([]byte, error) {
	salt := make([]byte, hp.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("error generating salt: %v", err)
	}

	pepperedPassword := hp.applyPepper([]byte(password))
	hash := pbkdf2.Key(pepperedPassword, salt, int(hp.iterations), int(hp.keyLength), sha256.New)

	// Binary format:
	// [1B version][1B algorithm][4B iterations][1B salt length][salt bytes][hash bytes]
	binaryHash := make([]byte, 0, 7+hp.saltLength+hp.keyLength)

	// Version and algorithm
	binaryHash = append(binaryHash, versionByte)
	binaryHash = append(binaryHash, algorithmByte)

	// Iterations (4 bytes, big endian)
	iterBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(iterBytes, hp.iterations)
	binaryHash = append(binaryHash, iterBytes...)

	// Salt length and salt
	binaryHash = append(binaryHash, byte(hp.saltLength))
	binaryHash = append(binaryHash, salt...)

	// Hash
	binaryHash = append(binaryHash, hash...)

	return binaryHash, nil
}

func (hp Profile) ComparePasswordAndHash(password, passwordHash []byte) (bool, error) {
	if len(passwordHash) < 7 {
		return false, fmt.Errorf("hash too short")
	}

	if passwordHash[0] != versionByte {
		return false, fmt.Errorf("unsupported hash version: %d", passwordHash[0])
	}

	if passwordHash[1] != algorithmByte {
		return false, fmt.Errorf("unsupported algorithm: %d", passwordHash[1])
	}

	// Extract iterations
	iterations := binary.BigEndian.Uint32(passwordHash[2:6])

	if iterations < minIterations {
		return false, fmt.Errorf("iterations below minimum security threshold")
	}

	// Extract salt length and validate
	saltLen := int(passwordHash[6])
	if saltLen < 32 {
		return false, fmt.Errorf("salt length below security minimum")
	}

	if len(passwordHash) < 7+saltLen+int(hp.keyLength) {
		return false, fmt.Errorf("invalid hash length")
	}

	salt := passwordHash[7 : 7+saltLen]
	storedHash := passwordHash[7+saltLen:]

	pepperedPassword := hp.applyPepper([]byte(password))

	newHash := pbkdf2.Key(pepperedPassword, salt, int(iterations), len(storedHash), sha256.New)

	return subtle.ConstantTimeCompare(storedHash, newHash) == 1, nil
}
