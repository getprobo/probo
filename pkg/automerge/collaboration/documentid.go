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

package collaboration

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"
)

// AutomergeURLPrefix is the scheme automerge-repo puts in front of a document id
// to form a document URL, for example "automerge:34YWzjYt5gPJpq5RfXAkPfPcUj1r".
const AutomergeURLPrefix = "automerge:"

// DocumentIDByteLength is the length of the binary document id automerge-repo
// base58check-encodes: a 16-byte (128-bit) identifier.
const DocumentIDByteLength = 16

// base58Alphabet is the Bitcoin base58 alphabet automerge-repo's bs58check uses.
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// EncodeDocumentID encodes a 16-byte identifier as an automerge-repo document id
// using base58check (base58 of the payload followed by the first four bytes of
// its double SHA-256), matching @automerge/automerge-repo's binaryToDocumentId.
func EncodeDocumentID(id [DocumentIDByteLength]byte) string {
	return base58CheckEncode(id[:])
}

// DecodeDocumentID decodes an automerge-repo document id back to its 16 bytes,
// rejecting a bad checksum or a payload that is not 16 bytes long.
func DecodeDocumentID(documentID string) ([DocumentIDByteLength]byte, error) {
	var id [DocumentIDByteLength]byte

	payload, err := base58CheckDecode(documentID)
	if err != nil {
		return id, fmt.Errorf("invalid automerge document id %q: %w", documentID, err)
	}

	if len(payload) != DocumentIDByteLength {
		return id, fmt.Errorf(
			"automerge document id %q decodes to %d bytes, want %d",
			documentID,
			len(payload),
			DocumentIDByteLength,
		)
	}

	copy(id[:], payload)

	return id, nil
}

// ValidDocumentID reports whether documentID is a well-formed automerge-repo
// document id (correct base58, checksum, and length).
func ValidDocumentID(documentID string) bool {
	_, err := DecodeDocumentID(documentID)

	return err == nil
}

// DeriveDocumentID derives a stable automerge-repo document id from an arbitrary
// seed string, such as a Probo document-version GID. It hashes the seed and
// takes the first 16 bytes, so every peer that knows the seed computes the same
// id without coordination. This is what lets browser clients and Go agents join
// the same repo document, and it is required for ephemeral gossip (presence and
// cursors) to line up, since a peer drops an ephemeral frame whose document id
// it does not recognise.
func DeriveDocumentID(seed string) string {
	digest := sha256.Sum256([]byte(seed))

	var id [DocumentIDByteLength]byte
	copy(id[:], digest[:DocumentIDByteLength])

	return EncodeDocumentID(id)
}

// AutomergeURL wraps a document id in the automerge: scheme.
func AutomergeURL(documentID string) string {
	return AutomergeURLPrefix + documentID
}

// DeriveAutomergeURL derives a stable automerge: URL from a seed string.
func DeriveAutomergeURL(seed string) string {
	return AutomergeURL(DeriveDocumentID(seed))
}

// ParseAutomergeURL extracts and validates the document id from an automerge:
// URL, rejecting a missing scheme or a malformed id.
func ParseAutomergeURL(url string) (string, error) {
	documentID, found := strings.CutPrefix(url, AutomergeURLPrefix)
	if !found {
		return "", fmt.Errorf("automerge url %q is missing the %q scheme", url, AutomergeURLPrefix)
	}

	if _, err := DecodeDocumentID(documentID); err != nil {
		return "", fmt.Errorf("automerge url %q has an invalid document id: %w", url, err)
	}

	return documentID, nil
}

// base58CheckEncode appends the 4-byte double-SHA-256 checksum and base58-encodes
// the result.
func base58CheckEncode(payload []byte) string {
	checked := make([]byte, 0, len(payload)+4)
	checked = append(checked, payload...)
	checked = append(checked, checksum(payload)...)

	return base58Encode(checked)
}

// base58CheckDecode base58-decodes the input and verifies its trailing checksum,
// returning the payload without it.
func base58CheckDecode(encoded string) ([]byte, error) {
	decoded, err := base58Decode(encoded)
	if err != nil {
		return nil, err
	}

	if len(decoded) < 4 {
		return nil, fmt.Errorf("base58check value is too short to contain a checksum")
	}

	payload := decoded[:len(decoded)-4]
	want := decoded[len(decoded)-4:]

	got := checksum(payload)
	if got[0] != want[0] || got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		return nil, fmt.Errorf("base58check checksum mismatch")
	}

	return payload, nil
}

// checksum is the first four bytes of the double SHA-256 of the payload.
func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])

	return second[:4]
}

func base58Encode(input []byte) string {
	value := new(big.Int).SetBytes(input)
	radix := big.NewInt(58)
	remainder := new(big.Int)
	zero := new(big.Int)

	var reversed []byte

	for value.Cmp(zero) > 0 {
		value.DivMod(value, radix, remainder)
		reversed = append(reversed, base58Alphabet[remainder.Int64()])
	}

	// Each leading zero byte is encoded as the alphabet's first character.
	for _, b := range input {
		if b != 0 {
			break
		}

		reversed = append(reversed, base58Alphabet[0])
	}

	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	return string(reversed)
}

func base58Decode(encoded string) ([]byte, error) {
	value := new(big.Int)
	radix := big.NewInt(58)

	for _, character := range encoded {
		index := strings.IndexRune(base58Alphabet, character)
		if index < 0 {
			return nil, fmt.Errorf("invalid base58 character %q", character)
		}

		value.Mul(value, radix)
		value.Add(value, big.NewInt(int64(index)))
	}

	decoded := value.Bytes()

	// Restore the leading zero bytes the encoder wrote as leading '1's.
	zeros := 0
	for zeros < len(encoded) && encoded[zeros] == base58Alphabet[0] {
		zeros++
	}

	result := make([]byte, zeros+len(decoded))
	copy(result[zeros:], decoded)

	return result, nil
}
