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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

type (
	CompliancePortalThirdParty struct {
		ID                 gid.GID   `db:"id"`
		OrganizationID     gid.GID   `db:"organization_id"`
		CompliancePortalID gid.GID   `db:"trust_center_id"`
		ThirdPartyID       gid.GID   `db:"third_party_id"`
		CreatedAt          time.Time `db:"created_at"`
		UpdatedAt          time.Time `db:"updated_at"`
	}

	CompliancePortalThirdParties []*CompliancePortalThirdParty
)

func (cptp *CompliancePortalThirdParty) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cp_third_parties WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (cptp *CompliancePortalThirdParty) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cp_third_parties (
	id,
	tenant_id,
	organization_id,
	trust_center_id,
	third_party_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@trust_center_id,
	@third_party_id,
	@created_at,
	@updated_at
)
ON CONFLICT (trust_center_id, third_party_id) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"id":              cptp.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": cptp.OrganizationID,
		"trust_center_id": cptp.CompliancePortalID,
		"third_party_id":  cptp.ThirdPartyID,
		"created_at":      cptp.CreatedAt,
		"updated_at":      cptp.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert trust center third party: %w", err)
	}

	return nil
}

func DeleteCompliancePortalThirdPartyByCompliancePortalIDAndThirdPartyID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	compliancePortalID gid.GID,
	thirdPartyID gid.GID,
) error {
	q := `
DELETE FROM cp_third_parties
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND third_party_id = @third_party_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": compliancePortalID,
		"third_party_id":  thirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center third party: %w", err)
	}

	return nil
}

func DeleteCompliancePortalThirdPartyByID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
DELETE FROM cp_third_parties
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center third party: %w", err)
	}

	return nil
}

func (cptp *CompliancePortalThirdParty) LoadByCompliancePortalIDAndThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	thirdPartyID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	trust_center_id,
	third_party_id,
	created_at,
	updated_at
FROM
	cp_third_parties
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND third_party_id = @third_party_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": compliancePortalID,
		"third_party_id":  thirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center third party: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalThirdParty])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center third party: %w", err)
	}

	*cptp = row

	return nil
}

// LoadThirdPartyPublishedByCompliancePortalIDAndThirdPartyIDs reports which
// third parties are published on one compliance portal for a bounded page.
func LoadThirdPartyPublishedByCompliancePortalIDAndThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	thirdPartyIDs []gid.GID,
) (map[gid.GID]bool, error) {
	q := `
SELECT
	third_party_id
FROM
	cp_third_parties
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND third_party_id = ANY(@third_party_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": compliancePortalID,
		"third_party_ids": thirdPartyIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query trust center third parties: %w", err)
	}
	defer rows.Close()

	publishedByThirdPartyID := map[gid.GID]bool{}

	for rows.Next() {
		var thirdPartyID gid.GID

		if err := rows.Scan(&thirdPartyID); err != nil {
			return nil, fmt.Errorf("cannot scan trust center third party: %w", err)
		}

		publishedByThirdPartyID[thirdPartyID] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate trust center third parties: %w", err)
	}

	return publishedByThirdPartyID, nil
}

// LoadCompliancePortalThirdPartiesByCompliancePortalIDAndThirdPartyIDs loads
// portal third-party link rows for a bounded set of third-party IDs.
func LoadCompliancePortalThirdPartiesByCompliancePortalIDAndThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalID gid.GID,
	thirdPartyIDs []gid.GID,
) (map[gid.GID]*CompliancePortalThirdParty, error) {
	if len(thirdPartyIDs) == 0 {
		return map[gid.GID]*CompliancePortalThirdParty{}, nil
	}

	q := `
SELECT
	id,
	organization_id,
	trust_center_id,
	third_party_id,
	created_at,
	updated_at
FROM
	cp_third_parties
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND third_party_id = ANY(@third_party_ids::text[]);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": compliancePortalID,
		"third_party_ids": thirdPartyIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query trust center third parties: %w", err)
	}
	defer rows.Close()

	rowsByThirdPartyID := map[gid.GID]*CompliancePortalThirdParty{}

	for rows.Next() {
		row, err := pgx.RowToStructByName[CompliancePortalThirdParty](rows)
		if err != nil {
			return nil, fmt.Errorf("cannot scan trust center third party: %w", err)
		}

		rowsByThirdPartyID[row.ThirdPartyID] = &row
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot collect trust center third parties: %w", err)
	}

	return rowsByThirdPartyID, nil
}
