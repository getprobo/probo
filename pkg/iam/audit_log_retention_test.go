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

package iam

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAuditLogRetentionHandler_DeletesEntriesOlderThanRetention(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	orgA := insertRetentionOrganization(t, client, now)
	orgB := insertRetentionOrganization(t, client, now)

	insertRetentionAuditLog(
		t,
		client,
		orgA,
		now.Add(-DefaultAuditLogRetention).Add(-time.Second),
	)
	boundaryA := insertRetentionAuditLog(
		t,
		client,
		orgA,
		now.Add(-DefaultAuditLogRetention),
	)
	recentA := insertRetentionAuditLog(
		t,
		client,
		orgA,
		now.Add(-time.Hour),
	)
	insertRetentionAuditLog(
		t,
		client,
		orgB,
		now.Add(-DefaultAuditLogRetention).Add(-time.Minute),
	)

	h := &auditLogRetentionHandler{
		pg:        client,
		logger:    log.NewLogger(log.WithOutput(io.Discard)),
		retention: DefaultAuditLogRetention,
		batchSize: 1,
		now:       func() time.Time { return now },
	}

	require.NoError(t, h.Run(t.Context()))

	assert.Equal(
		t,
		[]gid.GID{boundaryA, recentA},
		loadRetentionAuditLogIDs(t, client, orgA.organizationID),
	)
	assert.Empty(t, loadRetentionAuditLogIDs(t, client, orgB.organizationID))

	require.NoError(t, h.Run(t.Context()))

	assert.Equal(
		t,
		[]gid.GID{boundaryA, recentA},
		loadRetentionAuditLogIDs(t, client, orgA.organizationID),
	)
}

func TestAuditLogRetentionHandler_StopsWhenCanceled(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	// Ahead of the real clock. A parallel sweep deletes rows older than
	// 90 days from time.Now(), so this row stays out of its range.
	now := time.Now().UTC().Add(2 * DefaultAuditLogRetention).Round(time.Microsecond)
	org := insertRetentionOrganization(t, client, now)
	kept := insertRetentionAuditLog(
		t,
		client,
		org,
		now.Add(-DefaultAuditLogRetention).Add(-time.Second),
	)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	h := &auditLogRetentionHandler{
		pg:        client,
		logger:    log.NewLogger(log.WithOutput(io.Discard)),
		retention: DefaultAuditLogRetention,
		batchSize: 1,
		now:       func() time.Time { return now },
	}

	err := h.Run(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, []gid.GID{kept}, loadRetentionAuditLogIDs(t, client, org.organizationID))
}

type retentionOrganization struct {
	tenantID       gid.TenantID
	organizationID gid.GID
}

func insertRetentionOrganization(
	t *testing.T,
	client *pg.Client,
	now time.Time,
) retentionOrganization {
	t.Helper()

	tenantID := gid.NewTenantID()
	org := retentionOrganization{
		tenantID:       tenantID,
		organizationID: gid.New(tenantID, coredata.OrganizationEntityType),
	}

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := tx.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at, asset_inventory_enabled)
					 VALUES ($1, $2, $3, $4, $5, false)`,
					org.organizationID.String(),
					org.tenantID.String(),
					"audit-log-retention-"+org.organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					"DELETE FROM organizations WHERE id = $1",
					org.organizationID.String(),
				)

				return err
			},
		)
	})

	return org
}

func insertRetentionAuditLog(
	t *testing.T,
	client *pg.Client,
	org retentionOrganization,
	createdAt time.Time,
) gid.GID {
	t.Helper()

	entry := &coredata.AuditLogEntry{
		ID:             gid.New(org.tenantID, coredata.AuditLogEntryEntityType),
		OrganizationID: org.organizationID,
		ActorID:        gid.New(org.tenantID, coredata.IdentityEntityType),
		ActorType:      coredata.AuditLogActorTypeSystem,
		Action:         "iam:organization:get",
		ResourceType:   "organization",
		ResourceID:     org.organizationID,
		Outcome:        coredata.AuditLogOutcomeAllowed,
		Metadata:       []byte(`{}`),
		CreatedAt:      createdAt,
		RecordedAt:     createdAt,
	}

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return entry.Insert(ctx, tx, coredata.NewScope(org.tenantID, gid.Nil))
			},
		),
	)

	return entry.ID
}

func loadRetentionAuditLogIDs(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
) []gid.GID {
	t.Helper()

	var ids []gid.GID

	require.NoError(
		t,
		client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := conn.Query(
					ctx,
					`SELECT id
					 FROM audit_log_entries
					 WHERE organization_id = @organization_id
					 ORDER BY created_at, id`,
					pgx.StrictNamedArgs{"organization_id": organizationID},
				)
				if err != nil {
					return err
				}

				defer rows.Close()

				for rows.Next() {
					var id gid.GID
					if err := rows.Scan(&id); err != nil {
						return err
					}

					ids = append(ids, id)
				}

				return rows.Err()
			},
		),
	)

	return ids
}
