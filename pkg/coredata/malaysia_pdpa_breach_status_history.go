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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	MalaysiaPDPABreachStatusHistory struct {
		ID                 gid.GID                   `db:"id"`
		OrganizationID     gid.GID                   `db:"organization_id"`
		IncidentID         gid.GID                   `db:"incident_id"`
		FromStatus         *MalaysiaPDPABreachStatus `db:"from_status"`
		ToStatus           MalaysiaPDPABreachStatus  `db:"to_status"`
		ChangedByProfileID gid.GID                   `db:"changed_by_profile_id"`
		Reason             *string                   `db:"reason"`
		CreatedAt          time.Time                 `db:"created_at"`
	}

	MalaysiaPDPABreachStatusHistories []*MalaysiaPDPABreachStatusHistory
)

func (h *MalaysiaPDPABreachStatusHistory) CursorKey(field MalaysiaPDPABreachStatusHistoryOrderField) page.CursorKey {
	if field == MalaysiaPDPABreachStatusHistoryOrderFieldCreatedAt {
		return page.NewCursorKey(h.ID, h.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (h *MalaysiaPDPABreachStatusHistory) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM malaysia_pdpa_breach_status_history WHERE id = ANY(@resource_ids::text[])`
	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query Malaysia PDPA breach history authorization attributes: %w", err)
	}
	defer rows.Close()

	attributes := make(policy.AttributesByID)
	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan Malaysia PDPA breach history authorization attributes: %w", err)
		}
		attributes[id] = policy.Attributes{"organization_id": organizationID.String()}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate Malaysia PDPA breach history authorization attributes: %w", err)
	}

	return attributes, nil
}

func (h *MalaysiaPDPABreachStatusHistory) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO malaysia_pdpa_breach_status_history (
    id, tenant_id, organization_id, incident_id, from_status, to_status,
    changed_by_profile_id, reason, created_at
) VALUES (
    @id, @tenant_id, @organization_id, @incident_id, @from_status, @to_status,
    @changed_by_profile_id, @reason, @created_at
);`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{
		"id":                    h.ID,
		"tenant_id":             scope.GetTenantID(),
		"organization_id":       h.OrganizationID,
		"incident_id":           h.IncidentID,
		"from_status":           h.FromStatus,
		"to_status":             h.ToStatus,
		"changed_by_profile_id": h.ChangedByProfileID,
		"reason":                h.Reason,
		"created_at":            h.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("cannot insert Malaysia PDPA breach status history: %w", err)
	}

	return nil
}

func (hs *MalaysiaPDPABreachStatusHistories) CountByIncidentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	incidentID gid.GID,
) (int, error) {
	q := `SELECT COUNT(id) FROM malaysia_pdpa_breach_status_history WHERE %s AND incident_id = @incident_id;`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"incident_id": incidentID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count Malaysia PDPA breach status history: %w", err)
	}

	return count, nil
}

func (hs *MalaysiaPDPABreachStatusHistories) LoadByIncidentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	incidentID gid.GID,
	cursor *page.Cursor[MalaysiaPDPABreachStatusHistoryOrderField],
) error {
	q := `
SELECT
    id, organization_id, incident_id, from_status, to_status,
    changed_by_profile_id, reason, created_at
FROM malaysia_pdpa_breach_status_history
WHERE %s AND incident_id = @incident_id AND %s;`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"incident_id": incidentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Malaysia PDPA breach status history: %w", err)
	}

	history, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MalaysiaPDPABreachStatusHistory])
	if err != nil {
		return fmt.Errorf("cannot collect Malaysia PDPA breach status history: %w", err)
	}

	*hs = history

	return nil
}
