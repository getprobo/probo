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
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestFiles_LoadDeletedBefore_ExcludesRecentAndActive(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	cutoff := now.Add(-30 * 24 * time.Hour)

	expired := insertTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-31*24*time.Hour)))
	recent := insertTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-2*24*time.Hour)))
	active := insertTestFile(t, pgClient, scope, organizationID, now, nil)

	var files coredata.Files

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return files.LoadDeletedBefore(
					ctx,
					conn,
					scope,
					cutoff,
					nil,
					100,
				)
			},
		),
	)

	require.Len(t, files, 1)
	assert.Equal(t, expired.ID, files[0].ID)

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if err := recent.LoadByID(ctx, conn, scope, recent.ID); err != nil {
					return err
				}

				return active.LoadByID(ctx, conn, scope, active.ID)
			},
		),
	)
}

func TestFiles_LoadDeletedBefore_Paginates(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	cutoff := now.Add(-30 * 24 * time.Hour)
	deletedAt := now.Add(-31 * 24 * time.Hour)

	first := insertTestFile(t, pgClient, scope, organizationID, now, new(deletedAt))
	second := insertTestFile(t, pgClient, scope, organizationID, now, new(deletedAt))

	if first.ID.String() > second.ID.String() {
		first, second = second, first
	}

	var page coredata.Files

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return page.LoadDeletedBefore(ctx, conn, scope, cutoff, nil, 1)
			},
		),
	)
	require.Len(t, page, 1)
	assert.Equal(t, first.ID, page[0].ID)

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return page.LoadDeletedBefore(ctx, conn, scope, cutoff, &page[0].ID, 1)
			},
		),
	)
	require.Len(t, page, 1)
	assert.Equal(t, second.ID, page[0].ID)
}

func TestFile_LoadDeletedByIDForUpdateSkipLocked_RechecksCutoff(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	cutoff := now.Add(-30 * 24 * time.Hour)

	expired := insertTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-31*24*time.Hour)))
	redeleted := insertTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-31*24*time.Hour)))

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return redeleted.SoftDelete(ctx, tx, scope)
			},
		),
	)

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var locked coredata.File
				if err := locked.LoadDeletedByIDForUpdateSkipLocked(
					ctx,
					tx,
					scope,
					expired.ID,
					cutoff,
				); err != nil {
					return err
				}

				assert.Equal(t, expired.ID, locked.ID)

				var skipped coredata.File
				require.ErrorIs(
					t,
					skipped.LoadDeletedByIDForUpdateSkipLocked(
						ctx,
						tx,
						scope,
						redeleted.ID,
						cutoff,
					),
					coredata.ErrResourceNotFound,
				)

				return nil
			},
		),
	)
}

func TestFile_Delete(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	deleted := insertTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-31*24*time.Hour)))
	active := insertTestFile(t, pgClient, scope, organizationID, now, nil)

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := deleted.Delete(ctx, tx, scope); err != nil {
					return err
				}

				return active.Delete(ctx, tx, scope)
			},
		),
	)

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				var gone coredata.File
				require.ErrorIs(t, gone.LoadByID(ctx, conn, scope, deleted.ID), coredata.ErrResourceNotFound)

				return active.LoadByID(ctx, conn, scope, active.ID)
			},
		),
	)
}

func TestFile_Delete_WhenStillReferenced(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	file := insertTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-31*24*time.Hour)))

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				org := &coredata.Organization{
					ID:         organizationID,
					TenantID:   tenantID,
					Name:       "Referenced File Org",
					LogoFileID: &file.ID,
					CreatedAt:  now,
					UpdatedAt:  now,
				}

				return org.Insert(ctx, tx)
			},
		),
	)

	err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return file.Delete(ctx, tx, scope)
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceInUse)
	assert.ErrorContains(t, err, "file is still referenced")

	require.NoError(
		t,
		pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return file.LoadByID(ctx, conn, scope, file.ID)
			},
		),
	)
}

func insertTestFile(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	now time.Time,
	deletedAt *time.Time,
) *coredata.File {
	t.Helper()

	objectKey, err := uuid.NewV7()
	require.NoError(t, err)

	file := &coredata.File{
		ID:             gid.New(organizationID.TenantID(), coredata.FileEntityType),
		OrganizationID: organizationID,
		BucketName:     "uploads",
		MimeType:       "application/pdf",
		FileName:       "test.pdf",
		FileKey:        objectKey.String(),
		FileSize:       1,
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      deletedAt,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return file.Insert(ctx, tx, scope)
			},
		),
	)

	return file
}
