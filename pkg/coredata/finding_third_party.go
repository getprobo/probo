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
	FindingThirdParty struct {
		FindingID      gid.GID   `db:"finding_id"`
		ThirdPartyID   gid.GID   `db:"third_party_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	FindingThirdParties []*FindingThirdParty
)

func (fs FindingThirdParties) Merge(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	findingID gid.GID,
	organizationID gid.GID,
	thirdPartyIDs []gid.GID,
) error {
	q := `
WITH third_party_ids AS (
	SELECT
		unnest(@third_party_ids::text[]) AS third_party_id,
		@tenant_id AS tenant_id,
		@finding_id AS finding_id,
		@organization_id AS organization_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO finding_third_parties AS tgt
USING third_party_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.finding_id = src.finding_id
	AND tgt.third_party_id = src.third_party_id
WHEN NOT MATCHED THEN
	INSERT (tenant_id, finding_id, third_party_id, organization_id, created_at)
	VALUES (src.tenant_id, src.finding_id, src.third_party_id, src.organization_id, src.created_at)
WHEN NOT MATCHED BY SOURCE
	AND tgt.tenant_id = @tenant_id AND tgt.finding_id = @finding_id
	THEN DELETE
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"finding_id":      findingID,
		"organization_id": organizationID,
		"created_at":      time.Now(),
		"third_party_ids": thirdPartyIDs,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot merge finding third parties: %w", err)
	}

	return nil
}
