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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const documentPDFTestLease = 5 * time.Minute

type documentPDFClaimFixture struct {
	client  *pg.Client
	scope   coredata.Scoper
	version coredata.DocumentVersion
}

func newDocumentPDFClaimFixture(t *testing.T) documentPDFClaimFixture {
	t.Helper()

	// Migrate the real schema before copying it. Claims scan across tenants, so
	// connection-local tables keep parallel fixtures out of each other's queue.
	test.PGClient(t)
	opts, err := test.PGClientOptions()
	require.NoError(t, err)
	opts = append(
		opts,
		pg.WithPoolSize(1),
		pg.WithRegisterer(prometheus.NewRegistry()),
	)
	client, err := pg.NewClient(opts...)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	now := time.Now().UTC().Truncate(time.Microsecond)
	document := coredata.Document{
		ID:             gid.New(tenantID, coredata.DocumentEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		WriteMode:      coredata.DocumentWriteModeAuthored,
		Status:         coredata.DocumentStatusActive,
		Title:          "PDF claim test",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	version := coredata.DocumentVersion{
		ID:             gid.New(tenantID, coredata.DocumentVersionEntityType),
		OrganizationID: document.OrganizationID,
		DocumentID:     document.ID,
		Title:          document.Title,
		Major:          1,
		Classification: coredata.DocumentClassificationInternal,
		DocumentType:   coredata.DocumentTypePolicy,
		Content:        "policy",
		Status:         coredata.DocumentVersionStatusPublished,
		Orientation:    coredata.DocumentVersionOrientationPortrait,
		PublishedAt:    new(now),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := tx.Exec(
					ctx,
					`
CREATE TEMP TABLE documents (LIKE public.documents INCLUDING ALL);
CREATE TEMP TABLE document_versions (LIKE public.document_versions INCLUDING ALL);
`,
				); err != nil {
					return err
				}
				if err := document.Insert(ctx, tx, scope); err != nil {
					return err
				}
				return version.Insert(ctx, tx, scope)
			},
		),
	)

	return documentPDFClaimFixture{client: client, scope: scope, version: version}
}

func (f documentPDFClaimFixture) claim(t *testing.T, now time.Time) (coredata.DocumentVersion, error) {
	t.Helper()

	var claimed coredata.DocumentVersion
	err := f.client.WithTx(
		t.Context(),
		func(ctx context.Context, tx pg.Tx) error {
			return claimed.ClaimNextPublishedWithoutFileForUpdate(ctx, tx, 3, now, documentPDFTestLease)
		},
	)
	return claimed, err
}

func (f documentPDFClaimFixture) load(t *testing.T) coredata.DocumentVersion {
	t.Helper()

	var version coredata.DocumentVersion
	require.NoError(
		t,
		f.client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return version.LoadByID(ctx, conn, f.scope, f.version.ID)
			},
		),
	)
	return version
}

func (f documentPDFClaimFixture) release(t *testing.T, version *coredata.DocumentVersion, now time.Time) {
	t.Helper()
	require.NoError(
		t,
		f.client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return version.ReleasePDFClaim(ctx, tx, f.scope, now)
			},
		),
	)
}

func TestDocumentVersion_ClaimPDFLease(t *testing.T) {
	t.Parallel()

	f := newDocumentPDFClaimFixture(t)
	now := time.Date(2026, time.November, 10, 1, 0, 0, 123456789, time.UTC)
	first, err := f.claim(t, now)
	require.NoError(t, err)
	assert.Equal(t, f.version.ID, first.ID)
	assert.Equal(t, 1, first.PdfAttemptCount)
	require.NotNil(t, first.PdfClaimedAt)
	assert.Equal(t, now.Truncate(time.Microsecond), *first.PdfClaimedAt)
	persisted := f.load(t)
	require.NotNil(t, persisted.PdfClaimedAt)
	assert.WithinDuration(t, *first.PdfClaimedAt, *persisted.PdfClaimedAt, 0)

	// claim() commits before returning: no row lock can hide a missing lease.
	_, err = f.claim(t, now.Add(documentPDFTestLease/2))
	assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable)
	_, err = f.claim(t, first.PdfClaimedAt.Add(documentPDFTestLease))
	assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable)

	recoveredAt := first.PdfClaimedAt.Add(documentPDFTestLease + time.Microsecond)
	second, err := f.claim(t, recoveredAt)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, 2, second.PdfAttemptCount)

	f.release(t, &first, recoveredAt.Add(time.Second))
	persisted = f.load(t)
	require.NotNil(t, persisted.PdfClaimedAt)
	assert.WithinDuration(t, *second.PdfClaimedAt, *persisted.PdfClaimedAt, 0, "stale release must preserve the newer claim")
	_, err = f.claim(t, recoveredAt.Add(time.Second))
	assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable)

	// The owner can release using the returned timestamp, then retry at once.
	f.release(t, &second, recoveredAt.Add(time.Second))
	assert.Nil(t, f.load(t).PdfClaimedAt)
	third, err := f.claim(t, recoveredAt.Add(time.Second))
	require.NoError(t, err)
	assert.Equal(t, 3, third.PdfAttemptCount)
	_, err = f.claim(t, recoveredAt.Add(2*documentPDFTestLease))
	assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable, "expiry must not bypass the attempt cap")
}

