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

package complianceportal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

func TestLoadResources_SkipsReportAndFileFromOtherPortal(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	scope := coredata.NewScope(tenantID)
	now := time.Now()

	var (
		portalAID        gid.GID
		accessID         gid.GID
		keptAuditLinkID  gid.GID
		keptReportFileID gid.GID
		skippedReportID  gid.GID
		skippedFileID    gid.GID
		keptFileID       gid.GID
		identityID       gid.GID
	)

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				organization := coredata.Organization{
					ID:        organizationID,
					TenantID:  tenantID,
					Name:      "load-resources-" + organizationID.String(),
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}

				portalA, err := insertTestPortal(ctx, tx, scope, organizationID, "portal-a")
				if err != nil {
					return err
				}

				portalB, err := insertTestPortal(ctx, tx, scope, organizationID, "portal-b")
				if err != nil {
					return err
				}

				portalAID = portalA

				emailAddress, err := mail.ParseAddr(
					fmt.Sprintf("%s@example.com", organizationID),
				)
				if err != nil {
					return err
				}

				identityID = gid.New(gid.NilTenant, coredata.IdentityEntityType)

				identity := coredata.Identity{
					ID:                   identityID,
					EmailAddress:         emailAddress,
					FullName:             "Resource Loader",
					EmailAddressVerified: true,
					CreatedAt:            now,
					UpdatedAt:            now,
				}
				if err := identity.Insert(ctx, tx); err != nil {
					return err
				}

				access := coredata.CompliancePortalAccess{
					ID:                 gid.New(tenantID, coredata.CompliancePortalAccessEntityType),
					OrganizationID:     organizationID,
					TenantID:           tenantID,
					IdentityID:         identityID,
					CompliancePortalID: portalA,
					State:              coredata.CompliancePortalAccessStateActive,
					CreatedAt:          now,
					UpdatedAt:          now,
				}
				if err := access.Insert(ctx, tx, scope); err != nil {
					return err
				}

				accessID = access.ID

				keptReportFileID, keptAuditLinkID, err = insertTestReport(
					ctx,
					tx,
					scope,
					organizationID,
					portalA,
					"kept-report",
				)
				if err != nil {
					return err
				}

				var skippedAuditLinkID gid.GID

				skippedReportID, skippedAuditLinkID, err = insertTestReport(
					ctx,
					tx,
					scope,
					organizationID,
					portalB,
					"skipped-report",
				)
				if err != nil {
					return err
				}

				keptFileID, err = insertTestPortalFile(
					ctx,
					tx,
					scope,
					organizationID,
					portalA,
					"kept-file",
					coredata.CompliancePortalVisibilityPublic,
				)
				if err != nil {
					return err
				}

				skippedFileID, err = insertTestPortalFile(
					ctx,
					tx,
					scope,
					organizationID,
					portalB,
					"skipped-file",
					coredata.CompliancePortalVisibilityPublic,
				)
				if err != nil {
					return err
				}

				for _, auditLinkID := range []gid.GID{keptAuditLinkID, skippedAuditLinkID} {
					id := auditLinkID

					row := coredata.CompliancePortalDocumentAccess{
						ID:                       gid.New(tenantID, coredata.CompliancePortalDocumentAccessEntityType),
						OrganizationID:           organizationID,
						CompliancePortalAccessID: access.ID,
						CompliancePortalAuditID:  &id,
						Status:                   coredata.CompliancePortalDocumentAccessStatusRequested,
						CreatedAt:                now,
						UpdatedAt:                now,
					}
					if err := row.Insert(ctx, tx, scope); err != nil {
						return err
					}
				}

				for _, fileID := range []gid.GID{keptFileID, skippedFileID} {
					id := fileID

					row := coredata.CompliancePortalDocumentAccess{
						ID:                       gid.New(tenantID, coredata.CompliancePortalDocumentAccessEntityType),
						OrganizationID:           organizationID,
						CompliancePortalAccessID: access.ID,
						CompliancePortalFileID:   &id,
						Status:                   coredata.CompliancePortalDocumentAccessStatusRequested,
						CreatedAt:                now,
						UpdatedAt:                now,
					}
					if err := row.Insert(ctx, tx, scope); err != nil {
						return err
					}
				}

				return nil
			},
		),
	)

	t.Cleanup(
		func() {
			_ = client.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					_ = (&coredata.Identity{ID: identityID}).Delete(ctx, tx)

					return (&coredata.Organization{}).Delete(ctx, tx, organizationID)
				},
			)
		},
	)

	var (
		reports []messageReport
		files   []messageFile
	)

	require.NoError(
		t,
		client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				var err error

				_, reports, files, err = loadResources(ctx, conn, scope, portalAID, accessID)

				return err
			},
		),
	)

	require.Len(t, reports, 1)
	assert.Equal(t, keptReportFileID.String(), reports[0].ID)
	require.Len(t, files, 1)
	assert.Equal(t, keptFileID.String(), files[0].ID)
	assert.NotEqual(t, skippedReportID.String(), reports[0].ID)
	assert.NotEqual(t, skippedFileID.String(), files[0].ID)
}

