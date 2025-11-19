// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ProcessingActivityVendor struct {
		ProcessingActivityID gid.GID      `db:"processing_activity_id"`
		VendorID             gid.GID      `db:"vendor_id"`
		TenantID             gid.TenantID `db:"tenant_id"`
		SnapshotID           *gid.GID     `db:"snapshot_id"`
		CreatedAt            time.Time    `db:"created_at"`
	}

	ProcessingActivityVendors []*ProcessingActivityVendor

	ProcessingActivitySnapshotter interface {
		InsertProcessingActivitySnapshots(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error
	}
)

func (pav ProcessingActivityVendors) Merge(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	processingActivityID gid.GID,
	organizationID gid.GID,
	vendorIDs []gid.GID,
) error {
	q := `
WITH vendor_ids AS (
	SELECT
		unnest(@vendor_ids::text[]) AS vendor_id,
		@tenant_id AS tenant_id,
		@processing_activity_id AS processing_activity_id,
		@organization_id AS organization_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO processing_activity_vendors AS tgt
USING vendor_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.processing_activity_id = src.processing_activity_id
	AND tgt.vendor_id = src.vendor_id
WHEN NOT MATCHED
	THEN INSERT (tenant_id, processing_activity_id, vendor_id, organization_id, created_at)
		VALUES (src.tenant_id, src.processing_activity_id, src.vendor_id, src.organization_id, src.created_at)
	WHEN NOT MATCHED BY SOURCE
		AND tgt.tenant_id = @tenant_id AND tgt.processing_activity_id = @processing_activity_id
		THEN DELETE
	`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"processing_activity_id": processingActivityID,
		"organization_id":        organizationID,
		"created_at":             time.Now(),
		"vendor_ids":             vendorIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot merge processing activity vendors: %w", err)
	}

	return nil
}

func (pav ProcessingActivityVendors) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	processingActivityID gid.GID,
	organizationID gid.GID,
	vendorIDs []gid.GID,
) error {
	q := `
WITH vendor_ids AS (
	SELECT unnest(@vendor_ids::text[]) AS vendor_id
)
INSERT INTO processing_activity_vendors (tenant_id, processing_activity_id, vendor_id, organization_id, created_at)
SELECT
	@tenant_id AS tenant_id,
	@processing_activity_id AS processing_activity_id,
	vendor_id,
	@organization_id AS organization_id,
	@created_at AS created_at
FROM vendor_ids
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"processing_activity_id": processingActivityID,
		"organization_id":        organizationID,
		"created_at":             time.Now(),
		"vendor_ids":             vendorIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert processing activity vendors: %w", err)
	}

	return nil
}

func (pav ProcessingActivityVendors) InsertProcessingActivitySnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	source_processing_activities AS (
		SELECT id
		FROM processing_activities
		WHERE organization_id = @organization_id AND snapshot_id IS NULL
	),
	snapshot_processing_activities AS (
		SELECT id, source_id
		FROM processing_activities
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	),
	snapshot_vendors AS (
		SELECT id, source_id
		FROM vendors
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	),
	source_processing_activity_vendors AS (
		SELECT processing_activity_id, vendor_id, snapshot_id, created_at
		FROM processing_activity_vendors
		WHERE %s AND processing_activity_id = ANY(SELECT id FROM source_processing_activities) AND snapshot_id IS NULL
	)
INSERT INTO processing_activity_vendors (tenant_id, processing_activity_id, vendor_id, organization_id, snapshot_id, created_at)
SELECT
	@tenant_id,
	spa.id,
	sv.id,
	@organization_id,
	@snapshot_id,
	pav.created_at
FROM source_processing_activity_vendors pav
JOIN snapshot_processing_activities spa ON spa.source_id = pav.processing_activity_id
JOIN snapshot_vendors sv ON sv.source_id = pav.vendor_id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert processing activity vendor snapshots: %w", err)
	}

	return nil
}
