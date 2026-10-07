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
	"go.probo.inc/probo/pkg/page"
)

type portalAccessResourceFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
	portalID       gid.GID
	accessID       gid.GID
	restrictedID   gid.GID
	publicID       gid.GID
	documentID     gid.GID
	identityID     gid.GID
}

func seedPortalAccessResourceFixture(t *testing.T, ctx context.Context, client *pg.Client) portalAccessResourceFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	var (
		portalID     gid.GID
		accessID     gid.GID
		restrictedID gid.GID
		publicID     gid.GID
		documentID   gid.GID
		identityID   gid.GID
	)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		organization := coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "access-resource-" + organizationID.String(),
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

		portalID = gid.New(tenantID, coredata.CompliancePortalEntityType)

		portal := coredata.CompliancePortal{
			ID:                   portalID,
			OrganizationID:       organizationID,
			TenantID:             tenantID,
			Active:               true,
			Slug:                 strings.ToLower(portalID.String()),
			SearchEngineIndexing: coredata.SearchEngineIndexingNotIndexable,
			Capabilities:         coredata.DefaultCompliancePortalCapabilities(),
			MailingListID:        mailingList.ID,
			EntityName:           "access-resource-portal",
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
			FullName:             "Access Resource",
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

		restrictedID, err = insertPortalAccessResourceFile(
			ctx,
			tx,
			scope,
			organizationID,
			portalID,
			"restricted-file",
			coredata.CompliancePortalVisibilityRestricted,
		)
		if err != nil {
			return err
		}

		publicID, err = insertPortalAccessResourceFile(
			ctx,
			tx,
			scope,
			organizationID,
			portalID,
			"public-file",
			coredata.CompliancePortalVisibilityPublic,
		)
		if err != nil {
			return err
		}

		documentID, err = insertPortalAccessResourceDocument(
			ctx,
			tx,
			scope,
			organizationID,
			portalID,
			accessID,
		)
		if err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_ = (&coredata.Identity{ID: identityID}).Delete(ctx, tx)

			return (&coredata.Organization{}).Delete(ctx, tx, organizationID)
		})
	})

	return portalAccessResourceFixture{
		scope:          scope,
		organizationID: organizationID,
		portalID:       portalID,
		accessID:       accessID,
		restrictedID:   restrictedID,
		publicID:       publicID,
		documentID:     documentID,
		identityID:     identityID,
	}
}

func insertPortalAccessResourceFile(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	portalID gid.GID,
	name string,
	visibility coredata.CompliancePortalVisibility,
) (gid.GID, error) {
	now := time.Now()

	objectKey, err := uuid.NewV7()
	if err != nil {
		return gid.Nil, err
	}

	blob := coredata.File{
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
	if err := blob.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	file := coredata.CompliancePortalFile{
		ID:                         gid.New(organizationID.TenantID(), coredata.CompliancePortalFileEntityType),
		OrganizationID:             organizationID,
		CompliancePortalID:         portalID,
		Name:                       name,
		Category:                   "OTHER",
		FileID:                     blob.ID,
		CompliancePortalVisibility: visibility,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
	if err := file.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	return file.ID, nil
}

func insertPortalAccessResourceDocument(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	portalID gid.GID,
	accessID gid.GID,
) (gid.GID, error) {
	now := time.Now()
	publishedMajor := 1
	publishedMinor := 0
	publishedAt := now

	document := coredata.Document{
		ID:                    gid.New(organizationID.TenantID(), coredata.DocumentEntityType),
		OrganizationID:        organizationID,
		CurrentPublishedMajor: &publishedMajor,
		CurrentPublishedMinor: &publishedMinor,
		WriteMode:             coredata.DocumentWriteModeAuthored,
		Status:                coredata.DocumentStatusActive,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := document.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	version := coredata.DocumentVersion{
		ID:             gid.New(organizationID.TenantID(), coredata.DocumentVersionEntityType),
		OrganizationID: organizationID,
		DocumentID:     document.ID,
		Title:          "Restricted policy",
		Major:          publishedMajor,
		Minor:          publishedMinor,
		Classification: coredata.DocumentClassificationInternal,
		DocumentType:   coredata.DocumentTypePolicy,
		Content:        "policy",
		Status:         coredata.DocumentVersionStatusPublished,
		Orientation:    coredata.DocumentVersionOrientationPortrait,
		PublishedAt:    &publishedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := version.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	portalDocument := coredata.CompliancePortalDocument{
		ID:                 gid.New(organizationID.TenantID(), coredata.CompliancePortalDocumentEntityType),
		OrganizationID:     organizationID,
		CompliancePortalID: portalID,
		DocumentID:         document.ID,
		Visibility:         coredata.CompliancePortalVisibilityRestricted,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := portalDocument.Upsert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	documentAccess := coredata.CompliancePortalDocumentAccess{
		ID:                         gid.New(organizationID.TenantID(), coredata.CompliancePortalDocumentAccessEntityType),
		OrganizationID:             organizationID,
		CompliancePortalAccessID:   accessID,
		CompliancePortalDocumentID: &portalDocument.ID,
		Status:                     coredata.CompliancePortalDocumentAccessStatusRequested,
		RequestedAt:                &now,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
	if err := documentAccess.Insert(ctx, tx, scope); err != nil {
		return gid.Nil, err
	}

	return document.ID, nil
}

func TestCompliancePortalAccessResources_LoadByCompliancePortalAccessID_ExcludesPublicFiles(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := t.Context()
	fx := seedPortalAccessResourceFixture(t, ctx, client)

	var resources coredata.CompliancePortalAccessResources

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			50,
			nil,
			page.Head,
			page.OrderBy[coredata.CompliancePortalAccessResourceOrderField]{
				Field:     coredata.CompliancePortalAccessResourceOrderFieldAccessStatus,
				Direction: page.OrderDirectionAsc,
			},
		)

		return resources.LoadByCompliancePortalAccessID(
			ctx,
			conn,
			fx.scope,
			fx.accessID,
			fx.organizationID,
			fx.portalID,
			cursor,
			coredata.NewCompliancePortalAccessResourceFilter(nil),
		)
	}))

	require.Len(t, resources, 2)

	byKind := make(map[coredata.CompliancePortalAccessResourceKind]*coredata.CompliancePortalAccessResource, len(resources))
	for _, resource := range resources {
		byKind[resource.Kind] = resource
	}

	fileResource, ok := byKind[coredata.CompliancePortalAccessResourceKindFile]
	require.True(t, ok)
	assert.Equal(t, fx.restrictedID, fileResource.ID)
	assert.NotEqual(t, fx.publicID, fileResource.ID)
	assert.Nil(t, fileResource.Status)

	documentResource, ok := byKind[coredata.CompliancePortalAccessResourceKindDocument]
	require.True(t, ok)
	assert.Equal(t, fx.documentID, documentResource.ID)
	require.NotNil(t, documentResource.Status)
	assert.Equal(t, coredata.CompliancePortalDocumentAccessStatusRequested, *documentResource.Status)
}