func TestResolveAccessResourceIDs_MapsPortalIDsAndDropsForeignRows(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	scope := coredata.NewScope(tenantID)
	now := time.Now()

	var (
		portalID              gid.GID
		documentID            gid.GID
		documentLinkID        gid.GID
		reportFileID          gid.GID
		auditLinkID           gid.GID
		portalFileID          gid.GID
		missingCatalogDoc     gid.GID
		foreignDocumentID     gid.GID
		foreignDocumentLinkID gid.GID
		foreignReportFileID   gid.GID
		foreignPortalFileID   gid.GID
	)

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				organization := coredata.Organization{
					ID:        organizationID,
					TenantID:  tenantID,
					Name:      "resolve-ids-" + organizationID.String(),
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := organization.Insert(ctx, tx); err != nil {
					return err
				}

				var err error

				portalID, err = insertTestPortal(ctx, tx, scope, organizationID, "resolve-portal")
				if err != nil {
					return err
				}

				documentID = gid.New(tenantID, coredata.DocumentEntityType)

				document := coredata.Document{
					ID:             documentID,
					OrganizationID: organizationID,
					WriteMode:      coredata.DocumentWriteModeAuthored,
					Status:         coredata.DocumentStatusActive,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if err := document.Insert(ctx, tx, scope); err != nil {
					return err
				}

				portalDocument := coredata.CompliancePortalDocument{
					ID:                 gid.New(tenantID, coredata.CompliancePortalDocumentEntityType),
					OrganizationID:     organizationID,
					CompliancePortalID: portalID,
					DocumentID:         documentID,
					Visibility:         coredata.CompliancePortalVisibilityRestricted,
					CreatedAt:          now,
					UpdatedAt:          now,
				}
				if err := portalDocument.Upsert(ctx, tx, scope); err != nil {
					return err
				}

				documentLinkID = portalDocument.ID
				missingCatalogDoc = gid.New(tenantID, coredata.CompliancePortalDocumentEntityType)

				reportFileID, auditLinkID, err = insertTestReport(
					ctx,
					tx,
					scope,
					organizationID,
					portalID,
					"resolve-report",
				)
				if err != nil {
					return err
				}

				portalFileID, err = insertTestPortalFile(
					ctx,
					tx,
					scope,
					organizationID,
					portalID,
					"resolve-file",
					coredata.CompliancePortalVisibilityRestricted,
				)
				if err != nil {
					return err
				}

				otherPortalID, err := insertTestPortal(ctx, tx, scope, organizationID, "resolve-other-portal")
				if err != nil {
					return err
				}

				foreignDocumentID = gid.New(tenantID, coredata.DocumentEntityType)
				foreignDocument := coredata.Document{
					ID:             foreignDocumentID,
					OrganizationID: organizationID,
					WriteMode:      coredata.DocumentWriteModeAuthored,
					Status:         coredata.DocumentStatusActive,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if err := foreignDocument.Insert(ctx, tx, scope); err != nil {
					return err
				}

				foreignPortalDocument := coredata.CompliancePortalDocument{
					ID:                 gid.New(tenantID, coredata.CompliancePortalDocumentEntityType),
					OrganizationID:     organizationID,
					CompliancePortalID: otherPortalID,
					DocumentID:         foreignDocumentID,
					Visibility:         coredata.CompliancePortalVisibilityRestricted,
					CreatedAt:          now,
					UpdatedAt:          now,
				}
				if err := foreignPortalDocument.Upsert(ctx, tx, scope); err != nil {
					return err
				}

				foreignDocumentLinkID = foreignPortalDocument.ID

				foreignReportFileID, _, err = insertTestReport(
					ctx,
					tx,
					scope,
					organizationID,
					otherPortalID,
					"resolve-foreign-report",
				)
				if err != nil {
					return err
				}

				foreignPortalFileID, err = insertTestPortalFile(
					ctx,
					tx,
					scope,
					organizationID,
					otherPortalID,
					"resolve-foreign-file",
					coredata.CompliancePortalVisibilityRestricted,
				)

				return err
			},
		),
	)

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			return (&coredata.Organization{}).Delete(ctx, tx, organizationID)
		})
	})

	var (
		documentIDs []gid.GID
		auditIDs    []gid.GID
		fileIDs     []gid.GID
	)

	require.NoError(
		t,
		client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				var err error

				documentIDs, auditIDs, fileIDs, err = resolveAccessResourceIDs(
					ctx,
					conn,
					scope,
					portalID,
					[]gid.GID{
						documentID,
						documentLinkID,
						missingCatalogDoc,
						foreignDocumentID,
						foreignDocumentLinkID,
						reportFileID,
						foreignReportFileID,
						portalFileID,
						foreignPortalFileID,
					},
				)

				return err
			},
		),
	)

	assert.ElementsMatch(t, []gid.GID{documentLinkID}, documentIDs)
	assert.Equal(t, []gid.GID{auditLinkID}, auditIDs)
	assert.Equal(t, []gid.GID{portalFileID}, fileIDs)
}

