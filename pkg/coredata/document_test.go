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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestDocument_LoadPublishedByID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := t.Context()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()
	publishedMajor := 1
	publishedMinor := 0
	publishedAt := now

	document := coredata.Document{
		ID:                    gid.New(tenantID, coredata.DocumentEntityType),
		OrganizationID:        organizationID,
		CurrentPublishedMajor: &publishedMajor,
		CurrentPublishedMinor: &publishedMinor,
		WriteMode:             coredata.DocumentWriteModeAuthored,
		Status:                coredata.DocumentStatusActive,
		Title:                 "Working title",
		Classification:        coredata.DocumentClassificationConfidential,
		DocumentType:          coredata.DocumentTypeOther,
		Orientation:           coredata.DocumentVersionOrientationPortrait,
		Content:               "working draft",
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	version := coredata.DocumentVersion{
		ID:             gid.New(tenantID, coredata.DocumentVersionEntityType),
		OrganizationID: organizationID,
		DocumentID:     document.ID,
		Title:          "Published title",
		Major:          publishedMajor,
		Minor:          publishedMinor,
		Classification: coredata.DocumentClassificationPublic,
		DocumentType:   coredata.DocumentTypePolicy,
		Content:        "published body",
		Status:         coredata.DocumentVersionStatusPublished,
		Orientation:    coredata.DocumentVersionOrientationLandscape,
		PublishedAt:    &publishedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		organization := coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "published-document-" + organizationID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := organization.Insert(ctx, tx); err != nil {
			return err
		}

		if err := document.Insert(ctx, tx, scope); err != nil {
			return err
		}

		return version.Insert(ctx, tx, scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			return (&coredata.Organization{}).Delete(ctx, tx, organizationID)
		})
	})

	t.Run("returns the published snapshot", func(t *testing.T) {
		t.Parallel()

		var loaded coredata.Document
		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadPublishedByID(ctx, conn, scope, document.ID)
		}))

		assert.Equal(t, document.ID, loaded.ID)
		assert.Equal(t, organizationID, loaded.OrganizationID)
		assert.Equal(t, "Published title", loaded.Title)
		assert.Equal(t, "published body", loaded.Content)
		assert.Equal(t, coredata.DocumentClassificationPublic, loaded.Classification)
		assert.Equal(t, coredata.DocumentTypePolicy, loaded.DocumentType)
		assert.Equal(t, coredata.DocumentVersionOrientationLandscape, loaded.Orientation)
		require.NotNil(t, loaded.CurrentPublishedMajor)
		assert.Equal(t, publishedMajor, *loaded.CurrentPublishedMajor)
	})

	t.Run("hides the document from another tenant", func(t *testing.T) {
		t.Parallel()

		otherScope := coredata.NewScope(gid.NewTenantID(), gid.Nil)
		err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			var loaded coredata.Document
			return loaded.LoadPublishedByID(ctx, conn, otherScope, document.ID)
		})

		require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	})

	t.Run("loads across tenants when unscoped", func(t *testing.T) {
		t.Parallel()

		var loaded coredata.Document
		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadPublishedByID(ctx, conn, coredata.NewNoScope(), document.ID)
		}))

		assert.Equal(t, "Published title", loaded.Title)
	})
}
