// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
)

type (
	AccessEntryDecisionHistory struct {
		ID             gid.GID             `db:"id"`
		OrganizationID gid.GID             `db:"organization_id"`
		AccessEntry    gid.GID             `db:"access_entry_id"`
		Decision       AccessEntryDecision `db:"decision"`
		DecisionNote   *string             `db:"decision_note"`
		DecidedBy      *gid.GID            `db:"decided_by"`
		DecidedAt      time.Time           `db:"decided_at"`
		CreatedAt      time.Time           `db:"created_at"`
	}

	AccessEntryDecisionHistories []*AccessEntryDecisionHistory
)

func (h *AccessEntryDecisionHistory) Insert(
	ctx context.Context,
	conn pg.Tx,
	predicate Predicater,
) error {
	q := `
INSERT INTO access_entry_decision_history (
    id,
    tenant_id,
    organization_id,
    access_entry_id,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @access_entry_id,
    @decision,
    @decision_note,
    @decided_by,
    @decided_at,
    @created_at
);
`
	args := pgx.StrictNamedArgs{
		"id":              h.ID,
		"tenant_id":       predicate.GetTenantID(),
		"organization_id": h.OrganizationID,
		"access_entry_id": h.AccessEntry,
		"decision":        h.Decision,
		"decision_note":   h.DecisionNote,
		"decided_by":      h.DecidedBy,
		"decided_at":      h.DecidedAt,
		"created_at":      h.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert access entry decision history: %w", err)
	}

	return nil
}

func (h *AccessEntryDecisionHistory) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_entry_decision_history WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

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

func (hs *AccessEntryDecisionHistories) LoadByEntryID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	entryID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_entry_id,
    decision,
    decision_note,
    decided_by,
    decided_at,
    created_at
FROM
    access_entry_decision_history
WHERE
    %s
    AND access_entry_id = @access_entry_id
ORDER BY decided_at ASC;
`
	q = fmt.Sprintf(q, predicate.SQLFragment())

	args := pgx.StrictNamedArgs{"access_entry_id": entryID}
	maps.Copy(args, predicate.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access entry decision history: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessEntryDecisionHistory])
	if err != nil {
		return fmt.Errorf("cannot collect access entry decision history: %w", err)
	}

	*hs = result

	return nil
}
