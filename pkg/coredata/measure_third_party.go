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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	MeasureThirdParty struct {
		MeasureID      gid.GID   `db:"measure_id"`
		ThirdPartyID   gid.GID   `db:"third_party_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	MeasureThirdParties []*MeasureThirdParty
)

// Upsert links a measure to a third party. The organization_id stored in the
// junction row is derived from the measures table inside the INSERT, so a
// caller cannot place the mapping into a different organization than the
// measure actually belongs to. Idempotent: re-linking an existing pair is a
// no-op.
func (mtp MeasureThirdParty) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    measures_third_parties (
        measure_id,
        third_party_id,
        organization_id,
        tenant_id,
        created_at
    )
SELECT
    @measure_id,
    @third_party_id,
    m.organization_id,
    @tenant_id,
    @created_at
FROM
    measures m
WHERE
    m.id = @measure_id
    AND m.tenant_id = @tenant_id
ON CONFLICT (measure_id, third_party_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"measure_id":     mtp.MeasureID,
		"third_party_id": mtp.ThirdPartyID,
		"tenant_id":      scope.GetTenantID(),
		"created_at":     mtp.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert measure third party: %w", err)
	}

	return nil
}

func (mtp MeasureThirdParty) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	measureID gid.GID,
	thirdPartyID gid.GID,
) error {
	q := `
DELETE
FROM
    measures_third_parties
WHERE
    %s
    AND measure_id = @measure_id
    AND third_party_id = @third_party_id;
`

	args := pgx.StrictNamedArgs{
		"measure_id":     measureID,
		"third_party_id": thirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)

	return err
}
