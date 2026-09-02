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
	ThirdPartyAdministrator struct {
		ThirdPartyID           gid.GID   `db:"third_party_id"`
		AdministratorProfileID gid.GID   `db:"administrator_profile_id"`
		OrganizationID         gid.GID   `db:"organization_id"`
		CreatedAt              time.Time `db:"created_at"`
		UpdatedAt              time.Time `db:"updated_at"`
	}

	ThirdPartyAdministrators []*ThirdPartyAdministrator
)

// LoadByThirdPartyID loads all administrators for a third party.
func (as *ThirdPartyAdministrators) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
) error {
	q := `
SELECT
	third_party_id,
	administrator_profile_id,
	organization_id,
	created_at,
	updated_at
FROM third_party_administrators
WHERE
	%s
	AND third_party_id = @third_party_id
ORDER BY created_at ASC, administrator_profile_id ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query third party administrators: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyAdministrator])
	if err != nil {
		return fmt.Errorf("cannot collect third party administrators: %w", err)
	}

	*as = result

	return nil
}

// LoadByThirdPartyIDs loads all administrators for the given third parties.
func (as *ThirdPartyAdministrators) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*as = ThirdPartyAdministrators{}
		return nil
	}

	q := `
SELECT
	third_party_id,
	administrator_profile_id,
	organization_id,
	created_at,
	updated_at
FROM third_party_administrators
WHERE
	%s
	AND third_party_id = ANY(@third_party_ids)
ORDER BY created_at ASC, administrator_profile_id ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_ids": thirdPartyIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query third party administrators: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyAdministrator])
	if err != nil {
		return fmt.Errorf("cannot collect third party administrators: %w", err)
	}

	*as = result

	return nil
}

// MergeByThirdPartyID merges the given administrator profile IDs for a third party,
// inserting new ones, keeping existing ones, and deleting removed ones.
func (as *ThirdPartyAdministrators) MergeByThirdPartyID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	thirdPartyID gid.GID,
	organizationID gid.GID,
	administratorProfileIDs []gid.GID,
) error {
	q := `
MERGE INTO third_party_administrators AS target
USING (
	SELECT unnest(@administrator_profile_ids::text[]) AS administrator_profile_id
) AS source
ON
	%[1]s
	AND target.third_party_id = @third_party_id
	AND target.administrator_profile_id = source.administrator_profile_id
WHEN NOT MATCHED THEN
	INSERT (third_party_id, administrator_profile_id, tenant_id, organization_id, created_at, updated_at)
	VALUES (@third_party_id, source.administrator_profile_id, @tenant_id, @organization_id, @now, @now)
WHEN NOT MATCHED BY SOURCE
	AND %[1]s
	AND target.third_party_id = @third_party_id THEN
	DELETE;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	now := time.Now()

	ids := make([]string, len(administratorProfileIDs))
	for i, id := range administratorProfileIDs {
		ids[i] = id.String()
	}

	args := pgx.StrictNamedArgs{
		"third_party_id":            thirdPartyID,
		"administrator_profile_ids": ids,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           organizationID,
		"now":                       now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot merge third party administrators: %w", err)
	}

	result := make(ThirdPartyAdministrators, 0, len(administratorProfileIDs))
	for _, profileID := range administratorProfileIDs {
		result = append(
			result,
			&ThirdPartyAdministrator{
				ThirdPartyID:           thirdPartyID,
				AdministratorProfileID: profileID,
				OrganizationID:         organizationID,
				CreatedAt:              now,
				UpdatedAt:              now,
			},
		)
	}

	*as = result

	return nil
}

func (c *ThirdPartyAdministrators) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM third_party_administrators
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete third party administrators: %w", err)
	}

	return nil
}
