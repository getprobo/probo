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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	internaltest "go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestElectronicSignature_EmailSubjectRoundTrip(t *testing.T) {
	t.Parallel()

	completionSubject := "Your signed Non-Disclosure Agreement - Certificate of Completion"

	for _, tt := range []struct {
		name    string
		subject *string
	}{
		{name: "without completion email", subject: nil},
		{name: "with completion email", subject: &completionSubject},
	} {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				client := internaltest.PGClient(t)
				tenantID := gid.NewTenantID()
				scope := coredata.NewScope(tenantID)
				now := time.Now()
				organization := coredata.Organization{
					ID:        gid.New(tenantID, coredata.OrganizationEntityType),
					TenantID:  tenantID,
					Name:      "Electronic signature test",
					CreatedAt: now,
					UpdatedAt: now,
				}
				file := coredata.File{
					ID:             gid.New(tenantID, coredata.FileEntityType),
					OrganizationID: organization.ID,
					BucketName:     "test",
					MimeType:       "application/pdf",
					FileName:       "document.pdf",
					FileKey:        uuid.MustNewV4().String(),
					Visibility:     coredata.FileVisibilityPrivate,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				signature := coredata.ElectronicSignature{
					ID:             gid.New(tenantID, coredata.ElectronicSignatureEntityType),
					OrganizationID: organization.ID,
					Status:         coredata.ElectronicSignatureStatusPending,
					DocumentType:   coredata.ElectronicSignatureDocumentTypeNDA,
					FileID:         file.ID,
					SignerEmail:    "signer@example.com",
					ConsentText:    "I consent.",
					EmailSubject:   tt.subject,
					SealVersion:    1,
					MaxAttempts:    10,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				err := client.WithTx(
					context.Background(),
					func(ctx context.Context, tx pg.Tx) error {
						if err := organization.Insert(ctx, tx); err != nil {
							return fmt.Errorf("cannot insert organization: %w", err)
						}

						if err := file.Insert(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot insert file: %w", err)
						}

						if err := signature.Insert(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot insert signature: %w", err)
						}

						return nil
					},
				)
				require.NoError(t, err)

				t.Cleanup(func() {
					_ = client.WithTx(
						context.Background(),
						func(ctx context.Context, tx pg.Tx) error {
							for _, query := range []struct {
								sql string
								id  gid.GID
							}{
								{"DELETE FROM electronic_signatures WHERE id = $1", signature.ID},
								{"DELETE FROM files WHERE id = $1", file.ID},
								{"DELETE FROM organizations WHERE id = $1", organization.ID},
							} {
								if _, err := tx.Exec(ctx, query.sql, query.id); err != nil {
									return err
								}
							}

							return nil
						},
					)
				})

				var loaded coredata.ElectronicSignature

				err = client.WithConn(
					context.Background(),
					func(ctx context.Context, conn pg.Querier) error {
						return loaded.LoadByID(ctx, conn, scope, signature.ID)
					},
				)
				require.NoError(t, err)
				assert.Equal(t, tt.subject, loaded.EmailSubject)
			},
		)
	}
}