func insertTestPortal(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	name string,
) (gid.GID, error) {
	now := time.Now()
	portalID := gid.New(organizationID.TenantID(), coredata.CompliancePortalEntityType)

	mailingList := coredata.MailingList{
		ID:             gid.New(organizationID.TenantID(), coredata.MailingListEntityType),
		OrganizationID: organizationID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := mailingList.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	portal := coredata.CompliancePortal{
		ID:                   portalID,
		OrganizationID:       organizationID,
		TenantID:             organizationID.TenantID(),
		Active:               true,
		Slug:                 strings.ToLower(portalID.String()),
		SearchEngineIndexing: coredata.SearchEngineIndexingNotIndexable,
		Capabilities:         coredata.DefaultCompliancePortalCapabilities(),
		MailingListID:        mailingList.ID,
		EntityName:           name,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := portal.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	return portalID, nil
}

func insertTestBlobFile(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	name string,
) (gid.GID, error) {
	now := time.Now()

	objectKey, err := uuid.NewV7()
	if err != nil {
		return gid.Nil, err
	}

	file := coredata.File{
		ID:             gid.New(organizationID.TenantID(), coredata.FileEntityType),
		OrganizationID: organizationID,
		BucketName:     "uploads",
		MimeType:       "application/pdf",
		FileName:       name + ".pdf",
		FileKey:        objectKey.String(),
		FileSize:       1,
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := file.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	return file.ID, nil
}

func insertTestPortalFile(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	portalID gid.GID,
	name string,
	visibility coredata.CompliancePortalVisibility,
) (gid.GID, error) {
	now := time.Now()

	blobID, err := insertTestBlobFile(
		ctx,
		tx,
		scope,
		organizationID,
		name,
	)
	if err != nil {
		return gid.Nil, err
	}

	file := coredata.CompliancePortalFile{
		ID:                         gid.New(organizationID.TenantID(), coredata.CompliancePortalFileEntityType),
		OrganizationID:             organizationID,
		CompliancePortalID:         portalID,
		Name:                       name,
		Category:                   "OTHER",
		FileID:                     blobID,
		CompliancePortalVisibility: visibility,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
	if err := file.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	return file.ID, nil
}

func insertTestReport(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	portalID gid.GID,
	name string,
) (gid.GID, gid.GID, error) {
	now := time.Now()

	reportFileID, err := insertTestBlobFile(
		ctx,
		tx,
		scope,
		organizationID,
		name+"-report",
	)
	if err != nil {
		return gid.Nil, gid.Nil, err
	}

	framework := coredata.Framework{
		ID:             gid.New(organizationID.TenantID(), coredata.FrameworkEntityType),
		OrganizationID: organizationID,
		ReferenceID:    name,
		Name:           name,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := framework.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, gid.Nil, err
	}

	audit := coredata.Audit{
		ID:             gid.New(organizationID.TenantID(), coredata.AuditEntityType),
		OrganizationID: organizationID,
		FrameworkID:    framework.ID,
		ReportFileID:   &reportFileID,
		State:          coredata.AuditStateCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := audit.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, gid.Nil, err
	}

	portalAudit := coredata.CompliancePortalAudit{
		ID:                 gid.New(organizationID.TenantID(), coredata.CompliancePortalAuditEntityType),
		OrganizationID:     organizationID,
		CompliancePortalID: portalID,
		AuditID:            audit.ID,
		Visibility:         coredata.CompliancePortalVisibilityPublic,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := portalAudit.Upsert(ctx, tx, scope); err != nil {
		return gid.Nil, gid.Nil, err
	}

	return reportFileID, portalAudit.ID, nil
}
