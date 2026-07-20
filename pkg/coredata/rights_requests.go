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
	"errors"
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
	RightsRequest struct {
		ID             gid.GID            `db:"id"`
		OrganizationID gid.GID            `db:"organization_id"`
		RequestType    RightsRequestType  `db:"request_type"`
		RequestState   RightsRequestState `db:"request_state"`
		DataSubject    *string            `db:"data_subject"`
		Contact        *string            `db:"contact"`
		Details        *string            `db:"details"`
		Deadline       *time.Time         `db:"deadline"`
		ActionTaken    *string            `db:"action_taken"`
		CreatedAt      time.Time          `db:"created_at"`
		UpdatedAt      time.Time          `db:"updated_at"`
	}

	RightsRequests []*RightsRequest
)

func (rr *RightsRequest) CursorKey(field RightsRequestOrderField) page.CursorKey {
	switch field {
	case RightsRequestOrderFieldCreatedAt:
		return page.NewCursorKey(rr.ID, rr.CreatedAt)
	case RightsRequestOrderFieldDeadline:
		return page.NewCursorKey(rr.ID, rr.Deadline)
	case RightsRequestOrderFieldState:
		return page.NewCursorKey(rr.ID, rr.RequestState)
	case RightsRequestOrderFieldType:
		return page.NewCursorKey(rr.ID, rr.RequestType)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (rr *RightsRequest) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM rights_requests WHERE id = ANY(@resource_ids::text[])`

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

func (rr *RightsRequest) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	rightsRequestID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	request_type,
	request_state,
	data_subject,
	contact,
	details,
	deadline,
	action_taken,
	created_at,
	updated_at
FROM
	rights_requests
WHERE
	%s
	AND id = @rights_request_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"rights_request_id": rightsRequestID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query rights request: %w", err)
	}

	request, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RightsRequest])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect rights request: %w", err)
	}

	*rr = request

	return nil
}

func (rrs *RightsRequests) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	rights_requests
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count rights requests: %w", err)
	}

	return count, nil
}

func (rrs *RightsRequests) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[RightsRequestOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	request_type,
	request_state,
	data_subject,
	contact,
	details,
	deadline,
	action_taken,
	created_at,
	updated_at
FROM
	rights_requests
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query rights requests: %w", err)
	}

	requests, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RightsRequest])
	if err != nil {
		return fmt.Errorf("cannot collect rights requests: %w", err)
	}

	*rrs = requests

	return nil
}

func (rrs *RightsRequests) LoadByOrganizationIDAndContact(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	contact string,
	cursor *page.Cursor[RightsRequestOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	request_type,
	request_state,
	data_subject,
	contact,
	details,
	deadline,
	action_taken,
	created_at,
	updated_at
FROM
	rights_requests
WHERE
	%s
	AND organization_id = @organization_id
	AND contact = @contact
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"contact":         contact,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query rights requests: %w", err)
	}

	requests, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RightsRequest])
	if err != nil {
		return fmt.Errorf("cannot collect rights requests: %w", err)
	}

	*rrs = requests

	return nil
}

func (rr *RightsRequest) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO rights_requests (
	id,
	tenant_id,
	organization_id,
	request_type,
	request_state,
	data_subject,
	contact,
	details,
	deadline,
	action_taken,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@request_type,
	@request_state,
	@data_subject,
	@contact,
	@details,
	@deadline,
	@action_taken,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              rr.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": rr.OrganizationID,
		"request_type":    rr.RequestType,
		"request_state":   rr.RequestState,
		"data_subject":    rr.DataSubject,
		"contact":         rr.Contact,
		"details":         rr.Details,
		"deadline":        rr.Deadline,
		"action_taken":    rr.ActionTaken,
		"created_at":      rr.CreatedAt,
		"updated_at":      rr.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert rights request: %w", err)
	}

	return nil
}

func (rr *RightsRequest) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE rights_requests SET
	request_type = @request_type,
	request_state = @request_state,
	data_subject = @data_subject,
	contact = @contact,
	details = @details,
	deadline = @deadline,
	action_taken = @action_taken,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            rr.ID,
		"request_type":  rr.RequestType,
		"request_state": rr.RequestState,
		"data_subject":  rr.DataSubject,
		"contact":       rr.Contact,
		"details":       rr.Details,
		"deadline":      rr.Deadline,
		"action_taken":  rr.ActionTaken,
		"updated_at":    rr.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update rights request: %w", err)
	}

	return nil
}

func (rr *RightsRequest) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM rights_requests
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": rr.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete rights request: %w", err)
	}

	return nil
}
