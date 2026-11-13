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

package probo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func newPublicationPDFFile(version coredata.DocumentVersion) *coredata.File {
	now := time.Now()
	return &coredata.File{
		ID:             gid.New(version.ID.TenantID(), coredata.FileEntityType),
		OrganizationID: version.OrganizationID,
		BucketName:     "uploads",
		MimeType:       "application/pdf",
		FileName:       "publication.pdf",
		FileKey:        uuid.MustNewV4().String(),
		FileSize:       1,
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func newPublicationPDFVersion(
	t *testing.T,
	client *pg.Client,
	completed bool,
) (coredata.DocumentVersion, coredata.Scoper) {
	t.Helper()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	scope := coredata.NewScope(tenantID, gid.Nil)
	now := time.Now().UTC().Truncate(time.Microsecond)
	document := coredata.Document{
		ID:             gid.New(tenantID, coredata.DocumentEntityType),
		OrganizationID: organizationID,
		WriteMode:      coredata.DocumentWriteModeAuthored,
		Status:         coredata.DocumentStatusActive,
		Title:          "Publication PDF transaction test",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	version := coredata.DocumentVersion{
		ID:              gid.New(tenantID, coredata.DocumentVersionEntityType),
		OrganizationID:  organizationID,
		DocumentID:      document.ID,
		Title:           document.Title,
		Major:           1,
		Classification:  coredata.DocumentClassificationInternal,
		DocumentType:    coredata.DocumentTypePolicy,
		Content:         "policy",
		Status:          coredata.DocumentVersionStatusPublished,
		Orientation:     coredata.DocumentVersionOrientationPortrait,
		PdfAttemptCount: 2,
		PdfClaimedAt:    new(now),
		PublishedAt:     new(now),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := tx.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $4)`,
					organizationID,
					tenantID,
					"Publication PDF test",
					now,
				); err != nil {
					return err
				}
				if err := document.Insert(ctx, tx, scope); err != nil {
					return err
				}
				if completed {
					existing := newPublicationPDFFile(version)
					if err := existing.Insert(ctx, tx, scope); err != nil {
						return err
					}
					version.FileID = new(existing.ID)
				}
				return version.Insert(ctx, tx, scope)
			},
		),
	)
	t.Cleanup(
		func() {
			assert.NoError(
				t,
				client.WithTx(
					context.Background(),
					func(ctx context.Context, tx pg.Tx) error {
						if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
							return err
						}
						if _, err := tx.Exec(ctx, `DELETE FROM files WHERE organization_id = $1`, organizationID); err != nil {
							return err
						}
						var remaining int
						if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM files WHERE organization_id = $1`, organizationID).Scan(&remaining); err != nil {
							return err
						}
						assert.Zero(t, remaining, "publication PDF fixtures must not leave file records")
						return nil
					},
				),
			)
		},
	)
	return version, scope
}

func TestDocumentService_SavePublicationPDF(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		stale     bool
		completed bool
	}{
		"current owner":        {},
		"replaced claim":       {stale: true},
		"newer completed file": {stale: true, completed: true},
	} {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()
				client := test.PGClient(t)
				version, scope := newPublicationPDFVersion(t, client, tc.completed)
				expectedFileID := version.FileID
				if tc.stale {
					version.PdfClaimedAt = new(version.PdfClaimedAt.Add(-pdfClaimLease - time.Second))
				}
				fileRecord := newPublicationPDFFile(version)
				service := DocumentService{svc: &Service{pg: client}}
				require.NoError(t, service.savePublicationPDF(t.Context(), scope, &version, fileRecord, time.Now()))

				require.NoError(
					t,
					client.WithConn(
						t.Context(),
						func(ctx context.Context, conn pg.Querier) error {
							var saved coredata.File
							err := saved.LoadByID(ctx, conn, scope, fileRecord.ID)
							if tc.stale {
								assert.ErrorIs(t, err, coredata.ErrResourceNotFound, "a losing worker's file record must roll back")
							} else {
								assert.NoError(t, err)
								expectedFileID = new(fileRecord.ID)
							}
							var persisted coredata.DocumentVersion
							if err := persisted.LoadByID(ctx, conn, scope, version.ID); err != nil {
								return err
							}
							assert.Equal(t, expectedFileID, persisted.FileID)
							assert.Equal(t, 2, persisted.PdfAttemptCount)
							return nil
						},
					),
				)
			},
		)
	}
}
