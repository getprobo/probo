// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	AssetThirdParty struct {
		AssetID      gid.GID      `db:"asset_id"`
		ThirdPartyID gid.GID      `db:"third_party_id"`
		TenantID     gid.TenantID `db:"tenant_id"`
		CreatedAt    time.Time    `db:"created_at"`
	}

	AssetThirdParties []*AssetThirdParty
)

func (av AssetThirdParties) Merge(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
	organizationID gid.GID,
	thirdPartyIDs []gid.GID,
) error {
	q := `
WITH third_party_ids AS (
	SELECT
		unnest(@third_party_ids::text[]) AS third_party_id,
		@tenant_id AS tenant_id,
		@asset_id AS asset_id,
		@organization_id AS organization_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO asset_third_parties AS tgt
USING third_party_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.asset_id = src.asset_id
	AND tgt.third_party_id = src.third_party_id
WHEN NOT MATCHED
	THEN INSERT (tenant_id, asset_id, third_party_id, organization_id, created_at)
		VALUES (src.tenant_id, src.asset_id, src.third_party_id, src.organization_id, src.created_at)
	WHEN NOT MATCHED BY SOURCE
		AND tgt.tenant_id = @tenant_id AND tgt.asset_id = @asset_id
		THEN DELETE
	`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"asset_id":        assetID,
		"organization_id": organizationID,
		"created_at":      time.Now(),
		"third_party_ids": thirdPartyIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot merge asset third_parties: %w", err)
	}

	return nil
}

func (av AssetThirdParties) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	assetID gid.GID,
	organizationID gid.GID,
	thirdPartyIDs []gid.GID,
) error {
	q := `
WITH third_party_ids AS (
	SELECT unnest(@third_party_ids::text[]) AS third_party_id
)
INSERT INTO asset_third_parties (tenant_id, asset_id, third_party_id, organization_id, created_at)
SELECT
	@tenant_id AS tenant_id,
	@asset_id AS asset_id,
	third_party_id,
	@organization_id AS organization_id,
	@created_at AS created_at
FROM third_party_ids
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"asset_id":        assetID,
		"organization_id": organizationID,
		"created_at":      time.Now(),
		"third_party_ids": thirdPartyIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert asset third_parties: %w", err)
	}

	return nil
}
