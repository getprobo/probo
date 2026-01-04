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
	"errors"
	"fmt"
	"maps"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	SCIMEvent struct {
		ID                  gid.GID   `db:"id"`
		OrganizationID      gid.GID   `db:"organization_id"`
		SCIMConfigurationID gid.GID   `db:"scim_configuration_id"`
		Method              string    `db:"method"`
		Path                string    `db:"path"`
		RequestBody         *string   `db:"request_body"`
		ResponseBody        *string   `db:"response_body"`
		StatusCode          int       `db:"status_code"`
		ErrorMessage        *string   `db:"error_message"`
		MembershipID        *gid.GID  `db:"membership_id"`
		IPAddress           net.IP    `db:"ip_address"`
		CreatedAt           time.Time `db:"created_at"`
	}

	SCIMEvents []*SCIMEvent
)

func (s *SCIMEvent) CursorKey(orderBy SCIMEventOrderField) page.CursorKey {
	switch orderBy {
	case SCIMEventOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *SCIMEvent) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM iam_scim_events WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query scim event authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (s *SCIMEvent) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	eventID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    method,
    path,
    request_body,
    response_body,
    status_code,
    error_message,
    membership_id,
    ip_address,
    created_at
FROM
    iam_scim_events
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": eventID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_events: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_event: %w", err)
	}

	*s = event

	return nil
}

func (s *SCIMEvent) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO iam_scim_events (
    id,
    tenant_id,
    organization_id,
    scim_configuration_id,
    method,
    path,
    request_body,
    response_body,
    status_code,
    error_message,
    membership_id,
    ip_address,
    created_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @scim_configuration_id,
    @method,
    @path,
    @request_body,
    @response_body,
    @status_code,
    @error_message,
    @membership_id,
    @ip_address,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                    s.ID,
		"tenant_id":             scope.GetTenantID(),
		"organization_id":       s.OrganizationID,
		"scim_configuration_id": s.SCIMConfigurationID,
		"method":                s.Method,
		"path":                  s.Path,
		"request_body":          s.RequestBody,
		"response_body":         s.ResponseBody,
		"status_code":           s.StatusCode,
		"error_message":         s.ErrorMessage,
		"membership_id":         s.MembershipID,
		"ip_address":            s.IPAddress,
		"created_at":            s.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert scim_event: %w", err)
	}

	return nil
}

func (s *SCIMEvents) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[SCIMEventOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    method,
    path,
    request_body,
    response_body,
    status_code,
    error_message,
    membership_id,
    ip_address,
    created_at
FROM
    iam_scim_events
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
		return fmt.Errorf("cannot query iam_scim_events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SCIMEvent])
	if err != nil {
		return fmt.Errorf("cannot collect scim_events: %w", err)
	}

	*s = events

	return nil
}

func (s *SCIMEvents) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_scim_events
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count scim_events: %w", err)
	}

	return count, nil
}

func (s *SCIMEvents) LoadBySCIMConfigurationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	scimConfigurationID gid.GID,
	cursor *page.Cursor[SCIMEventOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    method,
    path,
    request_body,
    response_body,
    status_code,
    error_message,
    membership_id,
    ip_address,
    created_at
FROM
    iam_scim_events
WHERE
    %s
    AND scim_configuration_id = @scim_configuration_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"scim_configuration_id": scimConfigurationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SCIMEvent])
	if err != nil {
		return fmt.Errorf("cannot collect scim_events: %w", err)
	}

	*s = events

	return nil
}

func (s *SCIMEvents) CountBySCIMConfigurationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	scimConfigurationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_scim_events
WHERE
    %s
    AND scim_configuration_id = @scim_configuration_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"scim_configuration_id": scimConfigurationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count scim_events: %w", err)
	}

	return count, nil
}
