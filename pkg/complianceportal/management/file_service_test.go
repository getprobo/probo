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

package management

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

type updateFileAccessFixture struct {
	scope      *coredata.Scope
	accessID   gid.GID
	fileID     gid.GID
	identityID gid.GID
}

func seedUpdateFileAccessFixture(t *testing.T, ctx context.Context, client *pg.Client) updateFileAccessFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	var (
		accessID   gid.GID
		fileID     gid.GID
		identityID gid.GID
	)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		organization := coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "update-file-" + organizationID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := organization.Insert(ctx, tx); err != nil {
			return err
		}

		mailingList := coredata.MailingList{
			ID:             gid.New(tenantID, coredata.MailingListEntityType),
			OrganizationID: organizationID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := mailingList.Insert(ctx, tx, scope); err != nil {
			return err
		}

		portalID := gid.New(tenantID, coredata.CompliancePortalEntityType)
		portal := coredata.CompliancePortal{
			ID:                   portalID,
			OrganizationID:       organizationID,
			TenantID:             tenantID,
			Active:               true,
			Slug:                 strings.ToLower(portalID.String()),
			SearchEngineIndexing: coredata.SearchEngineIndexingNotIndexable,
			Capabilities:         coredata.DefaultCompliancePortalCapabilities(),
			MailingListID:        mailingList.ID,
			EntityName:           "update-file-portal",
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := portal.Insert(ctx, tx, scope); err != nil {
			return err
		}

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
			FullName:             "File Updater",
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
			CompliancePortalID: portalID,
			State:              coredata.CompliancePortalAccessStateActive,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := access.Insert(ctx, tx, scope); err != nil {
			return err
		}

		accessID = access.ID

		objectKey, err := uuid.NewV7()
		if err != nil {
			return err
		}

		blob := coredata.File{
			ID:             gid.New(tenantID, coredata.FileEntityType),
			OrganizationID: organizationID,
			BucketName:     "uploads",
			MimeType:       "application/pdf",
			FileName:       "restricted-file.pdf",
			FileKey:        objectKey.String(),
			FileSize:       1,
			Visibility:     coredata.FileVisibilityPrivate,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := blob.Insert(ctx, tx, scope); err != nil {
			return err
		}

		file := coredata.CompliancePortalFile{
			ID:                         gid.New(tenantID, coredata.CompliancePortalFileEntityType),
			OrganizationID:             organizationID,
			CompliancePortalID:         portalID,
			Name:                       "restricted-file",
			Category:                   "OTHER",
			FileID:                     blob.ID,
			CompliancePortalVisibility: coredata.CompliancePortalVisibilityRestricted,
			CreatedAt:                  now,
			UpdatedAt:                  now,
		}
		if err := file.Insert(ctx, tx, scope); err != nil {
			return err
		}

		fileID = file.ID
		requestedAt := now
		row := coredata.CompliancePortalDocumentAccess{
			ID:                       gid.New(tenantID, coredata.CompliancePortalDocumentAccessEntityType),
			OrganizationID:           organizationID,
			CompliancePortalAccessID: accessID,
			CompliancePortalFileID:   &fileID,
			Status:                   coredata.CompliancePortalDocumentAccessStatusRequested,
			RequestedAt:              &requestedAt,
			CreatedAt:                now,
			UpdatedAt:                now,
		}

		return row.Insert(ctx, tx, scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_ = (&coredata.Identity{ID: identityID}).Delete(ctx, tx)

			return (&coredata.Organization{}).Delete(ctx, tx, organizationID)
		})
	})

	return updateFileAccessFixture{
		scope:      scope,
		accessID:   accessID,
		fileID:     fileID,
		identityID: identityID,
	}
}

func loadFileAccesses(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	accessID gid.GID,
	fileID gid.GID,
) coredata.CompliancePortalDocumentAccesses {
	t.Helper()

	var accesses coredata.CompliancePortalDocumentAccesses

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return accesses.LoadByCompliancePortalAccessIDAndCompliancePortalFileIDs(
			ctx,
			conn,
			scope,
			accessID,
			[]gid.GID{fileID},
		)
	}))

	return accesses
}

func TestService_DeleteFile_DeletesAccesses(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := t.Context()
	fx := seedUpdateFileAccessFixture(t, ctx, client)
	svc := &Service{pg: client}

	require.NoError(t, svc.DeleteFile(ctx, fx.scope, fx.fileID))
	assert.Empty(t, loadFileAccesses(t, ctx, client, fx.scope, fx.accessID, fx.fileID))
}
