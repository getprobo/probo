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

package coredata_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func newSealTestSignature(signedAt time.Time) *coredata.ElectronicSignature {
	tenantID := gid.NewTenantID()

	fileHash := strings.Repeat("a", 64)
	signerFullName := "Test Signer"
	signerIP := "203.0.113.7"
	signerUA := "Mozilla/5.0 (X11; Linux x86_64)"

	return &coredata.ElectronicSignature{
		ID:              gid.New(tenantID, coredata.ElectronicSignatureEntityType),
		OrganizationID:  gid.New(tenantID, coredata.OrganizationEntityType),
		DocumentType:    coredata.ElectronicSignatureDocumentTypeOther,
		FileID:          gid.New(tenantID, coredata.FileEntityType),
		FileHash:        &fileHash,
		SignerFullName:  &signerFullName,
		SignerEmail:     "Test.Signer@example.com",
		SignerIPAddress: &signerIP,
		SignerUserAgent: &signerUA,
		ConsentText:     "I consent to sign this document electronically.",
		SignedAt:        &signedAt,
	}
}

// A verifier only ever sees the timestamp the certificate prints. If that
// rendering drops precision the seal cannot be recomputed from it, so the two
// must stay byte-identical.
func TestElectronicSignature_SealSignedAtIsWhatTheSealHashes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		signedAt time.Time
		want     string
	}{
		{
			name:     "microseconds",
			signedAt: time.Date(2026, 8, 11, 9, 36, 49, 961506000, time.UTC),
			want:     "2026-08-11T09:36:49.961506Z",
		},
		{
			name:     "trailing zeros stripped",
			signedAt: time.Date(2026, 8, 11, 9, 36, 49, 100000000, time.UTC),
			want:     "2026-08-11T09:36:49.1Z",
		},
		{
			name:     "whole second carries no fraction",
			signedAt: time.Date(2026, 8, 11, 9, 36, 49, 0, time.UTC),
			want:     "2026-08-11T09:36:49Z",
		},
		{
			name:     "nanoseconds truncated to microseconds",
			signedAt: time.Date(2026, 8, 11, 9, 36, 49, 961506789, time.UTC),
			want:     "2026-08-11T09:36:49.961506Z",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			signature := newSealTestSignature(tc.signedAt)

			assert.Equal(t, tc.want, signature.SealSignedAt())

			seal, err := signature.ComputeSeal(1)
			require.NoError(t, err)

			fields := []string{
				signature.ID.String(),
				signature.OrganizationID.String(),
				signature.DocumentType.String(),
				signature.FileID.String(),
				*signature.FileHash,
				*signature.SignerFullName,
				strings.ToLower(signature.SignerEmail),
				*signature.SignerIPAddress,
				*signature.SignerUserAgent,
				signature.ConsentText,
				signature.SealSignedAt(),
			}
			expected := sha256.Sum256([]byte(strings.Join(fields, "\n")))

			assert.Equal(t, hex.EncodeToString(expected[:]), seal)
		})
	}
}

func TestElectronicSignature_SealSignedAtNilIsEmpty(t *testing.T) {
	t.Parallel()

	signature := &coredata.ElectronicSignature{}

	assert.Empty(t, signature.SealSignedAt())
}
