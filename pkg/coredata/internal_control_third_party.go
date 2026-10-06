// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	InternalControlThirdParty struct {
		InternalControlID gid.GID   `db:"internal_control_id"`
		ThirdPartyID      gid.GID   `db:"third_party_id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		CreatedAt         time.Time `db:"created_at"`
	}

	InternalControlThirdParties []*InternalControlThirdParty
)

// Upsert links an internal control to a third party. The organization_id stored in the
// junction row is derived from the internal controls table inside the INSERT, so a
// caller cannot place the mapping into a different organization than the
// internal control actually belongs to. Idempotent: re-linking an existing pair is a
// no-op.
func (mtp InternalControlThirdParty) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    internal_controls_third_parties (
        internal_control_id,
        third_party_id,
        organization_id,
        tenant_id,
        created_at
    )
SELECT
    @internal_control_id,
    @third_party_id,
    m.organization_id,
    @tenant_id,
    @created_at
FROM
    internal_controls m
WHERE
    m.id = @internal_control_id
    AND m.tenant_id = @tenant_id
ON CONFLICT (internal_control_id, third_party_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"internal_control_id": mtp.InternalControlID,
		"third_party_id":      mtp.ThirdPartyID,
		"tenant_id":           scope.GetTenantID(),
		"created_at":          mtp.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert internal control third party: %w", err)
	}

	return nil
}

func (mtp InternalControlThirdParty) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	internalControlID gid.GID,
	thirdPartyID gid.GID,
) error {
	q := `
DELETE
FROM
    internal_controls_third_parties
WHERE
    %s
    AND internal_control_id = @internal_control_id
    AND third_party_id = @third_party_id;
`

	args := pgx.StrictNamedArgs{
		"internal_control_id": internalControlID,
		"third_party_id":      thirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)

	return err
}
