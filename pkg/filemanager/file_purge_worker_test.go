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

package filemanager

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestDeletedFilePurgeHandler_Run_PurgesExpiredFiles(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	expired := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))
	recent := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-30*time.Minute)))
	active := insertPurgeTestFile(t, pgClient, scope, organizationID, now, nil)

	var (
		mu          sync.Mutex
		deletedKeys []string
	)

	svc := newPurgeTestService(
		t,
		pgClient,
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)

			mu.Lock()

			deletedKeys = append(deletedKeys, r.URL.Path)

			mu.Unlock()

			w.WriteHeader(http.StatusNoContent)
		},
	)

	h := newPurgeHandler(svc, scope, now, time.Hour, 10, 10)
	require.NoError(t, h.Run(t.Context()))

	mu.Lock()
	assert.Contains(t, deletedKeys, "/"+expired.BucketName+"/"+expired.FileKey)
	assert.NotContains(t, deletedKeys, "/"+recent.BucketName+"/"+recent.FileKey)
	assert.NotContains(t, deletedKeys, "/"+active.BucketName+"/"+active.FileKey)
	mu.Unlock()

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				var gone coredata.File
				require.ErrorIs(
					t,
					gone.LoadByID(ctx, conn, scope, expired.ID),
					coredata.ErrResourceNotFound,
				)

				if err := recent.LoadByID(ctx, conn, scope, recent.ID); err != nil {
					return err
				}

				return active.LoadByID(ctx, conn, scope, active.ID)
			},
		),
	)
}

func TestDeletedFilePurgeHandler_Run_KeepsRowWhenS3DeleteFails(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	expired := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))

	svc := newPurgeTestService(
		t,
		pgClient,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		},
	)

	h := newPurgeHandler(svc, scope, now, time.Hour, 10, 10)
	err := h.Run(t.Context())
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot delete file from S3")

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return expired.LoadByID(ctx, conn, scope, expired.ID)
			},
		),
	)
}

func TestDeletedFilePurgeHandler_Run_SkipsS3WhenFileStillReferenced(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	expired := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				org := &coredata.Organization{
					ID:         organizationID,
					TenantID:   tenantID,
					Name:       "Referenced File Org",
					LogoFileID: &expired.ID,
					CreatedAt:  now,
					UpdatedAt:  now,
				}

				return org.Insert(ctx, tx)
			},
		),
	)

	var s3Deletes int

	svc := newPurgeTestService(
		t,
		pgClient,
		func(w http.ResponseWriter, _ *http.Request) {
			s3Deletes++

			w.WriteHeader(http.StatusNoContent)
		},
	)

	h := newPurgeHandler(svc, scope, now, time.Hour, 10, 10)
	err := h.Run(t.Context())
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot delete file from database")
	assert.ErrorContains(t, err, "file is still referenced")
	assert.Zero(t, s3Deletes)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return expired.LoadByID(ctx, conn, scope, expired.ID)
			},
		),
	)
}

func TestDeletedFilePurgeHandler_Run_StopsAtMaxPerTick(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	first := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))
	second := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))
	third := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))

	svc := newPurgeTestService(
		t,
		pgClient,
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	h := newPurgeHandler(svc, scope, now, time.Hour, 10, 2)
	require.NoError(t, h.Run(t.Context()))

	remaining := 0

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				for _, file := range []*coredata.File{first, second, third} {
					var loaded coredata.File

					err := loaded.LoadByID(ctx, conn, scope, file.ID)
					if err == nil {
						remaining++
					}
				}

				return nil
			},
		),
	)
	assert.Equal(t, 1, remaining)
}

func TestDeletedFilePurgeWorker_Run_CompletesInitialSweepWhenCanceled(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	expired := insertPurgeTestFile(t, pgClient, scope, organizationID, now, new(now.Add(-48*time.Hour)))

	var s3Deletes int

	svc := newPurgeTestService(
		t,
		pgClient,
		func(w http.ResponseWriter, _ *http.Request) {
			s3Deletes++

			w.WriteHeader(http.StatusNoContent)
		},
	)

	logger := log.NewLogger(log.WithOutput(io.Discard))
	handler := newPurgeHandler(svc, scope, now, time.Hour, 10, 10)
	w := &DeletedFilePurgeWorker{
		handler: handler,
		periodic: worker.NewPeriodic(
			"file-purge-worker-cancel-test",
			handler,
			logger,
			worker.WithInterval(time.Hour),
		),
		logger: logger,
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := w.Run(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, s3Deletes)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				var gone coredata.File

				require.ErrorIs(
					t,
					gone.LoadByID(ctx, conn, scope, expired.ID),
					coredata.ErrResourceNotFound,
				)

				return nil
			},
		),
	)
}

func newPurgeHandler(
	svc *Service,
	scope coredata.Scoper,
	now time.Time,
	retention time.Duration,
	batchSize int,
	maxPerTick int,
) *deletedFilePurgeHandler {
	return &deletedFilePurgeHandler{
		svc:        svc,
		logger:     log.NewLogger(log.WithOutput(io.Discard)),
		scope:      scope,
		retention:  retention,
		batchSize:  batchSize,
		maxPerTick: maxPerTick,
		now:        func() time.Time { return now },
	}
}

func newPurgeTestService(
	t *testing.T,
	pgClient *pg.Client,
	handler http.HandlerFunc,
) *Service {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	s3Client := awss3.NewFromConfig(
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("access-key", "secret-key", ""),
		},
		func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(srv.URL)
			o.UsePathStyle = true
		},
	)

	return NewService(pgClient, nil, s3Client, log.NewLogger(log.WithOutput(io.Discard)))
}

func insertPurgeTestFile(
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
		FileName:       "purge.pdf",
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