func TestDocumentVersion_UpdatePreservesPDFClaim(t *testing.T) {
	t.Parallel()

	f := newDocumentPDFClaimFixture(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	claimed, err := f.claim(t, now)
	require.NoError(t, err)
	snapshot := f.version
	snapshot.Title = "Updated title"
	snapshot.PdfAttemptCount = claimed.PdfAttemptCount
	snapshot.UpdatedAt = now.Add(time.Hour)
	require.NoError(
		t,
		f.client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return snapshot.Update(ctx, tx, f.scope)
			},
		),
	)
	persisted := f.load(t)
	require.NotNil(t, persisted.PdfClaimedAt)
	assert.WithinDuration(t, *claimed.PdfClaimedAt, *persisted.PdfClaimedAt, 0)
	assert.Equal(t, snapshot.Title, persisted.Title)
	_, err = f.claim(t, now.Add(documentPDFTestLease/2))
	assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable)
	_, err = f.claim(t, now.Add(documentPDFTestLease+time.Microsecond))
	assert.NoError(t, err, "ordinary updates must not extend the claim")
}

func TestDocumentVersion_ClaimPDFEligibility(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		query string
		args  []any
	}{
		"unpublished":      {query: `UPDATE document_versions SET status = $1`, args: []any{coredata.DocumentVersionStatusPendingApproval}},
		"completed":        {query: `UPDATE document_versions SET file_id = id`},
		"exhausted":        {query: `UPDATE document_versions SET pdf_attempt_count = 3`},
		"deleted document": {query: `UPDATE documents SET deleted_at = NOW()`},
	} {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()
				f := newDocumentPDFClaimFixture(t)
				require.NoError(
					t,
					f.client.WithConn(
						t.Context(),
						func(ctx context.Context, conn pg.Querier) error {
							_, err := conn.Exec(ctx, tc.query, tc.args...)
							return err
						},
					),
				)
				_, err := f.claim(t, time.Now())
				assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable)
			},
		)
	}
}

func TestDocumentVersion_AttachPDFFileRejectsStaleClaim(t *testing.T) {
	t.Parallel()

	f := newDocumentPDFClaimFixture(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	first, err := f.claim(t, now)
	require.NoError(t, err)
	second, err := f.claim(t, now.Add(documentPDFTestLease+time.Microsecond))
	require.NoError(t, err)

	firstFileID := gid.New(f.scope.GetTenantID(), coredata.FileEntityType)
	secondFileID := gid.New(f.scope.GetTenantID(), coredata.FileEntityType)
	attach := func(version *coredata.DocumentVersion, scope coredata.Scoper, fileID gid.GID) bool {
		t.Helper()
		var attached bool
		require.NoError(
			t,
			f.client.WithTx(
				t.Context(),
				func(ctx context.Context, tx pg.Tx) error {
					var err error
					attached, err = version.AttachPDFFile(ctx, tx, scope, fileID, now.Add(2*documentPDFTestLease))
					return err
				},
			),
		)
		return attached
	}

	assert.False(t, attach(&f.version, f.scope, firstFileID), "an unclaimed snapshot must not attach")
	assert.False(t, attach(&first, f.scope, firstFileID), "a replaced claim must not attach")
	wrongScope := coredata.NewScope(gid.NewTenantID(), gid.Nil)
	assert.False(t, attach(&second, wrongScope, secondFileID), "attachment must stay tenant scoped")
	setStatus := func(status coredata.DocumentVersionStatus) {
		t.Helper()
		require.NoError(
			t,
			f.client.WithConn(
				t.Context(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, `UPDATE document_versions SET status = $1`, status)
					return err
				},
			),
		)
	}
	setStatus(coredata.DocumentVersionStatusVoided)
	assert.False(t, attach(&second, f.scope, secondFileID), "a voided version must not attach")
	setStatus(coredata.DocumentVersionStatusPublished)
	assert.True(t, attach(&second, f.scope, secondFileID), "the current owner must attach")
	assert.False(t, attach(&first, f.scope, firstFileID), "a late owner must not overwrite the completed file")
	assert.False(t, attach(&second, f.scope, firstFileID), "even the same owner must not overwrite a completed file")
	persisted := f.load(t)
	assert.Equal(t, new(secondFileID), persisted.FileID)
	assert.Equal(t, 2, persisted.PdfAttemptCount)
	f.release(t, &second, now.Add(3*documentPDFTestLease))
	assert.NotNil(t, f.load(t).PdfClaimedAt, "release must leave completed rows alone")
	_, err = f.claim(t, now.Add(4*documentPDFTestLease))
	assert.ErrorIs(t, err, coredata.ErrNoDocumentPDFJobAvailable)
}

func TestDocumentVersion_ReleasePDFClaimOwner(t *testing.T) {
	t.Parallel()

	f := newDocumentPDFClaimFixture(t)
	now := time.Date(2026, time.November, 10, 1, 0, 0, 987654321, time.UTC)
	claimed, err := f.claim(t, now)
	require.NoError(t, err)
	f.release(t, &f.version, now)
	assert.NotNil(t, f.load(t).PdfClaimedAt, "an unclaimed snapshot must not release")
	other := claimed
	require.NoError(
		t,
		f.client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return other.ReleasePDFClaim(ctx, tx, coredata.NewScope(gid.NewTenantID(), gid.Nil), now)
			},
		),
	)
	assert.NotNil(t, f.load(t).PdfClaimedAt, "release must stay tenant scoped")
	f.release(t, &claimed, now)
	assert.Nil(t, f.load(t).PdfClaimedAt)
	retried, err := f.claim(t, now.Add(time.Microsecond))
	require.NoError(t, err)
	assert.Equal(t, 2, retried.PdfAttemptCount)
}
